module "avm_res_web_serverfarm" {
  source  = "Azure/avm-res-web-serverfarm/azurerm"
  version = "2.0.3"

  location  = var.location
  name      = var.name
  parent_id = var.parent_id

  os_type  = var.os_type
  sku_name = var.sku_name

  # Flex Consumption (FC1) is serverless: Azure manages worker count and zone
  # balancing, so zone balancing is disabled to match plain service-plan
  # behaviour. The module already special-cases FC1 (kind = functionapp,
  # capacity = 0, no maximumElasticWorkerCount).
  zone_balancing_enabled = var.zone_balancing_enabled
}
