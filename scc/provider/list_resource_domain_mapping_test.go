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

func TestListDomainMapping(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_domain_mapping")
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
					Config: providerConfig(user) + listDomainMappingQueryConfig("test", "scc", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_domain_mapping.test", 1),

						querycheck.ExpectIdentity(
							"scc_domain_mapping.test",
							map[string]knownvalue.Check{
								"region_host":     knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":      knownvalue.StringRegexp(regexpValidUUID),
								"internal_domain": knownvalue.StringExact("testterraforminternaldomain"),
							},
						),
					},
				},
				// Verify list results contain full resource schema data
				{
					Query:  true,
					Config: providerConfig(user) + listDomainMappingQueryConfigWithIncludeResource("test", "scc", "cf.eu12.hana.ondemand.com", "304492be-5f0f-4bb0-8f59-c982107bc878"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_domain_mapping.test", 1),

						querycheck.ExpectIdentity(
							"scc_domain_mapping.test",
							map[string]knownvalue.Check{
								"region_host":     knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":      knownvalue.StringRegexp(regexpValidUUID),
								"internal_domain": knownvalue.StringExact("testterraforminternaldomain"),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_domain_mapping.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"region_host":     knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":      knownvalue.StringExact("304492be-5f0f-4bb0-8f59-c982107bc878"),
								"internal_domain": knownvalue.StringExact("testterraforminternaldomain"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("region_host"),
									KnownValue: knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								},
								{
									Path:       tfjsonpath.New("subaccount"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("internal_domain"),
									KnownValue: knownvalue.StringExact("testterraforminternaldomain"),
								},
								{
									Path:       tfjsonpath.New("virtual_domain"),
									KnownValue: knownvalue.StringExact("testterraformvirtualdomain"),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - subaccount not found", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_domain_mapping_error_subaccount_not_found")
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
						listDomainMappingQueryConfig(
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

func listDomainMappingQueryConfig(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_domain_mapping" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}

func listDomainMappingQueryConfigWithIncludeResource(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_domain_mapping" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}
