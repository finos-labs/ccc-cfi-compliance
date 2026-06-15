# Support infrastructure: a virtual network providing the delegated "functions"
# subnet (Flex Consumption VNet integration) and a "vm" subnet for the virtual
# machine NIC. This is an enabler for the CCC capability services, not itself a
# CCC service under test, so it is not driven by a control-informed tfvars file.

module "avm_res_network_virtualnetwork" {
  source  = "Azure/avm-res-network-virtualnetwork/azurerm"
  version = "0.17.1"

  location      = var.location
  name          = var.name
  parent_id     = var.parent_id
  address_space = var.address_space

  # CCC.Core.CN01 (network-layer encryption). DropUnencrypted requires full-fleet
  # SKU/accelerated-networking support and can break traffic to unsupported
  # workloads, so the deployable baseline relaxes enforcement to AllowUnencrypted.
  encryption = {
    enabled     = true
    enforcement = "AllowUnencrypted"
  }

  subnets = {
    functions = {
      name             = "functions-integration"
      address_prefixes = [var.functions_subnet_prefix]
      delegations = [{
        name = "Microsoft.App.environments"
        service_delegation = {
          name = "Microsoft.App/environments"
        }
      }]
    }
    vm = {
      name             = "vm"
      address_prefixes = [var.vm_subnet_prefix]
    }
  }
}
