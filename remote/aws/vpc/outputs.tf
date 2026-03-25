locals {
  cn03_allowed_requester_vpc_ids = sort(distinct(concat(
    [for key in sort(keys(aws_vpc.cn03_allowed_peer)) : aws_vpc.cn03_allowed_peer[key].id],
    sort(data.aws_vpcs.cn03_tagged_allowed.ids),
  )))
  cn03_disallowed_requester_vpc_ids = sort(distinct(concat(
    [for key in sort(keys(aws_vpc.cn03_disallowed_peer)) : aws_vpc.cn03_disallowed_peer[key].id],
    sort(data.aws_vpcs.cn03_tagged_disallowed.ids),
  )))

  cn03_allowed_requester_csv    = join(",", local.cn03_allowed_requester_vpc_ids)
  cn03_disallowed_requester_csv = join(",", local.cn03_disallowed_requester_vpc_ids)

  cn03_allowed_account_ids_csv          = join(",", local.cn03_allowed_account_ids_effective)
  cn03_non_allowlisted_requester_vpc_id = try(aws_vpc.cn03_non_allowlisted_requester[0].id, "")
  cn03_default_requester_vpc_id = coalesce(
    try(local.cn03_disallowed_requester_vpc_ids[0], ""),
    try(local.cn03_allowed_requester_vpc_ids[0], ""),
    ""
  )

  cn03_peer_trial_matrix = {
    schema_version               = 1
    receiver_vpc_id              = module.vpc.vpc_id
    peer_owner_id                = var.cn03_peer_owner_id
    allowed_requester_vpc_ids    = local.cn03_allowed_requester_vpc_ids
    disallowed_requester_vpc_ids = local.cn03_disallowed_requester_vpc_ids
    metadata = {
      profile = var.cn03_profile
      region  = var.aws_region
      source  = "remote/aws/vpc"
    }
  }

  cn03_feature_env = {
    CN03_REQUESTER_VPC_ID                 = local.cn03_default_requester_vpc_id
    CN03_RECEIVER_VPC_ID                  = module.vpc.vpc_id
    CN03_PEER_OWNER_ID                    = var.cn03_peer_owner_id
    CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID = local.cn03_non_allowlisted_requester_vpc_id
    CN03_ALLOWED_REQUESTER_VPC_IDS        = local.cn03_allowed_requester_csv
    CN03_DISALLOWED_REQUESTER_VPC_IDS     = local.cn03_disallowed_requester_csv
    CN03_ALLOWED_REQUESTER_VPC_ID_1       = try(local.cn03_allowed_requester_vpc_ids[0], "")
    CN03_ALLOWED_REQUESTER_VPC_ID_2       = try(local.cn03_allowed_requester_vpc_ids[1], "")
    CN03_DISALLOWED_REQUESTER_VPC_ID_1    = try(local.cn03_disallowed_requester_vpc_ids[0], "")
    CN03_DISALLOWED_REQUESTER_VPC_ID_2    = try(local.cn03_disallowed_requester_vpc_ids[1], "")
  }
}

output "vpc_id" {
  description = "Created VPC ID"
  value       = module.vpc.vpc_id
}

output "public_subnet_ids" {
  description = "Created public subnet IDs"
  value       = module.vpc.public_subnets
}

output "common_resource_tags" {
  description = "Resolved common tag set applied to all resources."
  value       = local.common_resource_tags
}

output "cn03_allowed_requester_vpc_ids" {
  description = "CN03 requester VPC IDs expected to be allowed."
  value       = local.cn03_allowed_requester_vpc_ids
}

output "cn03_disallowed_requester_vpc_ids" {
  description = "CN03 requester VPC IDs expected to be denied."
  value       = local.cn03_disallowed_requester_vpc_ids
}

output "cn03_non_allowlisted_requester_vpc_id" {
  description = "CN03 requester VPC ID that is intentionally outside allow/disallow lists."
  value       = try(aws_vpc.cn03_non_allowlisted_requester[0].id, null)
}

output "cn03_guardrail_policy_arn" {
  description = "ARN of the CN03 guardrail policy when enabled."
  value       = local.cn03_guardrail_policy_arn
}

output "cn03_guardrail_status" {
  description = "CN03 guardrail declaration summary."
  value = {
    applied                       = var.cn03_apply_guardrail
    condition_key                 = "ec2:RequesterVpc"
    allowed_account_ids           = local.cn03_allowed_account_ids_effective
    allowed_accepter_vpc_arns     = local.cn03_allowed_accepter_vpc_arns
    disallowed_accepter_vpc_arns  = local.cn03_disallowed_accepter_vpc_arns
    discovered_allowed_vpc_ids    = sort(data.aws_vpcs.cn03_tagged_allowed.ids)
    discovered_disallowed_vpc_ids = sort(data.aws_vpcs.cn03_tagged_disallowed.ids)
    attached_role_names           = local.cn03_guardrail_role_names
    attached_user_names           = local.cn03_guardrail_user_names
    policy_mode                   = local.cn03_guardrail_policy_mode
    existing_policy_arn_input     = local.cn03_existing_guardrail_policy_arn
    guardrail_policy_arn          = local.cn03_guardrail_policy_arn
  }
}

output "cn03_peer_trial_matrix" {
  description = "Structured CN03 dry-run trial matrix for file export and batch API checks."
  value       = local.cn03_peer_trial_matrix
}

output "cn03_peer_trial_matrix_json" {
  description = "JSON string form of the CN03 trial matrix."
  value       = jsonencode(local.cn03_peer_trial_matrix)
}

output "cn03_feature_env" {
  description = "Feature-ready CN03 environment variables derived from IaC outputs."
  value       = local.cn03_feature_env
}

output "CN03_REQUESTER_VPC_ID" {
  description = "Default requester VPC ID for single-case CN03 dry-run checks."
  value       = local.cn03_default_requester_vpc_id
}

output "CN03_RECEIVER_VPC_ID" {
  description = "Receiver VPC ID for CN03 dry-run checks."
  value       = module.vpc.vpc_id
}

output "CN03_PEER_OWNER_ID" {
  description = "Optional peer owner account ID for CN03 dry-run checks."
  value       = var.cn03_peer_owner_id
}

output "CN03_ALLOWED_ACCOUNT_IDS" {
  description = "CSV allow-list of account IDs used by CN03 guardrail condition."
  value       = local.cn03_allowed_account_ids_csv
}

output "CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID" {
  description = "Requester VPC ID intentionally outside explicit allow/disallow lists."
  value       = local.cn03_non_allowlisted_requester_vpc_id
}

output "CN03_ALLOWED_REQUESTER_VPC_IDS" {
  description = "CSV allow-list of requester VPC IDs."
  value       = local.cn03_allowed_requester_csv
}

output "CN03_DISALLOWED_REQUESTER_VPC_IDS" {
  description = "CSV list of requester VPC IDs expected to be denied."
  value       = local.cn03_disallowed_requester_csv
}

output "CN03_ALLOWED_REQUESTER_VPC_ID_1" {
  description = "First allowed requester VPC ID."
  value       = try(local.cn03_allowed_requester_vpc_ids[0], "")
}

output "CN03_ALLOWED_REQUESTER_VPC_ID_2" {
  description = "Second allowed requester VPC ID."
  value       = try(local.cn03_allowed_requester_vpc_ids[1], "")
}

output "CN03_DISALLOWED_REQUESTER_VPC_ID_1" {
  description = "First disallowed requester VPC ID."
  value       = try(local.cn03_disallowed_requester_vpc_ids[0], "")
}

output "CN03_DISALLOWED_REQUESTER_VPC_ID_2" {
  description = "Second disallowed requester VPC ID."
  value       = try(local.cn03_disallowed_requester_vpc_ids[1], "")
}
