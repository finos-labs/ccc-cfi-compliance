# This file is CONTROL-INFORMED CANDIDATE configuration, not a compliance
# attestation. It covers only technical settings expressible as AVM module
# inputs. It does NOT establish FedRAMP / CCC / NIST compliance. Mapped
# controls may be only partially satisfied; unmapped and process/evidence
# controls are out of scope here. Human review and deployment-context
# validation are required before use.
#
# AVM module:   Azure/avm-res-containerservice-managedcluster/azurerm
# Module ver:   0.8.1
# Pattern from: https://github.com/Azure/terraform-azurerm-avm-res-containerservice-managedcluster/pull/333
#               (examples/finos-ccc/main.tf)

# Controls: CCC.Core.CN06, CCC.K8S.CN09, CCC.K8S.CN15  [coverage: partial]
# Limitation: Region approval, version support/patch evidence and tag/label
# enforcement are external. Pin a supported patch in the target region.
# westus2 default GA line as of 2026-10; 1.31 is LTS-only now.
aks_kubernetes_version = "1.36"

# Controls: CCC.K8S.CN02, CCC.K8S.CN16  [coverage: partial]
# Limitation: Supply Entra GROUP object IDs for cluster-admin. Application
# Roles/Bindings, MFA and revocation are external; local accounts are not
# disabled by this deployment.
# aks_cluster_admin_group_object_ids = ["00000000-0000-0000-0000-000000000000"]

# Controls: CCC.K8S.CN12, CCC.K8S.CN18  [coverage: partial]
# Limitation: No public node IPs; kubelet/IMDS/SSH restrictions and boot
# attestation are not established by these settings. Sizing is demonstration
# capacity only.
# v4+ SKUs cannot use API-server VNet integration on an encrypted VNet
# (VNetEncryptionNotCompatibleWithAPIServerVNetIntegration). D4s_v3 is
# available in westus2 zone 2 for this subscription.
aks_node_vm_size            = "Standard_D4s_v3"
# westus2 has 6 regional vCPUs left; two D4s_v3 nodes need 8.
aks_node_count              = 1
aks_node_availability_zones = ["2"]

# Controls: CCC.K8S.CN06  [coverage: supporting]
# Limitation: Hardened posture is userDefinedRouting with an approved
# UDR/firewall egress path. The deployable baseline uses loadBalancer.
aks_outbound_type = "loadBalancer"

# Overlay / service CIDRs — verify no overlap with the VNet (10.40.0.0/16).
aks_pod_cidr       = "172.18.0.0/16"
aks_service_cidr   = "172.19.0.0/16"
aks_dns_service_ip = "172.19.0.10"
