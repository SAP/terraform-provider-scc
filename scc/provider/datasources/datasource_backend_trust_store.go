package datasources

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &BackendTrustStoreDataSource{}

func NewBackendTrustStoreDataSource() datasource.DataSource {
	return &BackendTrustStoreDataSource{}
}

type BackendTrustStoreDataSource struct {
	Client *api.RestApiClient
}

func (d *BackendTrustStoreDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_trust_store"
}

func (d *BackendTrustStoreDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Backend Trust Store Data Source.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Associate Administrator
	* Subaccount Administrator 
	* Display
	* Support
	* Monitoring

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/proxy-settings>`,
		Attributes: map[string]schema.Attribute{
			"trust_all_backends": schema.BoolAttribute{
				MarkdownDescription: "Flag (boolean) indicating whether all backends are trusted (true), or only the backends represented by certificates are trusted (false).",
				Computed:            true,
			},
			"trusted_backends": schema.ListNestedAttribute{
				MarkdownDescription: "List of trusted backends represented by certificates.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"alias": schema.StringAttribute{
							MarkdownDescription: "Alias of the backend certificate.",
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
						"issuer": schema.StringAttribute{
							MarkdownDescription: "Issuer of the backend certificate.",
							Computed:            true,
						},
						"valid_to": schema.StringAttribute{
							MarkdownDescription: "Validity end date of the backend certificate.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *BackendTrustStoreDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.Client = client
}

func (d *BackendTrustStoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.BackendTrustStoreDataSourceConfig
	var respObj apiobjects.BackendTrustStoreConfiguration
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetBackendTrustStoreBaseEndpoint()

	diags = helpers.RequestAndUnmarshal(d.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := model.BackendTrustStoreDataSourceValueFrom(ctx, respObj)
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
