package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/google/uuid"
)

// AzureIAMService implements IAMService for Azure
type AzureIAMService struct {
	msiClient      *armmsi.UserAssignedIdentitiesClient
	authClient     *armauthorization.RoleAssignmentsClient
	ctx            context.Context
	credential     azcore.TokenCredential
	subscriptionID string
	resourceGroup  string
}

// NewAzureIAMService creates a new Azure IAM service using default credentials
func NewAzureIAMService(ctx context.Context, subscriptionID, resourceGroup string) (*AzureIAMService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	msiClient, err := armmsi.NewUserAssignedIdentitiesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MSI client: %w", err)
	}

	authClient, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization client: %w", err)
	}

	return &AzureIAMService{
		msiClient:      msiClient,
		authClient:     authClient,
		ctx:            ctx,
		credential:     cred,
		subscriptionID: subscriptionID,
		resourceGroup:  resourceGroup,
	}, nil
}

// NewAzureIAMServiceWithCredentials creates a new Azure IAM service with specific credentials
func NewAzureIAMServiceWithCredentials(ctx context.Context, subscriptionID, resourceGroup string, cred azcore.TokenCredential) (*AzureIAMService, error) {
	msiClient, err := armmsi.NewUserAssignedIdentitiesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MSI client: %w", err)
	}

	authClient, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization client: %w", err)
	}

	return &AzureIAMService{
		msiClient:      msiClient,
		authClient:     authClient,
		ctx:            ctx,
		credential:     cred,
		subscriptionID: subscriptionID,
		resourceGroup:  resourceGroup,
	}, nil
}

// ProvisionUser creates a new user-assigned managed identity (Azure's equivalent for service identities)
func (s *AzureIAMService) ProvisionUser(userName string) (*Identity, error) {
	// Managed identity names must be alphanumeric, hyphens, underscores
	identityName := sanitizeManagedIdentityName(userName)
	location := "eastus" // Default location, could be parameterized

	var managedIdentity *armmsi.Identity
	var identityAlreadyExists bool

	// Check if managed identity already exists
	getResp, err := s.msiClient.Get(s.ctx, s.resourceGroup, identityName, nil)
	if err == nil {
		// Identity exists - reuse it
		fmt.Printf("🔷 Managed identity %s already exists, reusing...\n", identityName)
		managedIdentity = &getResp.Identity
		identityAlreadyExists = true
	} else {
		// Identity doesn't exist - create it
		fmt.Printf("🔷 Creating managed identity %s...\n", identityName)

		tags := map[string]*string{
			"Purpose":   toPtr("CCC-Testing"),
			"ManagedBy": toPtr("CCC-CFI-Compliance-Framework"),
		}

		createResp, err := s.msiClient.CreateOrUpdate(s.ctx, s.resourceGroup, identityName, armmsi.Identity{
			Location: &location,
			Tags:     tags,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create managed identity %s: %w", identityName, err)
		}
		managedIdentity = &createResp.Identity
	}

	// Extract identity information
	if managedIdentity.Properties == nil {
		return nil, fmt.Errorf("managed identity properties are nil")
	}

	principalID := ""
	clientID := ""
	tenantID := ""
	resourceID := ""

	if managedIdentity.Properties.PrincipalID != nil {
		principalID = *managedIdentity.Properties.PrincipalID
	}
	if managedIdentity.Properties.ClientID != nil {
		clientID = *managedIdentity.Properties.ClientID
	}
	if managedIdentity.Properties.TenantID != nil {
		tenantID = *managedIdentity.Properties.TenantID
	}
	if managedIdentity.ID != nil {
		resourceID = *managedIdentity.ID
	}

	// Create identity with credentials
	identity := &Identity{
		UserName:    userName,
		Provider:    "azure",
		Credentials: make(map[string]string),
	}

	// Store Azure-specific fields in Credentials map
	identity.Credentials["principal_id"] = principalID
	identity.Credentials["client_id"] = clientID
	identity.Credentials["tenant_id"] = tenantID
	identity.Credentials["resource_id"] = resourceID
	identity.Credentials["subscription_id"] = s.subscriptionID
	identity.Credentials["resource_group"] = s.resourceGroup
	identity.Credentials["identity_name"] = identityName

	// For Azure managed identities, we don't have a secret key
	// They authenticate via Azure AD using the managed identity
	fmt.Printf("   ℹ️  Managed identities use Azure AD authentication (no keys)\n")

	// Log the created/retrieved identity details
	fmt.Printf("✅ Provisioned managed identity: %s\n", userName)
	fmt.Printf("   Principal ID: %s\n", identity.Credentials["principal_id"])
	fmt.Printf("   Client ID: %s\n", identity.Credentials["client_id"])
	fmt.Printf("   Resource ID: %s\n", identity.Credentials["resource_id"])

	// Store identity info as JSON for potential future use
	identityJSON, _ := json.Marshal(map[string]string{
		"principal_id": principalID,
		"client_id":    clientID,
		"tenant_id":    tenantID,
	})
	identity.Credentials["identity_json"] = string(identityJSON)

	if !identityAlreadyExists {
		fmt.Printf("   ⏳ Waiting for identity to propagate in Azure AD...\n")
		// Note: In production, you might want to add a sleep here to allow
		// the identity to fully propagate through Azure AD
	}

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

	// Get the principal ID from the identity
	principalID := identity.Credentials["principal_id"]
	if principalID == "" {
		return fmt.Errorf("principal ID not found in identity credentials")
	}

	// Parse the scope from serviceID
	scope := s.parseScope(serviceID)

	fmt.Printf("🔐 Granting %s access to %s for principal %s...\n", level, scope, principalID)

	// Create a unique name for the role assignment
	roleAssignmentName := uuid.New().String()

	// Create role assignment
	roleAssignmentParams := armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &principalID,
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

// DestroyUser removes a managed identity and all associated resources
func (s *AzureIAMService) DestroyUser(identity *Identity) error {
	identityName := identity.Credentials["identity_name"]
	if identityName == "" {
		// Try to extract from userName
		identityName = sanitizeManagedIdentityName(identity.UserName)
	}

	fmt.Printf("🗑️  Deleting managed identity %s...\n", identityName)

	// List and delete role assignments for this identity
	principalID := identity.Credentials["principal_id"]
	if principalID != "" {
		fmt.Printf("   🔍 Looking for role assignments for principal %s...\n", principalID)

		// List role assignments in the subscription
		filter := fmt.Sprintf("principalId eq '%s'", principalID)
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

	// Delete the managed identity
	_, err := s.msiClient.Delete(s.ctx, s.resourceGroup, identityName, nil)
	if err != nil {
		// Check if identity doesn't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "NotFound") {
			fmt.Printf("   ℹ️  Managed identity already deleted\n")
			return nil
		}
		return fmt.Errorf("failed to delete managed identity %s: %w", identityName, err)
	}

	fmt.Printf("   ✅ Managed identity deleted\n")
	return nil
}

// Helper functions

func (s *AzureIAMService) getRoleDefinitionForLevel(serviceID string, level string) (string, error) {
	// Azure built-in role definition IDs
	// Format: /subscriptions/{subscriptionId}/providers/Microsoft.Authorization/roleDefinitions/{roleId}

	baseRolePath := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions", s.subscriptionID)

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
				s.subscriptionID, s.resourceGroup, accountName)
		}
	}

	// Default: use resource group scope
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", s.subscriptionID, s.resourceGroup)
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

func sanitizeManagedIdentityName(userName string) string {
	// Managed identity names: alphanumeric, hyphens, underscores, 3-128 chars
	result := ""
	for _, char := range userName {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' {
			result += string(char)
		} else {
			result += "-"
		}
	}

	// Ensure minimum length
	if len(result) < 3 {
		result = result + "-test"
	}

	// Ensure maximum length
	if len(result) > 128 {
		result = result[:128]
	}

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	return result
}

func toPtr(s string) *string {
	return &s
}
