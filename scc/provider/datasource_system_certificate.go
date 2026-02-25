package provider

import (
	"context"
	"encoding/pem"
	"fmt"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &SystemCertificateDataSource{}

func NewSystemCertificateDataSource() datasource.DataSource {
	return &SystemCertificateDataSource{}
}

type SystemCertificateDataSource struct {
	client *api.RestApiClient
}

func (d *SystemCertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_certificate"
}

func (r *SystemCertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Certificate Data Source.
				
Provides metadata about the TLS certificate configured for the Cloud Connector system. 
This information can be used to verify certificate validity, issuer details, and hostname configuration.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-certificate-apis#get-description-for-a-system-certificate>
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-certificate-apis#get-binary-content-of-a-system-certificate>`,
		Attributes: map[string]schema.Attribute{
			"valid_to": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the end of the validity period.",
				Computed:            true,
			},
			"valid_from": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the beginning of the validity period.",
				Computed:            true,
			},
			"subject_dn": schema.SingleNestedAttribute{
				MarkdownDescription: "Subject Distinguished Name (DN) of the certificate, identifying the certificate owner.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cn": schema.StringAttribute{
						MarkdownDescription: "Common Name (CN) of the certificate, typically representing the domain name or identifier for which the certificate is issued.",
						Computed:            true,
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
			"issuer": schema.StringAttribute{
				MarkdownDescription: "Certificate authority (CA) that issued this certificate.",
				Computed:            true,
			},
			"serial_number": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the certificate, typically assigned by the CA.",
				Computed:            true,
			},
			"subject_alternative_names": schema.StringAttribute{
				MarkdownDescription: "List of Subject Alternative Names (SANs) included in the certificate, such as DNS names, IP addresses, or email addresses.",
				Computed:            true,
			},
			"certificate_pem": schema.StringAttribute{
				MarkdownDescription: "System certificate in PEM format.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *SystemCertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SystemCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemCertificateConfig
	var respObj apiobjects.SystemCertificate
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSystemCertificateEndpoint()

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

	responseModel, diags := SystemCertificateDataSourceValueFrom(ctx, respObj, pemBytes)
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
