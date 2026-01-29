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

func TestListSubaccountK8SServiceChannel(t *testing.T) {
	subaccount := "1de4ab49-1b7b-47ca-89bb-0a4d9da1d057"
	regionHost := "cf.eu12.hana.ondemand.com"

	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount_k8s_service_channel")
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
					Config: providerConfig(user) + listSubaccountK8SServiceChannelQueryConfig("test", "scc",
						regionHost, subaccount),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount_k8s_service_channel.test", 1),

						querycheck.ExpectIdentity(
							"scc_subaccount_k8s_service_channel.test",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
								"id":          knownvalue.Int64Exact(4),
							},
						),
					},
				},
				// Verify list results contain full resource schema data
				{
					Query: true,
					Config: providerConfig(user) + listSubaccountK8SServiceChannelQueryConfigWithIncludeResource("test", "scc",
						regionHost, subaccount),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount_k8s_service_channel.test", 1),

						querycheck.ExpectIdentity(
							"scc_subaccount_k8s_service_channel.test",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
								"id":          knownvalue.Int64Exact(4),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_subaccount_k8s_service_channel.test",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact(regionHost),
								"subaccount":  knownvalue.StringExact(subaccount),
								"id":          knownvalue.Int64Exact(4),
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
									Path:       tfjsonpath.New("k8s_cluster_host"),
									KnownValue: knownvalue.StringRegexp(regexp.MustCompile(`^(test_cluster_host|REDACTED_K8S_CLUSTER_HOST)$`)),
								},
								{
									Path:       tfjsonpath.New("k8s_service_id"),
									KnownValue: knownvalue.StringRegexp(regexp.MustCompile(`^test_service_id|REDACTED_K8S_SERVICE_ID$`)),
								},
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("K8S"),
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
									KnownValue: knownvalue.Int64Exact(4),
								},
								{
									Path:       tfjsonpath.New("enabled"),
									KnownValue: knownvalue.Bool(false),
								},
								{
									Path:       tfjsonpath.New("local_port"),
									KnownValue: knownvalue.Int64Exact(3000),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - subaccount not found", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount_k8s_service_channel_error_subaccount_not_found")
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
						listSubaccountK8SServiceChannelQueryConfig(
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

func listSubaccountK8SServiceChannelQueryConfig(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_subaccount_k8s_service_channel" "%s" {
               provider = "%s"
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}

func listSubaccountK8SServiceChannelQueryConfigWithIncludeResource(lable, providerName, regionHost, subaccount string) string {
	return fmt.Sprintf(`list "scc_subaccount_k8s_service_channel" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
				subaccount="%s"
			   }
             }`, lable, providerName, regionHost, subaccount)
}
