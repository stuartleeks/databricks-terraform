provider "azurerm" {
  version = "~> 2.3"
  features {}
}

provider "azuread" {
  version = "~> 0.8"

}

provider "random" {
  version = "~> 2.2"
}

resource "random_string" "naming" {
  special = false
  upper   = false
  length  = 6
}

resource "azurerm_resource_group" "example" {
  name     = "inttest${random_string.naming.result}"
  location = "eastus"
}

resource "azurerm_databricks_workspace" "example" {
  name                = "workspace${random_string.naming.result}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku                 = "standard"
}

resource "azurerm_storage_account" "account" {
  name                     = "${random_string.naming.result}datalake"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
  account_kind             = "StorageV2"
  is_hns_enabled           = "true"
}

output "workspace" {
  value = azurerm_databricks_workspace.example
}

output "datalake" {
  value = azurerm_storage_account.account
}
