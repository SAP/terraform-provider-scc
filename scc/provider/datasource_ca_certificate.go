package provider

import (
	"context"
	"encoding/pem"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &CACertificateDataSource{}

func NewCACertificateDataSource() datasource.DataSource {
	return &CACertificateDataSource{}
}

type CACertificateDataSource struct {
	client *api.RestApiClient
}

func (d *CACertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_certificate"
}

func (r *CACertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector CA Certificate Data Source.
				
Provides metadata about the Certificate Authority (CA) certificate trusted by the Cloud Connector.
This information can be used to verify the issuing authority, certificate validity period, and trust configuration.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/ca-certificate-for-principal-propagation-apis#get-description-for-a-ca-certificate-for-principal-propagation>
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/ca-certificate-for-principal-propagation-apis#get-binary-content-of-a-ca-certificate-for-principal-propagation>`,
		Attributes: map[string]schema.Attribute{
			"valid_to": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the end of the validity period.",
				Computed:            true,
			},
			"valid_from": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the beginning of the validity period.",
				Computed:            true,
			},
			"subject_dn": schema.StringAttribute{
				MarkdownDescription: "Subject Distinguished Name (DN) of the CA certificate, identifying the certificate authority.",
				Computed:            true,
			},
			"issuer": schema.StringAttribute{
				MarkdownDescription: "Distinguished Name (DN) of the issuing Certificate Authority. For self-signed root CAs, this is the same as the subject.",
				Computed:            true,
			},
			"serial_number": schema.StringAttribute{
				MarkdownDescription: "Serial number assigned to the CA certificate by its issuing authority.",
				Computed:            true,
			},
			"subject_alternative_names": schema.StringAttribute{
				MarkdownDescription: "Subject Alternative Names (SANs) present in the CA certificate, if any.",
				Computed:            true,
			},
			"certificate_pem": schema.StringAttribute{
				MarkdownDescription: "CA certificate in PEM format, which can be used to configure trust stores or verify certificate chains.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *CACertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *CACertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CACertificateConfig
	var respObj apiobjects.CACertificate
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetCACertificateEndpoint()

	// Get Certificate Metadata
	diags = requestAndUnmarshal(d.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Binary Certificate
	certBytes, diags := GetCertificateBinary(d.client, endpoint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	responseModel, diags := CACertificateDataSourceValueFrom(ctx, respObj, pemBytes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
