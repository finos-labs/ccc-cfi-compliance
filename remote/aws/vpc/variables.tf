# --- Infrastructure ---

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.20.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones to use for the public subnets (must align with public_subnet_cidrs length)"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]

  validation {
    condition     = length(var.availability_zones) == length(var.public_subnet_cidrs)
    error_message = "availability_zones length must match public_subnet_cidrs length."
  }
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets (one per AZ)"
  type        = list(string)
  default     = ["10.20.0.0/24", "10.20.1.0/24"]

  validation {
    condition     = length(var.public_subnet_cidrs) > 0
    error_message = "public_subnet_cidrs must include at least one subnet."
  }
}

# --- Naming & Tags ---

variable "target_id" {
  description = "CFI Target ID (e.g., aws-vpc)"
  type        = string
  default     = "local-test"
}

variable "owner" {
  description = "Owner tag value applied to all resources."
  type        = string
  default     = "cfi-owner"
}

variable "team" {
  description = "Team tag value applied to all resources."
  type        = string
  default     = "cfi-team"
}

variable "common_tags" {
  description = "Additional common tags merged into all resources."
  type        = map(string)
  default     = {}
}

variable "github_run_id" {
  description = "GitHub Actions run ID"
  type        = string
  default     = "local"
}

variable "github_repository" {
  description = "GitHub repository"
  type        = string
  default     = "local"
}

# --- CN01 ---

# CN01 is observational on AWS (default VPC existence is account/region state).
# This flag is documentation-only and exported via outputs for traceability.
variable "cn01_observation_only" {
  description = "Keep true to indicate CN01 is evaluated as an observational control in this module."
  type        = bool
  default     = true
}

# --- CN02 ---

variable "map_public_ip_on_launch" {
  description = "Whether instances launched in public subnets receive a public IP by default (set true to simulate CCC.VPC.CN02 failure; set false to satisfy CCC.VPC.CN02)."
  type        = bool
  default     = true
}

