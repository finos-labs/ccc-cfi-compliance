variable "location" {
  description = "Azure region where the managed cluster should be deployed."
  type        = string
}

variable "name" {
  description = "Managed cluster name."
  type        = string
}

variable "parent_id" {
  description = "Resource ID of the resource group in which to create the managed cluster."
  type        = string
}

variable "kubernetes_version" {
  description = <<-EOT
    Controls: CCC.Core.CN06, CCC.K8S.CN09, CCC.K8S.CN15 [coverage: partial]
    Kubernetes version supported in the selected region. Region approval, version
    support/patch evidence and tag/label enforcement are external.
  EOT
  type        = string
}

variable "control_plane_identity_id" {
  description = <<-EOT
    Controls: CCC.K8S.CN17 [coverage: partial]
    User-assigned identity ARM ID for the control plane. Must be distinct from the kubelet identity.
  EOT
  type        = string
}

variable "kubelet_identity_id" {
  description = <<-EOT
    Controls: CCC.K8S.CN17 [coverage: partial]
    User-assigned identity ARM ID for the kubelet (not an object/client ID).
  EOT
  type        = string
}

variable "cluster_admin_group_object_ids" {
  description = <<-EOT
    Controls: CCC.K8S.CN02, CCC.K8S.CN16 [coverage: partial]
    Entra GROUP object IDs for Kubernetes cluster-admin access. Application
    Roles/Bindings, MFA and revocation are external; local accounts are not
    disabled by this module.
  EOT
  type        = list(string)
}

variable "private_dns_zone_id" {
  description = <<-EOT
    Controls: CCC.K8S.CN01, CCC.Core.CN05 [coverage: partial]
    API private DNS zone ARM ID (e.g. a unique subzone of private.<location>.azmk8s.io).
  EOT
  type        = string
}

variable "node_subnet_id" {
  description = "Existing AKS node subnet ARM ID."
  type        = string
}

variable "api_server_subnet_id" {
  description = "Existing delegated API-server subnet ARM ID for VNet integration."
  type        = string
}

variable "disk_encryption_set_id" {
  description = <<-EOT
    Controls: CCC.Core.CN02, CCC.K8S.CN07 [coverage: partial]
    Disk encryption set ARM ID for node OS/data disks (CMK).
  EOT
  type        = string
}

variable "key_vault_key_id" {
  description = <<-EOT
    Controls: CCC.Core.CN02, CCC.K8S.CN07 [coverage: partial]
    Versioned HTTPS Key Vault key URL for Kubernetes secrets KMS encryption.
  EOT
  type        = string
}

variable "key_vault_resource_id" {
  description = "Key Vault ARM ID hosting the KMS key (private network access)."
  type        = string
}

variable "node_vm_size" {
  description = "VM size for the default system node pool."
  type        = string
  default     = "Standard_D4s_v5"
}

variable "node_count" {
  description = "Node count for the default system pool (demonstration capacity only)."
  type        = number
  default     = 2
}

variable "node_availability_zones" {
  description = "Availability zones for the default system pool."
  type        = list(string)
  default     = ["1", "2", "3"]
}

variable "outbound_type" {
  description = <<-EOT
    Controls: CCC.K8S.CN06 [coverage: supporting]
    AKS outbound type. Hardened posture is userDefinedRouting with an approved
    egress path; the deployable baseline defaults to loadBalancer.
  EOT
  type        = string
  default     = "loadBalancer"
}

variable "pod_cidr" {
  description = "Overlay pod CIDR; must not overlap the VNet or service CIDR."
  type        = string
  default     = "172.18.0.0/16"
}

variable "service_cidr" {
  description = "Service CIDR; must not overlap the VNet or pod CIDR."
  type        = string
  default     = "172.19.0.0/16"
}

variable "dns_service_ip" {
  description = "DNS service IP within service_cidr."
  type        = string
  default     = "172.19.0.10"
}

variable "tags" {
  description = "Tags applied to the managed cluster."
  type        = map(string)
  default     = {}
}
