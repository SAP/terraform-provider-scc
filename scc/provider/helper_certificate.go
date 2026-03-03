package provider

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type CertificateSubjectDNConfig struct {
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

func ExpandSubjectDN(ctx context.Context, subjectDN types.Object) (*CertificateSubjectDNConfig, diag.Diagnostics) {
	if subjectDN.IsNull() || subjectDN.IsUnknown() {
		return nil, diag.Diagnostics{}
	}

	var result CertificateSubjectDNConfig
	diags := subjectDN.As(ctx, &result, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &result, diags
}

func BuildSubjectDN(subjectDN *CertificateSubjectDNConfig) string {
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

func parseSubjectDN(dn string) *CertificateSubjectDNConfig {
	result := &CertificateSubjectDNConfig{
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

func GetCertificateBinary(client *api.RestApiClient, endpoint string) ([]byte, diag.Diagnostics) {
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
