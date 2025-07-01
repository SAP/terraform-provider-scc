package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataSourceSubaccounts(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccounts")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSubaccounts("test"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.ComposeTestCheckFunc(
							func(s *terraform.State) error {
								rs, ok := s.RootModule().Resources["data.scc_subaccounts.test"]
								if !ok {
									return fmt.Errorf("Not found: %s", "data.scc_subaccounts.test")
								}

								subaccounts := rs.Primary.Attributes["subaccounts.#"]
								if subaccounts != "2" && subaccounts != "3" {
									return fmt.Errorf("Expected subaccounts to be 2 or 3, got: %s", subaccounts)
								}
								return nil
							},
						),
						resource.TestMatchResourceAttr("data.scc_subaccounts.test", "subaccounts.0.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_subaccounts.test", "subaccounts.0.region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestCheckResourceAttr("data.scc_subaccounts.test", "subaccounts.0.location_id", ""),
					),
				},
			},
		})

	})

}

func DataSourceSubaccounts(datasourceName string) string {
	return fmt.Sprintf(`
	data "scc_subaccounts" "%s"{}
	`, datasourceName)
}
