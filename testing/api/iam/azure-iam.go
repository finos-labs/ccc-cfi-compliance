package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	authClient  *armauthorization.RoleAssignmentsClient
	ctx         context.Context
	credential  azcore.TokenCredential
	cloudParams environment.CloudParams
	httpClient  *http.Client
	tenantID    string
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
		authClient:  authClient,
		ctx:         ctx,
		credential:  cred,
		cloudParams: cloudParams,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		tenantID:    tenantID,
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

// ProvisionUser creates a new service principal with a client secret
func (s *AzureIAMService) ProvisionUser(userName string) (*Identity, error) {
	// Service principal display names can be more flexible than managed identity names
	displayName := sanitizeServicePrincipalName(userName)

	fmt.Printf("🔷 Creating service principal: %s\n", displayName)

	// Step 1: Create an Azure AD application
	appID, objectID, err := s.createApplication(displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	fmt.Printf("   📱 Application created: %s (ObjectID: %s)\n", appID, objectID)

	// Step 2: Create a service principal for the application
	spObjectID, err := s.createServicePrincipal(appID)
	if err != nil {
		// Try to clean up the application if service principal creation fails
		_ = s.deleteApplication(objectID)
		return nil, fmt.Errorf("failed to create service principal: %w", err)
	}

	fmt.Printf("   🔑 Service principal created (ObjectID: %s)\n", spObjectID)

	// Step 3: Create a client secret
	clientSecret, secretID, err := s.addApplicationPassword(objectID, displayName)
	if err != nil {
		// Try to clean up on failure
		_ = s.deleteServicePrincipal(spObjectID)
		_ = s.deleteApplication(objectID)
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

	fmt.Printf("✅ Provisioned service principal: %s\n", userName)
	fmt.Printf("   Client ID: %s\n", identity.Credentials["client_id"])
	fmt.Printf("   Tenant ID: %s\n", identity.Credentials["tenant_id"])
	fmt.Printf("   💡 Client secret can be used from anywhere (not just Azure)\n")

	// Small delay to allow Azure AD propagation
	time.Sleep(2 * time.Second)

	return identity, nil
}

// SetAccess grants an identity access to a specific Azure resource at the specified level
func (s *AzureIAMService) SetAccess(identity *Identity, serviceID string, level string) error {
	// Get the role definition ID based on access level
	roleDefinitionID, err := s.getRoleDefinitionForLevel(serviceID, level)
	if err != nil {
		return fmt.Errorf("failed to determine role: %w", err)
	}

	if roleDefinitionID == "" {
		// "none" level - no role to assign
		return nil
	}

	// Get the service principal object ID from the identity
	objectID := identity.Credentials["object_id"]
	if objectID == "" {
		return fmt.Errorf("object_id not found in identity credentials")
	}

	// Parse the scope from serviceID
	scope := s.parseScope(serviceID)

	fmt.Printf("🔐 Granting %s access to %s for service principal %s...\n", level, scope, objectID)

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
			return nil
		}
		return fmt.Errorf("failed to create role assignment: %w", err)
	}

	fmt.Printf("   ✅ Access granted\n")
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
func (s *AzureIAMService) GetTestableResources() ([]environment.TestParams, error) {
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
