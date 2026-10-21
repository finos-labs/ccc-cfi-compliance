# Support infrastructure: a virtual network providing the delegated "functions"
# subnet (Flex Consumption VNet integration), a "vm" subnet for the virtual
# machine NIC, and AKS node + API-server subnets for the managed cluster. This
# is an enabler for the CCC capability services, not itself a CCC service under
# test, so it is not driven by a control-informed tfvars file.

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
    enforcement = var.encryption_enforcement
  }

  subnets = {
    # Dedicated subnet for the storage / key-vault / function private endpoints.
    pe = {
      name                              = "private-endpoints"
      address_prefixes                  = [var.pe_subnet_prefix]
      private_endpoint_network_policies = "Enabled"
    }
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
    # Node subnet for the private AKS cluster (no public node IPs).
    aks = {
      name             = "aks-nodes"
      address_prefixes = [var.aks_subnet_prefix]
    }
    # Delegated API-server subnet for AKS VNet integration + private KMS.
    aks_api = {
      name             = "aks-api"
      address_prefixes = [var.aks_api_subnet_prefix]
      delegations = [{
        name = "Microsoft.ContainerService.managedClusters"
        service_delegation = {
          name = "Microsoft.ContainerService/managedClusters"
          actions = [
            "Microsoft.Network/virtualNetworks/subnets/join/action",
          ]
        }
      }]
    }
  }
}
