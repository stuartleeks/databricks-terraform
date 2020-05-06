package databricks

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/joho/godotenv"
	azuread "github.com/terraform-providers/terraform-provider-azuread/azuread"
	azurerm "github.com/terraform-providers/terraform-provider-azurerm/azurerm"
	"log"
	"os"
	"testing"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider("").(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"databricks": testAccProvider,
		"azurerm":    azurerm.Provider(),
		"azuread":    azuread.Provider(),
	}
}

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Failed to load environment")
	}
	code := m.Run()
	os.Exit(code)
}

func azurePreCheck(t *testing.T) {
	variables := []string{
		"ARM_CLIENT_ID",
		"ARM_CLIENT_SECRET",
		"ARM_SUBSCRIPTION_ID",
		"ARM_TENANT_ID",
		"ARM_TEST_LOCATION",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for Azure acceptance tests!", variable)
		}
	}
}
