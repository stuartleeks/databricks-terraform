package databricks

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAzureAdlsGen2Mount_capture_error(t *testing.T) {
	randPrefix := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		// CheckDestroy: testAccCheckShellScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccAzureAdlsGen2Mount_capture_error(randPrefix),
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("Something went wrong!"),
			},
		},
	})
}

func testAccAzureAdlsGen2Mount_capture_error(rg string) string {
	// return fmt.Sprintf(`
	// 							resource "azure" "my_scope" {
	// 							  name = "%s"
	// 							}
	// 							`, scopeName)
	return rg
}
