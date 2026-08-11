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

func TestDataSourceSubjectPatternRule_Read_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ds := &datasources.SubjectPatternRuleDataSource{Client: tfutils.NewTestClient(t, srv)}

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)

	conditionAttrTypes := map[string]tftypes.Type{
		"variable": tftypes.String,
		"operator": tftypes.String,
		"value":    tftypes.String,
	}
	subjectPatternAttrTypes := map[string]tftypes.Type{
		"cn":    tftypes.String,
		"email": tftypes.String,
		"l":     tftypes.String,
		"ou":    tftypes.String,
		"o":     tftypes.String,
		"st":    tftypes.String,
		"c":     tftypes.String,
	}

	raw := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"index":           tftypes.Number,
				"description":     tftypes.String,
				"condition":       tftypes.Object{AttributeTypes: conditionAttrTypes},
				"subject_pattern": tftypes.Object{AttributeTypes: subjectPatternAttrTypes},
			},
		},
		map[string]tftypes.Value{
			"index":           tftypes.NewValue(tftypes.Number, 0),
			"description":     tftypes.NewValue(tftypes.String, nil),
			"condition":       tftypes.NewValue(tftypes.Object{AttributeTypes: conditionAttrTypes}, nil),
			"subject_pattern": tftypes.NewValue(tftypes.Object{AttributeTypes: subjectPatternAttrTypes}, nil),
		},
	)

	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := &datasource.ReadResponse{}
	ds.Read(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
