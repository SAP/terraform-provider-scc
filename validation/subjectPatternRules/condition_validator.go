package subjectpatternrules

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type subjectPatternRuleConditionValidator struct{}

func SubjectPatternRuleCondition() validator.Object {
	return subjectPatternRuleConditionValidator{}
}

func (v subjectPatternRuleConditionValidator) Description(context.Context) string {
	return "Validates the subject pattern rule condition."
}

func (v subjectPatternRuleConditionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v subjectPatternRuleConditionValidator) ValidateObject(
	ctx context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var condition struct {
		Variable string `tfsdk:"variable"`
		Operator string `tfsdk:"operator"`
		Value    string `tfsdk:"value"`
	}

	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &condition, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// user_type only supports is/is_not
	if condition.Variable == "user_type" &&
		condition.Operator != "is" &&
		condition.Operator != "is_not" {

		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("operator"),
			"Invalid operator",
			`For variable "user_type", operator must be either "is" or "is_not".`,
		)
	}

	// value required for is/is_not
	if (condition.Operator == "is" || condition.Operator == "is_not") &&
		condition.Value == "" {

		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("value"),
			"Missing value",
			`The "value" attribute is required when operator is "is" or "is_not".`,
		)
	}

	// value must not be set for exist/exist_not
	if (condition.Operator == "exist" || condition.Operator == "exist_not") &&
		condition.Value != "" {

		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("value"),
			"Unexpected value",
			`The "value" attribute must not be set when operator is "exist" or "exist_not".`,
		)
	}

	// value must be either "Business" or "Technical" for user_type
	if condition.Variable == "user_type" &&
		(condition.Operator == "is" || condition.Operator == "is_not") {

		if condition.Value != "Business" && condition.Value != "Technical" {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("value"),
				"Invalid value",
				`For variable "user_type", value must be either "Business" or "Technical".`,
			)
		}
	}
}
