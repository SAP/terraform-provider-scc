package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestListSystemMappingResource(t *testing.T) {
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	regionHost := "cf.eu12.hana.ondemand.com"

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_system_mapping_resource")
		if user.CloudUsername == "" || user.CloudPassword == "" {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: providerConfig(user) + listSystemMappingResourceQueryConfig("test", "scc",
						regionHost, subaccount, "testterraformvirtual", "900"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_system_mapping_resource.test", 1),

						querycheck.ExpectIdentity(
							"scc_system_mapping_resource.test",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact("testterraformvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
								"url_path":     knownvalue.StringExact("/"),
							},
						),
					},
				},
				// Verify list results contain full resource schema data
				{
					Query: true,
					Config: providerConfig(user) + listSystemMappingResourceQueryConfigWithIncludeResource("test", "scc",
						regionHost, subaccount, "testterraformvirtual",
						"900"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_system_mapping_resource.test", 1),

						querycheck.ExpectIdentity(
							"scc_system_mapping_resource.test",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact("testterraformvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
								"url_path":     knownvalue.StringExact("/"),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_system_mapping_resource.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringExact(subaccount),
								"virtual_host": knownvalue.StringExact("testterraformvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
								"url_path":     knownvalue.StringExact("/"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("region_host"),
									KnownValue: knownvalue.StringExact(regionHost),
								},
								{
									Path:       tfjsonpath.New("subaccount"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("virtual_host"),
									KnownValue: knownvalue.StringExact("testterraformvirtual"),
								},
								{
									Path:       tfjsonpath.New("virtual_port"),
									KnownValue: knownvalue.StringExact("900"),
								},

								{
									Path:       tfjsonpath.New("url_path"),
									KnownValue: knownvalue.StringExact("/"),
								},
								{
									Path:       tfjsonpath.New("path_only"),
									KnownValue: knownvalue.Bool(true),
								},
								{
									Path:       tfjsonpath.New("enabled"),
									KnownValue: knownvalue.Bool(true),
								},
								{
									Path:       tfjsonpath.New("websocket_upgrade_allowed"),
									KnownValue: knownvalue.Bool(false),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - subaccount not found", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_system_mapping_resource_error_subaccount_not_found")
		if user.CloudUsername == "" || user.CloudPassword == "" {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password")
		}
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: providerConfig(user) +
						listSystemMappingResourceQueryConfig(
							"test",
							"scc",
							"cf.eu12.hana.ondemand.com",
							"224492be-5f0f-4bb0-8f59-c982107bc878",
							"testterraformvirtual",
							"900",
						),

					ExpectError: regexp.MustCompile(`(?i)404.*subaccount.*does not exist`),
				},
			},
		})
	})

}

func listSystemMappingResourceQueryConfig(lable, providerName, regionHost, subaccount, virtualHost, virtualPort string) string {
	return fmt.Sprintf(`list "scc_system_mapping_resource" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
				virtual_host="%s"
				virtual_port="%s"
			   }
             }`, lable, providerName, regionHost, subaccount, virtualHost, virtualPort)
}

func listSystemMappingResourceQueryConfigWithIncludeResource(lable, providerName, regionHost, subaccount, virtualHost, virtualPort string) string {
	return fmt.Sprintf(`list "scc_system_mapping_resource" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
				subaccount="%s"
				virtual_host="%s"
				virtual_port="%s"
			   }
             }`, lable, providerName, regionHost, subaccount, virtualHost, virtualPort)
}
