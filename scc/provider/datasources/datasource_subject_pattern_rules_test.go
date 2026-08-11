package datasources_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SAP/terraform-provider-scc/scc/provider/datasources"
	"github.com/SAP/terraform-provider-scc/scc/provider/tfutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
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

func TestDataSourceSubjectPatternRules_Read_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ds := &datasources.SubjectPatternRulesDataSource{Client: tfutils.NewTestClient(t, srv)}

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)

	schemaType := schemaResp.Schema.Type().TerraformType(context.Background())
	raw := tftypes.NewValue(schemaType, nil)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := &datasource.ReadResponse{}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
