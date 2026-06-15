variable "instance_id" {
  description = "Unique ID for this run; used in globally unique resource names."
  type        = string
  default     = "20260611"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "westus2"
}

variable "subscription_id" {
  description = "Azure subscription ID (optional if using az cli or ARM_SUBSCRIPTION_ID)"
  type        = string
  default     = null
}

variable "allow_nested_items_to_be_public" {
  type    = bool
  default = false
}

variable "shared_access_key_enabled" {
  type    = bool
  default = false
}

variable "default_to_oauth_authentication" {
  type    = bool
  default = true
}

variable "blob_properties" {
  type = object({
    automatic_snapshot_policy_enabled = optional(bool)
    change_feed = optional(object({
      enabled           = optional(bool)
      retention_in_days = optional(number)
    }))
    container_delete_retention_policy = optional(object({
      allow_permanent_delete = optional(bool)
      days                   = optional(number)
      enabled                = optional(bool)
    }))
    cors_rules = optional(list(object({
      allowed_headers    = list(string)
      allowed_methods    = list(string)
      allowed_origins    = list(string)
      exposed_headers    = list(string)
      max_age_in_seconds = number
    })))
    default_service_version = optional(string)
    delete_retention_policy = optional(object({
      allow_permanent_delete = optional(bool)
      days                   = optional(number)
      enabled                = optional(bool)
    }))
    last_access_time_tracking_policy = optional(object({
      blob_type                    = optional(list(string))
      enable                       = bool
      name                         = optional(string)
      tracking_granularity_in_days = optional(number)
    }))
    restore_policy = optional(object({
      days    = optional(number)
      enabled = bool
    }))
    versioning_enabled = optional(bool)
  })
  default = {
    versioning_enabled = true
    change_feed = {
      enabled           = true
      retention_in_days = 1
    }
    restore_policy = {
      enabled = false
      days    = 1
    }
    last_access_time_tracking_policy = {
      enable = false
      name   = "AccessTimeTracking"
    }
    delete_retention_policy = {
      enabled = true
      days    = 30
    }
    container_delete_retention_policy = {
      enabled = true
      days    = 30
    }
  }
}

variable "infrastructure_encryption_enabled" {
  type    = bool
  default = true
}

variable "network_rules" {
  type = object({
    bypass                     = optional(set(string), ["AzureServices"])
    default_action             = optional(string, "Deny")
    ip_rules                   = optional(set(string), [])
    virtual_network_subnet_ids = optional(set(string), [])
    private_link_access = optional(list(object({
      endpoint_resource_id = string
      endpoint_tenant_id   = optional(string)
    })))
    timeouts = optional(object({
      create = optional(string)
      delete = optional(string)
      read   = optional(string)
      update = optional(string)
    }))
  })
  default = {
    default_action = "Deny"
    bypass         = ["AzureServices"]
  }
}

variable "public_network_access_enabled" {
  type    = bool
  default = false
}

variable "https_traffic_only_enabled" {
  type    = bool
  default = true
}

variable "min_tls_version" {
  type    = string
  default = "TLS1_2"
}

# ---------------------------------------------------------------------------
# Key Vault (avm-res-keyvault-vault) — see remote/azure/avm/key-vault.tfvars
# for the control-informed values and CCC annotations.
# Note: public_network_access_enabled (declared above) is shared with the
# storage account module; both adopt the private posture (false).
# ---------------------------------------------------------------------------

variable "enabled_for_deployment" {
  type    = bool
  default = false
}

variable "enabled_for_disk_encryption" {
  type    = bool
  default = false
}

variable "enabled_for_template_deployment" {
  type    = bool
  default = false
}

variable "legacy_access_policies_enabled" {
  type    = bool
  default = false
}

variable "purge_protection_enabled" {
  type    = bool
  default = true
}

variable "soft_delete_retention_days" {
  type    = number
  default = 90
}

variable "sku_name" {
  type    = string
  default = "premium"
}

variable "network_acls" {
  type = object({
    bypass                     = optional(string, "None")
    default_action             = optional(string, "Deny")
    ip_rules                   = optional(list(string), [])
    virtual_network_subnet_ids = optional(list(string), [])
  })
  default = {
    default_action = "Deny"
    bypass         = "AzureServices"
  }
}

# ---------------------------------------------------------------------------
# Shared across the serverless-function and virtual-machine modules. Both adopt
# a system-assigned managed identity (CCC.Core.CN03/CN05). The function also
# receives a user-assigned identity for backing-storage access, injected in the
# root module block rather than via this variable.
# ---------------------------------------------------------------------------
variable "managed_identities" {
  type = object({
    system_assigned            = optional(bool, false)
    user_assigned_resource_ids = optional(set(string), [])
  })
  default = {
    system_assigned = true
  }
}

# ---------------------------------------------------------------------------
# Log Analytics workspace (avm-res-operationalinsights-workspace) — see
# remote/azure/avm/log-analytics-workspace.tfvars for CCC annotations.
# ---------------------------------------------------------------------------
variable "log_analytics_workspace_retention_in_days" {
  type    = number
  default = 365
}

variable "log_analytics_workspace_internet_ingestion_enabled" {
  type    = string
  default = "false"
}

variable "log_analytics_workspace_internet_query_enabled" {
  type    = string
  default = "false"
}

# ---------------------------------------------------------------------------
# Serverless function (avm-res-web-site) — see
# remote/azure/avm/serverless-function.tfvars for CCC annotations.
# Note: public_network_access_enabled (declared above) is shared with storage
# and key vault; all three adopt the private posture (false).
# ---------------------------------------------------------------------------
variable "client_certificate_enabled" {
  type    = bool
  default = true
}

variable "client_certificate_mode" {
  type    = string
  default = "Required"
}

variable "function_app_uses_fc1" {
  type    = bool
  default = true
}

variable "https_only" {
  type    = bool
  default = true
}

variable "maximum_instance_count" {
  type    = number
  default = 40
}

variable "site_config" {
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

# ---------------------------------------------------------------------------
# Virtual machine (avm-res-compute-virtualmachine) — see
# remote/azure/avm/virtual-machine.tfvars for CCC annotations. vm_zone and
# vm_sku_size are deployment knobs (Trusted Launch + encryption-at-host capable).
# ---------------------------------------------------------------------------
variable "account_credentials" {
  type = object({
    password_authentication_disabled = optional(bool, true)
    admin_credentials = optional(object({
      username                           = optional(string, "azureuser")
      generate_admin_password_or_ssh_key = optional(bool, true)
    }), {})
  })
  default = {
    password_authentication_disabled = true
  }
}

variable "encryption_at_host_enabled" {
  type    = bool
  default = true
}

variable "secure_boot_enabled" {
  type    = bool
  default = true
}

variable "vtpm_enabled" {
  type    = bool
  default = true
}

variable "boot_diagnostics" {
  type    = bool
  default = true
}

variable "vm_zone" {
  type    = string
  default = "1"
}

variable "vm_sku_size" {
  type    = string
  default = "Standard_D2ds_v5"
}
