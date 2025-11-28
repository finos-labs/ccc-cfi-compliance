variable "storage_account_name" {
  description = "Azure storage account name (set via TF_VAR_storage_account_name or AZURE_STORAGE_ACCOUNT)"
  type        = string
  default     = "storagecfitesting2025"
}

# Storage account for compliance testing
module "storage_account" {
  source = "git::https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount.git?ref=main"
  location = "eastus"
  resource_group_name = "cfi_test"
  name = var.storage_account_name

  account_tier             = "Standard"
  account_replication_type = "LRS"  # Locally redundant - cheapest option
  access_tier              = "Hot"  # Hot is default, but explicit

}
