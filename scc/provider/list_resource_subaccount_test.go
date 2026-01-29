package provider

import (
	"fmt"
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
					Config: providerConfig(user) + listSubaccountQueryConfig("scc_sa", "scc"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount.scc_sa", 2),

						querycheck.ExpectIdentity(
							"scc_subaccount.scc_sa",
							map[string]knownvalue.Check{
								"region_host": knownvalue.StringExact("cf.eu12.hana.ondemand.com"),
								"subaccount":  knownvalue.StringRegexp(regexpValidUUID),
							},
						),

						querycheck.ExpectIdentity(
							"scc_subaccount.scc_sa",
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
					Config: providerConfig(user) + listSubaccountQueryConfigWithFilter("scc_sa", "scc", "cf.us10.hana.ondemand.com"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount.scc_sa", 0),
					},
				},
			},
		})
	})
}

func listSubaccountQueryConfig(lable, providerName string) string {
	return fmt.Sprintf(`list "scc_subaccount" "%s" {
               provider = "%s"
			   include_resource = true
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
