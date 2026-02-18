package systemMapping

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validateListProtocol(protocol string, allowed []string) diag.Diagnostics {
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

// --- List validator ---
type ProtocolListValidator struct {
	AllowedProtocols []string
}

func (v ProtocolListValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Only valid when protocol is one of: %v", v.AllowedProtocols)
}

func (v ProtocolListValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ProtocolListValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var protocol string
	diags := req.Config.GetAttribute(ctx, path.Root("protocol"), &protocol)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Append the diagnostics from the core function
	resp.Diagnostics.Append(validateListProtocol(protocol, v.AllowedProtocols)...)
}

// --- Constructor ---
func ValidateProtocolList(allowed []string) validator.List {
	return ProtocolListValidator{AllowedProtocols: allowed}
}
