package model

import (
	"context"
	"fmt"
	"strconv"

	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProxySettingsDataSourceConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

type ProxySettingsResourceConfig struct {
	// INPUT
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	// OUTPUT
	ID types.String `tfsdk:"id"` // The ID of the proxy settings resource. Used for import and identity purposes. The value is always `proxy-settings`.
}

func ProxySettingsDataSourceValueFrom(ctx context.Context, value apiobjects.ProxySettings) (ProxySettingsDataSourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var portValue types.Int64

	if value.Port == "" {
		portValue = types.Int64Null()
	} else {
		port, err := strconv.ParseInt(value.Port, 10, 64)
		if err != nil {
			diags.AddError(
				"Invalid proxy port",
				fmt.Sprintf("Unable to parse proxy port %q: %s", value.Port, err),
			)
			return ProxySettingsDataSourceConfig{}, diags
		}
		portValue = types.Int64Value(port)
	}

	model := &ProxySettingsDataSourceConfig{
		Host:     valueOrNullString(value.Host),
		Port:     portValue,
		User:     valueOrNullString(value.User),
		Password: valueOrNullString(value.Password), // SCC returns a masked password value (e.g. "xxx")
	}

	return *model, diags
}

func valueOrNullString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func ProxySettingsResourceValueFrom(ctx context.Context, plan ProxySettingsResourceConfig, value apiobjects.ProxySettings) (ProxySettingsResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var portValue types.Int64
	var passwordValue types.String

	if value.Port == "" {
		portValue = types.Int64Null()
	} else {
		port, err := strconv.ParseInt(value.Port, 10, 64)
		if err != nil {
			diags.AddError(
				"Invalid proxy port",
				fmt.Sprintf("Unable to parse proxy port %q: %s", value.Port, err),
			)
			return ProxySettingsResourceConfig{}, diags
		}
		portValue = types.Int64Value(port)
	}

	const maskedPassword = "***"
	if value.Password == "" {
		passwordValue = types.StringNull()
	} else if value.Password == maskedPassword && !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		// API masks the real value; preserve what the caller already knows.
		passwordValue = plan.Password
	} else if value.Password == maskedPassword {
		// Import case: API indicates a password is configured but the real value is
		// unrecoverable. Store the masked sentinel so state reflects that a password
		// exists rather than falsely indicating none is set.
		passwordValue = types.StringValue(maskedPassword)
	} else {
		passwordValue = types.StringValue(value.Password)
	}

	model := &ProxySettingsResourceConfig{
		ID:       types.StringValue("proxy-settings"),
		Host:     valueOrNullString(value.Host),
		Port:     portValue,
		User:     valueOrNullString(value.User),
		Password: passwordValue,
	}

	return *model, diags
}
