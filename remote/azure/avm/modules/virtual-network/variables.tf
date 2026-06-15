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

variable "vm_subnet_prefix" {
  description = "Address prefix for the virtual machine subnet."
  type        = string
  default     = "10.40.3.0/24"
}
