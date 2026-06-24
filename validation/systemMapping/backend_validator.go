package systemMapping

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validateProtocolBackend(protocol, backendType string) diag.Diagnostics {
	var diags diag.Diagnostics

	allowed := map[string][]string{
		"HTTP": {
			"abapSys",
			"hana",
			"applServerJava",
			"netweaverCE",
			"BC",
			"PI",
			"netweaverGW",
			"otherSAPsys",
			"nonSAPsys",
		},
		"HTTPS": {
			"abapSys",
			"hana",
			"applServerJava",
			"netweaverCE",
			"BC",
			"PI",
			"netweaverGW",
			"otherSAPsys",
			"nonSAPsys",
		},
		"RFC": {
			"abapSys",
			"netweaverGW",
		},
		"RFCS": {
			"abapSys",
			"netweaverGW",
		},
		"RFCWS": {
			"abapSys",
			"netweaverGW",
		},
		"LDAP": {
			"nonSAPsys",
		},
		"LDAPS": {
			"nonSAPsys",
		},
		"TCP": {
			"abapSys",
			"hana",
			"netweaverGW",
			"otherSAPsys",
			"nonSAPsys",
		},
		"TCPS": {
			"abapSys",
			"hana",
			"netweaverGW",
			"otherSAPsys",
			"nonSAPsys",
		},
	}

	validBackends, ok := allowed[protocol]
	if !ok {
		return diags
	}

	if !slices.Contains(validBackends, backendType) {
		diags.AddAttributeError(
			path.Empty(),
			"Invalid backend_type for protocol",
			fmt.Sprintf(
				"backend_type %q is not valid for protocol %q. Allowed values: %v",
				backendType,
				protocol,
				validBackends,
			),
		)
	}

	return diags
}

// Wrapper
type ProtocolBackendValidator struct{}

func (v ProtocolBackendValidator) Description(_ context.Context) string {
	return "Ensures the selected protocol is supported by the chosen backend_type."
}
func (v ProtocolBackendValidator) MarkdownDescription(_ context.Context) string {
	return "Validates that the selected **protocol** is supported by the configured **backend_type**."
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
