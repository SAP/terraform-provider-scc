package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = &UICertificatePKCS12CertificateResource{}

func NewUICertificatePKCS12CertificateResource() resource.Resource {
	return &UICertificatePKCS12CertificateResource{}
}

type UICertificatePKCS12CertificateResource struct {
	client *api.RestApiClient
}

func (r *UICertificatePKCS12CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ui_certificate_pkcs12_certificate"
}

func (r *UICertificatePKCS12CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **PKCS#12 (P12) UI Certificate** for the SAP BTP Connectivity service. 
The PKCS#12 file must be created from a CSR generated in SAP Cloud Connector and signed by a trusted Certificate Authority (CA).
		
**Supports:**
- PKCS#12 Certificate: A certificate bundle that is signed by an external Certificate Authority (CA) and includes bundle containing private key and full certificate chain.

**Required Workflow:**
1. Generate a Certificate Signing Request (CSR) from the SAP Cloud Connector.
2. Submit the CSR to a trusted Certificate Authority (CA).
3. Obtain the signed certificate (leaf certificate) and the CA chain.
4. Create a PKCS#12 bundle that includes:
   - The signed leaf certificate.
   - The private key corresponding to the CSR (exported from SAP Cloud Connector).
   - Intermediate CA certificate(s) (if applicable)
   - Root CA certificate
5. Provide the chain to Terraform using either:
   - filebase64("certificate.p12")
   - Inline base64-encoded PKCS#12 string

**Notes:**
- Cloud Connector accepts **only the latest CSR**
- Certificate must match the CSR's public key and subject.
- The PKCS#12 file must include the private key.
- On deleting the UI certificate resource, Terraform only removes the resource from the state. The UI certificate remains configured in SAP Cloud Connector because the connector does not provide an API to delete UI certificates and will continue to be used until it is replaced by uploading a new certificate (for example, from a new CSR).
- Any change to the PKCS#12 content forces replacement since SAP Cloud Connector supports only one system certificate.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/authentication-and-ui-settings#upload-a-pkcs#12-certificate-as-ui-certificate>`,
		Attributes: map[string]schema.Attribute{
			"pkcs12_certificate": schema.StringAttribute{
				MarkdownDescription: `PKCS#12 (.p12) certificate bundle.
This value may be provided as:
- Raw binary using filebase64("certificate.p12")
- Base64-encoded string
- Inline Base64 multi-line string

The bundle must contain:
- Leaf certificate
- Private key
- Full certificate chain

This attribute is sensitive and forces replacement when changed.`,
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password used to decrypt the PKCS#12 file.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_password": schema.StringAttribute{
				MarkdownDescription: `Password used to encrypt the private key within the PKCS#12 file. 
This is often the same as the main password but can be different depending on how the PKCS#12 file was created.
If not set, the provider will omit this form field.`,
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subject_dn": schema.SingleNestedAttribute{
				MarkdownDescription: "Subject Distinguished Name (DN) of the certificate. The Common Name (CN) is mandatory, while other fields like L, OU, O, ST, C, or Email may be present depending on the issuing CA.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cn": schema.StringAttribute{
						MarkdownDescription: "Common Name (CN) of the certificate, typically representing the domain name or identifier for which the certificate is issued.",
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Email address associated with the certificate subject.",
						Computed:            true,
					},
					"l": schema.StringAttribute{
						MarkdownDescription: "Locality (L) of the certificate subject, such as a city or town.",
						Computed:            true,
					},
					"ou": schema.StringAttribute{
						MarkdownDescription: "Organizational Unit (OU) of the certificate subject, representing a department or division within an organization.",
						Computed:            true,
					},
					"o": schema.StringAttribute{
						MarkdownDescription: "Organization (O) of the certificate subject, representing the name of the organization.",
						Computed:            true,
					},
					"st": schema.StringAttribute{
						MarkdownDescription: "State or Province (ST) of the certificate subject.",
						Computed:            true,
					},
					"c": schema.StringAttribute{
						MarkdownDescription: "Country (C) of the certificate subject, typically represented as a two-letter ISO country code.",
						Computed:            true,
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
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of SAN, such as DNS, IP, RFC822 or URI.",
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("DNS", "IP", "RFC822", "URI"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The value of the SAN, such as a domain name for DNS, an IP address for IP, an email address for RFC822, or a URI for URI.",
							Computed:            true,
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

func (r *UICertificatePKCS12CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UICertificatePKCS12CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PKCS12UICertificateResourceConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, d := r.createInternal(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificatePKCS12CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// If there is no state, there is nothing to read (in case of mock testing, the state can be null but the resource still needs to be read to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state PKCS12UICertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetUICertificateEndpoint()

	// Get Certificate Metadata
	d := requestAndUnmarshalFunc(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, d := pkcs12UICertificateResourceValueFromFunc(ctx, respObj)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.PKCS12Certificate = state.PKCS12Certificate
	responseModel.Password = state.Password
	responseModel.KeyPassword = state.KeyPassword

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificatePKCS12CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating a PKCS#12 UI certificate is not supported because SAP Cloud Connector only accepts the latest uploaded certificate. To update the certificate, you must create a new resource with the updated PKCS#12 content, which will replace the existing certificate in SAP Cloud Connector.",
	)
}

func (r *UICertificatePKCS12CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *UICertificatePKCS12CertificateResource) createInternal(ctx context.Context, plan PKCS12UICertificateResourceConfig) (*PKCS12UICertificateResourceConfig, diag.Diagnostics) {

	var diags diag.Diagnostics
	var respObj apiobjects.Certificate

	rawCertificate, d := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	endpoint := endpoints.GetUICertificateEndpoint()

	keyPassword := ""
	if !plan.KeyPassword.IsNull() && !plan.KeyPassword.IsUnknown() {
		keyPassword = plan.KeyPassword.ValueString()
	}

	// Upload PKCS#12 Certificate
	d = uploadPKCS12CertificateFunc(r.client, endpoint, rawCertificate, plan.Password.ValueString(), keyPassword)
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

	responseModel, d := pkcs12UICertificateResourceValueFromFunc(ctx, respObj)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	responseModel.PKCS12Certificate = plan.PKCS12Certificate
	responseModel.Password = plan.Password
	responseModel.KeyPassword = plan.KeyPassword

	return &responseModel, diags
}
