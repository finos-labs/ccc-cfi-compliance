# Used by CN03 guardrail to scope the allow-list to the deploying account when
# no explicit cn03_allowed_account_ids are provided.
data "aws_caller_identity" "current" {}

locals {
  name = "cfi-${var.instance_id}-vpc"

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
  cn03_caller_arn                    = data.aws_caller_identity.current.arn
  cn03_caller_arn_resource           = try(split(":", local.cn03_caller_arn)[5], "")
  cn03_caller_is_user                = startswith(local.cn03_caller_arn_resource, "user/")
  cn03_caller_is_assumed_role        = startswith(local.cn03_caller_arn_resource, "assumed-role/")
  cn03_caller_is_role                = startswith(local.cn03_caller_arn_resource, "role/")
  cn03_caller_user_name              = local.cn03_caller_is_user ? element(reverse(split("/", local.cn03_caller_arn_resource)), 0) : ""
  cn03_caller_role_name = local.cn03_caller_is_assumed_role ? try(split("/", local.cn03_caller_arn_resource)[1], "") : (
    local.cn03_caller_is_role ? element(reverse(split("/", local.cn03_caller_arn_resource)), 0) : ""
  )
  cn03_guardrail_role_names = compact(distinct(concat(
    var.cn03_guardrail_attach_role_names,
    var.cn03_attach_guardrail_to_caller_principal ? [local.cn03_caller_role_name] : [],
    [try(aws_iam_role.cn03_test_role[0].name, "")]
  )))
  cn03_guardrail_user_names = compact(distinct(concat(
    var.cn03_guardrail_attach_user_names,
    var.cn03_attach_guardrail_to_caller_principal ? [local.cn03_caller_user_name] : []
  )))
  cn03_existing_guardrail_policy_arn = trimspace(var.cn03_existing_guardrail_policy_arn)
  cn03_guardrail_policy_name         = trimspace(var.cn03_guardrail_policy_name)
  cn03_target_guardrail_policy_arn = local.cn03_existing_guardrail_policy_arn != "" ? local.cn03_existing_guardrail_policy_arn : format(
    "arn:aws:iam::%s:policy/%s",
    data.aws_caller_identity.current.account_id,
    local.cn03_guardrail_policy_name,
  )
  cn03_target_guardrail_policy_name = local.cn03_existing_guardrail_policy_arn != "" ? element(reverse(split("/", local.cn03_existing_guardrail_policy_arn)), 0) : local.cn03_guardrail_policy_name
  cn03_guardrail_policy_arn         = var.cn03_apply_guardrail ? local.cn03_target_guardrail_policy_arn : null
  cn03_guardrail_policy_mode        = var.cn03_apply_guardrail ? "reconcile" : "disabled"

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
    allowed_accepter_vpc_arns_json  = jsonencode([module.vpc.vpc_arn])
  })

  # ---------------------------------------------------------------------------
  # CN04 — VPC flow logs must capture all traffic (CCC.VPC.CN04)
  # ---------------------------------------------------------------------------

  cn04_flow_log_role_name = substr("cfi-${var.instance_id}-cn04-flowlogs-role", 0, 64)
  cn04_log_group_name     = "${var.cn04_flow_log_log_group_prefix}/${local.name}"
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
# BAD VPC — Intentionally non-compliant VPC for CN02/CN04 negative test scenarios.
# map_public_ip_on_launch=true violates CCC.VPC.CN02 (no flow logs = CN04 violation).
# No aws_flow_log resources are attached to this VPC by design.
# =============================================================================

module "bad_vpc" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v5.7.0"

  name = "${local.name}-bad"
  cidr = var.bad_vpc_cidr

  azs            = var.availability_zones
  public_subnets = [for i in range(length(var.availability_zones)) : cidrsubnet(var.bad_vpc_cidr, 8, i)]

  enable_dns_support   = true
  enable_dns_hostnames = true

  enable_nat_gateway = false
  enable_vpn_gateway = false

  create_igw              = true
  map_public_ip_on_launch = true

  tags = merge(local.common_resource_tags, {
    CFIControlSet = "CCC.VPC"
    CFIVpcRole    = "bad"
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
    CFIVpcRole = "cn03-peer-test-vpc"
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
    CFIVpcRole = "cn03-peer-test-vpc"
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
    CFIVpcRole = "cn03-peer-test-vpc"
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

resource "aws_iam_role" "cn03_test_role" {
  count = var.cn03_apply_guardrail && var.cn03_create_test_role ? 1 : 0

  name        = "${local.name}-cn03-test-role"
  description = "Dedicated IAM role for CN03 guardrail enforcement testing."
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn03-test-role"
    CFIControl = "CCC.VPC.CN03"
  })
}

resource "null_resource" "cn03_guardrail_reconcile" {
  count = var.cn03_apply_guardrail ? 1 : 0

  triggers = {
    policy_arn  = local.cn03_target_guardrail_policy_arn
    policy_name = local.cn03_target_guardrail_policy_name
    policy_sha  = sha256(local.cn03_guardrail_policy_json)
    policy_b64  = base64encode(local.cn03_guardrail_policy_json)
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
      POLICY_NAME='${self.triggers.policy_name}'
      TMP_POLICY_FILE="$$(mktemp)"
      echo '${self.triggers.policy_b64}' | base64 --decode > "$${TMP_POLICY_FILE}"
      if aws iam get-policy --policy-arn "$${POLICY_ARN}" >/dev/null 2>&1; then
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
      else
        aws iam create-policy \
          --policy-name "$${POLICY_NAME}" \
          --policy-document "file://$${TMP_POLICY_FILE}" \
          >/dev/null
      fi

      rm -f "$${TMP_POLICY_FILE}"
    EOT
  }
}

resource "aws_iam_role_policy_attachment" "cn03_guardrail" {
  count = var.cn03_apply_guardrail ? length(local.cn03_guardrail_role_names) : 0

  role       = local.cn03_guardrail_role_names[count.index]
  policy_arn = local.cn03_guardrail_policy_arn
  depends_on = [null_resource.cn03_guardrail_reconcile]
}

resource "aws_iam_user_policy_attachment" "cn03_guardrail" {
  count = var.cn03_apply_guardrail ? length(local.cn03_guardrail_user_names) : 0

  user       = local.cn03_guardrail_user_names[count.index]
  policy_arn = local.cn03_guardrail_policy_arn
  depends_on = [null_resource.cn03_guardrail_reconcile]
}

# =============================================================================
# CN04 — VPC flow logs must capture all traffic (CCC.VPC.CN04)
# Creates a CloudWatch log group, IAM role/policy, and aws_flow_log resource
# scoped to the base VPC. All resources are gated by cn04_enable_flow_logs
# flag — disabled by default.
# =============================================================================

resource "aws_cloudwatch_log_group" "cn04_flow_logs" {
  count = var.cn04_enable_flow_logs ? 1 : 0

  name              = local.cn04_log_group_name
  retention_in_days = var.cn04_flow_log_retention_days

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn04-flow-logs"
    CFIControl = "CCC.VPC.CN04"
  })
}

resource "aws_iam_role" "cn04_flow_logs" {
  count = var.cn04_enable_flow_logs ? 1 : 0

  name = local.cn04_flow_log_role_name
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "vpc-flow-logs.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn04-flow-logs-role"
    CFIControl = "CCC.VPC.CN04"
  })
}

resource "aws_iam_role_policy" "cn04_flow_logs" {
  count = var.cn04_enable_flow_logs ? 1 : 0

  name = "${local.cn04_flow_log_role_name}-policy"
  role = aws_iam_role.cn04_flow_logs[0].id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:DescribeLogGroups",
          "logs:DescribeLogStreams",
          "logs:PutLogEvents"
        ]
        Resource = [
          aws_cloudwatch_log_group.cn04_flow_logs[0].arn,
          "${aws_cloudwatch_log_group.cn04_flow_logs[0].arn}:*"
        ]
      }
    ]
  })
}

resource "aws_flow_log" "cn04_vpc" {
  count = var.cn04_enable_flow_logs ? 1 : 0

  vpc_id               = module.vpc.vpc_id
  log_destination_type = "cloud-watch-logs"
  log_destination      = aws_cloudwatch_log_group.cn04_flow_logs[0].arn
  iam_role_arn         = aws_iam_role.cn04_flow_logs[0].arn
  traffic_type         = upper(var.cn04_flow_log_traffic_type)

  tags = merge(local.common_resource_tags, {
    Name       = "${local.name}-cn04-flow-log"
    CFIControl = "CCC.VPC.CN04"
  })

  depends_on = [aws_iam_role_policy.cn04_flow_logs]
}

