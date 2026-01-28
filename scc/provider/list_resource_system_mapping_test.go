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

func TestListSystemMapping(t *testing.T) {
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	regionHost := "cf.eu12.hana.ondemand.com"

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_system_mapping")
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
					Query:  true,
					Config: providerConfig(user) + listSystemMappingQueryConfig("test", "scc", regionHost, subaccount),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_system_mapping.test", 1),

						querycheck.ExpectIdentity(
							"scc_system_mapping.test",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact("testterraformvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
							},
						),
					},
				},
				// Verify list results contain full resource schema data
				{
					Query:  true,
					Config: providerConfig(user) + listSystemMappingQueryConfigWithIncludeResource("test", "scc", regionHost, subaccount),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_system_mapping.test", 1),

						querycheck.ExpectIdentity(
							"scc_system_mapping.test",
							map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringRegexp(regexpValidUUID),
								"virtual_host": knownvalue.StringExact("testterraformvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_system_mapping.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"region_host":  knownvalue.StringExact(regionHost),
								"subaccount":   knownvalue.StringExact(subaccount),
								"virtual_host": knownvalue.StringExact("testterraformvirtual"),
								"virtual_port": knownvalue.StringExact("900"),
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
									Path:       tfjsonpath.New("internal_host"),
									KnownValue: knownvalue.StringExact("testterraforminternal"),
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
									Path:       tfjsonpath.New("internal_port"),
									KnownValue: knownvalue.StringExact("900"),
								},
								{
									Path:       tfjsonpath.New("protocol"),
									KnownValue: knownvalue.StringExact("HTTP"),
								},
								{
									Path:       tfjsonpath.New("authentication_mode"),
									KnownValue: knownvalue.StringExact("KERBEROS"),
								},
								{
									Path:       tfjsonpath.New("backend_type"),
									KnownValue: knownvalue.StringExact("abapSys"),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - subaccount not found", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_system_mapping_error_subaccount_not_found")
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
						listSystemMappingQueryConfig(
							"test",
							"scc",
							"cf.eu12.hana.ondemand.com",
							"224492be-5f0f-4bb0-8f59-c982107bc878",
						),

					ExpectError: regexp.MustCompile(`(?i)404.*subaccount.*does not exist`),
				},
			},
		})
	})

}

func listSystemMappingQueryConfig(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_system_mapping" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}

func listSystemMappingQueryConfigWithIncludeResource(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_system_mapping" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}
