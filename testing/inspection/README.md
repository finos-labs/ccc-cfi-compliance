# Inspection Package

This package provides the core data structures for CCC compliance testing.

## Components

### `types.go`

Contains the primary data structures:

- **`TestParams`**: Holds parameters for service testing including provider, resource name, catalog type, region, and UID
- **`AllCatalogTypes`**: List of all known CCC catalog types for filtering tests

## Catalog Types

The following catalog types are supported:

- `CCC.ObjStor` - Object Storage (S3, Azure Blob, GCS)
- `CCC.RDMS` - Relational Database Management System
- `CCC.VM` - Virtual Machines
- `CCC.Serverless` - Serverless Computing
- `CCC.Batch` - Batch Processing
- `CCC.Message` - Message Queue
- `CCC.GenAI` - Generative AI
- `CCC.MLDE` - Machine Learning Development Environment
- `CCC.KeyMgmt` - Key Management
- `CCC.Secrets` - Secrets Management
- `CCC.Vector` - Vector Database
- `CCC.Warehouse` - Data Warehouse
- `CCC.ContReg` - Container Registry
- `CCC.Build` - Build Service
- `CCC.IAM` - Identity and Access Management
- `CCC.AuditLog` - Audit Logging
- `CCC.Logging` - Logging
- `CCC.Monitoring` - Monitoring
- `CCC.VPC` - Virtual Private Cloud
- `CCC.LB` - Load Balancer
