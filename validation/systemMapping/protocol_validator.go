package systemMapping

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validateStringProtocol(protocol string, allowed []string) diag.Diagnostics {
	var diags diag.Diagnostics

	if slices.Contains(allowed, protocol) {
		return diags // valid
	}

	diags.AddAttributeError(
		path.Empty(),
		"Attribute Not Valid for Protocol",
		fmt.Sprintf("Protocol %q is not allowed. Allowed: %v", protocol, allowed),
	)
	return diags
}

// --- Wrapper ---
type ProtocolValidatorCore struct {
	AllowedProtocols []string
}

func (v ProtocolValidatorCore) Description(_ context.Context) string {
	return fmt.Sprintf("Only valid when protocol is one of: %v", v.AllowedProtocols)
}
func (v ProtocolValidatorCore) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ProtocolValidatorCore) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var protocol string
	diags := req.Config.GetAttribute(ctx, path.Root("protocol"), &protocol)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateStringProtocol(protocol, v.AllowedProtocols)...)
}

func ValidateProtocolString(allowed []string) validator.String {
	return ProtocolValidatorCore{AllowedProtocols: allowed}
}
