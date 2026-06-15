# Support infrastructure: a private DNS zone linked to the virtual network so
# that private-endpoint FQDNs resolve to their private IPs. This is the
# dns.resolution.vnet-link enabler for the private-endpoint posture of the
# storage account, key vault and function app; it is not itself a CCC service
# under test, so it is not driven by a control-informed tfvars file.

module "avm_res_network_privatednszone" {
  source  = "Azure/avm-res-network-privatednszone/azurerm"
  version = "0.5.0"

  domain_name = var.domain_name
  parent_id   = var.parent_id

  virtual_network_links = {
    link = {
      vnetlinkname       = var.link_name
      virtual_network_id = var.virtual_network_id
    }
  }
}
