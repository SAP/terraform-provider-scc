package provider

import (
	"context"
	"encoding/pem"
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

var _ resource.Resource = &SystemCertificatePKCS12CertificateResource{}

func NewSystemCertificatePKCS12CertificateResource() resource.Resource {
	return &SystemCertificatePKCS12CertificateResource{}
}

type SystemCertificatePKCS12CertificateResource struct {
	client *api.RestApiClient
}

func (r *SystemCertificatePKCS12CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_certificate_pkcs12_certificate"
}

func (r *SystemCertificatePKCS12CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates and manages a **PKCS#12 (P12) System Certificate** for the SAP BTP Connectivity service. 
The PKCS#12 file must be created from a CSR generated in SAP Cloud Connector and signed by a trusted Certificate Authority (CA).
		
**Supports:**
• PKCS#12 Certificate: A certificate bundle that is signed by an external Certificate Authority (CA) and includes bundle containing private key and full certificate chain.

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

**Behavior:**
- This resource supports **in-place certificate rotation**.
- Updating the pkcs12_certificate, password or key_password will **upload a new certificate**, replacing the existing certificate without deleting it.
- This avoids downtime and aligns with the Cloud Connector certificate lifecycle (CSR → sign → upload).

**Renewal Note:**
- To renew a certificate, a **new CSR must be generated** from SAP Cloud Connector.
- The signed certificate must correspond to the **most recently generated CSR**, otherwise the upload will fail.

**Notes:**
- Cloud Connector accepts **only the latest CSR**
- Certificate must match the CSR's public key and subject.
- The PKCS#12 file must include the private key.
- On deleting the system certificate resource, the certificate is removed from the SAP Cloud Connector, and any existing connections that rely on that certificate will be disrupted until a new certificate is uploaded using a new CSR.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-certificate-apis#upload-a-pkcs#12-certificate-as-system-certificate-(master-only)>`,
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
- Full certificate chain`,
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password used to decrypt the PKCS#12 file.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"key_password": schema.StringAttribute{
				MarkdownDescription: `Password used to encrypt the private key within the PKCS#12 file. 
This is often the same as the main password but can be different depending on how the PKCS#12 file was created.
If not set, the provider will omit this form field.`,
				Optional:  true,
				Sensitive: true,
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

func (r *SystemCertificatePKCS12CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SystemCertificatePKCS12CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PKCS12SystemCertificateResourceConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, d := createPKCS12SystemCertificateFunc(r, ctx, plan)
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

func (r *SystemCertificatePKCS12CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// If there is no state, there is nothing to read (in case of mock testing, the state can be null but the resource still needs to be read to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state PKCS12SystemCertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

	// Get Certificate Metadata
	requestDiags := requestAndUnmarshalFunc(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(requestDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Binary Certificate
	certBytes, binaryDiags := getCertificateBinaryFunc(r.client, endpoint)
	resp.Diagnostics.Append(binaryDiags...)
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

	responseModel, modelDiags := pkcs12SystemCertificateResourceValueFromFunc(ctx, respObj)
	resp.Diagnostics.Append(modelDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.PKCS12Certificate = state.PKCS12Certificate
	responseModel.Password = state.Password
	responseModel.KeyPassword = state.KeyPassword
	responseModel.CertificatePEM = types.StringValue(string(pemBytes))

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemCertificatePKCS12CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PKCS12SystemCertificateResourceConfig
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

	if !shouldUpdatePKCS12(plan.PKCS12Certificate, state.PKCS12Certificate, plan.Password, state.Password, plan.KeyPassword, state.KeyPassword) {
		// No changes to the certificate or passwords, so skip the update
		return
	}

	responseModel, d := createPKCS12SystemCertificateFunc(r, ctx, plan)
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

func (r *SystemCertificatePKCS12CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// If there is no state, there is nothing to delete (in case of mock testing, the state can be null but the resource still needs to be deleted to set the response)
	if req.State.Raw.IsNull() {
		return
	}

	var state PKCS12SystemCertificateResourceConfig
	var respObj apiobjects.Certificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

	diags = requestAndUnmarshalFunc(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

var createPKCS12SystemCertificateFunc = func(r *SystemCertificatePKCS12CertificateResource, ctx context.Context, plan PKCS12SystemCertificateResourceConfig) (*PKCS12SystemCertificateResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var respObj apiobjects.Certificate

	rawCertificate, d := validatePKCS12Inputs(plan.PKCS12Certificate, plan.KeyPassword)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

	keyPassword := ""
	if !plan.KeyPassword.IsNull() && !plan.KeyPassword.IsUnknown() {
		keyPassword = plan.KeyPassword.ValueString()
	}

	// Upload PKCS#12 Certificate
	uploadDiags := uploadPKCS12CertificateFunc(r.client, endpoint, rawCertificate, plan.Password.ValueString(), keyPassword)
	diags.Append(uploadDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// Get Certificate Metadata
	requestDiags := requestAndUnmarshalFunc(r.client, &respObj, "GET", endpoint, nil, true)
	diags.Append(requestDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// Generate Binary Certificate
	certBytes, binaryDiags := getCertificateBinaryFunc(r.client, endpoint)
	diags.Append(binaryDiags...)
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

	responseModel, modelDiags := pkcs12SystemCertificateResourceValueFromFunc(ctx, respObj)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return nil, diags
	}

	responseModel.PKCS12Certificate = plan.PKCS12Certificate
	responseModel.Password = plan.Password
	responseModel.KeyPassword = plan.KeyPassword
	responseModel.CertificatePEM = types.StringValue(string(pemBytes))

	return &responseModel, diags
}
