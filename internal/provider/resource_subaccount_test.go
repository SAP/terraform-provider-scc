package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceSubaccount(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + ResourceSubaccount("test", "cf.us10.hana.ondemand.com", "6d4250ed-bd75-4a54-a329-46221823c74b", "sarthak.goyal01@sap.com", "9896Sg@20196829", "subaccount added via terraform"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "cloud_user", "sarthak.goyal01@sap.com"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "cloud_password", "9896Sg@20196829"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "description", "subaccount added via terraform"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "location_id", ""),

						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.connected_since_time_stamp", regexValidTimeStamp),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "tunnel.state", "Connected"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "tunnel.user", "sarthak.goyal01@sap.com"),

						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.not_after_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.not_before_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
					),
				},
			},
		})

	})

}

func ResourceSubaccount(datasourceName string, regionHost string, subaccount string, cloudUser string, cloudPassword string, description string) string {
	return fmt.Sprintf(`
	resource "cloudconnector_subaccount" "%s" {
    region_host= "%s"
    subaccount= "%s"
    cloud_user= "%s"
    cloud_password= "%s" 
    description= "%s"
	}
	`, datasourceName, regionHost, subaccount, cloudUser, cloudPassword, description)
}
