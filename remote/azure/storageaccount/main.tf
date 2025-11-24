module "storage_account" {
  source = "git::https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount.git?ref=main"

  # Required parameters
  name                     = "stcfi${random_string.suffix.result}"
  resource_group_name      = azurerm_resource_group.this.name
  location                 = azurerm_resource_group.this.location
  
  # Cost optimization settings
  account_tier             = "Standard"
  account_replication_type = "LRS"  # Locally redundant - cheapest option
  access_tier              = "Hot"  # Hot is default, but explicit
  
  # Disable expensive features
  enable_https_traffic_only           = true   # Security best practice, no extra cost
  min_tls_version                     = "TLS1_2"  # Security, no extra cost
  allow_nested_items_to_be_public     = false  # Security, no extra cost
  shared_access_key_enabled           = true   # Needed for basic access
  
  # Disable costly data protection features
  blob_properties = {
    versioning_enabled            = false  # 💰 EXPENSIVE: Stores every version
    change_feed_enabled           = false  # 💰 EXPENSIVE: Logs all changes
    last_access_time_enabled      = false  # 💰 Can incur costs for tracking
    
    delete_retention_policy = {
      days = 1  # 💰 Minimum retention (7 days default is expensive)
    }
    
    container_delete_retention_policy = {
      days = 1  # 💰 Minimum retention (7 days default is expensive)
    }
  }
  
  # Disable file/queue/table soft delete if not needed
  share_properties = {
    retention_policy = {
      days = 1  # 💰 Minimum retention for file shares
    }
  }
  
  # Disable advanced threat protection
  # Note: This might be a separate resource in the module
  # advanced_threat_protection_enabled = false  # 💰 EXPENSIVE security feature
  
  # Don't enable geo-replication or cross-region features
  # cross_tenant_replication_enabled = false  # Already using LRS
  
  # Disable large file shares if not needed (premium feature)
  large_file_share_enabled = false
}

resource "azurerm_resource_group" "this" {
  name     = "rg-cfi-storage"
  location = "East US"
}

resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}
