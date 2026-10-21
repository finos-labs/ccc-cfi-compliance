variable "domain_name" {
  description = "Private DNS zone name (e.g. privatelink.blob.core.windows.net)."
  type        = string
}

variable "parent_id" {
  description = "Resource ID of the resource group in which to create the private DNS zone."
  type        = string
}

variable "virtual_network_id" {
  description = "Resource ID of the virtual network to link to the private DNS zone."
  type        = string
}

variable "link_name" {
  description = "Name of the virtual network link."
  type        = string
  default     = "to-vnet"
}
