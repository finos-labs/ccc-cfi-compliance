variable "location" {
  description = "Azure region where the function app should be deployed."
  type        = string
}

variable "name" {
  description = "Function app name."
  type        = string
}

variable "parent_id" {
  description = "Resource ID of the resource group in which to create the function app."
  type        = string
}

variable "service_plan_resource_id" {
  description = "Resource ID of the Flex Consumption (FC1) service plan that hosts the function app."
  type        = string
}

variable "fc1_runtime_name" {
  description = "Flex Consumption runtime name (deployment-specific; FC1 requires an explicit runtime)."
  type        = string
  default     = "dotnet-isolated"
}

variable "fc1_runtime_version" {
  description = "Flex Consumption runtime version."
  type        = string
  default     = "8.0"
}

variable "virtual_network_subnet_id" {
  description = "Resource ID of the delegated subnet used for outbound VNet integration."
  type        = string
}

variable "storage_container_endpoint" {
  description = "Blob container endpoint URL used as the Flex Consumption deployment package store."
  type        = string
}

variable "storage_user_assigned_identity_id" {
  description = "Resource ID of the user-assigned identity authorised to read the deployment container."
  type        = string
}

variable "function_app_uses_fc1" {
  description = <<-EOT
    Controls: CCC.SvlsComp.CN01, CCC.SvlsComp.CN02 [coverage: supporting]
    Host the app on the Flex Consumption (FC1) plan.
    Limitation: FC1 requires a matching Flex Consumption service plan (service_plan_resource_id) - deployment-specific.
    Limitation: Hosting-plan choice; not itself a complete expression of either control.
  EOT
  type        = bool
  default     = true
}

variable "client_certificate_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: supporting]
    Require incoming client certificates (mutual TLS) at the front door.
    Limitation: Caller cert distribution/validation is an operational concern; some PaaS-internal callers may not present a cert.
  EOT
  type        = bool
  default     = true
}

variable "client_certificate_mode" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: supporting]
    Client certificate enforcement mode (Required | Optional | OptionalInteractiveUser).
    Limitation: Caller cert distribution/validation is an operational concern; some PaaS-internal callers may not present a cert.
  EOT
  type        = string
  default     = "Required"
}

variable "https_only" {
  description = <<-EOT
    Controls: CCC.Core.CN01 [coverage: partial]
    Redirect all traffic to HTTPS.
    Limitation: Module default is already true; pinned here so the baseline is explicit and drift is caught.
  EOT
  type        = bool
  default     = true
}

variable "maximum_instance_count" {
  description = <<-EOT
    Controls: CCC.SvlsComp.CN02 [coverage: partial]
    Upper bound on the number of concurrent Flex Consumption instances.
    Limitation: Caps total concurrency, NOT per-entity invocation rate. True per-caller throttling (CN02 AR01) must be enforced upstream (APIM rate-limit policy or Front Door). 40 is an example ceiling - set per workload.
  EOT
  type        = number
  default     = 40
}

variable "public_network_access_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN05, CCC.SvlsComp.CN01 [coverage: partial]
    Whether the function app is reachable over the public internet.
    Limitation: Disabling public access is necessary but not sufficient for CN01 - a private endpoint must also be wired up for the app to remain reachable.
  EOT
  type        = bool
  default     = false
}

variable "managed_identities" {
  description = <<-EOT
    Controls: CCC.Core.CN03, CCC.Core.CN05 [coverage: supporting]
    Managed identities assigned to the function app.
    Limitation: Managed identity is the auth principal, not MFA; CN03 (MFA for human access) is satisfied via Entra Conditional Access, not this module.
  EOT
  type = object({
    system_assigned            = optional(bool, false)
    user_assigned_resource_ids = optional(set(string), [])
  })
  default = {
    system_assigned = true
  }
}

variable "site_config" {
  description = <<-EOT
    Controls: CCC.Core.CN01, CCC.Core.CN05, CCC.Core.CN10 [coverage: partial, supporting]
    Site configuration for the function app.
    Limitation: TLS floor only; does not itself prove cipher suite posture or PQC readiness.
    Limitation: Covers the management/deployment endpoint, not the request data path.
    Limitation: Disables FTPS entirely; if FTPS deployment is required, use 'FtpsOnly' instead.
    Limitation: vnet_route_all_enabled requires VNet integration (virtual_network_subnet_id) to take effect; egress policy itself lives in the VNet/firewall, not this module.
  EOT
  type = object({
    minimum_tls_version     = optional(string)
    scm_minimum_tls_version = optional(string)
    ftps_state              = optional(string)
    vnet_route_all_enabled  = optional(bool)
  })
  default = {
    minimum_tls_version     = "1.3"
    scm_minimum_tls_version = "1.3"
    ftps_state              = "Disabled"
    vnet_route_all_enabled  = true
  }
}

variable "private_endpoints" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: partial]
    Map of private endpoints to create (passed through to the AVM module). Each
    entry typically specifies subnet_resource_id and private_dns_zone_resource_ids.
  EOT
  type        = any
  default     = {}
}
