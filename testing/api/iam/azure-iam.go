package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/finos-labs/ccc-cfi-compliance/testing/environment"
	"github.com/google/uuid"
)

// AzureIAMService implements IAMService for Azure using Service Principals
type AzureIAMService struct {
	authClient       *armauthorization.RoleAssignmentsClient
	ctx              context.Context
	credential       azcore.TokenCredential
	cloudParams      environment.CloudParams
	httpClient       *http.Client
	tenantID         string
	provisionedUsers map[string]*Identity // Cache of provisioned users by userName
	accessLevels     map[string]string    // Cache of access levels by "userName:serviceID"
}

// NewAzureIAMService creates a new Azure IAM service using default credentials
func NewAzureIAMService(ctx context.Context, cloudParams environment.CloudParams) (*AzureIAMService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	return newAzureIAMServiceInternal(ctx, cloudParams, cred)
}

// NewAzureIAMServiceWithCredentials creates a new Azure IAM service with specific credentials
func NewAzureIAMServiceWithCredentials(ctx context.Context, cloudParams environment.CloudParams, cred azcore.TokenCredential) (*AzureIAMService, error) {
	return newAzureIAMServiceInternal(ctx, cloudParams, cred)
}

func newAzureIAMServiceInternal(ctx context.Context, cloudParams environment.CloudParams, cred azcore.TokenCredential) (*AzureIAMService, error) {
	authClient, err := armauthorization.NewRoleAssignmentsClient(cloudParams.AzureSubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization client: %w", err)
	}

	// Get tenant ID from the credential
	tenantID, err := getTenantID(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant ID: %w", err)
	}

	return &AzureIAMService{
		authClient:       authClient,
		ctx:              ctx,
		credential:       cred,
		cloudParams:      cloudParams,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		tenantID:         tenantID,
		provisionedUsers: make(map[string]*Identity),
		accessLevels:     make(map[string]string),
	}, nil
}

// getTenantID retrieves the tenant ID from the credential
func getTenantID(ctx context.Context, cred azcore.TokenCredential) (string, error) {
	// Get a token to extract tenant ID
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return "", err
	}

	// Parse the JWT token to extract tenant ID
	// The token is in format: header.payload.signature
	parts := strings.Split(token.Token, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid token format")
	}

	// For now, we'll use a simpler approach - get it from the Azure CLI config
	// In production, you'd parse the JWT properly
	return getAzureTenantID(), nil
}

// getAzureTenantID gets the tenant ID from environment or Azure CLI
func getAzureTenantID() string {
	// Try to get from Azure CLI
	// In production, this should be passed as a parameter or read from config
	// For now, return a placeholder that will be populated by the credential
	return "" // Will be populated when we make Graph API calls
}

// ProvisionUser creates a new service principal with a client secret, or returns existing one
func (s *AzureIAMService) ProvisionUser(userName string) (*Identity, error) {
	// Check cache first - if we've already provisioned this user in this session, return it
	if cachedIdentity, exists := s.provisionedUsers[userName]; exists {
		fmt.Printf("♻️  Using cached identity for user %s (skipping propagation delay)\n", userName)
		return cachedIdentity, nil
	}

	// Service principal display names can be more flexible than managed identity names
	displayName := sanitizeServicePrincipalName(userName)

	fmt.Printf("🔷 Provisioning service principal: %s\n", displayName)

	// Check if application already exists
	existingAppID, existingObjectID, err := s.findApplicationByDisplayName(displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing application: %w", err)
	}

	var appID, objectID, spObjectID string
	var isExisting bool

	if existingAppID != "" {
		// Application already exists
		fmt.Printf("   ℹ️  Application already exists: %s\n", existingAppID)
		appID = existingAppID
		objectID = existingObjectID
		isExisting = true

		// Get or create service principal for existing app
		spObjectID, err = s.getOrCreateServicePrincipal(appID)
		if err != nil {
			return nil, fmt.Errorf("failed to get service principal: %w", err)
		}
	} else {
		// Create new application
		fmt.Printf("   📱 Creating new application...\n")
		appID, objectID, err = s.createApplication(displayName)
		if err != nil {
			return nil, fmt.Errorf("failed to create application: %w", err)
		}
		fmt.Printf("   📱 Application created: %s (ObjectID: %s)\n", appID, objectID)

		// Create service principal for the application
		spObjectID, err = s.createServicePrincipal(appID)
		if err != nil {
			// Try to clean up the application if service principal creation fails
			_ = s.deleteApplication(objectID)
			return nil, fmt.Errorf("failed to create service principal: %w", err)
		}
		fmt.Printf("   🔑 Service principal created (ObjectID: %s)\n", spObjectID)
	}

	// Always create a new client secret (we can't retrieve existing ones)
	clientSecret, secretID, err := s.addApplicationPassword(objectID, displayName)
	if err != nil {
		if !isExisting {
			// Clean up if this was a new resource
			_ = s.deleteServicePrincipal(spObjectID)
			_ = s.deleteApplication(objectID)
		}
		return nil, fmt.Errorf("failed to create client secret: %w", err)
	}

	fmt.Printf("   🔐 Client secret created\n")

	// Get tenant ID
	tenantID, err := s.getActualTenantID()
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant ID: %w", err)
	}

	// Create identity with credentials
	identity := &Identity{
		UserName:    userName,
		Provider:    "azure",
		Credentials: make(map[string]string),
	}

	// Store Azure-specific fields in Credentials map
	identity.Credentials["client_id"] = appID            // Application (client) ID
	identity.Credentials["client_secret"] = clientSecret // Client secret (works from anywhere!)
	identity.Credentials["tenant_id"] = tenantID         // Tenant ID
	identity.Credentials["object_id"] = spObjectID       // Service principal object ID
	identity.Credentials["app_object_id"] = objectID     // Application object ID
	identity.Credentials["secret_id"] = secretID         // Secret ID for cleanup
	identity.Credentials["subscription_id"] = s.cloudParams.AzureSubscriptionID
	identity.Credentials["display_name"] = displayName

	if isExisting {
		fmt.Printf("✅ Using existing service principal with new secret: %s\n", userName)
	} else {
		fmt.Printf("✅ Provisioned new service principal: %s\n", userName)
	}
	fmt.Printf("   Client ID: %s\n", identity.Credentials["client_id"])
	fmt.Printf("   Tenant ID: %s\n", identity.Credentials["tenant_id"])
	fmt.Printf("   💡 Client secret can be used from anywhere (not just Azure)\n")

	fmt.Printf("   ⏳ Waiting 15s for Azure AD propagation (new service principal)...\n")
	time.Sleep(15 * time.Second)

	// Cache the identity for future requests
	s.provisionedUsers[userName] = identity

	return identity, nil
}

// SetAccess grants an identity access to a specific Azure resource at the specified level
func (s *AzureIAMService) SetAccess(identity *Identity, serviceID string, level string) error {
	// Check cache first - if we've already set this access level, skip it
	cacheKey := fmt.Sprintf("%s:%s", identity.UserName, serviceID)
	if cachedLevel, exists := s.accessLevels[cacheKey]; exists && cachedLevel == level {
		fmt.Printf("♻️  Access level already set to %s for %s (skipping propagation delay)\n", level, identity.UserName)
		return nil
	}

	// Get the role definition ID based on access level
	roleDefinitionID, err := s.getRoleDefinitionForLevel(serviceID, level)
	if err != nil {
		return fmt.Errorf("failed to determine role: %w", err)
	}

	if roleDefinitionID == "" {
		// "none" level - no role to assign
		// Cache this state
		s.accessLevels[cacheKey] = level
		return nil
	}

	// Get the service principal object ID from the identity
	objectID := identity.Credentials["object_id"]
	if objectID == "" {
		return fmt.Errorf("object_id not found in identity credentials")
	}

	// Parse the scope from serviceID
	scope := s.parseScope(serviceID)

	fmt.Printf("🔐 Granting %s access for service principal %s\n", level, objectID)
	fmt.Printf("   Scope: %s\n", scope)
	fmt.Printf("   Role: %s\n", roleDefinitionID)

	// Create a unique name for the role assignment
	roleAssignmentName := uuid.New().String()

	// Create role assignment
	roleAssignmentParams := armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &objectID,
			RoleDefinitionID: &roleDefinitionID,
		},
	}

	_, err = s.authClient.Create(s.ctx, scope, roleAssignmentName, roleAssignmentParams, nil)
	if err != nil {
		// Check if assignment already exists
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "RoleAssignmentExists") {
			fmt.Printf("   ℹ️  Role assignment already exists\n")
			// Cache this state
			s.accessLevels[cacheKey] = level
			return nil
		}
		return fmt.Errorf("failed to create role assignment: %w", err)
	}

	fmt.Printf("   ✅ Access granted\n")

	// Azure RBAC propagation delay: Role assignments need time to take effect
	// Data plane permissions can take 30-60 seconds to propagate
	fmt.Printf("   ⏳ Waiting 30s for RBAC propagation...\n")
	time.Sleep(30 * time.Second)

	// Cache the access level for future requests
	s.accessLevels[cacheKey] = level

	return nil
}

// DestroyUser removes a service principal and all associated resources
func (s *AzureIAMService) DestroyUser(identity *Identity) error {
	displayName := identity.Credentials["display_name"]
	if displayName == "" {
		displayName = identity.UserName
	}

	fmt.Printf("🗑️  Deleting service principal: %s\n", displayName)

	// Step 1: Delete role assignments for this identity
	objectID := identity.Credentials["object_id"]
	if objectID != "" {
		fmt.Printf("   🔍 Looking for role assignments for principal %s...\n", objectID)

		// List role assignments in the subscription
		filter := fmt.Sprintf("principalId eq '%s'", objectID)
		pager := s.authClient.NewListForSubscriptionPager(&armauthorization.RoleAssignmentsClientListForSubscriptionOptions{
			Filter: &filter,
		})

		for pager.More() {
			page, err := pager.NextPage(s.ctx)
			if err != nil {
				fmt.Printf("   ⚠️  Failed to list role assignments: %v\n", err)
				break
			}

			for _, assignment := range page.Value {
				if assignment.Name != nil {
					fmt.Printf("   🗑️  Deleting role assignment %s...\n", *assignment.Name)

					// Extract scope from assignment ID
					scope := extractScopeFromAssignmentID(*assignment.ID)

					_, err := s.authClient.Delete(s.ctx, scope, *assignment.Name, nil)
					if err != nil {
						fmt.Printf("   ⚠️  Failed to delete role assignment: %v\n", err)
					}
				}
			}
		}
	}

	// Step 2: Delete the service principal
	spObjectID := identity.Credentials["object_id"]
	if spObjectID != "" {
		err := s.deleteServicePrincipal(spObjectID)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to delete service principal: %v\n", err)
		} else {
			fmt.Printf("   ✅ Service principal deleted\n")
		}
	}

	// Step 3: Delete the application
	appObjectID := identity.Credentials["app_object_id"]
	if appObjectID != "" {
		err := s.deleteApplication(appObjectID)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to delete application: %v\n", err)
		} else {
			fmt.Printf("   ✅ Application deleted\n")
		}
	}

	fmt.Printf("✅ Service principal cleanup complete\n")
	return nil
}

// Helper functions

func (s *AzureIAMService) getRoleDefinitionForLevel(serviceID string, level string) (string, error) {
	// Azure built-in role definition IDs
	// Format: /subscriptions/{subscriptionId}/providers/Microsoft.Authorization/roleDefinitions/{roleId}

	baseRolePath := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions", s.cloudParams.AzureSubscriptionID)

	switch level {
	case "none":
		return "", nil
	case "read":
		// Reader role for general read access
		// Storage Blob Data Reader for blob storage
		if strings.Contains(serviceID, "storage") || strings.Contains(serviceID, "blob") {
			return fmt.Sprintf("%s/2a2b9908-6ea1-4ae2-8e65-a410df84e7d1", baseRolePath), nil // Storage Blob Data Reader
		}
		return fmt.Sprintf("%s/acdd72a7-3385-48ef-bd42-f606fba81ae7", baseRolePath), nil // Reader
	case "write":
		// Contributor role for write access
		// Storage Blob Data Contributor for blob storage
		if strings.Contains(serviceID, "storage") || strings.Contains(serviceID, "blob") {
			return fmt.Sprintf("%s/ba92f5b4-2d11-453d-a403-e96b0029c9fe", baseRolePath), nil // Storage Blob Data Contributor
		}
		return fmt.Sprintf("%s/b24988ac-6180-42a0-ab88-20f7382dd24c", baseRolePath), nil // Contributor
	case "admin":
		// Owner role for admin access
		// Storage Blob Data Owner for blob storage
		if strings.Contains(serviceID, "storage") || strings.Contains(serviceID, "blob") {
			return fmt.Sprintf("%s/b7e6dc6d-f1e8-4753-8033-0f276bb0955b", baseRolePath), nil // Storage Blob Data Owner
		}
		return fmt.Sprintf("%s/8e3af657-a8ff-443c-a75c-2fe8c4bcb635", baseRolePath), nil // Owner
	default:
		return "", fmt.Errorf("unsupported access level: %s", level)
	}
}

func (s *AzureIAMService) parseScope(serviceID string) string {
	// If serviceID is already a full resource ID, use it as scope
	if strings.HasPrefix(serviceID, "/subscriptions/") {
		return serviceID
	}

	// If it's a storage account or container reference
	if strings.Contains(serviceID, "storage") {
		// Try to extract storage account name and build resource ID
		// Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Storage/storageAccounts/{name}
		parts := strings.Split(serviceID, "/")
		if len(parts) > 0 {
			accountName := parts[len(parts)-1]
			return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s",
				s.cloudParams.AzureSubscriptionID, s.cloudParams.AzureResourceGroup, accountName)
		}
	}

	// Default: use resource group scope
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", s.cloudParams.AzureSubscriptionID, s.cloudParams.AzureResourceGroup)
}

func extractScopeFromAssignmentID(assignmentID string) string {
	// Assignment ID format: {scope}/providers/Microsoft.Authorization/roleAssignments/{name}
	// Extract scope by removing the role assignment suffix
	parts := strings.Split(assignmentID, "/providers/Microsoft.Authorization/roleAssignments/")
	if len(parts) > 0 {
		return parts[0]
	}
	return assignmentID
}

func sanitizeServicePrincipalName(userName string) string {
	// Service principal display names can be more flexible
	// Just ensure it's a valid display name
	result := userName

	// Add CCC prefix for easy identification
	if !strings.HasPrefix(result, "CCC-") {
		result = "CCC-Test-" + result
	}

	// Ensure maximum length (120 chars is safe)
	if len(result) > 120 {
		result = result[:120]
	}

	return result
}

func toPtr(s string) *string {
	return &s
}

// Fill this later when we are writing tests for IAM
func (s *AzureIAMService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
	return []environment.TestParams{}, nil
}

// Microsoft Graph API helper methods

func (s *AzureIAMService) callGraphAPI(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	graphURL := "https://graph.microsoft.com/v1.0" + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(s.ctx, method, graphURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token for Microsoft Graph
	token, err := s.credential.GetToken(s.ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("graph API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return result, nil
}

func (s *AzureIAMService) createApplication(displayName string) (appID, objectID string, err error) {
	requestBody := map[string]interface{}{
		"displayName":    displayName,
		"signInAudience": "AzureADMyOrg",
	}

	result, err := s.callGraphAPI("POST", "/applications", requestBody)
	if err != nil {
		return "", "", err
	}

	appID, _ = result["appId"].(string)
	objectID, _ = result["id"].(string)

	if appID == "" || objectID == "" {
		return "", "", fmt.Errorf("failed to extract application IDs from response")
	}

	return appID, objectID, nil
}

func (s *AzureIAMService) createServicePrincipal(appID string) (objectID string, err error) {
	requestBody := map[string]interface{}{
		"appId": appID,
	}

	result, err := s.callGraphAPI("POST", "/servicePrincipals", requestBody)
	if err != nil {
		return "", err
	}

	objectID, _ = result["id"].(string)
	if objectID == "" {
		return "", fmt.Errorf("failed to extract service principal object ID from response")
	}

	return objectID, nil
}

func (s *AzureIAMService) addApplicationPassword(appObjectID, displayName string) (secret, secretID string, err error) {
	requestBody := map[string]interface{}{
		"passwordCredential": map[string]interface{}{
			"displayName": displayName + "-secret",
		},
	}

	result, err := s.callGraphAPI("POST", "/applications/"+appObjectID+"/addPassword", requestBody)
	if err != nil {
		return "", "", err
	}

	secret, _ = result["secretText"].(string)
	secretID, _ = result["keyId"].(string)

	if secret == "" || secretID == "" {
		return "", "", fmt.Errorf("failed to extract secret from response")
	}

	return secret, secretID, nil
}

func (s *AzureIAMService) deleteServicePrincipal(objectID string) error {
	_, err := s.callGraphAPI("DELETE", "/servicePrincipals/"+objectID, nil)
	return err
}

func (s *AzureIAMService) deleteApplication(objectID string) error {
	_, err := s.callGraphAPI("DELETE", "/applications/"+objectID, nil)
	return err
}

func (s *AzureIAMService) getActualTenantID() (string, error) {
	// Get the organization details to extract tenant ID
	result, err := s.callGraphAPI("GET", "/organization", nil)
	if err != nil {
		return "", err
	}

	// Extract tenant ID from the organization response
	if value, ok := result["value"].([]interface{}); ok && len(value) > 0 {
		if org, ok := value[0].(map[string]interface{}); ok {
			if tenantID, ok := org["id"].(string); ok {
				return tenantID, nil
			}
		}
	}

	return "", fmt.Errorf("failed to extract tenant ID from organization response")
}

func (s *AzureIAMService) findApplicationByDisplayName(displayName string) (appID, objectID string, err error) {
	// Search for applications by display name
	filter := fmt.Sprintf("displayName eq '%s'", displayName)
	endpoint := fmt.Sprintf("/applications?$filter=%s", url.QueryEscape(filter))

	result, err := s.callGraphAPI("GET", endpoint, nil)
	if err != nil {
		return "", "", err
	}

	// Check if any applications were found
	if value, ok := result["value"].([]interface{}); ok && len(value) > 0 {
		if app, ok := value[0].(map[string]interface{}); ok {
			appID, _ = app["appId"].(string)
			objectID, _ = app["id"].(string)

			if appID != "" && objectID != "" {
				return appID, objectID, nil
			}
		}
	}

	// No application found
	return "", "", nil
}

func (s *AzureIAMService) getOrCreateServicePrincipal(appID string) (objectID string, err error) {
	// Try to find existing service principal
	filter := fmt.Sprintf("appId eq '%s'", appID)
	endpoint := fmt.Sprintf("/servicePrincipals?$filter=%s", url.QueryEscape(filter))

	result, err := s.callGraphAPI("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	// Check if service principal exists
	if value, ok := result["value"].([]interface{}); ok && len(value) > 0 {
		if sp, ok := value[0].(map[string]interface{}); ok {
			objectID, _ = sp["id"].(string)
			if objectID != "" {
				fmt.Printf("   ℹ️  Service principal already exists (ObjectID: %s)\n", objectID)
				return objectID, nil
			}
		}
	}

	// Service principal doesn't exist, create it
	fmt.Printf("   🔑 Creating service principal...\n")
	objectID, err = s.createServicePrincipal(appID)
	if err != nil {
		return "", err
	}

	fmt.Printf("   🔑 Service principal created (ObjectID: %s)\n", objectID)
	return objectID, nil
}
