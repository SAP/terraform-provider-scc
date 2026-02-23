package provider

import (
	"context"
	"encoding/pem"
	"fmt"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = &SystemCertificateSelfSignedResource{}

func NewSystemCertificateSelfSignedResource() resource.Resource {
	return &SystemCertificateSelfSignedResource{}
}

type SystemCertificateSelfSignedResource struct {
	client *api.RestApiClient
}

func (r *SystemCertificateSelfSignedResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_certificate_self_signed"
}

func (r *SystemCertificateSelfSignedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **Self-Signed System Certificate** in SAP Cloud Connector.
		
**Supports:**
• Self-signed certificates

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-certificate-apis#create-a-self-signed-system-certificate-(master-only)>`,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "Certificate type. Allowed values: `selfsigned`",
				Computed:            true,
				Default:             stringdefault.StaticString("selfsigned"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits. Allowed values: 2048 or 4096.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(2048, 4096),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"subject_dn": schema.SingleNestedAttribute{
				MarkdownDescription: "Subject Distinguished Name (DN) of the certificate. The Common Name (CN) is mandatory, while other fields like L, OU, O, ST, C, or Email may be present depending on the issuing CA.",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
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
							stringvalidator.LengthBetween(2,2),
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^[^,=\\]+$`),
								"C must not contain ',', '=', or '\\'",
							),
						},
					},
				},
			},
			"valid_to": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the end of the validity period.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"valid_from": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the beginning of the validity period.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "Certificate authority (CA) that issued this certificate.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"serial_number": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the certificate, typically assigned by the CA.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_pem": schema.StringAttribute{
				MarkdownDescription: "System certificate in PEM format.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SystemCertificateSelfSignedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *SystemCertificateSelfSignedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemCertificateSelfSignedResourceConfig
	var respObj apiobjects.SystemCertificate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.SubjectDN == nil || plan.SubjectDN.CommonName.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Subject DN",
			"Subject DN with a non-empty Common Name (CN) is required to create a self-signed certificate.",
		)
		return
	}

	subjectDN := BuildSubjectDN(plan.SubjectDN)
	planBody := map[string]any{
		"type":      "selfsigned",
		"keySize":   plan.KeySize.ValueInt64(),
		"subjectDN": subjectDN,
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

	// Create Self-Signed Certificate
	diags = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Certificate Metadata
	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Binary Certificate
	certBytes, diags := GetCertificateBinary(r.client, endpoint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certDiags := validatePEMData(string(pemBytes))
	resp.Diagnostics.Append(certDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemCertificateSelfSignedResourceValueFrom(ctx, respObj, pemBytes, plan.SubjectDN)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.Type = plan.Type
	responseModel.KeySize = plan.KeySize

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemCertificateSelfSignedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemCertificateSelfSignedResourceConfig
	var respObj apiobjects.SystemCertificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

	// Get Certificate Metadata
	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Binary Certificate
	certBytes, diags := GetCertificateBinary(r.client, endpoint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certDiags := validatePEMData(string(pemBytes))
	resp.Diagnostics.Append(certDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemCertificateSelfSignedResourceValueFrom(ctx, respObj, pemBytes, state.SubjectDN)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.Type = state.Type
	responseModel.KeySize = state.KeySize

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemCertificateSelfSignedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Changing a self-signed system certificate requires resource replacement.",
	)
}

func (r *SystemCertificateSelfSignedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemCertificateSelfSignedResourceConfig
	var respObj apiobjects.SystemCertificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
