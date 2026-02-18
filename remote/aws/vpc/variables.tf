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

variable "map_public_ip_on_launch" {
  description = "Whether instances launched in public subnets receive a public IP by default (set true to simulate CCC.VPC.CN02 failure; set false to satisfy CCC.VPC.CN02)."
  type        = bool
  default     = true
}

# CN01 is observational on AWS (default VPC existence is account/region state).
# This flag is documentation-only and exported via outputs for traceability.
variable "cn01_observation_only" {
  description = "Keep true to indicate CN01 is evaluated as an observational control in this module."
  type        = bool
  default     = true
}

# CN03: dry-run peering matrix and optional guardrail declaration.
variable "cn03_profile" {
  description = "CN03 fixture profile. Use create to provision requester trial VPCs; use external to provide external requester IDs."
  type        = string
  default     = "create"

  validation {
    condition     = contains(["create", "external"], lower(var.cn03_profile))
    error_message = "cn03_profile must be either create or external."
  }
}

variable "cn03_create_peers" {
  description = "Whether to create dedicated requester VPCs for CN03 allow/disallow dry-run trials."
  type        = bool
  default     = true
}

variable "cn03_allowed_requester_vpc_cidrs" {
  description = "CIDR blocks for requester VPCs expected to be allowed by peering guardrails."
  type        = list(string)
  default     = ["10.40.0.0/20", "10.40.16.0/20"]

  validation {
    condition     = (!var.cn03_create_peers) || length(var.cn03_allowed_requester_vpc_cidrs) > 0
    error_message = "cn03_allowed_requester_vpc_cidrs must contain at least one CIDR when cn03_create_peers is true."
  }
}

variable "cn03_disallowed_requester_vpc_cidrs" {
  description = "CIDR blocks for requester VPCs expected to be denied by peering guardrails."
  type        = list(string)
  default     = ["10.30.0.0/20", "10.30.16.0/20"]

  validation {
    condition     = (!var.cn03_create_peers) || length(var.cn03_disallowed_requester_vpc_cidrs) > 0
    error_message = "cn03_disallowed_requester_vpc_cidrs must contain at least one CIDR when cn03_create_peers is true."
  }
}

variable "cn03_create_non_allowlisted_requester" {
  description = "Whether to create a dedicated requester VPC intentionally outside both allowed/disallowed lists for CN03 MUST coverage."
  type        = bool
  default     = true
}

variable "cn03_non_allowlisted_requester_vpc_cidr" {
  description = "CIDR block for dedicated CN03 requester VPC that is intentionally not in the explicit allow/disallow lists."
  type        = string
  default     = "10.50.0.0/20"
}

variable "cn03_apply_guardrail" {
  description = "Whether to create a CN03 IAM deny policy for CreateVpcPeeringConnection."
  type        = bool
  default     = true
}

variable "cn03_allowed_account_ids" {
  description = "Compatibility-only output input from older CN03 account-based guardrail mode."
  type        = list(string)
  default     = []
}

variable "cn03_guardrail_attach_role_names" {
  description = "IAM role names that should receive the CN03 guardrail policy attachment."
  type        = list(string)
  default     = []
}

variable "cn03_guardrail_attach_user_names" {
  description = "IAM user names that should receive the CN03 guardrail policy attachment."
  type        = list(string)
  default     = []
}

variable "cn03_existing_guardrail_policy_arn" {
  description = "Existing managed IAM policy ARN to update in-place for CN03 guardrail. When set, Terraform updates this policy document instead of creating a new managed policy."
  type        = string
  default     = ""
}

variable "cn03_peer_class_tag_key" {
  description = "Tag key used to classify CN03 requester VPCs as allowed/disallowed."
  type        = string
  default     = "PeerClass"
}

variable "cn03_allowed_peer_tag_value" {
  description = "Tag value identifying VPCs allowed as requester sources in CN03 guardrail."
  type        = string
  default     = "allowed"
}

variable "cn03_disallowed_peer_tag_value" {
  description = "Tag value identifying VPCs classified as disallowed for CN03 test evidence."
  type        = string
  default     = "disallowed"
}

variable "cn03_allowed_accepter_vpc_arns" {
  description = "Optional explicit requester VPC ARN allow-list additions for CN03 guardrail (in addition to tag discovery). Variable name is kept for backward compatibility."
  type        = list(string)
  default     = []
}

variable "cn03_peer_owner_id" {
  description = "Optional peer owner account ID used by CN03 dry-run calls for cross-account peering."
  type        = string
  default     = ""
}

# CN04 (Policy/Behavior): flow-log configuration evidence.
variable "cn04_enable_flow_logs" {
  description = "Enable VPC flow logs on the primary VPC (recommended true for CN04 policy checks)."
  type        = bool
  default     = true
}

variable "cn04_flow_log_traffic_type" {
  description = "Traffic type for CN04 flow logs: ALL, ACCEPT, or REJECT."
  type        = string
  default     = "ALL"

  validation {
    condition     = contains(["ALL", "ACCEPT", "REJECT"], upper(var.cn04_flow_log_traffic_type))
    error_message = "cn04_flow_log_traffic_type must be one of: ALL, ACCEPT, REJECT."
  }
}

variable "cn04_flow_log_retention_days" {
  description = "Retention period for CN04 CloudWatch log group."
  type        = number
  default     = 7
}

variable "cn04_flow_log_log_group_prefix" {
  description = "Prefix for CN04 CloudWatch log group name."
  type        = string
  default     = "/aws/vpc/flow-logs"
}
