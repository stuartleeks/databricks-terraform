package databricks

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAzureAdlsGen2Mount_capture_error(t *testing.T) {
	randPrefix := acctest.RandString(8)
	randSpPassword := acctest.RandString(24)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { azurePreCheck(t) },
		Providers: testAccProviders,
		ProviderFactories: map[string]terraform.ResourceProviderFactory,
		// CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccAzureAdlsGen2Mount_capture_error(randPrefix, randSpPassword),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("Something went wrong!"),
			},
		},
	})
}

func testAccAzureAdlsGen2Mount_capture_error(randPrefix string, randSpPassword string) string {
	clientID := os.Getenv("ARM_CLIENT_ID")
	clientSecret := os.Getenv("ARM_CLIENT_SECRET")
	tenantID := os.Getenv("ARM_TENANT_ID")
	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	definition := fmt.Sprintf(`
	provider "azurerm" {
	  version = "~> 2.3"
	  features {}
	}

	provider "azuread" {
	  version = "~> 0.8"
	}

	resource "azurerm_resource_group" "example" {
	  name     = "%[1]srg"
	  location = "eastus" # note must be lower without spaces not verbose style
	}
	
	resource "azurerm_databricks_workspace" "example" {
	  name                = "databricks-test"
	  resource_group_name = azurerm_resource_group.example.name
	  location            = azurerm_resource_group.example.location
	  sku                 = "standard"
	}
	
	# Configure Datalake and SP for access
	resource "azurerm_storage_account" "account" {
	  name                     = "%[1]sdatalake"
	  resource_group_name      = azurerm_resource_group.example.name
	  location                 = azurerm_resource_group.example.location
	  account_tier             = "Standard"
	  account_replication_type = "GRS"
	  account_kind             = "StorageV2"
	  is_hns_enabled           = "true"
	}
	
	resource "azuread_application" "datalake" {
	  name                       = "%[1]sdatalake"
	  identifier_uris            = ["http://%[1]sdatalake"]
	  available_to_other_tenants = false
	  oauth2_allow_implicit_flow = true
	
	}
	
	resource "azuread_service_principal" "datalake" {
	  application_id               = azuread_application.datalake.application_id
	  app_role_assignment_required = false
	}

	resource "azuread_service_principal_password" "datalake" {
	  service_principal_id = azuread_service_principal.datalake.id
	  value                = "%[2]s"
	  # Review best way forward with this setting
	  end_date = "2050-01-01T01:02:03Z"
	}
	
	resource "azurerm_role_assignment" "datalake" {
	  scope = azurerm_storage_account.account.id
	  #https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#storage-blob-data-contributor
	  role_definition_name = "Storage Blob Data Contributor"
	  principal_id         = azuread_service_principal.datalake.id
	}
	
	provider "databricks" {
	  azure_auth = {
		managed_resource_group = azurerm_databricks_workspace.example.managed_resource_group_name
		azure_region           = azurerm_databricks_workspace.example.location
		workspace_name         = azurerm_databricks_workspace.example.name
		resource_group         = azurerm_databricks_workspace.example.resource_group_name
	
		client_id       = "%[3]s"
		client_secret   = "%[4]s"
		tenant_id       = "%[5]s"
		subscription_id = "%[6]s"
	  }

	  initial_manage_principal = "users"
	}
	
	resource "databricks_secret" "client_secret" {
	  key          = "datalake_sp_secret"
	  string_value = "%[2]s"
	  scope        = databricks_secret_scope.terraform.name
	}
	
	resource "databricks_cluster" "cluster" {
		num_workers = 1
		spark_version = "6.4.x-scala2.11"
		node_type_id = "Standard_D3_v2"
	}
	
	data "azuread_client_config" "current" {
	}
	
	resource "databricks_azure_adls_gen2_mount" "mount" {
	  cluster_id           = databricks_cluster.cluster.id
	  container_name       = "dev" #todo: replace with env...
	  storage_account_name = azurerm_storage_account.account.name
	  directory            = "/dir"
	  mount_name           = "localdir"
	  tenant_id            = data.azuread_client_config.current.tenant_id
	  client_id            = azuread_application.datalake.application_id
	  client_secret_scope  = databricks_secret_scope.terraform.name
	  client_secret_key    = databricks_secret.client_secret.key
	}

`, randPrefix, randSpPassword, clientID, clientSecret, tenantID, subscriptionID)
	return definition
}
