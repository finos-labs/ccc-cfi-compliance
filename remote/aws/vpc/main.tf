locals {
  name = "cfi-${var.target_id}-vpc"
}

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

  public_subnet_tags = {
    Tier = "public"
  }
}
