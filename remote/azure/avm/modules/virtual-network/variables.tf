variable "location" {
  description = "Azure region where the virtual network should be deployed."
  type        = string
}

variable "name" {
  description = "Virtual network name."
  type        = string
}

variable "parent_id" {
  description = "Resource ID of the resource group in which to create the virtual network."
  type        = string
}

variable "address_space" {
  description = "Address space for the virtual network."
  type        = list(string)
  default     = ["10.40.0.0/16"]
}

variable "functions_subnet_prefix" {
  description = "Address prefix for the delegated functions integration subnet."
  type        = string
  default     = "10.40.2.0/24"
}

variable "pe_subnet_prefix" {
  description = "Address prefix for the private-endpoints subnet."
  type        = string
  default     = "10.40.1.0/24"
}

variable "encryption_enforcement" {
  description = <<-EOT
    Controls: CCC.Core.CN01 [coverage: supporting]
    VNet encryption enforcement mode (DropUnencrypted | AllowUnencrypted).
    Limitation: DropUnencrypted requires all VMs in the VNet to run supported
    SKUs/accelerated networking and can break traffic to unsupported workloads,
    so the deployable baseline defaults to AllowUnencrypted. Set to
    DropUnencrypted for a hardened posture once fleet support is validated.
  EOT
  type        = string
  default     = "AllowUnencrypted"
}

variable "vm_subnet_prefix" {
  description = "Address prefix for the virtual machine subnet."
  type        = string
  default     = "10.40.3.0/24"
}
