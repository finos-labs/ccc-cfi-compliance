locals {
  name = "cfi-${var.target_id}-vpc"

  # ---------------------------------------------------------------------------
  # Naming & Tags — common tags applied to all resources via provider default_tags
  # ---------------------------------------------------------------------------

  common_resource_tags = merge(
    {
      Owner            = var.owner
      team             = var.team
      Environment      = "cfi-test"
      ManagedBy        = "Terraform"
      Project          = "CCC-CFI-Compliance"
      AutoCleanup      = "true"
      CFITargetID      = var.target_id
      GitHubWorkflow   = "CFI-Build"
      GitHubRunID      = var.github_run_id
      GitHubRepository = var.github_repository
    },
    var.common_tags
  )
}

# =============================================================================
# BASE — Shared test VPC (used by all CN01–CN04 controls)
# Provisions the primary VPC and public subnets that serve as the test target
# across all CCC.VPC controls. CN02 behaviour is controlled by
# map_public_ip_on_launch.
# =============================================================================

module "vpc" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v5.7.0"

  name = local.name
  cidr = var.vpc_cidr

  azs            = var.availability_zones
  public_subnets = var.public_subnet_cidrs

  enable_dns_support   = true
  enable_dns_hostnames = true

  enable_nat_gateway = false
  enable_vpn_gateway = false

  create_igw              = true
  map_public_ip_on_launch = var.map_public_ip_on_launch

  tags = merge(local.common_resource_tags, {
    CFIControlSet = "CCC.VPC"
  })

  public_subnet_tags = merge(local.common_resource_tags, {
    Tier = "public"
  })
}

