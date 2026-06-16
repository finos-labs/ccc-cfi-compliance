variable "location" {
  description = "Azure region where the virtual machine should be deployed."
  type        = string
}

variable "name" {
  description = "Virtual machine name."
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group in which to create the virtual machine."
  type        = string
}

variable "zone" {
  description = "Availability zone for the virtual machine. Null deploys regionally (no zone pin)."
  type        = string
  default     = null
  nullable    = true
}

variable "sku_size" {
  description = "VM size. Must be a Gen2 / Trusted Launch-capable size that supports encryption at host."
  type        = string
  default     = "Standard_D2ds_v5"
}

variable "subnet_resource_id" {
  description = "Resource ID of the subnet the VM NIC is attached to."
  type        = string
}

variable "source_image_reference" {
  description = "Marketplace image reference. Must be a Gen2 image for Trusted Launch."
  type = object({
    publisher = string
    offer     = string
    sku       = string
    version   = string
  })
  default = {
    publisher = "Canonical"
    offer     = "ubuntu-24_04-lts"
    sku       = "server"
    version   = "latest"
  }
}

variable "os_disk" {
  description = <<-EOT
    Controls: CCC.Core.CN02, CCC.Core.CN11 [coverage: partial]
    OS managed disk configuration. Set disk_encryption_set_id to encrypt the OS
    disk with a customer-managed key (CMK) backed by Key Vault; omit it to use
    platform-managed keys.
  EOT
  type = object({
    caching                = string
    storage_account_type   = string
    disk_encryption_set_id = optional(string)
  })
  default = {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }
}

variable "account_credentials" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: partial]
    Admin credential configuration. Disables password authentication in favour of an SSH key.
    Limitation: Applies to Linux SSH; Windows admin access is governed differently. Key distribution is an operational concern.
  EOT
  type = object({
    password_authentication_disabled = optional(bool, true)
    admin_credentials = optional(object({
      username                           = optional(string, "azureuser")
      generate_admin_password_or_ssh_key = optional(bool, true)
    }), {})
  })
  default = {
    password_authentication_disabled = true
    admin_credentials = {
      username                           = "azureuser"
      generate_admin_password_or_ssh_key = true
    }
  }
}

variable "encryption_at_host_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN02 [coverage: partial]
    Encrypt the VM's temp disk and OS/data disk caches on the host.
    Limitation: Requires the EncryptionAtHost feature registered on the subscription and a supported VM size.
    Limitation: Uses platform-managed keys unless combined with a disk encryption set.
  EOT
  type        = bool
  default     = true
}

variable "managed_identities" {
  description = <<-EOT
    Controls: CCC.Core.CN03, CCC.Core.CN05 [coverage: supporting]
    Managed identities assigned to the virtual machine.
    Limitation: Managed identity is the auth principal, not MFA; CN03 (MFA for human access) is enforced via Entra Conditional Access / JIT, not this module.
  EOT
  type = object({
    system_assigned            = optional(bool, false)
    user_assigned_resource_ids = optional(set(string), [])
  })
  default = {
    system_assigned = true
  }
}

variable "secure_boot_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: supporting]
    Enable Secure Boot (Trusted Launch).
    Limitation: Requires a Gen2 image and a Trusted Launch-capable VM size; forces VM recreation if changed.
  EOT
  type        = bool
  default     = true
}

variable "vtpm_enabled" {
  description = <<-EOT
    Controls: CCC.Core.CN05 [coverage: supporting]
    Enable the virtual TPM (Trusted Launch).
    Limitation: Requires a Gen2 image and Trusted Launch-capable size; forces VM recreation if changed.
    Limitation: vTPM enables measured boot/attestation but attestation itself is configured separately.
  EOT
  type        = bool
  default     = true
}

variable "boot_diagnostics" {
  description = <<-EOT
    Controls: CCC.Core.CN04 [coverage: supporting]
    Enable boot diagnostics (managed storage).
    Limitation: Boot diagnostics is platform/serial telemetry only; full access/change auditing also needs diagnostic_settings + guest-level logging (Azure Monitor Agent).
  EOT
  type        = bool
  default     = true
}
