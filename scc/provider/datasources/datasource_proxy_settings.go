package datasources

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &ProxySettingsDataSource{}

func NewProxySettingsDataSource() datasource.DataSource {
	return &ProxySettingsDataSource{}
}

type ProxySettingsDataSource struct {
	Client *api.RestApiClient
}

func (d *ProxySettingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_settings"
}

func (d *ProxySettingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Proxy Settings Data Source.

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
			"host": schema.StringAttribute{
				MarkdownDescription: "The name of the proxy host.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the proxy host.",
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The username for the proxy authentication.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for the proxy authentication.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *ProxySettingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProxySettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.ProxySettingsDataSourceConfig
	var respObj apiobjects.ProxySettings
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetProxySettingsEndpoint()

	diags = helpers.RequestAndUnmarshal(d.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := model.ProxySettingsDataSourceValueFrom(ctx, respObj)
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
