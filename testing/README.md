# CCC CFI Compliance Testing

This directory contains the testing infrastructure for running CCC (Common Cloud Controls) compliance tests against cloud resources.

## Overview

The testing system discovers cloud resources using native cloud provider APIs and runs appropriate Cucumber/Gherkin tests against them based on their catalog type.

## Architecture

### 0. `run-compliance-tests.sh`

The main entry point for running compliance tests. This shell script:

- Loads environment variables from `compliance-testing.env`
- Parses command-line arguments for provider, region, and cloud-specific options
- Builds the Go test runner binary (`ccc-compliance`)
- Executes the runner with the configured parameters

### 1. Test Runner (`runner/`)

The test runner orchestrates compliance test execution:

- **`main.go`**: CLI entry point that:

  - Parses flags and builds `CloudParams` configuration
  - Iterates over all `ServiceTypes` defined in `environment/types.go`
  - Creates a `ServiceRunner` for each service type

- **`ServiceRunner.go`**: Interface that all service runners implement

- **`BasicServiceRunner.go`**: Default implementation that:

  1. Creates a cloud factory for the target provider (see below)
  2. Gets the service API via `factory.GetServiceAPI(serviceName)`
  3. Calls `GetOrProvisionTestableResources()` to discover resources. Each resource is captured in a `TestParams` object.
  4. For each returned `TestParams`, runs godog tests filtered by `CatalogTypes`
  5. Generates an HTML and OCSF report per resource

### 2. Cloud APIs (`api/`)

Abstractions for interacting with cloud services:

- **`factory/`**: Factory pattern for creating cloud service clients
  - `factory.go`: Main factory interface
  - `aws_factory.go`, `azure_factory.go`, `gcp_factory.go`: Provider implementations
- **`generic/`**: Base `Service` interface with methods like `GetOrProvisionTestableResources()`
- **`iam/`**: Identity and Access Management operations
- **`object-storage/`**: Object storage operations (S3, Azure Blob, GCS)
  - `elevation/`: Access elevation for testing locked-down resources
- _more to follow_

### 3. Features (`features/`)

Gherkin feature files organized by CCC catalog type:

- **`CCC.Core/`**: Core control feature files (e.g., `CCC-Core-CN01-AR01.feature`)
- **`CCC.ObjStor/`**: Object storage feature files (e.g., `CCC-ObjStor-CN01-AR01.feature`)

Features are tagged with their catalog type (e.g., `@CCC.ObjStor`) for automatic filtering.

There is one file per assessment requirement defined in CCC. Where we have different implementations of tests for different types of services, we tag them with the service type they are implemented for.

### 4. Test Language (`language/`)

Step definitions and utilities for BDD tests:

- **`generic/`**: Generic BDD steps for Gherkin tests, allowing you to call API methods and test results. See [`language/generic/README.md`](language/generic/README.md) for details
- **`cloud/`**: Cloud-specific test steps. See [`language/cloud/README.md`](language/cloud/README.md) for details.
- **`attachments/`**: Test attachment handling
- **`reporters/`**: HTML and OCSF formatters for test output
  - `html-formatter.go`: HTML report generation
  - `ocsf-formatter.go`: OCSF-compliant JSON output

### 5. Environment (`environment/`)

Core data structures and configuration:

- **`types.go`**: Core types including:
  - `TestParams`: Parameters for resource testing
  - `CloudParams`: Cloud provider configuration
  - `ServiceTypes`: List of supported service types
  - `PolicyDefinition`, `PolicyResult`: Policy evaluation structures

### 6. Policy (`policy/`)

Policy definitions for compliance checks:

- **`CCC.Core/`**: Core control policies organized by control number
- **`CCC.VPC/`**: VPC-specific policies
- Each policy is a YAML file specifying queries and validation rules

### 7. Output (`output/`)

Test results are written here:

- `resource-<name>.html`: HTML reports per resource
- `resource-<name>.ocsf.json`: OCSF JSON output per resource
- `combined.ocsf.json`: Combined OCSF output from all resources

## Usage

#### 1. Cloud Provider Login

**Cloud credentials** must be configured for the provider you're testing:

- AWS: `aws configure` or environment variables
- Azure: `az login`
- GCP: `gcloud auth login`

#### 2. Deploy Object Storage Terraform Modules

Install some terraform to test against. Some examples below:

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

#### 3. Run Compliance Tests

After deploying infrastructure:

```bash
./testing/run-compliance-tests.sh --provider aws
./testing/run-compliance-tests.sh --provider azure
./testing/run-compliance-tests.sh --provider gcp
```

All required variables are auto-loaded from `compliance-testing.env`, but you can override with command-line options if you want.

```
./run-compliance-tests.sh --help
```

#### 4. Review outputs

After completion, the `output` directory will contain an HTML and OCSF for each resource tested.

## Adding Support for New Services

To add support for a new cloud service:

1. **Add the service type** to `ServiceTypes` in `environment/types.go`:

```go
var ServiceTypes = []string{
    // ... existing types ...
    "new-service", // Your new service type
}
```

2. **Implement the Service interface** in `api/new-service/`:

```go
type NewService struct {
    // provider-specific clients
}

func (s *NewService) GetOrProvisionTestableResources() ([]environment.TestParams, error) {
    // Discover resources and return TestParams
}
```

3. **Register in the factory** (`api/factory/`):

```go
func (f *AWSFactory) GetServiceAPI(serviceName string) (generic.Service, error) {
    switch serviceName {
    case "new-service":
        return NewAWSNewService(f.cloudParams), nil
    // ...
    }
}
```

4. **Add feature files** in `features/CCC.NewCatalog/`:

```gherkin
@CCC.NewCatalog
Feature: CCC.NewCatalog.CN01 - Control Name
  Scenario: AR01 - Validation scenario
    Given the resource is configured
    Then the control requirement should be met
```

## Development

### Building Manually

```bash
cd testing
go build -o ccc-compliance ./runner/
```

### Running Tests Directly

```bash
./ccc-compliance \
  -provider aws \
  -region us-east-1 \
  -timeout 30m
```

### Adding New Test Steps

Ordinarily, you shouldn't need to add new steps to the framework as you can use the ones in generic to call all the API functions you need. Test step definitions are in:

- `language/cloud/cloud_steps.go` - Cloud-specific steps
- `language/generic/generic_steps.go` - Reusable generic steps

See the documentation in `language/generic/README.md` and `language/cloud/README.md`.
