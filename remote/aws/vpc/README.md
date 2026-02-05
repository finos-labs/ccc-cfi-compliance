# AWS VPC Test Environment

This Terraform configuration provisions a minimal AWS VPC suitable for running `CCC.VPC` compliance tests.
It uses the upstream `terraform-aws-modules/terraform-aws-vpc` module (pinned) to stay consistent with the other AWS targets in this repo (e.g., S3/RDS).

## What it creates

- 1 VPC
- 1 Internet Gateway
- public routing for the public subnets (`0.0.0.0/0` → IGW)
- N public subnets (default 2) with configurable default external IP assignment (`map_public_ip_on_launch`, used to simulate PASS/FAIL for `CCC.VPC.CN02.AR01`)

## How to apply locally

```bash
cd remote/aws/vpc
terraform init
terraform apply
```

To change region or subnet/AZ layout, set variables (example):

```bash
export TF_VAR_aws_region=us-east-1
export TF_VAR_availability_zones='["us-east-1a","us-east-1b"]'
export TF_VAR_public_subnet_cidrs='["10.20.0.0/24","10.20.1.0/24"]'
```

To simulate `CCC.VPC.CN02.AR01` outcomes:

```bash
# Fail case (default): public subnets assign external IPs by default
export TF_VAR_map_public_ip_on_launch=true

# Pass case: public subnets do NOT assign external IPs by default
export TF_VAR_map_public_ip_on_launch=false
```

## How to destroy

```bash
terraform destroy
```
