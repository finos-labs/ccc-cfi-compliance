# CCC CFI Compliance Testing

This directory contains the testing infrastructure for running CCC (Common Cloud Controls) compliance tests against cloud resources.

## Overview

The testing system discovers cloud resources using native cloud provider APIs and runs appropriate Cucumber/Gherkin tests against them based on their catalog type.

## Architecture

### 1. Service Runners (`services/`)

Each CCC catalog type has its own service runner:

- **`ServiceRunner.go`**: Interface that all service runners implement
- **`AbstractServiceRunner.go`**: Base implementation with common test execution logic
- **`CCC.ObjStor/`**: Object Storage service runner
  - `CCCObjStorServiceRunner.go`: Implements resource discovery for object storage
  - `features/`: Gherkin feature files for object storage tests

### 2. Cloud APIs (`api/`)

Abstractions for interacting with cloud services:

- **`factory/`**: Factory pattern for creating cloud service clients
- **`iam/`**: Identity and Access Management operations
- **`object-storage/`**: Object storage operations (S3, Blob, GCS)
- **`generic/`**: Base service interface

### 3. Test Language (`language/`)

- **`cloud/`**: Cloud-specific test steps and runners
- **`generic/`**: Generic BDD steps for Gherkin tests
- **`reporters/`**: HTML and OCSF formatters for test output

### 4. Inspection (`inspection/`)

- **`types.go`**: Core data structures (`TestParams`, `AllCatalogTypes`)

## Usage

### Prerequisites

**Cloud credentials** must be configured for the provider you're testing:

- AWS: `aws configure` or environment variables
- Azure: `az login`
- GCP: `gcloud auth login`

### Quick Start: Object Storage Testing

Quick reference for setting up cloud storage infrastructure for CCC compliance testing.

#### 1. Cloud Provider Login

**AWS**

```bash
aws configure
# Or verify existing session:
aws sts get-caller-identity
```

**Azure**

```bash
az login
# Verify:
az account show
```

**GCP**

```bash
gcloud auth login
gcloud auth application-default login
# Verify:
gcloud auth list
```

---

#### 2. Deploy Object Storage Terraform Modules

**AWS S3 Bucket**

Module: [terraform-aws-modules/terraform-aws-s3-bucket](https://github.com/terraform-aws-modules/terraform-aws-s3-bucket)

```bash
cd remote/aws/s3bucket
terraform init
terraform plan
terraform apply
```

**Azure Storage Account**

Module: [Azure/terraform-azurerm-avm-res-storage-storageaccount](https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount)

```bash
cd remote/azure/storageaccount
terraform init
terraform plan
terraform apply
```

**GCP Cloud Storage**

Module: [terraform-google-modules/terraform-google-cloud-storage](https://github.com/terraform-google-modules/terraform-google-cloud-storage)

```bash
cd remote/gcp/cloudstorage
terraform init
terraform plan
terraform apply
```

---

#### 3. Run Compliance Tests

After deploying infrastructure:

```bash
./testing/run-compliance-tests.sh --provider aws
./testing/run-compliance-tests.sh --provider azure
./testing/run-compliance-tests.sh --provider gcp
```

All required variables are auto-loaded from `compliance-testing.env`.

## Adding New Service Mappings

To add support for a new cloud service:

1. Add an entry to the appropriate CSV file:

   ```csv
   provider_service_type,catalog_type,description
   new-service,CCC.NewCatalog,Description of service
   ```

2. If creating a new catalog type, add it to `AllCatalogTypes` in `inspection/types.go`:

   ```go
   var AllCatalogTypes = []string{
       // ... existing types ...
       "CCC.NewCatalog", // New Catalog Type
   }
   ```

3. Run tests to verify:
   ```bash
   cd inspection
   go test -v -run TestLookupCatalogType
   ```

## Troubleshooting

### Authentication Errors

If you encounter authentication errors, ensure your cloud credentials are properly configured:

### No Resources Found

```
Warning: Found 0 accessible port(s)
Warning: Found 0 service(s)
```

**Solution**:

1. Verify cloud credentials are configured correctly:
   - AWS: `aws sts get-caller-identity`
   - Azure: `az account show`
   - GCP: `gcloud auth list`
2. Ensure resources exist in the cloud provider
3. Check that your IAM permissions allow listing resources

### No Catalog Type Mapping

```
Skipping service (no catalog type mapping)
```

**Solution**: Add the service type to the appropriate CSV file in `inspection/`

## Development

### Running Unit Tests

```bash
# Test service mappings
cd inspection
go test -v

# Test specific functionality
go test -v -run TestLookupCatalogType
```

### Adding New Test Steps

Test step definitions are in `language/cloud/` and `language/generic/`.
