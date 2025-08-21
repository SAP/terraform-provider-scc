package systemMapping

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validateProtocolBackend(protocol, backendType string) diag.Diagnostics {
	var diags diag.Diagnostics
	if (protocol == "LDAP" || protocol == "LDAPS") && backendType != "nonSAPsys" {
		diags.AddAttributeError(
			path.Empty(),
			"Invalid protocol for backend_type",
			fmt.Sprintf("Protocol %q is only valid when backend_type is \"nonSAPsys\" (got %q).", protocol, backendType),
		)
	}
	return diags
}

// Wrapper
type ProtocolBackendValidator struct{}

func (v ProtocolBackendValidator) Description(_ context.Context) string {
	return "Ensures LDAP/LDAPS protocols are only allowed when backend_type is nonSAPsys"
}
func (v ProtocolBackendValidator) MarkdownDescription(_ context.Context) string {
	return "Ensures **protocol** = `LDAP`/`LDAPS` is only allowed when **backend_type** = `nonSAPsys`."
}
func (v ProtocolBackendValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var backendType string
	diags := req.Config.GetAttribute(ctx, path.Root("backend_type"), &backendType)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateProtocolBackend(req.ConfigValue.ValueString(), backendType)...)
}

func ValidateProtocolBackend() validator.String {
	return ProtocolBackendValidator{}
}
