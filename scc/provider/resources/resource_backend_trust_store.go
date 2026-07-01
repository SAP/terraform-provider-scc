package resources

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BackendTrustStoreResource{}
var _ resource.ResourceWithImportState = &BackendTrustStoreResource{}
var _ resource.ResourceWithIdentity = &BackendTrustStoreResource{}

func NewBackendTrustStoreResource() resource.Resource {
	return &BackendTrustStoreResource{}
}

type BackendTrustStoreResource struct {
	Client *api.RestApiClient
}

type backendTrustStoreResourceIdentityModel struct {
	Alias types.String `tfsdk:"alias"`
}

func (r *BackendTrustStoreResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_trust_store"
}

func (r *BackendTrustStoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Adds and manages a **Back-End CA Certificate** in the SAP Cloud Connector **Back-End Trust Store Allowlist**. 
The Back-End Trust Store contains trusted Certificate Authority (CA) certificates that SAP Cloud Connector uses to validate TLS certificates presented by on-premise back-end systems.
		
**Supports:**
- Root CA certificates
- Intermediate CA certificates
- Self-signed CA certificates used as trust anchors

**Usage:**
Provide the certificate in PEM format using either:
- file("rootCA.pem")
- or by directly pasting the PEM-encoded certificate into the Terraform configuration.

**Behavior:**
- The certificate is uploaded to the Back-End Trust Store when the resource is created.
- Updating the certificate is **not supported** because SAP Cloud Connector does not provide an API to replace an existing trust store certificate.

**Notes:**
- Certificates must be PEM-encoded.
- The provider validates the PEM format before uploading.
- Deleting this resource removes the corresponding certificate from the Back-End Trust Store.
- Removing a certificate that is required to validate a back-end system's TLS certificate may cause SSL/TLS connections to that system to fail.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Associate Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/truststore-ca-certificates#add-a-back-end-certificate-to-truststore-(master-only)>`,
		Attributes: map[string]schema.Attribute{
			"certificate": schema.StringAttribute{
				MarkdownDescription: `PEM-encoded CA certificate to add to the Back-End Trust Store.
The certificate may be a root CA certificate, an intermediate CA certificate, or another trusted CA certificate required to validate TLS connections to back-end systems.

This value can be provided using:
- file("rootCA.pem")
- Inline multi-line string.

The provider validates that the value is a valid PEM-encoded certificate before uploading.`,
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Alias that uniquely identifies the certificate in the SAP Cloud Connector Back-End Trust Store.",
				Computed:            true,
			},
			"subject_dn": schema.SingleNestedAttribute{
				MarkdownDescription: "Subject Distinguished Name (DN) of the certificate. This contains the identity information associated with the certificate subject.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cn": schema.StringAttribute{
						MarkdownDescription: "Common Name (CN) of the certificate subject.",
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Email address (EMAIL) associated with the certificate subject.",
						Computed:            true,
					},
					"l": schema.StringAttribute{
						MarkdownDescription: "Locality (L) of the certificate subject, such as a city or town.",
						Computed:            true,
					},
					"ou": schema.StringAttribute{
						MarkdownDescription: "Organizational Unit (OU) of the certificate subject.",
						Computed:            true,
					},
					"o": schema.StringAttribute{
						MarkdownDescription: "Organization (O) of the certificate subject.",
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
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "Certificate authority (CA) that issued this certificate.",
				Computed:            true,
			},
		},
	}
}

func (rs *BackendTrustStoreResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"alias": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *BackendTrustStoreResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.Client = client
}

func (r *BackendTrustStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.BackendTrustStoreResourceConfig
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, diags := UploadBackendCertificateFunc(r, ctx, plan.Certificate.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Identity != nil {
		identity := backendTrustStoreResourceIdentityModel{
			Alias: model.Alias,
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *BackendTrustStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if req.State.Raw.IsNull() {
		return
	}

	var state model.BackendTrustStoreResourceConfig
	var trustStore apiobjects.BackendTrustStoreConfiguration

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alias := state.Alias.ValueString()
	endpoint := endpoints.GetBackendTrustStoreBaseEndpoint()

	diags = helpers.RequestAndUnmarshal(
		r.Client,
		&trustStore,
		"GET",
		endpoint,
		nil,
		true,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, trustedBackend := range trustStore.TrustedBackends {
		if trustedBackend.Alias != alias {
			continue
		}

		resourceModel, d := r.buildBackendTrustStoreModel(trustedBackend)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		diags = resp.State.Set(ctx, &resourceModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if resp.Identity != nil {
			identity := backendTrustStoreResourceIdentityModel{
				Alias: resourceModel.Alias,
			}

			diags = resp.Identity.Set(ctx, identity)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		return
	}

	resp.Diagnostics.AddError(
		"Backend trust store certificate not found",
		fmt.Sprintf("No backend trust store certificate with alias %q exists.", alias),
	)
}

func (r *BackendTrustStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Updating Backend Trust Store Certificates is Not Supported",
		"SAP Cloud Connector does not provide an API to update an existing backend trust store certificate. Delete the resource and create it again with the new certificate.",
	)
}

func (r *BackendTrustStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to delete
	if req.State.Raw.IsNull() {
		return
	}

	var state model.BackendTrustStoreResourceConfig
	var respObj apiobjects.BackendTrustStoreConfiguration
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetBackendTrustStoreCertificateEndpoint() + "/" + state.Alias.ValueString()

	diags = helpers.RequestAndUnmarshal(
		r.Client,
		&respObj,
		"DELETE",
		endpoint,
		nil,
		false,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

var UploadBackendCertificateFunc = func(r *BackendTrustStoreResource, ctx context.Context, certificate string) (*model.BackendTrustStoreResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Parse uploaded certificate
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		diags.AddError(
			"Invalid Certificate",
			"Failed to decode PEM certificate.",
		)
		return nil, diags
	}

	uploadedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		diags.AddError(
			"Invalid Certificate",
			fmt.Sprintf("The certificate must be a valid PEM-encoded X.509 certificate: %v", err),
		)
		return nil, diags
	}

	uploadEndpoint := endpoints.GetBackendTrustStoreCertificateEndpoint()

	d := helpers.UploadBackendTrustStoreCertificateFunc(r.Client, uploadEndpoint, certificate)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// Read trusted backends to get the alias and other metadata of the uploaded certificate
	var trustStore apiobjects.BackendTrustStoreConfiguration

	readEndpoint := endpoints.GetBackendTrustStoreBaseEndpoint()
	d = helpers.RequestAndUnmarshal(r.Client, &trustStore, "GET", readEndpoint, nil, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	for _, trustedBackend := range trustStore.TrustedBackends {
		if !helpers.MatchesBackendCertificate(trustedBackend, uploadedCert) {
			continue
		}

		resourceModel, d := r.buildBackendTrustStoreModel(trustedBackend)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		return &resourceModel, diags
	}

	diags.AddError(
		"Failed to Find Uploaded Certificate",
		"The uploaded certificate was not found in the Back-End Trust Store. Ensure that the certificate was uploaded successfully.",
	)

	return nil, diags
}

func (rs *BackendTrustStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		idParts := strings.Split(req.ID, ",")

		if len(idParts) != 1 || idParts[0] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: alias. Got: %q", req.ID),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("alias"), idParts[0])...)

		return
	}

	var identity backendTrustStoreResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("alias"), identity.Alias)...)
}

func (r *BackendTrustStoreResource) buildBackendTrustStoreModel(trustedBackend apiobjects.TrustedBackends) (model.BackendTrustStoreResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	certEndpoint := endpoints.GetBackendTrustStoreCertificateEndpoint() + "/" + trustedBackend.Alias

	certBytes, d := helpers.GetCertificateBinaryFunc(r.Client, certEndpoint)
	diags.Append(d...)
	if diags.HasError() {
		return model.BackendTrustStoreResourceConfig{}, diags
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	resourceModel, d := model.BackendTrustStoreResourceValueFrom(string(pemBytes), trustedBackend)
	diags.Append(d...)
	if diags.HasError() {
		return model.BackendTrustStoreResourceConfig{}, diags
	}

	return resourceModel, diags
}
