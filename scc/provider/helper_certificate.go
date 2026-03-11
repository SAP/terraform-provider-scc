package provider

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Wrappers for testing purposes (allows mocking in tests)
var getCertificateBinaryFunc = getCertificateBinary
var validatePEMChainFunc = validatePEMChain
var uploadSignedChainFunc = uploadSignedChain

type CertificateConfig struct {
	SubjectDN      types.Object `tfsdk:"subject_dn"`
	Issuer         types.String `tfsdk:"issuer"`
	ValidFrom      types.String `tfsdk:"valid_from"`
	ValidTo        types.String `tfsdk:"valid_to"`
	SerialNumber   types.String `tfsdk:"serial_number"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
}

type CertificateWithSANConfig struct {
	CertificateConfig
	SubjectAltNames types.List `tfsdk:"subject_alternative_names"`
}

type certificateSubjectDNConfig struct {
	CommonName         types.String `tfsdk:"cn"`
	Email              types.String `tfsdk:"email"`
	Locality           types.String `tfsdk:"l"`
	OrganizationalUnit types.String `tfsdk:"ou"`
	Organization       types.String `tfsdk:"o"`
	State              types.String `tfsdk:"st"`
	Country            types.String `tfsdk:"c"`
}

var subjectDNAttrTypes = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"cn":    types.StringType,
		"email": types.StringType,
		"l":     types.StringType,
		"ou":    types.StringType,
		"o":     types.StringType,
		"st":    types.StringType,
		"c":     types.StringType,
	},
}

type subjectAlternativeNames struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

var subjectAlternativeNamesType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	},
}

func expandSubjectDN(ctx context.Context, subjectDN types.Object) (*certificateSubjectDNConfig, diag.Diagnostics) {
	if subjectDN.IsNull() || subjectDN.IsUnknown() {
		return nil, diag.Diagnostics{}
	}

	var result certificateSubjectDNConfig
	diags := subjectDN.As(ctx, &result, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &result, diags
}

func buildSubjectDN(subjectDN *certificateSubjectDNConfig) string {
	if subjectDN == nil || subjectDN.CommonName.IsNull() {
		return ""
	}

	clean := func(value string) string {
		return strings.TrimSpace(value)
	}
	parts := []string{
		fmt.Sprintf("CN=%s", clean(subjectDN.CommonName.ValueString())),
	}

	if !subjectDN.Email.IsNull() && strings.TrimSpace(subjectDN.Email.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("EMAIL=%s", clean(subjectDN.Email.ValueString())))
	}
	if !subjectDN.Locality.IsNull() && strings.TrimSpace(subjectDN.Locality.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("L=%s", clean(subjectDN.Locality.ValueString())))
	}
	if !subjectDN.OrganizationalUnit.IsNull() && strings.TrimSpace(subjectDN.OrganizationalUnit.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("OU=%s", clean(subjectDN.OrganizationalUnit.ValueString())))
	}
	if !subjectDN.Organization.IsNull() && strings.TrimSpace(subjectDN.Organization.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("O=%s", clean(subjectDN.Organization.ValueString())))
	}
	if !subjectDN.State.IsNull() && strings.TrimSpace(subjectDN.State.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("ST=%s", clean(subjectDN.State.ValueString())))
	}
	if !subjectDN.Country.IsNull() && strings.TrimSpace(subjectDN.Country.ValueString()) != "" {
		parts = append(parts, fmt.Sprintf("C=%s", clean(subjectDN.Country.ValueString())))
	}
	return strings.Join(parts, ",")
}

func parseSubjectDN(dn string) *certificateSubjectDNConfig {
	result := &certificateSubjectDNConfig{
		CommonName:         types.StringNull(),
		Email:              types.StringNull(),
		Locality:           types.StringNull(),
		OrganizationalUnit: types.StringNull(),
		Organization:       types.StringNull(),
		State:              types.StringNull(),
		Country:            types.StringNull(),
	}

	for part := range strings.SplitSeq(dn, ",") {
		part = strings.TrimSpace(part)

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.ToUpper(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])

		switch key {
		case "CN":
			result.CommonName = types.StringValue(val)
		case "EMAIL":
			result.Email = types.StringValue(val)
		case "L":
			result.Locality = types.StringValue(val)
		case "OU":
			result.OrganizationalUnit = types.StringValue(val)
		case "O":
			result.Organization = types.StringValue(val)
		case "ST":
			result.State = types.StringValue(val)
		case "C":
			result.Country = types.StringValue(val)
		}
	}

	return result
}

func buildSubjectDNObject(dn *certificateSubjectDNConfig) types.Object {
	if dn == nil {
		return types.ObjectNull(subjectDNAttrTypes.AttrTypes)
	}

	attrs := map[string]attr.Value{
		"cn":    dn.CommonName,
		"email": dn.Email,
		"l":     dn.Locality,
		"ou":    dn.OrganizationalUnit,
		"o":     dn.Organization,
		"st":    dn.State,
		"c":     dn.Country,
	}

	obj, diags := types.ObjectValue(subjectDNAttrTypes.AttrTypes, attrs)
	if diags.HasError() {
		panic("failed to build subject_dn object")
	}
	return obj
}

func uploadSignedChain(c *api.RestApiClient, endpoint, cert string) diag.Diagnostics {
	var diags diag.Diagnostics
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("signedCertificate", "signed_chain.pem")
	if err != nil {
		diags.AddError(
			"Failed to Create Multipart Form",
			fmt.Sprintf("error creating multipart form: %v", err),
		)
		return diags
	}

	_, err = part.Write([]byte(cert))
	if err != nil {
		diags.AddError(
			"Failed to Write Certificate to Multipart Form",
			fmt.Sprintf("error writing certificate to multipart form: %v", err),
		)
		return diags
	}

	if err := writer.Close(); err != nil {
		diags.AddError(
			"Failed to Finalize Multipart Form",
			fmt.Sprintf("error closing multipart writer: %v", err),
		)
		return diags
	}

	resp, diags := c.DoRequest(http.MethodPatch, endpoint, body.Bytes(), "", writer.FormDataContentType())
	if diags.HasError() {
		return diags
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			diags.AddWarning(
				"Response Body Close Failed",
				fmt.Sprintf("error closing response body: %v", cerr),
			)
		}
	}()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		diags.AddError(
			"Failed to Upload Signed Certificate Chain",
			fmt.Sprintf("status code: %d, response: %s", resp.StatusCode, string(bodyBytes)),
		)
	}

	return diags
}

func uploadPKCS12Certificate(c *api.RestApiClient, endpoint string, pkcs12Bytes []byte, password, keyPassword string) diag.Diagnostics {
	var diags diag.Diagnostics
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("password", password); err != nil {
		diags.AddError(
			"Failed to Write Password Field",
			fmt.Sprintf("error writing password field: %v", err),
		)
		return diags
	}

	if keyPassword != "" {
		if err := writer.WriteField("keyPassword", keyPassword); err != nil {
			diags.AddError(
				"Failed to Write Key Password Field",
				fmt.Sprintf("error writing key password field: %v", err),
			)
			return diags
		}
	}

	part, err := writer.CreateFormFile("pkcs12", "certificate.p12")
	if err != nil {
		diags.AddError(
			"Failed to Create Multipart Form",
			fmt.Sprintf("error creating multipart form: %v", err),
		)
		return diags
	}

	_, err = part.Write(pkcs12Bytes)
	if err != nil {
		diags.AddError(
			"Failed to Write Certificate to Multipart Form",
			fmt.Sprintf("error writing certificate to multipart form: %v", err),
		)
		return diags
	}

	if err := writer.Close(); err != nil {
		diags.AddError(
			"Failed to Finalize Multipart Form",
			fmt.Sprintf("error closing multipart writer: %v", err),
		)
		return diags
	}

	resp, diags := c.DoRequest(http.MethodPut, endpoint, body.Bytes(), "", writer.FormDataContentType())
	if diags.HasError() {
		return diags
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			diags.AddWarning(
				"Response Body Close Failed",
				fmt.Sprintf("error closing response body: %v", cerr),
			)
		}
	}()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		diags.AddError(
			"Failed to Upload PKCS#12 Certificate",
			fmt.Sprintf("status code: %d, response: %s", resp.StatusCode, string(bodyBytes)),
		)

		return diags
	}

	return diags
}

func getCertificateBinary(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
	response, diags := client.DoRequest(http.MethodGet, endpoint, nil, "application/pkix-cert", "")
	if diags.HasError() {
		return nil, diags
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			diags.AddWarning(
				"Failed to Close Response Body",
				fmt.Sprintf("error closing response body: %v", err),
			)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Failed to Read Response Body", fmt.Sprintf("failed to read response body: %v", err))
		return nil, diags
	}
	return body, diags
}

func validatePEMData(data string) diag.Diagnostics {
	var diags diag.Diagnostics

	if strings.TrimSpace(data) == "" {
		diags.AddError(
			"Empty PEM Data",
			"No certificate data provided.",
		)
		return diags
	}

	block, _ := pem.Decode([]byte(data))
	if block == nil {
		diags.AddError(
			"Invalid PEM Block",
			"Failed to decode PEM block. Ensure certificate is valid PEM format.",
		)
		return diags
	}

	// Only check supported types
	switch block.Type {
	case "CERTIFICATE",
		"PRIVATE KEY",
		"RSA PRIVATE KEY",
		"EC PRIVATE KEY":
		return diags
	default:
		diags.AddError(
			"Unsupported PEM Type",
			fmt.Sprintf("Unsupported PEM block type: %s", block.Type),
		)
	}

	return diags
}

func validatePEMChain(data string) diag.Diagnostics {
	var diags diag.Diagnostics

	data = strings.TrimSpace(data)
	if data == "" {
		diags.AddError(
			"Empty Certificate Chain Data",
			"No certificate chain provided.",
		)
		return diags
	}

	result := []byte(data)
	certCount := 0

	for {
		var block *pem.Block
		block, result = pem.Decode(result)

		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			diags.AddError(
				"Invalid PEM Block in Chain",
				fmt.Sprintf("Failed to decode PEM block. Ensure certificate chain is valid PEM format. Unsupported block type: %s", block.Type),
			)
			return diags
		}

		_, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			diags.AddError(
				"Invalid Certificate in Chain",
				fmt.Sprintf("Failed to parse certificate in chain: %v", err),
			)
			return diags
		}

		certCount++
	}

	if certCount == 0 {
		diags.AddError(
			"No Valid Certificates Found",
			"Failed to find any valid certificates in the provided chain.",
		)
		return diags
	}

	return diags
}

func validatePKCS12Inputs(plan PKCS12SystemCertificateResourceConfig) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	rawCertificate := []byte(plan.PKCS12Certificate.ValueString())
	if decoded, err := base64.StdEncoding.DecodeString(plan.PKCS12Certificate.ValueString()); err == nil {
		rawCertificate = decoded
	}

	if len(rawCertificate) == 0 {
		diags.AddError(
			"Invalid PKCS#12 Certificate",
			"Provided PKCS#12 certificate is empty after decoding.",
		)
		return nil, diags
	}

	if !plan.KeyPassword.IsNull() &&
		!plan.KeyPassword.IsUnknown() &&
		plan.KeyPassword.ValueString() == "" {

		diags.AddError(
			"Invalid Key Password",
			"If key_password is set, it must not be empty.",
		)
		return nil, diags
	}

	return rawCertificate, diags
}

// buildCertificateModelWithSAN maps API certificate data to the Terraform model
// for CA/UI certificates, including Subject Alternative Names (SAN).
func buildCertificateModelWithSAN(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (CertificateWithSANConfig, diag.Diagnostics) {
	subjectAltNamesValue := []subjectAlternativeNames{}
	for _, san := range value.SubjectAltNames {
		subjectAltNamesValue = append(subjectAltNamesValue, subjectAlternativeNames{
			Type:  types.StringValue(san.Type),
			Value: types.StringValue(san.Value),
		})
	}

	var subjectAltNames types.List

	if len(subjectAltNamesValue) == 0 {
		subjectAltNames = types.ListNull(subjectAlternativeNamesType)
	} else {
		var diags diag.Diagnostics
		subjectAltNames, diags = types.ListValueFrom(ctx, subjectAlternativeNamesType, subjectAltNamesValue)
		if diags.HasError() {
			return CertificateWithSANConfig{}, diags
		}
	}

	certificateConfig, diags := buildCertificateModel(ctx, value, pemBytes)
	if diags.HasError() {
		return CertificateWithSANConfig{}, diags
	}

	model := &CertificateWithSANConfig{
		CertificateConfig: certificateConfig,
		SubjectAltNames:   subjectAltNames,
	}

	return *model, diag.Diagnostics{}
}

// buildCertificateModel maps API certificate data to the Terraform
// model for system certificates. SAN values are ignored because
// system certificates do not support Subject Alternative Names.
func buildCertificateModel(ctx context.Context, value apiobjects.Certificate, pemBytes []byte) (CertificateConfig, diag.Diagnostics) {
	model := &CertificateConfig{
		ValidTo:        ConvertMillisToTimes(value.NotAfterTimeStamp).WithTimezone,
		ValidFrom:      ConvertMillisToTimes(value.NotBeforeTimeStamp).WithTimezone,
		Issuer:         types.StringValue(value.Issuer),
		SerialNumber:   types.StringValue(value.SerialNumber),
		CertificatePEM: types.StringValue(string(pemBytes)),
		SubjectDN:      types.ObjectNull(subjectDNAttrTypes.AttrTypes),
	}

	if value.SubjectDN != "" {
		dn := parseSubjectDN(value.SubjectDN)
		model.SubjectDN = buildSubjectDNObject(dn)
	}

	return *model, diag.Diagnostics{}

}

func buildSignedChainPlan(
	ctx context.Context,
	r resource.Resource,
	chain string,
	includeSAN bool,
) tfsdk.Plan {

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	subjectDNType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"cn":    tftypes.String,
			"email": tftypes.String,
			"l":     tftypes.String,
			"ou":    tftypes.String,
			"o":     tftypes.String,
			"st":    tftypes.String,
			"c":     tftypes.String,
		},
	}

	attrTypes := map[string]tftypes.Type{
		"signed_chain":    tftypes.String,
		"subject_dn":      subjectDNType,
		"valid_to":        tftypes.String,
		"valid_from":      tftypes.String,
		"issuer":          tftypes.String,
		"serial_number":   tftypes.String,
		"certificate_pem": tftypes.String,
	}

	values := map[string]tftypes.Value{
		"signed_chain":    tftypes.NewValue(tftypes.String, chain),
		"subject_dn":      tftypes.NewValue(subjectDNType, nil),
		"valid_to":        tftypes.NewValue(tftypes.String, nil),
		"valid_from":      tftypes.NewValue(tftypes.String, nil),
		"issuer":          tftypes.NewValue(tftypes.String, nil),
		"serial_number":   tftypes.NewValue(tftypes.String, nil),
		"certificate_pem": tftypes.NewValue(tftypes.String, nil),
	}

	if includeSAN {
		sanType := tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"type":  tftypes.String,
				"value": tftypes.String,
			},
		}

		attrTypes["subject_alternative_names"] = tftypes.List{
			ElementType: sanType,
		}

		values["subject_alternative_names"] = tftypes.NewValue(
			tftypes.List{ElementType: sanType},
			nil,
		)
	}

	return tfsdk.Plan{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			values,
		),
	}
}
