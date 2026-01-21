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

	t.Run("happy path - multiple results", func(t *testing.T) {
		rec, user := setupVCR(t, "fixtures/list_resource_subaccount")
		if user.CloudUsername == "" || user.CloudPassword == "" {
			t.Fatalf("Missing TF_VAR_cloud_user or TF_VAR_cloud_password")
		}
		defer stopQuietly(rec)

		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query:  true,
					Config: providerConfig(user) + listSubaccountQueryConfig("test", "scc"),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subaccount.test", 3),

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
}

func listSubaccountQueryConfig(lable, providerName string) string {
	return fmt.Sprintf(`list "scc_subaccount" "%s" {
               provider = "%s"
             }`, lable, providerName)
}
