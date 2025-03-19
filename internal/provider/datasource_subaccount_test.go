package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubaccount(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_subaccount")
		rec.SetRealTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + DataSourceSubaccount("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_subaccount.test", "subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "display_name", "Terraform Subaccount Datasource"),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "description", "This subaccount has all the configurations for data source."),
						resource.TestCheckResourceAttr("data.cloudconnector_subaccount.test", "tunnel.user", "DL_65DDA8EBA97EAA0134EEB5DC@global.corp.sap"),
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
