variable "location" {
  description = "Azure region where the Log Analytics workspace should be deployed."
  type        = string
}

variable "name" {
  description = "Log Analytics workspace name (4-63 chars; alphanumerics and hyphens, must start and end with alphanumeric)."
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group in which to create the workspace."
  type        = string
}

variable "log_analytics_workspace_retention_in_days" {
  description = <<-EOT
    Controls: CCC.Core.CN04, CCC.Core.CN09 [coverage: partial]
    Number of days log data is retained in the workspace.
    Limitation: Retention period is regulatory/policy-specific - adjust to the workload's obligation. v0.5.1 supports 30-730 days (7 only on the Free tier).
    Limitation: Retention alone is not immutability; CN09 tamper-resistance still needs workspace RBAC and (for the strongest posture) a dedicated cluster / immutable export.
  EOT
  type        = number
  default     = 365
}

variable "log_analytics_workspace_internet_ingestion_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: partial]
    Whether ingestion of logs over the public internet is allowed.
    Limitation: Both variables are strings on v0.5.1 ("true" | "false" | "SecuredByPerimeter"), not booleans.
    Limitation: Disabling internet access only delivers reachability when private connectivity (AMPLS) is wired up.
  EOT
  type        = string
  default     = "false"
}

variable "log_analytics_workspace_internet_query_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: partial]
    Whether querying of logs over the public internet is allowed.
    Limitation: Both variables are strings on v0.5.1 ("true" | "false" | "SecuredByPerimeter"), not booleans.
    Limitation: Disabling internet access only delivers reachability when private connectivity (AMPLS) is wired up.
  EOT
  type        = string
  default     = "false"
}
