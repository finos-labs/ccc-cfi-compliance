module "storage_account" {
  source = "git::https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount.git?ref=main"
  location = "eastus"
  resource_group_name = "cfi_test"
  name = "storagecfitesting2025"

  account_tier             = "Standard"
  account_replication_type = "LRS"  # Locally redundant - cheapest option
  access_tier              = "Hot"  # Hot is default, but explicit

}
