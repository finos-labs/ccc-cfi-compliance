# Cloud Service API

This package provides a unified interface for interacting with cloud service APIs across AWS, Azure, and GCP.

## Architecture

### Factory Pattern (`factory/`)

The factory pattern provides a consistent way to create cloud service clients:

```go
// Create a factory for a specific cloud provider
factory, err := factory.NewFactory(factory.ProviderAWS, cloudProps)

// Get a service API client
service, err := factory.GetServiceAPI("object-storage")
service, err := factory.GetServiceAPI("iam")

// Get a service API with a specific identity
identity, err := iamService.ProvisionUser("test-user")
service, err := factory.GetServiceAPIWithIdentity("object-storage", identity)
```

### Generic Service Interface (`generic/`)

The `Service` interface provides a common abstraction for all cloud services. Currently empty but will be extended with common operations.

### IAM Service (`iam/`)

The `IAMService` interface provides identity and access management operations:

- **ProvisionUser**: Create a new user/identity
- **SetAccess**: Grant access to a service at a specific level (read/write/admin)
- **DestroyUser**: Remove an identity and all associated access

```go
// Get IAM service from factory
iamService, err := factory.GetIAMService()

// Provision a new user
identity, err := iamService.ProvisionUser("test-user")

// Grant access to a service
err = iamService.SetAccess("test-user", "service-id", iam.AccessLevelRead)

// Remove the user
err = iamService.DestroyUser("test-user")
```

## Usage in Tests

These APIs will be used by the compliance test framework to:

1. Provision test users/identities
2. Grant specific access levels to test privilege escalation
3. Interact with services using different identities
4. Clean up test resources after testing

## VPC CN01-CN04 Test Reference (AWS)

Use these commands from repo root.

### Prerequisites

- AWS account access with permissions for VPC/IAM/Flow Logs operations used by your target CN.
- AWS CLI configured (`aws sts get-caller-identity` should succeed).
- Terraform installed (for IaC-based setup).
- `jq` installed (used by CN03 artifact/export tooling).
- Executable test runner: `testing/run-compliance-tests.sh`.

### Do I need IaC and an AWS account?

- `AWS account`: yes, for `--provider aws` runs.
- `IaC required`: not strictly for every CN if your account already has suitable VPC resources and guardrails.
- `IaC recommended`: yes, for repeatable/local validation and CI parity.

What IaC in `remote/aws/vpc` orchestrates:

- Base VPC and public subnets (shared test target).
- CN02 input shape via `map_public_ip_on_launch`.
- CN03 requester fixture VPCs (allowed/disallowed/non-allowlisted), plus optional IAM guardrail policy create/update/attach.
- CN03 trial artifacts export: `cn03-feature.env` and `cn03-peer-trials.json`.
- CN04 flow-log infrastructure (log group, IAM role/policy, VPC flow log) when enabled.

Typical sequence:

1. Configure AWS auth (`AWS_PROFILE`, region).
2. Apply IaC once for the control(s) you want to test.
3. Export CN03 artifacts when testing CN03.
4. Run `./testing/run-compliance-tests.sh ...` with CN tags.

### Quick start (IaC + tests)

```bash
# 1) Auth/context
export AWS_PROFILE=default
export REGION=us-east-1

# 2) IaC apply (shared base + CN fixtures controlled by TF_VARs)
cd remote/aws/vpc
terraform init
terraform apply -auto-approve -input=false

# 3) CN03 artifacts (only needed for CN03)
./export-cn03-artifacts.sh
source ./cn03-feature.env
export CN03_PEER_TRIAL_MATRIX_FILE="$(pwd)/cn03-peer-trials.json"
cd ../../..

# 4) Run tests
./testing/run-compliance-tests.sh --provider aws --region "$REGION" --service vpc --tag 'CCC.VPC.CN01.AR01 && MAIN'
./testing/run-compliance-tests.sh --provider aws --region "$REGION" --service vpc --tag 'CCC.VPC.CN02.AR01 && MAIN'
./testing/run-compliance-tests.sh --provider aws --region "$REGION" --service vpc --tag 'CCC.VPC.CN03.AR01 && MAIN'
./testing/run-compliance-tests.sh --provider aws --region "$REGION" --service vpc --tag 'CCC.VPC.CN04.AR01 && MAIN'
```

### Baseline environment

```bash
export AWS_PROFILE=default
export REGION=us-east-1
```

### CN01 - Default network resources absent (`CCC.VPC.CN01.AR01`)

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN01.AR01 && MAIN'
```

IaC note:
- Optional. CN01 is observational on AWS account default-network state.

### CN02 - No default external IP in public subnets (`CCC.VPC.CN02.AR01`)

Main policy check:

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN02.AR01 && MAIN'
```

Opt-in behavior check (creates and deletes test resource):

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN02.AR01 && OPT_IN'
```

IaC note:
- Recommended for deterministic pass/fail by setting `TF_VAR_map_public_ip_on_launch=true|false`.

### CN03 - Restrict peering from non-allowlisted requesters (`CCC.VPC.CN03.AR01`)

Preferred setup path:

- Use IaC-generated env + matrix file via `./export-cn03-artifacts.sh` and `source ./cn03-feature.env`.
- Use manual exports only when you intentionally need to override generated values.

Required env:

```bash
export CN03_RECEIVER_VPC_ID="<target-vpc-id>"
export CN03_ALLOWED_REQUESTER_VPC_ID_1="<allowed-requester-vpc-id-1>"
export CN03_ALLOWED_REQUESTER_VPC_ID_2="<allowed-requester-vpc-id-2>"
export CN03_DISALLOWED_REQUESTER_VPC_ID_1="<disallowed-requester-vpc-id-1>"
export CN03_DISALLOWED_REQUESTER_VPC_ID_2="<disallowed-requester-vpc-id-2>"
export CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID="<non-allowlisted-requester-vpc-id>"
```

Optional env:

```bash
export CN03_ALLOWED_REQUESTER_VPC_IDS="<csv-allowed-requester-vpc-ids>"
export CN03_PEER_OWNER_ID="<peer-account-id>"
export CN03_PEER_TRIAL_MATRIX_FILE="<abs-path-to-cn03-peer-trials.json>"
```

If using IaC artifacts from `remote/aws/vpc`:

```bash
cd remote/aws/vpc
./export-cn03-artifacts.sh
source ./cn03-feature.env
cd ../../..
```

Main enforcement checks:

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN03.AR01 && MAIN'
```

Opt-in sanity and matrix checks:

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN03.AR01 && SANITY && OPT_IN'
```

IaC note:
- Strongly recommended. CN03 depends on controlled requester sets and guardrail policy behavior.

CN03 result diagnostics (for failure reasoning):

- `AllowListDefined`: whether CN03 allow-list input was resolved.
- `RequesterInAllowList`: whether tested requester VPC is in the allow-list.
- `GuardrailExpectation`: expected runtime decision from allow-list (`allow` or `deny`).
- `GuardrailMismatch`: `true` when dry-run runtime outcome does not match allow-list expectation.
- `Reason`: includes explicit suffix:
  - `CN03 guardrail aligned: ...` when behavior matches expectation.
  - `CN03 guardrail mismatch: ...` when behavior differs (missing/misconfigured enforcement).

### CN04 - Flow logs capture all traffic (`CCC.VPC.CN04.AR01`)

Main policy check:

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN04.AR01 && MAIN'
```

Opt-in behavior check (generates traffic; may incur cloud cost):

```bash
./testing/run-compliance-tests.sh \
  --provider aws \
  --region "$REGION" \
  --service vpc \
  --tag 'CCC.VPC.CN04.AR01 && OPT_IN'
```

IaC note:
- Recommended. Enable with:
  - `TF_VAR_cn04_enable_flow_logs=true`
  - `TF_VAR_cn04_flow_log_traffic_type=ALL`

### Common failures

- `AuthFailure` / `could not validate provided access credentials`:
  - Re-check `AWS_PROFILE`, session/token expiry, and `aws sts get-caller-identity`.
- `UnauthorizedOperation` with explicit deny on `CN03PeeringGuardrail`:
  - Expected for disallowed/non-allowlisted CN03 requester tests.
- CN03 `GuardrailMismatch=true`:
  - Runtime peering enforcement does not match declared allow-list input; verify IAM/SCP guardrail policy and attachment to the executing identity.
- CN03 nil env placeholders in feature run:
  - Re-run `./export-cn03-artifacts.sh`, then `source ./cn03-feature.env`.
- CN04 behavior check fails to observe records:
  - Verify flow logs are enabled for target VPC and `traffic_type=ALL`.
