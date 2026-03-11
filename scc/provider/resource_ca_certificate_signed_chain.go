package provider

import (
	"context"
	"encoding/pem"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = &CACertificateSignedChainResource{}

func NewCACertificateSignedChainResource() resource.Resource {
	return &CACertificateSignedChainResource{}
}

type CACertificateSignedChainResource struct {
	client *api.RestApiClient
}

func (r *CACertificateSignedChainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_certificate_signed_chain"
}

func (r *CACertificateSignedChainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **Signed Chain CA Certificate for Principal Propagation** for the SAP BTP Connectivity service. 
This resource uploads a certificate chain that was generated from a CSR downloaded from SAP Cloud Connector for **Principal Propagation**.
The uploaded certificate chain becomes the **Principal Propagation CA certificate** used by the connector.
		
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


**Notes:**
- Cloud Connector accepts **only the latest CSR**
- Certificate must match the CSR's public key and subject.
- Chain must be PEM-encoded.
- On deleting the CA certificate resource, the certificate is removed from the SAP Cloud Connector, and any existing connections that rely on that certificate will be disrupted until a new certificate is uploaded using a new CSR.
- Any change to signed_chain forces replacement since SAP Cloud Connector supports only one CA certificate.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/ca-certificate-for-principal-propagation-apis#upload-a-signed-certificate-chain-as-ca-certificate-for-principal-propagation-(master-only)>`,
		Attributes: map[string]schema.Attribute{
			"signed_chain": schema.StringAttribute{
				MarkdownDescription: `PEM-encoded signed certificate chain for the Principal Propagation CA certificate.
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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
			"certificate_pem": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded certificate data. This is the leaf certificate extracted from the provided signed chain.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CACertificateSignedChainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CACertificateSignedChainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SignedChainCertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.SignedChain.IsNull() && !plan.SignedChain.IsUnknown() {
		certDiags := validatePEMChainFunc(plan.SignedChain.ValueString())
		resp.Diagnostics.Append(certDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	endpoint := endpoints.GetCACertificateEndpoint()

	// Upload Signed Certificate Chain
	diags = uploadSignedChainFunc(r.client, endpoint, plan.SignedChain.ValueString())
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

	responseModel, diags := signedChainCertificateResourceValueFromFunc(ctx, respObj, pemBytes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.SignedChain = plan.SignedChain

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *CACertificateSignedChainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// If there is no state, there is nothing to read (in case of mock testing, the state can be null but the resource still needs to be read to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state SignedChainCertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetCACertificateEndpoint()

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

	responseModel, diags := signedChainCertificateResourceValueFromFunc(ctx, respObj, pemBytes)
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

func (r *CACertificateSignedChainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Changing a signed CA certificate requires resource replacement.",
	)
}

func (r *CACertificateSignedChainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// If there is no state, there is nothing to delete (in case of mock testing, the state can be null but the resource still needs to be deleted to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state SignedChainCertificateResourceConfig
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
