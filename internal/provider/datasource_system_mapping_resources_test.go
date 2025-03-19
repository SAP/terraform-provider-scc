package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMappingResources(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping_resources")
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
					Config: providerConfig("", user) + DataSourceSystemMappingResources("test", "cf.eu12.hana.ondemand.com", "0bcb0012-a982-42f9-bda4-0a5cb15f88c8", "testterraformvirtual", "900"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.region_host", "cf.eu12.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.subaccount", regexpValidUUID),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.virtual_host", "testterraformvirtual"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "credentials.virtual_port", "900"),

						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.#", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.id", "/google.com"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.enabled", "true"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.exact_match_only", "true"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.websocket_upgrade_allowed", "false"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping_resources.test", "system_mapping_resources.0.description", ""),
					),
				},
			},
		})

	})

}

func DataSourceSystemMappingResources(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "cloudconnector_system_mapping_resources" "%s" {
    credentials= {
        region_host= "%s"
        subaccount= "%s"
        virtual_host= "%s"
        virtual_port= "%s"
    }
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort)
}
