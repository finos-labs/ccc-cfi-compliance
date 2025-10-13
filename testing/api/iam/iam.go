package iam

// AccessLevel defines the level of access for a service
type AccessLevel string

const (
	AccessLevelRead  AccessLevel = "read"
	AccessLevelWrite AccessLevel = "write"
	AccessLevelAdmin AccessLevel = "admin"
)

// IAMService provides identity and access management operations
type IAMService interface {
	// ProvisionUser creates a new user/identity in the cloud provider
	// Returns the created Identity with credentials
	// returns an existing user if it already exists
	ProvisionUser(userName string) (*Identity, error)

	// SetAccess grants an identity access to a specific service at the specified level
	// serviceID is the cloud service identifier (ARN, resource ID, etc.)
	// level specifies the access level: "none", "read", "write", or "admin"
	// makes no change if the level is already set.
	SetAccess(identity *Identity, serviceID string, level string) error

	// DestroyUser removes the identity and all associated access
	// does nothing if the user does not exist
	DestroyUser(identity *Identity) error
}
