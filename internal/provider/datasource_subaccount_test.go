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
		rec, user := setupVCR(t, "fixtures/datasource_subaccount")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + DataSourceSubaccount("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "display_name", "7e8b3cba-d0af-4989-9407-bcad93929ae7"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "description", "Subaccount for Data Source..DO NOT DELETE!!!"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "tunnel.user", "sarthak.goyal01@sap.com"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "tunnel.state", "Connected"),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "tunnel.connected_since_time_stamp", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "tunnel.connections", "0"),

						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "tunnel.subaccount_certificate.not_after_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "tunnel.subaccount_certificate.not_before_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
					),
				},
			},
		})

	})

}

func DataSourceSubaccount(datasourceName string, regionHost string, subaccountID string) string {
	return fmt.Sprintf(`
	data "cloudconnector_subaccount" "%s"{
    region_host= "%s"
    subaccount= "%s"	
	}
	`, datasourceName, regionHost, subaccountID)
}
