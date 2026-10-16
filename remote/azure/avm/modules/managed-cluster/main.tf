# Thin wrapper around Azure/avm-res-containerservice-managedcluster/azurerm
# with FINOS CCC-informed defaults from the upstream finos-ccc example
# (PR #333 on terraform-azurerm-avm-res-containerservice-managedcluster).
# This is control-informed candidate configuration, not a compliance attestation.

module "avm_res_containerservice_managedcluster" {
  source  = "Azure/avm-res-containerservice-managedcluster/azurerm"
  # 0.8.1 is the newest release that still uses avm-utl-interfaces 0.6.0.
  # 0.8.2+ pulls interfaces 0.7.0, whose customer_managed_key_v2 validation
  # breaks terraform validate when that input is null (TF 1.11).
  version = "0.8.1"

  name               = var.name
  location           = var.location
  parent_id          = var.parent_id
  kubernetes_version = var.kubernetes_version
  tags               = var.tags
  enable_telemetry   = false

  sku = {
    name = "Base"
    tier = "Free"
  }

  # Explicit disabled ingress profile: default null trips a coalesce() validation
  # bug in 0.8.0/0.8.1 when TF evaluates both sides of `== null || ...`.
  ingress_profile = {
    gateway_api = {
      installation = "Disabled"
    }
    web_app_routing = {
      gateway_api_implementations = {
        app_routing_istio = {
          mode = "Disabled"
        }
      }
    }
  }

  # Controls: CCC.K8S.CN17 [coverage: partial]
  # Separate control-plane and kubelet identities; grants are wired at the root.
  managed_identities = {
    system_assigned            = false
    user_assigned_resource_ids = [var.control_plane_identity_id]
  }

  identity_profile = {
    kubeletidentity = {
      resource_id = var.kubelet_identity_id
    }
  }

  # Controls: CCC.K8S.CN02, CCC.K8S.CN16 [coverage: partial]
  enable_rbac = true
  aad_profile = {
    managed                = true
    enable_azure_rbac      = false
    admin_group_object_ids = var.cluster_admin_group_object_ids
  }

  # Controls: CCC.K8S.CN01, CCC.Core.CN05 [coverage: partial]
  api_server_access_profile = {
    enable_private_cluster             = true
    enable_private_cluster_public_fqdn = false
    enable_vnet_integration            = true
    private_dns_zone                   = var.private_dns_zone_id
    subnet_id                          = var.api_server_subnet_id
  }

  # Controls: CCC.K8S.CN03 [coverage: supporting]
  oidc_issuer_profile = {
    enabled = true
  }

  # Controls: CCC.Core.CN02, CCC.K8S.CN07 [coverage: partial]
  disk_encryption_set_id = var.disk_encryption_set_id
  security_profile = {
    workload_identity = {
      enabled = true
    }
    azure_key_vault_kms = {
      enabled                  = true
      key_id                   = var.key_vault_key_id
      key_vault_network_access = "Private"
      key_vault_resource_id    = var.key_vault_resource_id
    }
  }

  # Controls: CCC.K8S.CN12, CCC.K8S.CN18 [coverage: partial]
  default_agent_pool = {
    name                  = "systempool"
    vm_size               = var.node_vm_size
    count_of              = var.node_count
    max_pods              = 30
    vnet_subnet_id        = var.node_subnet_id
    os_sku                = "Ubuntu"
    os_disk_type          = "Managed"
    enable_node_public_ip = false
    availability_zones    = var.node_availability_zones
  }

  # Controls: CCC.K8S.CN06 [coverage: supporting]
  # Hardened posture uses userDefinedRouting with an approved egress path; the
  # deployable baseline defaults to loadBalancer (see var.outbound_type).
  network_profile = {
    load_balancer_sku   = "standard"
    outbound_type       = var.outbound_type
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    network_dataplane   = "cilium"
    network_policy      = "cilium"
    pod_cidr            = var.pod_cidr
    service_cidr        = var.service_cidr
    dns_service_ip      = var.dns_service_ip
  }

  # Controls: CCC.K8S.CN08 [coverage: partial]
  addon_profile_azure_policy = {
    enabled = false
  }
}
