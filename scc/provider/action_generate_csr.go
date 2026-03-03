package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type GenerateCSRAction struct {
	client *api.RestApiClient
}

var _ action.Action = &GenerateCSRAction{}

func NewGenerateCSRAction() action.Action {
	return &GenerateCSRAction{}
}

func (a *GenerateCSRAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generate_csr"
}

func (a *GenerateCSRAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a Certificate Signing Request (CSR) based on the type of Certificate.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of Certificate for which to generate the CSR.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ca", "system", "ui"),
				},
			},
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "The size of the key to generate for the CSR.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(2048, 4096),
				},
			},
			"subject_dn": schema.SingleNestedAttribute{
				MarkdownDescription: "The subject distinguished name (DN) for the CSR.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cn": schema.StringAttribute{
						MarkdownDescription: "Common Name (CN) of the certificate, typically representing the domain name or identifier for which the certificate is issued.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"CN must not contain ',', '=', or '\\'",
							),
						},
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Email address associated with the certificate subject.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"EMAIL must not contain ',', '=', or '\\'",
							),
						},
					},
					"l": schema.StringAttribute{
						MarkdownDescription: "Locality (L) of the certificate subject, such as a city or town.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"L must not contain ',', '=', or '\\'",
							),
						},
					},
					"ou": schema.StringAttribute{
						MarkdownDescription: "Organizational Unit (OU) of the certificate subject, representing a department or division within an organization.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"OU must not contain ',', '=', or '\\'",
							),
						},
					},
					"o": schema.StringAttribute{
						MarkdownDescription: "Organization (O) of the certificate subject, representing the name of the organization.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"O must not contain ',', '=', or '\\'",
							),
						},
					},
					"st": schema.StringAttribute{
						MarkdownDescription: "State or Province (ST) of the certificate subject.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"ST must not contain ',', '=', or '\\'",
							),
						},
					},
					"c": schema.StringAttribute{
						MarkdownDescription: "Country (C) of the certificate subject, typically represented as a two-letter ISO country code.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(2, 2),
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"C must not contain ',', '=', or '\\'",
							),
						},
					},
				},
			},
			"subject_alternative_names": schema.ListNestedAttribute{
				MarkdownDescription: "Subject Alternative Names (SANs) for the certificate, allowing additional identities to be associated with the certificate beyond the Common Name (CN).",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of SAN, such as DNS, IP, RFC822 or URI.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("DNS", "IP", "RFC822", "URI"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The value of the SAN, such as a domain name for DNS, an IP address for IP, an email address for RFC822, or a URI for URI.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

func (a *GenerateCSRAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.client = client
}

func safeProgress(resp *action.InvokeResponse, msg string) {
	if resp == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = r
		}
	}()

	resp.SendProgress(action.InvokeProgressEvent{Message: msg})
}

func (a *GenerateCSRAction) InvokeWithPlan(ctx context.Context, plan CSRActionConfig, resp *action.InvokeResponse) {
	if plan.SubjectDN.IsNull() || plan.SubjectDN.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Subject DN",
			"Subject DN with a non-empty Common Name (CN) is required to create a self-signed certificate.",
		)
		return
	}

	dnStruct, diags := ExpandSubjectDN(ctx, plan.SubjectDN)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subjectDN := BuildSubjectDN(dnStruct)
	planBody := map[string]any{
		"type":      "csr",
		"keySize":   plan.KeySize.ValueInt64(),
		"subjectDN": subjectDN,
	}

	if !plan.SubjectAlternativeNames.IsNull() &&
		!plan.SubjectAlternativeNames.IsUnknown() {
		var sanList []SubjectAlternativeNames
		diags = plan.SubjectAlternativeNames.ElementsAs(ctx, &sanList, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(sanList) > 0 {
			sanFields := []map[string]string{}
			for _, san := range sanList {
				sanFields = append(sanFields, map[string]string{
					"type":  san.Type.ValueString(),
					"value": san.Value.ValueString(),
				})
			}

			planBody["subjectAlternativeNames"] = sanFields
		}
	}

	certType := strings.ToLower(plan.Type.ValueString())
	var endpoint string

	switch certType {
	case "ca":
		endpoint = endpoints.GetCACertificateEndpoint()
	case "system":
		endpoint = endpoints.GetSystemCertificateEndpoint()
	case "ui":
		endpoint = endpoints.GetUICertificateEndpoint()
	default:
		resp.Diagnostics.AddError(
			"Invalid Certificate Type",
			"Certificate type must be one of 'ca', 'system', or 'ui'.",
		)
		return
	}

	// Create CSR by calling the appropriate API endpoint based on the certificate type
	csrResponse, diags := sendRequestFunc(a.client, planBody, endpoint, actionCreateRequest)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if csrResponse == nil || csrResponse.Body == nil {
		resp.Diagnostics.AddError("Invalid API Response", "CSR response body is nil")
		return
	}

	csrBytes, err := io.ReadAll(csrResponse.Body)
	defer func() {
		if err := csrResponse.Body.Close(); err != nil {
			// log but don’t fail action
			resp.Diagnostics.AddWarning(
				"Failed to close response body",
				err.Error(),
			)
		}
	}()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read CSR Response",
			fmt.Sprintf("An error occurred while reading the CSR response body: %v", err),
		)
		return
	}

	csr := strings.TrimSpace(string(csrBytes))
	if csr == "" {
		resp.Diagnostics.AddError(
			"Empty CSR Response",
			"The API response did not contain a valid CSR.",
		)
		return
	}

	filePath := filepath.Join(".", certType+"_csr.pem")

	err = os.WriteFile(filePath, []byte(csr), 0644)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Write CSR to File",
			fmt.Sprintf("An error occurred while writing the CSR to file: %v", err),
		)
		return
	}

	safeProgress(resp, "CSR generated successfully")
	safeProgress(resp, fmt.Sprintf("CSR saved to %s", filePath))
}

func (a *GenerateCSRAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var plan CSRActionConfig
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	a.InvokeWithPlan(ctx, plan, resp)
}
