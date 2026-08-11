package systemMapping

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// ---------------------------------------------------------------------------
// helper: build a tfsdk.Config with a single string "protocol" attribute
// ---------------------------------------------------------------------------

func buildProtocolConfig(protocol string) tfsdk.Config {
	s := fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"protocol": fwschema.StringAttribute{},
		},
	}
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"protocol": tftypes.String}},
		map[string]tftypes.Value{
			"protocol": tftypes.NewValue(tftypes.String, protocol),
		},
	)
	return tfsdk.Config{Schema: s, Raw: raw}
}

// helper: build a tfsdk.Config with "protocol" and "backend_type"
func buildBackendConfig(protocol, backendType string) tfsdk.Config {
	s := fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"protocol":     fwschema.StringAttribute{},
			"backend_type": fwschema.StringAttribute{},
		},
	}
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"protocol":     tftypes.String,
			"backend_type": tftypes.String,
		}},
		map[string]tftypes.Value{
			"protocol":     tftypes.NewValue(tftypes.String, protocol),
			"backend_type": tftypes.NewValue(tftypes.String, backendType),
		},
	)
	return tfsdk.Config{Schema: s, Raw: raw}
}

// ---------------------------------------------------------------------------
// ProtocolAuthenticationModeValidator wrapper
// ---------------------------------------------------------------------------

func TestProtocolAuthenticationModeValidator_Description(t *testing.T) {
	v := ProtocolAuthenticationModeValidator{}
	assert := func(s string) {
		if s == "" {
			t.Error("expected non-empty description")
		}
	}
	assert(v.Description(context.Background()))
	assert(v.MarkdownDescription(context.Background()))
}

func TestProtocolAuthenticationModeValidator_NullSkipped(t *testing.T) {
	v := ProtocolAuthenticationModeValidator{}
	req := validator.StringRequest{ConfigValue: types.StringNull()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value")
	}
}

func TestProtocolAuthenticationModeValidator_UnknownSkipped(t *testing.T) {
	v := ProtocolAuthenticationModeValidator{}
	req := validator.StringRequest{ConfigValue: types.StringUnknown()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown value")
	}
}

func TestProtocolAuthenticationModeValidator_ValidCombination(t *testing.T) {
	v := ProtocolAuthenticationModeValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("NONE"),
		Config:      buildProtocolConfig("HTTP"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestProtocolAuthenticationModeValidator_InvalidCombination(t *testing.T) {
	v := ProtocolAuthenticationModeValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("X509_GENERAL"),
		Config:      buildProtocolConfig("HTTP"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid auth mode")
	}
}

func TestValidateAuthenticationMode_Constructor(t *testing.T) {
	v := ValidateAuthenticationMode()
	if v == nil {
		t.Error("expected non-nil validator")
	}
}

// ---------------------------------------------------------------------------
// ProtocolBackendValidator wrapper
// ---------------------------------------------------------------------------

func TestProtocolBackendValidator_Description(t *testing.T) {
	v := ProtocolBackendValidator{}
	if v.Description(context.Background()) == "" {
		t.Error("expected non-empty description")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestProtocolBackendValidator_NullSkipped(t *testing.T) {
	v := ProtocolBackendValidator{}
	req := validator.StringRequest{ConfigValue: types.StringNull()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value")
	}
}

func TestProtocolBackendValidator_UnknownSkipped(t *testing.T) {
	v := ProtocolBackendValidator{}
	req := validator.StringRequest{ConfigValue: types.StringUnknown()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown value")
	}
}

func TestProtocolBackendValidator_ValidCombination(t *testing.T) {
	v := ProtocolBackendValidator{}
	// protocol is the ConfigValue here (note: backend reads backend_type from Config)
	req := validator.StringRequest{
		ConfigValue: types.StringValue("HTTP"),
		Config:      buildBackendConfig("HTTP", "abapSys"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestProtocolBackendValidator_InvalidCombination(t *testing.T) {
	v := ProtocolBackendValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("RFC"),
		Config:      buildBackendConfig("RFC", "hana"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid backend type")
	}
}

func TestValidateProtocolBackend_Constructor(t *testing.T) {
	v := ValidateProtocolBackend()
	if v == nil {
		t.Error("expected non-nil validator")
	}
}

// ---------------------------------------------------------------------------
// PortValidator wrapper
// ---------------------------------------------------------------------------

func TestPortValidator_Description(t *testing.T) {
	v := PortValidator{}
	if v.Description(context.Background()) == "" {
		t.Error("expected non-empty description")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestPortValidator_NullSkipped(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{ConfigValue: types.StringNull()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value")
	}
}

func TestPortValidator_UnknownSkipped(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{ConfigValue: types.StringUnknown()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown value")
	}
}

func TestPortValidator_ValidHTTPPort(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("8080"),
		Config:      buildProtocolConfig("HTTP"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestPortValidator_InvalidHTTPPort(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("99999"),
		Config:      buildProtocolConfig("HTTPS"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for out-of-range port")
	}
}

func TestPortValidator_ValidRFCValue(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("sapgw00"),
		Config:      buildProtocolConfig("RFC"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestPortValidator_InvalidRFCValue(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("bad"),
		Config:      buildProtocolConfig("RFC"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid RFC value")
	}
}

func TestPortValidator_ValidRFCSValue(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("sapgw00s"),
		Config:      buildProtocolConfig("RFCS"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestPortValidator_InvalidRFCSValue(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("bad"),
		Config:      buildProtocolConfig("RFCS"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid RFCS value")
	}
}

func TestValidatePort_Constructor(t *testing.T) {
	v := ValidatePort()
	if v == nil {
		t.Error("expected non-nil validator")
	}
}

// ---------------------------------------------------------------------------
// ProtocolValidatorCore wrapper
// ---------------------------------------------------------------------------

func TestProtocolValidatorCore_Description(t *testing.T) {
	v := ProtocolValidatorCore{AllowedProtocols: []string{"HTTP", "HTTPS"}}
	if v.Description(context.Background()) == "" {
		t.Error("expected non-empty description")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestProtocolValidatorCore_NullSkipped(t *testing.T) {
	v := ProtocolValidatorCore{AllowedProtocols: []string{"HTTP"}}
	req := validator.StringRequest{ConfigValue: types.StringNull()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value")
	}
}

func TestProtocolValidatorCore_UnknownSkipped(t *testing.T) {
	v := ProtocolValidatorCore{AllowedProtocols: []string{"HTTP"}}
	req := validator.StringRequest{ConfigValue: types.StringUnknown()}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown value")
	}
}

func TestProtocolValidatorCore_ValidProtocol(t *testing.T) {
	v := ProtocolValidatorCore{AllowedProtocols: []string{"HTTP", "HTTPS"}}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("someValue"),
		Config:      buildProtocolConfig("HTTP"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestProtocolValidatorCore_InvalidProtocol(t *testing.T) {
	v := ProtocolValidatorCore{AllowedProtocols: []string{"HTTP", "HTTPS"}}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("someValue"),
		Config:      buildProtocolConfig("RFC"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for disallowed protocol")
	}
}

func TestValidateProtocolString_Constructor(t *testing.T) {
	v := ValidateProtocolString([]string{"HTTP"})
	if v == nil {
		t.Error("expected non-nil validator")
	}
}

// ---------------------------------------------------------------------------
// ProtocolListValidator wrapper
// ---------------------------------------------------------------------------

func buildListProtocolConfig(protocol string) tfsdk.Config {
	s := fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"protocol": fwschema.StringAttribute{},
			"items":    fwschema.ListAttribute{ElementType: types.StringType},
		},
	}
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"protocol": tftypes.String,
			"items":    tftypes.List{ElementType: tftypes.String},
		}},
		map[string]tftypes.Value{
			"protocol": tftypes.NewValue(tftypes.String, protocol),
			"items":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
		},
	)
	return tfsdk.Config{Schema: s, Raw: raw}
}

func TestProtocolListValidator_Description(t *testing.T) {
	v := ProtocolListValidator{AllowedProtocols: []string{"HTTP"}}
	if v.Description(context.Background()) == "" {
		t.Error("expected non-empty description")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestProtocolListValidator_NullSkipped(t *testing.T) {
	v := ProtocolListValidator{AllowedProtocols: []string{"HTTP"}}
	req := validator.ListRequest{
		ConfigValue: types.ListNull(types.StringType),
	}
	resp := &validator.ListResponse{}
	v.ValidateList(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null list")
	}
}

func TestProtocolListValidator_UnknownSkipped(t *testing.T) {
	v := ProtocolListValidator{AllowedProtocols: []string{"HTTP"}}
	req := validator.ListRequest{
		ConfigValue: types.ListUnknown(types.StringType),
	}
	resp := &validator.ListResponse{}
	v.ValidateList(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for unknown list")
	}
}

func TestProtocolListValidator_ValidProtocol(t *testing.T) {
	v := ProtocolListValidator{AllowedProtocols: []string{"HTTP", "HTTPS"}}
	listVal, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("val")})
	req := validator.ListRequest{
		ConfigValue: listVal,
		Config:      buildListProtocolConfig("HTTP"),
	}
	resp := &validator.ListResponse{}
	v.ValidateList(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}

func TestProtocolListValidator_InvalidProtocol(t *testing.T) {
	v := ProtocolListValidator{AllowedProtocols: []string{"HTTP", "HTTPS"}}
	listVal, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("val")})
	req := validator.ListRequest{
		ConfigValue: listVal,
		Config:      buildListProtocolConfig("RFC"),
	}
	resp := &validator.ListResponse{}
	v.ValidateList(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for disallowed protocol")
	}
}

func TestValidateProtocolList_Constructor(t *testing.T) {
	v := ValidateProtocolList([]string{"HTTP"})
	if v == nil {
		t.Error("expected non-nil validator")
	}
}

// ---------------------------------------------------------------------------
// GetAttribute error branches — pass an empty-schema config so GetAttribute
// fails, covering the early-return path in each ValidateString/ValidateList
// ---------------------------------------------------------------------------

func emptyConfig() tfsdk.Config {
	s := fwschema.Schema{Attributes: map[string]fwschema.Attribute{}}
	raw := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{})
	return tfsdk.Config{Schema: s, Raw: raw}
}

func TestProtocolAuthenticationModeValidator_GetAttributeError(t *testing.T) {
	v := ProtocolAuthenticationModeValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("NONE"),
		Config:      emptyConfig(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error when protocol attribute is missing from config")
	}
}

func TestProtocolBackendValidator_GetAttributeError(t *testing.T) {
	v := ProtocolBackendValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("HTTP"),
		Config:      emptyConfig(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error when backend_type attribute is missing from config")
	}
}

func TestPortValidator_GetAttributeError(t *testing.T) {
	v := PortValidator{}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("8080"),
		Config:      emptyConfig(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error when protocol attribute is missing from config")
	}
}

func TestProtocolValidatorCore_GetAttributeError(t *testing.T) {
	v := ProtocolValidatorCore{AllowedProtocols: []string{"HTTP"}}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("val"),
		Config:      emptyConfig(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error when protocol attribute is missing from config")
	}
}

func TestProtocolListValidator_GetAttributeError(t *testing.T) {
	v := ProtocolListValidator{AllowedProtocols: []string{"HTTP"}}
	listVal, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("val")})
	req := validator.ListRequest{
		ConfigValue: listVal,
		Config:      emptyConfig(),
	}
	resp := &validator.ListResponse{}
	v.ValidateList(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error when protocol attribute is missing from config")
	}
}
