package model

import (
	"context"
	"fmt"
	"strings"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubjectPatternRulesDataSourceConfig struct {
	SubjectPatternRules []SubjectPatternRule `tfsdk:"subject_pattern_rules"`
}

type SubjectPatternRule struct {
	Description    types.String `tfsdk:"description"`
	SubjectPattern types.Object `tfsdk:"subject_pattern"`
	Condition      types.Object `tfsdk:"condition"`
}

type SubjectPatternRuleConfig struct {
	Index types.Int64 `tfsdk:"index"`
	SubjectPatternRule
}

type SubjectPatternCondition struct {
	Variable types.String `tfsdk:"variable"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
}

var SubjectPatternConditionType = map[string]attr.Type{
	"variable": types.StringType,
	"operator": types.StringType,
	"value":    types.StringType,
}

type SubjectPattern struct {
	CommonName       types.String `tfsdk:"cn"`
	Email            types.String `tfsdk:"email"`
	Locality         types.String `tfsdk:"l"`
	OrganizationUnit types.String `tfsdk:"ou"`
	Organization     types.String `tfsdk:"o"`
	State            types.String `tfsdk:"st"`
	Country          types.String `tfsdk:"c"`
}

var SubjectPatternType = map[string]attr.Type{
	"cn":    types.StringType,
	"email": types.StringType,
	"l":     types.StringType,
	"ou":    types.StringType,
	"o":     types.StringType,
	"st":    types.StringType,
	"c":     types.StringType,
}

type SubjectPatternRuleListFilterModel struct {
	SubjectPatternRuleConfig
}

func SubjectPatternRuleValueFrom(ctx context.Context, plan SubjectPatternRuleConfig, value apiobjects.SubjectPatternRule) (SubjectPatternRuleConfig, diag.Diagnostics) {
	sp := SubjectPattern{
		CommonName:       helpers.StringValueOrNull(value.SubjectPattern.CommonName),
		Email:            helpers.StringValueOrNull(value.SubjectPattern.Email),
		Locality:         helpers.StringValueOrNull(value.SubjectPattern.Locality),
		OrganizationUnit: helpers.StringValueOrNull(value.SubjectPattern.OrganizationUnit),
		Organization:     helpers.StringValueOrNull(value.SubjectPattern.Organization),
		State:            helpers.StringValueOrNull(value.SubjectPattern.State),
		Country:          helpers.StringValueOrNull(value.SubjectPattern.Country),
	}
	subjectPattern, diags := types.ObjectValueFrom(ctx, SubjectPatternType, sp)
	if diags.HasError() {
		return SubjectPatternRuleConfig{}, diags
	}

	condition, diags := ParseSubjectPatternRuleCondition(ctx, value.Condition)
	if diags.HasError() {
		return SubjectPatternRuleConfig{}, diags
	}

	sprModel := SubjectPatternRuleConfig{
		Index: plan.Index,
		SubjectPatternRule: SubjectPatternRule{
			Description:    types.StringValue(value.Description),
			SubjectPattern: subjectPattern,
			Condition:      condition,
		},
	}

	return sprModel, nil
}

func SubjectPatternRulesValueFrom(ctx context.Context, plan SubjectPatternRulesDataSourceConfig, value apiobjects.SubjectPatternRules) (SubjectPatternRulesDataSourceConfig, diag.Diagnostics) {
	subjectPatternRules := make([]SubjectPatternRule, 0, len(value))
	for _, spr := range value {
		sp := SubjectPattern{
			CommonName:       helpers.StringValueOrNull(spr.SubjectPattern.CommonName),
			Email:            helpers.StringValueOrNull(spr.SubjectPattern.Email),
			Locality:         helpers.StringValueOrNull(spr.SubjectPattern.Locality),
			OrganizationUnit: helpers.StringValueOrNull(spr.SubjectPattern.OrganizationUnit),
			Organization:     helpers.StringValueOrNull(spr.SubjectPattern.Organization),
			State:            helpers.StringValueOrNull(spr.SubjectPattern.State),
			Country:          helpers.StringValueOrNull(spr.SubjectPattern.Country),
		}
		subjectPattern, diags := types.ObjectValueFrom(ctx, SubjectPatternType, sp)
		if diags.HasError() {
			return SubjectPatternRulesDataSourceConfig{}, diags
		}

		condition, diags := ParseSubjectPatternRuleCondition(ctx, spr.Condition)
		if diags.HasError() {
			return SubjectPatternRulesDataSourceConfig{}, diags
		}

		sprModel := SubjectPatternRule{
			Description:    types.StringValue(spr.Description),
			SubjectPattern: subjectPattern,
			Condition:      condition,
		}
		subjectPatternRules = append(subjectPatternRules, sprModel)
	}
	model := SubjectPatternRulesDataSourceConfig{
		SubjectPatternRules: subjectPatternRules,
	}

	return model, nil
}

func ParseSubjectPatternRuleCondition(ctx context.Context, condition string) (types.Object, diag.Diagnostics) {
	var c SubjectPatternCondition

	switch {
	case strings.TrimSpace(condition) == "always true":
		c = SubjectPatternCondition{
			Variable: types.StringNull(),
			Operator: types.StringValue("always_true"),
			Value:    types.StringNull(),
		}

	case strings.HasSuffix(condition, " does not exist"):
		variable := strings.TrimSpace(strings.TrimSuffix(condition, " does not exist"))
		c = SubjectPatternCondition{
			Variable: types.StringValue(variable),
			Operator: types.StringValue("exist_not"),
			Value:    types.StringNull(),
		}

	case strings.HasSuffix(condition, " exists"):
		variable := strings.TrimSpace(strings.TrimSuffix(condition, " exists"))
		c = SubjectPatternCondition{
			Variable: types.StringValue(variable),
			Operator: types.StringValue("exist"),
			Value:    types.StringNull(),
		}

	case strings.Contains(condition, " is not "):
		parts := strings.SplitN(condition, " is not ", 2)
		c = SubjectPatternCondition{
			Variable: types.StringValue(strings.TrimSpace(parts[0])),
			Operator: types.StringValue("is_not"),
			Value:    types.StringValue(strings.TrimSpace(parts[1])),
		}

	case strings.HasSuffix(condition, " is not"):
		variable := strings.TrimSpace(strings.TrimSuffix(condition, " is not"))
		c = SubjectPatternCondition{
			Variable: types.StringValue(variable),
			Operator: types.StringValue("is_not"),
			Value:    types.StringValue(""),
		}

	case strings.Contains(condition, " is "):
		parts := strings.SplitN(condition, " is ", 2)
		c = SubjectPatternCondition{
			Variable: types.StringValue(strings.TrimSpace(parts[0])),
			Operator: types.StringValue("is"),
			Value:    types.StringValue(strings.TrimSpace(parts[1])),
		}

	case strings.HasSuffix(condition, " is"):
		variable := strings.TrimSpace(strings.TrimSuffix(condition, " is"))
		c = SubjectPatternCondition{
			Variable: types.StringValue(variable),
			Operator: types.StringValue("is"),
			Value:    types.StringValue(""),
		}

	default:
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid subject pattern condition",
				fmt.Sprintf("Unsupported condition %q", condition),
			),
		}
	}

	return types.ObjectValueFrom(ctx, SubjectPatternConditionType, c)
}
