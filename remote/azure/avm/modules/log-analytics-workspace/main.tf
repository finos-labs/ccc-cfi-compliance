module "avm_res_operationalinsights_workspace" {
  source  = "Azure/avm-res-operationalinsights-workspace/azurerm"
  version = "0.5.1"

  location            = var.location
  name                = var.name
  resource_group_name = var.resource_group_name

  log_analytics_workspace_retention_in_days          = var.log_analytics_workspace_retention_in_days
  log_analytics_workspace_internet_ingestion_enabled = var.log_analytics_workspace_internet_ingestion_enabled
  log_analytics_workspace_internet_query_enabled     = var.log_analytics_workspace_internet_query_enabled
}
