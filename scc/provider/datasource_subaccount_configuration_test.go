package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccount(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount_configuration")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + DataSourceSubaccountConfiguration("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "display_name", "Terraform Subaccount Datasource"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "description", "This subaccount has all the configurations for data source."),

						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.user", user.CloudUsername),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.state", "Connected"),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "tunnel.connected_since_time_stamp", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.application_connections.#", "0"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.service_channels.#", "1"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.service_channels.0.type", "K8S"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.service_channels.0.state", "Connected"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.service_channels.0.details", "cp.da2b3e1.stage.kyma.ondemand.com:443/bd64665f-060a-47b6-8aba-f406703f0acf"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.test", "tunnel.service_channels.0.comment", ""),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "tunnel.subaccount_certificate.not_after_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "tunnel.subaccount_certificate.not_before_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.test", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
					),
				},
			},
		})

	})

	t.Run("error path - region host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountConfigurationWoRegionHost("test", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8"),
					ExpectError: regexp.MustCompile(`The argument "region_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - subaccount id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountConfigurationWoSubaccount("test", "cf.eu12.hana.ondemand.com"),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

}

func DataSourceSubaccountConfiguration(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_configuration" "%s"{
    region_host= "%s"
    subaccount= "%s"	
	}
	`, datasourceName, regionHost, subaccountID)
}

func DataSourceSubaccountConfigurationWoRegionHost(datasourceName string, subaccountID string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_configuration" "%s" {
    subaccount= "%s"
	}
	`, datasourceName, subaccountID)
}

func DataSourceSubaccountConfigurationWoSubaccount(datasourceName string, regionHost string) string {
	return fmt.Sprintf(`
	data "scc_subaccount_configuration" "%s" {
	region_host= "%s"
	}
	`, datasourceName, regionHost)
}
