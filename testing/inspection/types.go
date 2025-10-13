package inspection

// TestParams holds the parameters for port / service testing
// This is the single shared structure used by both inspection and reporters
type TestParams struct {
	PortNumber          string   // Leave blank if not applicable (e.g., for services without specific ports)
	HostName            string   // Hostname or endpoint
	Protocol            string   // Protocol (e.g., "tcp", "https")
	ServiceType         string   // Type of service (e.g., "s3", "rds", "storage") - DEPRECATED, use ProviderServiceType
	ProviderServiceType string   // Cloud provider-specific service type (e.g., "s3", "rds", "Microsoft.Storage/storageAccounts")
	CatalogType         string   // CCC catalog type (e.g., "CCC.ObjStor", "CCC.RDMS", "CCC.VM")
	Region              string   // Cloud region
	Provider            string   // Cloud provider ("aws", "azure", "gcp")
	Labels              []string // Tags/labels from the resource
	UID                 string   // Unique identifier (ARN, resource ID, etc.)
	ResourceName        string   // Human-readable resource name extracted from ARN or resource ID
}

// AllCatalogTypes contains all known CCC catalog types for tag filtering
var AllCatalogTypes = []string{
	"CCC.ObjStor",    // Object Storage
	"CCC.RDMS",       // Relational Database Management System
	"CCC.VM",         // Virtual Machines
	"CCC.Serverless", // Serverless Computing
	"CCC.Batch",      // Batch Processing
	"CCC.Message",    // Message Queue
	"CCC.GenAI",      // Generative AI
	"CCC.MLDE",       // Machine Learning Development Environment
	"CCC.KeyMgmt",    // Key Management
	"CCC.Secrets",    // Secrets Management
	"CCC.Vector",     // Vector Database
	"CCC.Warehouse",  // Data Warehouse
	"CCC.ContReg",    // Container Registry
	"CCC.Build",      // Build Service
	"CCC.IAM",        // Identity and Access Management
	"CCC.AuditLog",   // Audit Logging
	"CCC.Logging",    // Logging
	"CCC.Monitoring", // Monitoring
	"CCC.VPC",        // Virtual Private Cloud
	"CCC.LB",         // Load Balancer
}
