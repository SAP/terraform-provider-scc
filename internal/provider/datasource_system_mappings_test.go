package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappings(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mappings")
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
					Config: providerConfig("", user) + DataSourceSystemMappings("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "credentials.region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mappings.test", "credentials.subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.#", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.virtual_host", "testterraformvirtual"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.virtual_port", "900"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.local_host", "testterraforminternal"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.local_port", "900"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.protocol", "HTTP"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.backend_type", "abapSys"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.authentication_mode", "KERBEROS"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.sid", ""),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.total_resources_count", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.enabled_resources_count", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.description", ""),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mappings.test", "system_mappings.0.sap_router", ""),
					),
				},
			},
		})

	})

}

func DataSourceSystemMappings(datasourceName string, regionHost string, subaccount string) string {
	return fmt.Sprintf(`
	data "cloudconnector_system_mappings" "%s" {
    credentials= {
        region_host= "%s"
        subaccount= "%s"
    }
	}
	`, datasourceName, regionHost, subaccount)
}
