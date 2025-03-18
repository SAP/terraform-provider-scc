package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSystemMapping(t *testing.T) {

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/datasource_system_mapping")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: providerConfig("", user) + DataSourceSystemMapping("test", "cf.us10.hana.ondemand.com", "7e8b3cba-d0af-4989-9407-bcad93929ae7", "testterraformvirtual", "900"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "credentials.region_host", "cf.us10.hana.ondemand.com"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mapping.test", "credentials.subaccount", regexpValidUUID),

						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.virtual_host", "testterraformvirtual"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.virtual_port", "900"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.local_host", "testterraforminternal"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.local_port", "900"),
						resource.TestMatchResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.creation_date", regexValidTimeStamp),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.protocol", "HTTP"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.backend_type", "abapSys"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.authentication_mode", "KERBEROS"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.host_in_header", "VIRTUAL"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.sid", ""),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.total_resources_count", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.enabled_resources_count", "1"),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.description", ""),
						resource.TestCheckResourceAttr("data.cloudconnector_system_mapping.test", "system_mapping.sap_router", ""),
					),
				},
			},
		})

	})

}

func DataSourceSystemMapping(datasourceName string, regionHost string, subaccount string, virtualHost string, virtualPort string) string {
	return fmt.Sprintf(`
	data "cloudconnector_system_mapping" "%s" {
    credentials= {
        region_host= "%s"
        subaccount= "%s"
    }
	system_mapping={
		virtual_host= "%s"
		virtual_port= "%s"
	}	
	}
	`, datasourceName, regionHost, subaccount, virtualHost, virtualPort)
}
