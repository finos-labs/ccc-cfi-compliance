# Used by CN03 guardrail to scope the allow-list to the deploying account when
# no explicit cn03_allowed_account_ids are provided.
data "aws_caller_identity" "current" {}

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

  # ---------------------------------------------------------------------------
  # CN03 — Restrict VPC peering to allowed requesters (CCC.VPC.CN03)
  # Locals used to build requester VPC fixtures, discover tagged peers, and
  # render the IAM guardrail policy that enforces the allow-list at runtime.
  # ---------------------------------------------------------------------------

  cn03_allowed_requester_cidr_map = var.cn03_create_peers ? {
    for index, cidr in var.cn03_allowed_requester_vpc_cidrs : format("%02d", index + 1) => cidr
  } : {}

  cn03_disallowed_requester_cidr_map = var.cn03_create_peers ? {
    for index, cidr in var.cn03_disallowed_requester_vpc_cidrs : format("%02d", index + 1) => cidr
  } : {}

  cn03_allowed_account_ids_effective = length(var.cn03_allowed_account_ids) > 0 ? var.cn03_allowed_account_ids : [data.aws_caller_identity.current.account_id]
  cn03_guardrail_role_names          = compact(var.cn03_guardrail_attach_role_names)
  cn03_guardrail_user_names          = compact(var.cn03_guardrail_attach_user_names)
  cn03_existing_guardrail_policy_arn = trimspace(var.cn03_existing_guardrail_policy_arn)
  cn03_create_guardrail_policy       = var.cn03_apply_guardrail && local.cn03_existing_guardrail_policy_arn == ""
  cn03_update_guardrail_policy       = var.cn03_apply_guardrail && local.cn03_existing_guardrail_policy_arn != ""
  cn03_guardrail_policy_arn          = local.cn03_update_guardrail_policy ? local.cn03_existing_guardrail_policy_arn : try(aws_iam_policy.cn03_guardrail[0].arn, null)
  cn03_guardrail_policy_mode         = local.cn03_update_guardrail_policy ? "existing" : (local.cn03_create_guardrail_policy ? "create" : "disabled")

  cn03_allowed_accepter_vpc_arns = distinct(concat(
    [for key in sort(keys(aws_vpc.cn03_allowed_peer)) : aws_vpc.cn03_allowed_peer[key].arn],
    [for key in sort(keys(data.aws_vpc.cn03_tagged_allowed)) : data.aws_vpc.cn03_tagged_allowed[key].arn],
    var.cn03_allowed_accepter_vpc_arns
  ))

  cn03_disallowed_accepter_vpc_arns = distinct(concat(
    [for key in sort(keys(aws_vpc.cn03_disallowed_peer)) : aws_vpc.cn03_disallowed_peer[key].arn],
    [for key in sort(keys(data.aws_vpc.cn03_tagged_disallowed)) : data.aws_vpc.cn03_tagged_disallowed[key].arn]
  ))

  cn03_guardrail_policy_json = templatefile("${path.module}/policies/cn03-guardrail-policy.json.tftpl", {
    allowed_requester_vpc_arns_json = jsonencode(local.cn03_allowed_accepter_vpc_arns)
  })
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

# =============================================================================
# CN03 — Restrict VPC peering to allowed requesters (CCC.VPC.CN03)
# Creates requester fixture VPCs (allowed/disallowed/non-allowlisted) and an
# optional IAM guardrail policy that denies CreateVpcPeeringConnection from
# requesters not in the allow-list. All resources are gated by cn03_create_peers
# and cn03_apply_guardrail flags — disabled by default.
# =============================================================================

resource "aws_vpc" "cn03_allowed_peer" {
  for_each = local.cn03_allowed_requester_cidr_map

  cidr_block           = each.value
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn03-allowed-requester-${each.key}"
    CFIControl = "CCC.VPC.CN03"
    PeerClass  = var.cn03_allowed_peer_tag_value
  })
}

resource "aws_vpc" "cn03_disallowed_peer" {
  for_each = local.cn03_disallowed_requester_cidr_map

  cidr_block           = each.value
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn03-disallowed-requester-${each.key}"
    CFIControl = "CCC.VPC.CN03"
    PeerClass  = var.cn03_disallowed_peer_tag_value
  })
}

resource "aws_vpc" "cn03_non_allowlisted_requester" {
  count = var.cn03_create_peers && var.cn03_create_non_allowlisted_requester ? 1 : 0

  cidr_block           = var.cn03_non_allowlisted_requester_vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn03-non-allowlisted-requester-01"
    CFIControl = "CCC.VPC.CN03"
  })
}

data "aws_vpcs" "cn03_tagged_allowed" {
  filter {
    name   = "tag:${var.cn03_peer_class_tag_key}"
    values = [var.cn03_allowed_peer_tag_value]
  }
}

data "aws_vpc" "cn03_tagged_allowed" {
  for_each = toset(data.aws_vpcs.cn03_tagged_allowed.ids)
  id       = each.value
}

data "aws_vpcs" "cn03_tagged_disallowed" {
  filter {
    name   = "tag:${var.cn03_peer_class_tag_key}"
    values = [var.cn03_disallowed_peer_tag_value]
  }
}

data "aws_vpc" "cn03_tagged_disallowed" {
  for_each = toset(data.aws_vpcs.cn03_tagged_disallowed.ids)
  id       = each.value
}

resource "aws_iam_policy" "cn03_guardrail" {
  count = local.cn03_create_guardrail_policy ? 1 : 0

  name_prefix = "${local.name}-cn03-guardrail-"
  description = "CN03 guardrail policy for CreateVpcPeeringConnection requester VPC allow-list enforcement."
  policy      = local.cn03_guardrail_policy_json

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn03-guardrail"
    CFIControl = "CCC.VPC.CN03"
  })

  lifecycle {
    precondition {
      condition     = length(local.cn03_allowed_accepter_vpc_arns) > 0
      error_message = "CN03 guardrail requires at least one allowlisted requester VPC ARN. Add the allowed tag or set cn03_allowed_accepter_vpc_arns."
    }
  }
}

resource "null_resource" "cn03_guardrail_existing" {
  count = local.cn03_update_guardrail_policy ? 1 : 0

  triggers = {
    policy_arn = local.cn03_existing_guardrail_policy_arn
    policy_sha = sha256(local.cn03_guardrail_policy_json)
    policy_b64 = base64encode(local.cn03_guardrail_policy_json)
  }

  lifecycle {
    precondition {
      condition     = length(local.cn03_allowed_accepter_vpc_arns) > 0
      error_message = "CN03 guardrail requires at least one allowlisted requester VPC ARN. Add the allowed tag or set cn03_allowed_accepter_vpc_arns."
    }
  }

  provisioner "local-exec" {
    when        = create
    interpreter = ["/bin/bash", "-c"]
    command     = <<-EOT
      set -euo pipefail

      POLICY_ARN='${self.triggers.policy_arn}'
      TMP_POLICY_FILE="$$(mktemp)"
      echo '${self.triggers.policy_b64}' | base64 --decode > "$${TMP_POLICY_FILE}"

      VERSION_COUNT="$$(aws iam list-policy-versions --policy-arn "$${POLICY_ARN}" --query 'length(Versions)' --output text)"
      if [ "$${VERSION_COUNT}" -ge 5 ]; then
        OLDEST_NON_DEFAULT="$$(aws iam list-policy-versions --policy-arn "$${POLICY_ARN}" --query 'Versions[?IsDefaultVersion==`false`]|sort_by(@,&CreateDate)[0].VersionId' --output text)"
        if [ "$${OLDEST_NON_DEFAULT}" != "None" ] && [ -n "$${OLDEST_NON_DEFAULT}" ]; then
          aws iam delete-policy-version --policy-arn "$${POLICY_ARN}" --version-id "$${OLDEST_NON_DEFAULT}"
        fi
      fi

      aws iam create-policy-version \
        --policy-arn "$${POLICY_ARN}" \
        --policy-document "file://$${TMP_POLICY_FILE}" \
        --set-as-default \
        >/dev/null

      rm -f "$${TMP_POLICY_FILE}"
    EOT
  }
}

resource "aws_iam_role_policy_attachment" "cn03_guardrail" {
  count = var.cn03_apply_guardrail ? length(local.cn03_guardrail_role_names) : 0

  role       = local.cn03_guardrail_role_names[count.index]
  policy_arn = local.cn03_guardrail_policy_arn
}

resource "aws_iam_user_policy_attachment" "cn03_guardrail" {
  count = var.cn03_apply_guardrail ? length(local.cn03_guardrail_user_names) : 0

  user       = local.cn03_guardrail_user_names[count.index]
  policy_arn = local.cn03_guardrail_policy_arn
}

