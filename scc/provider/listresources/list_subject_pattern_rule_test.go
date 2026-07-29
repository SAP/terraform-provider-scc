package listresources_test

import (
	"fmt"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestListSubjectPatternRule(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/list_resource_subject_pattern_rule")

		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query:  true,
					Config: tfutils.ProviderConfig(user) + listSubjectPatternRuleQueryConfig("scc_spr", "scc", 0),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subject_pattern_rule.scc_spr", 1),

						querycheck.ExpectIdentity(
							"scc_subject_pattern_rule.scc_spr",
							map[string]knownvalue.Check{
								"index": knownvalue.Int64Exact(0),
							},
						),
					},
				},
				// Verify list results contain full resource schema data
				{
					Query:  true,
					Config: tfutils.ProviderConfig(user) + listSubjectPatternRuleQueryConfigWithIncludeResource("scc_spr", "scc", 0),

					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("scc_subject_pattern_rule.scc_spr", 1),

						querycheck.ExpectIdentity(
							"scc_subject_pattern_rule.scc_spr",
							map[string]knownvalue.Check{
								"index": knownvalue.Int64Exact(0),
							},
						),

						// Resource data check (ONLY because include_resource = true)
						querycheck.ExpectResourceKnownValues(
							"scc_subject_pattern_rule.scc_spr",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"index": knownvalue.Int64Exact(0),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("index"),
									KnownValue: knownvalue.Int64Exact(0),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listSubjectPatternRuleQueryConfig(lable, providerName string, index int64) string {
	return fmt.Sprintf(`list "scc_subject_pattern_rule" "%s" {
               provider = "%s"
			   config {
			    index="%d"
			   }
             }`, lable, providerName, index)
}

func listSubjectPatternRuleQueryConfigWithIncludeResource(lable, providerName string, index int64) string {
	return fmt.Sprintf(`list "scc_subject_pattern_rule" "%s" {
               provider = "%s"
			   include_resource = true
			   config {
			    index="%d"
			   }
             }`, lable, providerName, index)
}
