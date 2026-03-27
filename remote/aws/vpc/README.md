# AWS VPC Test Environment

This Terraform configuration provisions a minimal AWS VPC environment for `CCC.VPC` testing.

## Minimum orchestration policy (per CN)

- `CN01` (default network resources): observational only on AWS.
  - This module does not create/remove AWS account default VPC resources.
  - Declaration flag: `cn01_observation_only=true`.
- `CN02` (default public IP in public subnet): declarative subnet setting.
  - Controlled by `map_public_ip_on_launch`.
- `CN03` (restrict peering destinations): fixture VPCs and guardrail policy — added in a separate module extension.
- `CN04` (flow logs): flow-log infrastructure — added in a separate module extension.

## What it creates

- Primary VPC with public subnets and IGW (always)

## Cleanup tagging

Resources are tagged with:

- `Owner=<value from TF_VAR_owner>`
- `CFITargetID=<target id>`

Use these tags for targeted cleanup when runs fail mid-way.

You can extend the shared tag set:

```bash
export TF_VAR_owner="platform-security"
export TF_VAR_common_tags='{"CostCenter":"1234","DataClass":"internal"}'
```

## Local apply

```bash
cd remote/aws/vpc
terraform init
terraform apply -auto-approve -input=false
```

## Common control profiles

### CN02 pass/fail simulation

```bash
# Fail case: public subnets auto-assign public IP
export TF_VAR_map_public_ip_on_launch=true

# Pass case: public subnets do not auto-assign public IP
export TF_VAR_map_public_ip_on_launch=false
```

## Destroy

```bash
terraform destroy -auto-approve -input=false
```
