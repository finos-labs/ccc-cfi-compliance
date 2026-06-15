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
