package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					Config: providerConfig("", user) + DataSourceSubaccounts("test"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_subaccounts.test", "subaccounts.#", "1"),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccounts.test", "subaccounts.0.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccounts.test", "subaccounts.0.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccounts.test", "subaccounts.0.location_id", ""),
					),
				},
			},
		})

	})

}

func DataSourceSubaccounts(datasourceName string) string {
	return fmt.Sprintf(`
	data "cloudconnector_subaccounts" "%s"{}
	`, datasourceName)
}
