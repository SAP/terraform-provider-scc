package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSubaccount(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/resource_subaccount")
		rec.SetRealTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})

		if len(user.CloudUsername) == 0 || len(user.CloudPassword) == 0 {
			t.Fatalf("Missing TF_CLOUD_USER or TF_CLOUD_PASSWORD for recording test fixtures")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig(user) + ResourceSubaccount("test", "cf.eu12.hana.ondemand.com", "7480ee65-e039-41cf-ba72-6aaf56c312df", user.CloudUsername, user.CloudPassword, "subaccount added via terraform tests"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "cloud_user", user.CloudUsername),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "cloud_password", user.CloudPassword),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "description", "subaccount added via terraform tests"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "location_id", ""),

						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.connected_since_time_stamp", regexValidTimeStamp),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "tunnel.connections", "0"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "tunnel.state", "Connected"),
						resource.TestCheckResourceAttr("cloudconnector_subaccount.test", "tunnel.user", user.CloudUsername),

						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.issuer", regexp.MustCompile(`CN=.*?,OU=S.*?,O=.*?,L=.*?,C=.*?`)),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.not_after_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.not_before_time_stamp", regexValidTimeStamp),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.serial_number", regexValidSerialNumber),
						resource.TestMatchResourceAttr("cloudconnector_subaccount.test", "tunnel.subaccount_certificate.subject_dn", regexp.MustCompile(`CN=.*?,L=.*?,OU=.*?,OU=.*?,O=.*?,C=.*?`)),
					),
				},
				{
					ResourceName:      "cloudconnector_subaccount.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: getImportStateForSubaccount("cloudconnector_subaccount.test"),
					ImportStateVerifyIdentifierAttribute: "subaccount",
					ImportStateVerifyIgnore: []string{
						"cloud_user",
						"cloud_password",
					},
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

func getImportStateForSubaccount(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s",
			rs.Primary.Attributes["region_host"],
			rs.Primary.Attributes["subaccount"],
		), nil
	}
}
