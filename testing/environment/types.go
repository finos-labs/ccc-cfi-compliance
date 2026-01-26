package environment

// TestParams holds the parameters for port / service testing
// This is the single shared structure used by both cloud api and reporters
type TestParams struct {
	PortNumber          string      // Leave blank if not applicable (e.g., for services without specific ports)
	HostName            string      // Hostname or endpoint
	Protocol            string      // Protocol (e.g., "tcp", "https")
	ServiceType         string      // Type of service (e.g., "s3", "rds", "storage") - DEPRECATED, use ProviderServiceType
	ProviderServiceType string      // Cloud provider-specific service type (e.g., "s3", "rds", "Microsoft.Storage/storageAccounts")
	CatalogTypes        []string    // CCC catalog types to test with (e.g., "CCC.ObjStor", "CCC.RDMS", "CCC.VM", "CCC.Core")
	Labels              []string    // Tags/labels from the resource
	UID                 string      // Unique identifier (ARN, resource ID, etc.)
	ResourceName        string      // Human-readable resource name extracted from ARN or resource ID
	CloudParams         CloudParams // details of the cloud environment
}

type CloudParams struct {
	Provider            string // Cloud provider ("aws", "azure", "gcp")
	Region              string // Cloud region
	AzureResourceGroup  string // Azure resource group name
	AzureSubscriptionID string // Azure subscription ID
	AzureStorageAccount string // Azure storage account name
	GCPProjectID        string // GCP Project ID
}

// ServiceTypes contains all known service types
var ServiceTypes = []string{
	"object-storage",
	"block-storage",
	"relational-database",
	"iam",
	"load-balancer",
	"security-group",
	"vpc",
}
