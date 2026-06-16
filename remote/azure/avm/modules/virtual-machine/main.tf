module "avm_res_compute_virtualmachine" {
  source  = "Azure/avm-res-compute-virtualmachine/azurerm"
  version = "0.21.0"

  location            = var.location
  name                = var.name
  resource_group_name = var.resource_group_name
  zone                = var.zone
  os_type             = "Linux"
  sku_size            = var.sku_size

  source_image_reference = var.source_image_reference
  os_disk                = var.os_disk

  account_credentials        = var.account_credentials
  boot_diagnostics           = var.boot_diagnostics
  encryption_at_host_enabled = var.encryption_at_host_enabled
  managed_identities         = var.managed_identities
  secure_boot_enabled        = var.secure_boot_enabled
  vtpm_enabled               = var.vtpm_enabled

  network_interfaces = {
    primary = {
      name = "${var.name}-nic"
      ip_configurations = {
        ipconfig1 = {
          name                          = "ipconfig1"
          private_ip_subnet_resource_id = var.subnet_resource_id
        }
      }
    }
  }
}
