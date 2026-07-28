package subjectpatternrules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type subjectPatternValidator struct{}

func SubjectPatternAtLeastOneField() validator.Object {
	return subjectPatternValidator{}
}

func (v subjectPatternValidator) Description(context.Context) string {
	return "Validates that at least one subject pattern field is set."
}

func (v subjectPatternValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v subjectPatternValidator) ValidateObject(
	ctx context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var sp struct {
		CN    string `tfsdk:"cn"`
		Email string `tfsdk:"email"`
		L     string `tfsdk:"l"`
		OU    string `tfsdk:"ou"`
		O     string `tfsdk:"o"`
		ST    string `tfsdk:"st"`
		C     string `tfsdk:"c"`
	}

	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &sp, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if sp.CN == "" && sp.Email == "" && sp.L == "" && sp.OU == "" && sp.O == "" && sp.ST == "" && sp.C == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Empty subject pattern",
			`At least one field must be set in "subject_pattern" (cn, email, l, ou, o, st, or c).`,
		)
	}
}
