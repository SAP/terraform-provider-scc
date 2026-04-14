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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UICertificateSignedChainResource{}

func NewUICertificateSignedChainResource() resource.Resource {
	return &UICertificateSignedChainResource{}
}

type UICertificateSignedChainResource struct {
	client *api.RestApiClient
}

func (r *UICertificateSignedChainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ui_certificate_signed_chain"
}

func (r *UICertificateSignedChainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **Signed Chain UI Certificate** for the SAP BTP Connectivity service. 
This resource uploads a certificate chain that was generated from a CSR downloaded from SAP Cloud Connector for **Principal Propagation**.
The uploaded certificate chain becomes the **UI certificate** used by the connector.
		
**Supports:**
- Signed Chain Certificate: A certificate that is signed by an external Certificate Authority (CA) and includes the full certificate chain up to the root CA.

**Required Workflow:**
1. Generate a **Certificate Signing Request (CSR)** from SAP Cloud Connector for Principal Propagation.
2. Submit the CSR to a trusted Certificate Authority (CA) to obtain a signed certificate chain.
3. Construct the PEM chain in the following order:
   - Signed certificate (generated from the CSR)
   - Intermediate CA certificate(s) (if applicable)
   - Root CA certificate
4. Provide the chain to Terraform using either:
   - file("signed_chain.pem")
   - or by directly pasting the PEM-encoded chain in the configuration.

**Behavior:**
- This resource supports **in-place certificate rotation**.
- Updating the signed_chain will **upload a new certificate**, replacing the existing certificate without deleting it.
- This avoids downtime and aligns with the Cloud Connector certificate lifecycle (CSR → sign → upload).

**Renewal Note:**
- To renew a certificate, a **new CSR must be generated** from SAP Cloud Connector.
- The signed certificate must correspond to the **most recently generated CSR**, otherwise the upload will fail.

**Notes:**
- Cloud Connector accepts **only the latest CSR**
- Certificate must match the CSR's public key and subject.
- Chain must be PEM-encoded.
- On deleting the UI certificate resource, Terraform only removes the resource from the state. The UI certificate remains configured in SAP Cloud Connector because the connector does not provide an API to delete UI certificates and will continue to be used until it is replaced by uploading a new certificate (for example, from a new CSR).

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/authentication-and-ui-settings#upload-a-signed-certificate-chain-as-ui-certificate>`,
		Attributes: map[string]schema.Attribute{
			"signed_chain": schema.StringAttribute{
				MarkdownDescription: `PEM-encoded signed certificate chain for the UI certificate.
The certificate chain must be ordered as follows:
1. **Signed Certificate**  
   The certificate issued from the Cloud Connector CSR.
2. **Intermediate CA Certificate(s)** (optional)  
   Certificates that link the signed certificate to the root CA.
3. **Root CA Certificate**  
   The trust anchor of the certificate hierarchy.
This value can be provided using:
- file("signed_chain.pem")
- Inline multi-line string.

The provider validates PEM format before uploading.`,
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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

func (r *UICertificateSignedChainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UICertificateSignedChainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SignedChainUICertificateResourceConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, d := createSignedChainUICertificateFunc(r, ctx, plan.SignedChain.ValueString())
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificateSignedChainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// If there is no state, there is nothing to read (in case of mock testing, the state can be null but the resource still needs to be read to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state SignedChainUICertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetUICertificateEndpoint()

	// Get Certificate Metadata
	diags = requestAndUnmarshalFunc(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := signedChainUICertificateResourceValueFromFunc(ctx, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.SignedChain = state.SignedChain

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificateSignedChainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SignedChainUICertificateResourceConfig
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

	// If the signed chain is not changing, there is no need to update
	if !shouldUpdateSignedChain(plan.SignedChain, state.SignedChain) {
		return
	}

	model, diags := createSignedChainUICertificateFunc(r, ctx, plan.SignedChain.ValueString())
	if diags.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UICertificateSignedChainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

var createSignedChainUICertificateFunc = func(r *UICertificateSignedChainResource, ctx context.Context, signedChain string) (*SignedChainUICertificateResourceConfig, diag.Diagnostics) {
	var respObj apiobjects.Certificate
	var diags diag.Diagnostics
	if signedChain != "" {
		certDiags := validatePEMChainFunc(signedChain)
		diags.Append(certDiags...)
		if diags.HasError() {
			return nil, diags
		}
	}

	endpoint := endpoints.GetUICertificateEndpoint()

	// Upload Signed Certificate Chain
	d := uploadSignedChainFunc(r.client, endpoint, signedChain)
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

	responseModel, d := signedChainUICertificateResourceValueFromFunc(ctx, respObj)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	responseModel.SignedChain = types.StringValue(signedChain)

	return &responseModel, diags
}
