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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CACertificateSelfSignedResource{}

func NewCACertificateSelfSignedResource() resource.Resource {
	return &CACertificateSelfSignedResource{}
}

type CACertificateSelfSignedResource struct {
	client *api.RestApiClient
}

func (r *CACertificateSelfSignedResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_certificate_self_signed"
}

func (r *CACertificateSelfSignedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **Self-Signed CA Certificate** in SAP Cloud Connector.
		
**Supports:**
• Self-signed certificates

**Behavior:**
- This resource creates a self-signed CA certificate directly on the SAP Cloud Connector.
- Any change to key_size or subject_dn will result in **replacement of the existing certificate**, as only one CA certificate is supported.
- Replacement will create a new certificate and remove the existing one.

**Notes:**
- SAP Cloud Connector supports only a single CA certificate for this purpose.
- Changing certificate properties (such as key size or subject) requires generating a new certificate.
- On terraform destroy, the CA certificate is removed from the SAP Cloud Connector, which may disrupt dependent configurations until a new certificate is created.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/ca-certificate-for-principal-propagation-apis#create-a-self-signed-ca-certificate-for-principal-propagation-(master-only)>`,
		Attributes: map[string]schema.Attribute{
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits. Allowed values: 2048 or 4096.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(2048, 4096),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Default: int64default.StaticInt64(4096),
			},
			"subject_dn": schema.SingleNestedAttribute{
				MarkdownDescription: "Subject Distinguished Name (DN) of the certificate. The Common Name (CN) is mandatory, while other fields like L, OU, O, ST, C, or Email may be present depending on the issuing CA.",
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
			"certificate_pem": schema.StringAttribute{
				MarkdownDescription: "CA certificate in PEM format.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CACertificateSelfSignedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CACertificateSelfSignedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SelfSignedCACertificateResourceConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, diags := createSelfSignedCACertificateFunc(r, ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CACertificateSelfSignedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	var state SelfSignedCACertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetCACertificateEndpoint()

	dnStruct, diags := expandSubjectDN(ctx, state.SubjectDN)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Certificate Metadata
	diags = requestAndUnmarshalFunc(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Binary Certificate
	certBytes, diags := getCertificateBinaryFunc(r.client, endpoint)
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

	responseModel, diags := selfSignedCACertificateResourceValueFromFunc(ctx, respObj, dnStruct)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.KeySize = state.KeySize
	responseModel.CertificatePEM = types.StringValue(string(pemBytes))

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CACertificateSelfSignedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SelfSignedCACertificateResourceConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.KeySize == state.KeySize && plan.SubjectDN.Equal(state.SubjectDN) && plan.SubjectAltNames.Equal(state.SubjectAltNames) {
		return
	}

	if !shouldUpdateSelfSignedCertificate(plan.KeySize, state.KeySize, plan.SubjectDN, state.SubjectDN, plan.SubjectAltNames, state.SubjectAltNames) {
		return
	}

	model, diags := createSelfSignedCACertificateFunc(r, ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CACertificateSelfSignedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	var state SelfSignedCACertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetCACertificateEndpoint()

	diags = requestAndUnmarshalFunc(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

var createSelfSignedCACertificateFunc = func(r *CACertificateSelfSignedResource, ctx context.Context, plan SelfSignedCACertificateResourceConfig) (*SelfSignedCACertificateResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var respObj apiobjects.Certificate

	if plan.SubjectDN.IsNull() || plan.SubjectDN.IsUnknown() {
		diags.AddError(
			"Missing Subject DN",
			"Subject DN with a non-empty Common Name (CN) is required to create a self-signed certificate.",
		)
		return nil, diags
	}

	dnStruct, d := expandSubjectDN(ctx, plan.SubjectDN)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	subjectDN := buildSubjectDN(dnStruct)
	planBody := map[string]any{
		"type":      "selfsigned",
		"keySize":   plan.KeySize.ValueInt64(),
		"subjectDN": subjectDN,
	}

	if !plan.SubjectAltNames.IsNull() &&
		!plan.SubjectAltNames.IsUnknown() {
		var sanList []subjectAlternativeNames
		d = plan.SubjectAltNames.ElementsAs(ctx, &sanList, false)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		if len(sanList) > 0 {
			sanFields := []map[string]string{}
			for _, san := range sanList {
				sanFields = append(sanFields, map[string]string{
					"type":  san.Type.ValueString(),
					"value": san.Value.ValueString(),
				})
			}

			planBody["subjectAltNames"] = sanFields
		}
	}

	endpoint := endpoints.GetCACertificateEndpoint()

	// Create Self-Signed Certificate
	d = requestAndUnmarshalFunc(r.client, &respObj, "POST", endpoint, planBody, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// Get Certificate Metadata
	d = requestAndUnmarshalFunc(r.client, &respObj, "GET", endpoint, nil, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// Generate Binary Certificate
	certBytes, d := getCertificateBinaryFunc(r.client, endpoint)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certDiags := validatePEMData(string(pemBytes))
	diags.Append(certDiags...)
	if diags.HasError() {
		return nil, diags
	}

	responseModel, d := selfSignedCACertificateResourceValueFromFunc(ctx, respObj, dnStruct)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	responseModel.KeySize = plan.KeySize
	responseModel.CertificatePEM = types.StringValue(string(pemBytes))

	return &responseModel, diags
}
