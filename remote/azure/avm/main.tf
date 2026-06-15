locals {
  storage_account_name  = "avmstor${var.instance_id}"
  default_container     = "ccc-avm-test-container-${var.instance_id}"
  key_vault_name        = "avmkv${var.instance_id}"
  log_analytics_name    = "avmlaw${var.instance_id}"
  virtual_network_name  = "avmvnet${var.instance_id}"
  function_storage_name = "avmfnstor${var.instance_id}"
  function_plan_name    = "avmplan${var.instance_id}"
  function_app_name     = "avmfunc${var.instance_id}"
  virtual_machine_name  = "avmvm${var.instance_id}"
}

data "azurerm_client_config" "current" {}

# Resource group for AVM testing.
# Managed as a resource to allow creation, but we import it if it already exists
# because it is excluded from the automated cleanup (nuke).
resource "azurerm_resource_group" "this" {
  name     = "avm-testing"
  location = var.location
}

module "storage_account" {
  source = "./modules/storage-account"

  location  = var.location
  parent_id = azurerm_resource_group.this.id
  name      = local.storage_account_name

  allow_nested_items_to_be_public   = var.allow_nested_items_to_be_public
  blob_properties                   = var.blob_properties
  default_to_oauth_authentication   = var.default_to_oauth_authentication
  https_traffic_only_enabled        = var.https_traffic_only_enabled
  infrastructure_encryption_enabled = var.infrastructure_encryption_enabled
  min_tls_version                   = var.min_tls_version
  network_rules                     = var.network_rules
  public_network_access_enabled     = var.public_network_access_enabled
  shared_access_key_enabled         = var.shared_access_key_enabled
  default_container                 = local.default_container
}

module "key_vault" {
  source = "./modules/key-vault"

  location            = var.location
  resource_group_name = azurerm_resource_group.this.name
  name                = local.key_vault_name
  tenant_id           = data.azurerm_client_config.current.tenant_id

  enabled_for_deployment          = var.enabled_for_deployment
  enabled_for_disk_encryption     = var.enabled_for_disk_encryption
  enabled_for_template_deployment = var.enabled_for_template_deployment
  legacy_access_policies_enabled  = var.legacy_access_policies_enabled
  network_acls                    = var.network_acls
  public_network_access_enabled   = var.public_network_access_enabled
  purge_protection_enabled        = var.purge_protection_enabled
  sku_name                        = var.sku_name
  soft_delete_retention_days      = var.soft_delete_retention_days
}

# ---------------------------------------------------------------------------
# Log Analytics workspace (avm-res-operationalinsights-workspace)
# ---------------------------------------------------------------------------
module "log_analytics_workspace" {
  source = "./modules/log-analytics-workspace"

  location            = var.location
  name                = local.log_analytics_name
  resource_group_name = azurerm_resource_group.this.name

  log_analytics_workspace_retention_in_days          = var.log_analytics_workspace_retention_in_days
  log_analytics_workspace_internet_ingestion_enabled = var.log_analytics_workspace_internet_ingestion_enabled
  log_analytics_workspace_internet_query_enabled     = var.log_analytics_workspace_internet_query_enabled
}

# ---------------------------------------------------------------------------
# Virtual network (support) — provides the delegated functions subnet and the
# VM subnet used by the serverless-function and virtual-machine modules.
# ---------------------------------------------------------------------------
module "virtual_network" {
  source = "./modules/virtual-network"

  location  = var.location
  name      = local.virtual_network_name
  parent_id = azurerm_resource_group.this.id
}

# ---------------------------------------------------------------------------
# Serverless function backing infrastructure (deployment enablers, not CCC
# attestation surfaces): Flex Consumption plan, a deployment-package storage
# account, the user-assigned identity the app uses to read it, and the RBAC
# grants + propagation wait that make identity-based access work at apply time.
# ---------------------------------------------------------------------------
resource "azurerm_service_plan" "functions" {
  name                = local.function_plan_name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location
  os_type             = "Linux"
  sku_name            = "FC1"
}

resource "azurerm_user_assigned_identity" "function" {
  name                = "${local.function_app_name}-uami"
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location
}

# Deployment-package store for the Flex Consumption app. Public access stays
# enabled because the Terraform runner and the Flex backing pull need data-plane
# reachability; access is still gated by Entra (shared keys disabled).
resource "azurerm_storage_account" "function" {
  name                            = local.function_storage_name
  resource_group_name             = azurerm_resource_group.this.name
  location                        = var.location
  account_tier                    = "Standard"
  account_replication_type        = "LRS"
  min_tls_version                 = "TLS1_2"
  https_traffic_only_enabled      = true
  public_network_access_enabled   = true
  shared_access_key_enabled       = false
  allow_nested_items_to_be_public = false
}

resource "azurerm_role_assignment" "deployer_function_blob" {
  scope                = azurerm_storage_account.function.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "azurerm_role_assignment" "function_blob" {
  scope                = azurerm_storage_account.function.id
  role_definition_name = "Storage Blob Data Owner"
  principal_id         = azurerm_user_assigned_identity.function.principal_id
}

resource "azurerm_storage_container" "deploy" {
  name               = "deploymentpackage"
  storage_account_id = azurerm_storage_account.function.id

  depends_on = [azurerm_role_assignment.deployer_function_blob]
}

resource "time_sleep" "wait_for_function_rbac" {
  create_duration = "60s"

  depends_on = [azurerm_role_assignment.function_blob]
}

# ---------------------------------------------------------------------------
# Serverless function (avm-res-web-site) — Flex Consumption, VNet-integrated.
# ---------------------------------------------------------------------------
module "serverless_function" {
  source = "./modules/serverless-function"

  location  = var.location
  name      = local.function_app_name
  parent_id = azurerm_resource_group.this.id

  service_plan_resource_id          = azurerm_service_plan.functions.id
  virtual_network_subnet_id         = module.virtual_network.functions_subnet_resource_id
  storage_container_endpoint        = "https://${azurerm_storage_account.function.name}.blob.core.windows.net/${azurerm_storage_container.deploy.name}"
  storage_user_assigned_identity_id = azurerm_user_assigned_identity.function.id

  client_certificate_enabled    = var.client_certificate_enabled
  client_certificate_mode       = var.client_certificate_mode
  function_app_uses_fc1         = var.function_app_uses_fc1
  https_only                    = var.https_only
  maximum_instance_count        = var.maximum_instance_count
  public_network_access_enabled = var.public_network_access_enabled
  site_config                   = var.site_config
  managed_identities = {
    system_assigned            = var.managed_identities.system_assigned
    user_assigned_resource_ids = [azurerm_user_assigned_identity.function.id]
  }

  depends_on = [time_sleep.wait_for_function_rbac, azurerm_storage_container.deploy]
}

# ---------------------------------------------------------------------------
# Virtual machine (avm-res-compute-virtualmachine) — Trusted Launch, encryption
# at host, no-password (generated SSH key), NIC on the VM subnet.
# ---------------------------------------------------------------------------
module "virtual_machine" {
  source = "./modules/virtual-machine"

  location            = var.location
  name                = local.virtual_machine_name
  resource_group_name = azurerm_resource_group.this.name
  zone                = var.vm_zone
  sku_size            = var.vm_sku_size
  subnet_resource_id  = module.virtual_network.vm_subnet_resource_id

  account_credentials        = var.account_credentials
  boot_diagnostics           = var.boot_diagnostics
  encryption_at_host_enabled = var.encryption_at_host_enabled
  managed_identities         = var.managed_identities
  secure_boot_enabled        = var.secure_boot_enabled
  vtpm_enabled               = var.vtpm_enabled
}
