package datasources_test

import (
	"fmt"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceSubjectPatternRule(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		rec, user := tfutils.SetupVCR(t, "fixtures/datasource_subject_pattern_rule")
		defer tfutils.StopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: tfutils.ProviderConfig(user) + DataSourceSubjectPatternRuleConfiguration("scc_spr", 0),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rule.scc_spr", "index", "0"),
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rule.scc_spr", "description", ""),
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rule.scc_spr", "condition.operator", "always_true"),
						resource.TestCheckResourceAttr("data.scc_subject_pattern_rule.scc_spr", "subject_pattern.cn", "${name}"),
					),
				},
			},
		})

	})
}

func DataSourceSubjectPatternRuleConfiguration(datasourceName string, index int64) string {
	return fmt.Sprintf(`
	data "scc_subject_pattern_rule" "%s" {
    index = %d
	}
	`, datasourceName, index)
}
