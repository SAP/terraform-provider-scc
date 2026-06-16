package listresources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
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
		rec, user := tfutils.SetupVCR(t, "fixtures/list_resource_subaccount_abap_service_channel")

		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: tfutils.ProviderConfig(user) + listSubaccountABAPServiceChannelQueryConfigWithIncludeResource("scc_abap_sc", "scc",
						regionHost, subaccount, false),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount_abap_service_channel.scc_abap_sc", 1),

						querycheck.ExpectIdentity(
							"scc_subaccount_abap_service_channel.scc_abap_sc",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(tfutils.RegexpValidUUID),
								"id":          knownvalue.Int64Exact(1),
								"type":        knownvalue.StringExact("ABAPCloud"),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_subaccount_abap_service_channel.scc_abap_sc",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringExact(subaccount),
								"id":          knownvalue.Int64Exact(1),
								"type":        knownvalue.StringExact("ABAPCloud"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("region_host"),
									KnownValue: knownvalue.StringExact(regionHost),
								},
								{
									Path:       tfjsonpath.New("subaccount"),
									KnownValue: knownvalue.StringRegexp(tfutils.RegexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("abap_cloud_tenant_host"),
									KnownValue: knownvalue.NotNull(),
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
									KnownValue: knownvalue.Int64Exact(1),
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
								{
									Path:       tfjsonpath.New("comment"),
									KnownValue: knownvalue.StringExact(""),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - subaccount not found", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/list_resource_subaccount_abap_service_channel_error_subaccount_not_found")

		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: tfutils.ProviderConfig(user) +
						listSubaccountABAPServiceChannelQueryConfig(
							"scc_abap_sc",
							"scc",
							"cf.eu12.hana.ondemand.com",
							"224492be-5f0f-4bb0-8f59-c982107bc878",
							false,
						),

					ExpectError: regexp.MustCompile(`(?i)404.*subaccount.*does not exist`),
				},
			},
		})
	})

	t.Run("error path - region_host mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query:       true,
					Config:      listSubaccountABAPServiceChannelQueryConfigWoRegionHost("scc_abap_sc", "scc", subaccount, false),
					ExpectError: regexp.MustCompile(`The argument "region_host" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - subaccount mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query:       true,
					Config:      listSubaccountABAPServiceChannelQueryConfigWoSubaccount("scc_abap_sc", "scc", regionHost, false),
					ExpectError: regexp.MustCompile(`The argument "subaccount" is required, but no definition was found.`),
				},
			},
		})
	})

	t.Run("error path - snc_encrypted mandatory", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(nil),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query:       true,
					Config:      listSubaccountABAPServiceChannelQueryConfigWoSNCEncrypted("scc_abap_sc", "scc", regionHost, subaccount),
					ExpectError: regexp.MustCompile(`The argument "snc_encrypted" is required, but no definition was found.`),
				},
			},
		})
	})

}

func listSubaccountABAPServiceChannelQueryConfig(lable, providerName, regionHost, subaccount string, sncEncrypted bool) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
				snc_encrypted=%t
			   }
             }`, lable, providerName, regionHost, subaccount, sncEncrypted)
}

func listSubaccountABAPServiceChannelQueryConfigWithIncludeResource(lable, providerName, regionHost, subaccount string, sncEncrypted bool) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
				subaccount="%s"
				snc_encrypted=%t
			   }
             }`, lable, providerName, regionHost, subaccount, sncEncrypted)
}

func listSubaccountABAPServiceChannelQueryConfigWoRegionHost(lable, providerName, subaccount string, sncEncrypted bool) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   config {
				subaccount="%s"
				snc_encrypted=%t
			   }
             }`, lable, providerName, subaccount, sncEncrypted)
}

func listSubaccountABAPServiceChannelQueryConfigWoSubaccount(lable, providerName, regionHost string, sncEncrypted bool) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				snc_encrypted=%t
			   }
             }`, lable, providerName, regionHost, sncEncrypted)
}

func listSubaccountABAPServiceChannelQueryConfigWoSNCEncrypted(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_subaccount_abap_service_channel" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}
