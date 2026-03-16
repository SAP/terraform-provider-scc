package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = &UICertificateSelfSignedResource{}

func NewUICertificateSelfSignedResource() resource.Resource {
	return &UICertificateSelfSignedResource{}
}

type UICertificateSelfSignedResource struct {
	client *api.RestApiClient
}

func (r *UICertificateSelfSignedResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ui_certificate_self_signed"
}

func (r *UICertificateSelfSignedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **Self-Signed UI Certificate** in SAP Cloud Connector.
		
**Supports:**
- Self-signed certificates

**Note:**
- Any change to key_size or subject_dn forces replacement since SAP Cloud Connector supports only one principal propagation UI certificate.
- On deleting the UI certificate resource, Terraform only removes the resource from the state. The UI certificate remains configured in SAP Cloud Connector because the connector does not provide an API to delete UI certificates and will continue to be used until it is replaced by creating a new self-signed certificate.


__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/authentication-and-ui-settings#create-a-self-signed-ui-certificate>`,
		Attributes: map[string]schema.Attribute{
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits. Allowed values: 2048 or 4096.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(2048, 4096),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
				Default: int64default.StaticInt64(4096),
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
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *UICertificateSelfSignedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UICertificateSelfSignedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SelfSignedUICertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.SubjectDN.IsNull() || plan.SubjectDN.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Subject DN",
			"Subject DN with a non-empty Common Name (CN) is required to create a self-signed certificate.",
		)
		return
	}

	dnStruct, diags := expandSubjectDN(ctx, plan.SubjectDN)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
		diags = plan.SubjectAltNames.ElementsAs(ctx, &sanList, false)
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

			planBody["subjectAltNames"] = sanFields
		}
	}

	endpoint := endpoints.GetUICertificateEndpoint()

	// Create Self-Signed Certificate
	diags = requestAndUnmarshalFunc(r.client, &respObj, "POST", endpoint, planBody, false)
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

	responseModel, diags := selfSignedUICertificateResourceValueFromFunc(ctx, respObj, dnStruct)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.KeySize = plan.KeySize

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificateSelfSignedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// If there is no state, there is nothing to read (in case of mock testing, the state can be null but the resource still needs to be read to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state SelfSignedUICertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetUICertificateEndpoint()

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

	responseModel, diags := selfSignedUICertificateResourceValueFromFunc(ctx, respObj, dnStruct)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.KeySize = state.KeySize

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificateSelfSignedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Changing a self-signed UI certificate requires resource replacement.",
	)
}

func (r *UICertificateSelfSignedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// If there is no state, there is nothing to delete (in case of mock testing, the state can be null but the resource still needs to be deleted to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	resp.Diagnostics.AddWarning(
		"UI Certificate Not Deleted in SAP Cloud Connector",
		"SAP Cloud Connector does not provide an API to delete UI certificates. "+
			"Terraform will remove this resource from its state, but the certificate "+
			"will remain configured in SAP Cloud Connector until it is replaced by "+
			"another certificate.",
	)
	resp.State.RemoveResource(ctx)
}
