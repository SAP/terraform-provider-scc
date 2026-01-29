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

func TestListSubaccountABAPServiceChannel(t *testing.T) {
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	regionHost := "cf.eu12.hana.ondemand.com"

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount_abap_service_channel")

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
					Config: providerConfig(user) + listSubaccountABAPServiceChannelQueryConfig("scc_abap_sc", "scc",
						regionHost, subaccount),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount_abap_service_channel.scc_abap_sc", 1),

						querycheck.ExpectIdentity(
							"scc_subaccount_abap_service_channel.scc_abap_sc",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
								"id":          knownvalue.Int64Exact(3),
							},
						),
					},
				},
				// Verify list results contain full resource schema data
				{
					Query: true,
					Config: providerConfig(user) + listSubaccountABAPServiceChannelQueryConfigWithIncludeResource("scc_abap_sc", "scc",
						regionHost, subaccount),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount_abap_service_channel.scc_abap_sc", 1),

						querycheck.ExpectIdentity(
							"scc_subaccount_abap_service_channel.scc_abap_sc",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
								"id":          knownvalue.Int64Exact(3),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_subaccount_abap_service_channel.scc_abap_sc",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringExact(subaccount),
								"id":          knownvalue.Int64Exact(3),
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
									Path:       tfjsonpath.New("abap_cloud_tenant_host"),
									KnownValue: knownvalue.StringRegexp(regexp.MustCompile(`.*abap\.region\.hana\.ondemand\.com|REDACTED.*`)),
								},
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("ABAPCloud"),
								},
								{
									Path: tfjsonpath.New("state"),
									KnownValue: knownvalue.ObjectExact(map[string]knownvalue.Check{
										"connected":                  knownvalue.Bool(false),
										"opened_connections":         knownvalue.Int64Exact(0),
										"connected_since_time_stamp": knownvalue.Int64Exact(0),
									}),
								},
								{
									Path:       tfjsonpath.New("connections"),
									KnownValue: knownvalue.Int64Exact(1),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.Int64Exact(3),
								},
								{
									Path:       tfjsonpath.New("enabled"),
									KnownValue: knownvalue.Bool(false),
								},
								{
									Path:       tfjsonpath.New("instance_number"),
									KnownValue: knownvalue.Int64Exact(50),
								},
								{
									Path:       tfjsonpath.New("port"),
									KnownValue: knownvalue.Int64Exact(3350),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - subaccount not found", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount_abap_service_channel_error_subaccount_not_found")

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
						listSubaccountABAPServiceChannelQueryConfig(
							"scc_abap_sc",
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

func listSubaccountABAPServiceChannelQueryConfig(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}

func listSubaccountABAPServiceChannelQueryConfigWithIncludeResource(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}
