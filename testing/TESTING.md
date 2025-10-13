# CCC CFI Compliance Testing

This directory contains the testing infrastructure for running CCC (Common Cloud Controls) compliance tests against cloud resources.

## Overview

The testing system discovers cloud resources (ports and services) using Steampipe and runs appropriate Cucumber/Gherkin tests against them based on their catalog type.

## Components

### 1. Service Mapping System (`inspection/`)

Maps cloud provider-specific service types to CCC catalog types:

- **`types.go`**: Core data structures including `TestParams` with `ProviderServiceType` and `CatalogType`
- **`aws-services.csv`**: AWS service type → CCC catalog type mappings (e.g., s3 → CCC.ObjStor)
- **`azure-services.csv`**: Azure service type mappings (e.g., Microsoft.Storage/storageAccounts → CCC.ObjStor)
- **`gcp-services.csv`**: GCP service type mappings (e.g., storage.googleapis.com/Bucket → CCC.ObjStor)
- **`service_mappings.go`**: Registry for loading and looking up mappings
- **`steampipe.go`**: Functions to discover ports and services from cloud providers

### 2. Test Runners (`language/cloud/`)

- **`test_port.go`**: Runs `@PerPort` tests for discovered ports
- **`test_service.go`**: Runs `@PerService` tests for discovered services

### 3. Test Orchestration

- **`runner/main.go`**: Go CLI tool that discovers resources and orchestrates test execution
- **`run-compliance-tests.sh`**: Shell wrapper with user-friendly interface

## Usage

### Prerequisites

1. **Steampipe** must be installed and running:

   ```bash
   steampipe service start
   ```

2. **Cloud credentials** must be configured for the provider you're testing:
   - AWS: `aws configure` or environment variables
   - Azure: `az login`
   - GCP: `gcloud auth login`

### Running Tests

#### Using the Shell Script (Recommended)

```bash
# Test AWS resources
./testing/run-compliance-tests.sh --provider aws

# Test Azure resources with custom output directory
./testing/run-compliance-tests.sh --provider azure --output results

# Test GCP resources, skip port tests
./testing/run-compliance-tests.sh --provider gcp --skip-ports

# Custom features path and timeout
./testing/run-compliance-tests.sh \
  --provider aws \
  --features ./my-features \
  --timeout 1h
```

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

### Steampipe Not Running

```
Error: Steampipe is not running or not accessible
```

**Solution**: Start Steampipe:

```bash
steampipe service start
```

### No Resources Found

```
Warning: Found 0 accessible port(s)
Warning: Found 0 service(s)
```

**Solution**:

1. Verify cloud credentials are configured
2. Ensure resources exist in the cloud provider
3. Check Steampipe plugin installation:
   ```bash
   steampipe plugin list
   steampipe plugin install aws azure gcp
   ```

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
