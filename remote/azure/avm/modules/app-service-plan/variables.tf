variable "location" {
  description = "Azure region where the app service plan should be deployed."
  type        = string
}

variable "name" {
  description = "App service plan name."
  type        = string
}

variable "parent_id" {
  description = "Resource ID of the resource group in which to create the app service plan."
  type        = string
}

variable "os_type" {
  description = "Operating system type of the plan. Possible values are Windows, Linux or WindowsContainer."
  type        = string
  default     = "Linux"
}

variable "sku_name" {
  description = "SKU name of the plan. Defaults to FC1 (Flex Consumption) for serverless function hosting."
  type        = string
  default     = "FC1"
}

variable "zone_balancing_enabled" {
  description = "Whether zone balancing is enabled. Left disabled for FC1 (Flex Consumption manages this)."
  type        = bool
  default     = false
}
