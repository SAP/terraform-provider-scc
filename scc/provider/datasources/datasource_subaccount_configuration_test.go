package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccount(t *testing.T) {
	regionHost := "cf.eu12.hana.ondemand.com"
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_subaccount_configuration")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceSubaccountConfiguration("scc_sa", regionHost, subaccount),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "region_host", regionHost),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "subaccount", tfutils.RegexpValidUUID),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "display_name", "Terraform Subaccount Datasource"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "description", "Subaccount used for all data sources in Cloud Connector Instance. DO NOT DELETE!!!"),

						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.user", user.CloudUsername),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.state", "Connected"),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.connected_since", tfutils.RegexpValidTimeStamp),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.application_connections.#", "0"),
						resource.TestCheckResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.service_channels.#", "0"),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.subaccount_certificate.valid_to", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.subaccount_certificate.valid_from", tfutils.RegexpValidTimeStamp),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("data.scc_subaccount_configuration.scc_sa", "tunnel.subaccount_certificate.serial_number", tfutils.RegexpValidSerialNumber),
					),
				},
			},
		})

	})

	t.Run("error path - region host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountConfigurationWoRegionHost("scc_sa", subaccount),
					ExpectError: regexp.MustCompile(`The argument "region_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - subaccount id mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      DataSourceSubaccountConfigurationWoSubaccount("scc_sa", regionHost),
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
