package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestListSubaccount(t *testing.T) {
	t.Parallel()

	t.Run("happy path - without any filter", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount")
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
					Config: providerConfig(user) + listSubaccountQueryConfig("test", "scc"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount.test", 2),

						querycheck.ExpectIdentity(
							"scc_subaccount.test",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),

						querycheck.ExpectIdentity(
							"scc_subaccount.test",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
			},
		})
	})

	t.Run("happy path - with region_host filter", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount_with_filter")
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
					Config: providerConfig(user) + listSubaccountQueryConfigWithFilter("test", "scc", "cf.us10.hana.ondemand.com"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount.test", 0),
					},
				},
			},
		})
	})

	t.Run("error - invalid filter type", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: getTestProviders(nil),
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: `
						list "scc_subaccount" "test_err" {
							provider = "scc"
							config {
								region_host = 12345 # This triggers the error in Config.Get
							}
						}`,
					// This covers: if diags := req.Config.Get(ctx, &filter); diags.HasError()
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Type`),
				},
			},
		})
	})

}

func listSubaccountQueryConfig(lable, providerName string) string {
	return fmt.Sprintf(`list "scc_subaccount" "%s" {
               provider = "%s"
             }`, lable, providerName)
}

func listSubaccountQueryConfigWithFilter(lable, providerName, regionHost string) string {
	return fmt.Sprintf(`list "scc_subaccount" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    region_host="%s"
			   }
             }`, lable, providerName, regionHost)
}
