package datasources_test

import (
	"fmt"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubjectPatternRules(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_subject_pattern_rules")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceSubjectPatternRulesConfiguration("scc_sprs", 0),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rules.scc_sprs", "subject_pattern_rules.#", "1"),
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rules.scc_sprs", "subject_pattern_rules.0.description", ""),
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rules.scc_sprs", "subject_pattern_rules.0.condition.operator", "always_true"),
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rules.scc_sprs", "subject_pattern_rules.0.subject_pattern.cn", "${name}"),
					),
				},
			},
		})

	})
}

func DataSourceSubjectPatternRulesConfiguration(datasourceName string, index int64) string {
	return fmt.Sprintf(`
	data "scc_subject_pattern_rules" "%s" {}
	`, datasourceName)
}
