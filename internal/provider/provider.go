package provider

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &cloudConnectorProvider{}
)

func New() provider.Provider {
	return NewWithClient(http.DefaultClient)
}

func NewWithClient(httpClient *http.Client) provider.Provider {
	return &cloudConnectorProvider{
		httpClient: httpClient,
	}
}

type cloudConnectorProvider struct {
	httpClient *http.Client
}

type cloudConnectorProviderData struct {
	InstanceURL types.String `tfsdk:"instance_url"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
}

func (c *cloudConnectorProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cloudconnector"
}

func (c *cloudConnectorProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Terraform Provider for SAP Cloud Connector allows users to manage and configure SAP Cloud Connector instances within SAP BTP (Business Technology Platform). It enables automation of connectivity between SAP BTP subaccounts and on-premise systems using Terraform.",
		Attributes: map[string]schema.Attribute{
			"instance_url": schema.StringAttribute{
				MarkdownDescription: "The URL of Cloud Connector Instance.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username used to connect to Cloud Connector Instance.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password used to connect to Cloud Connector Instance.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (c *cloudConnectorProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config cloudConnectorProviderData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.InstanceURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Unknown Cloud Connector Instance URL",
			"The provider cannot create the Cloud Connector client as there is an unknown configuration value for the Cloud Connector Instance URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CC_INSTANCE_URL environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Cloud Connector Instance Username",
			"The provider cannot create the Cloud Connector client as there is an unknown configuration value for the Cloud Connector Instance Username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CC_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Cloud Connector Instance Password",
			"The provider cannot create the Cloud Connector client as there is an unknown configuration value for the Cloud Connector Instance Password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CC_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	instance_url := os.Getenv("CC_INSTANCE_URL")
	username := os.Getenv("CC_USERNAME")
	password := os.Getenv("CC_PASSWORD")

	if !config.InstanceURL.IsNull() {
		instance_url = config.InstanceURL.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if instance_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Missing Cloud Connector Instance URL",
			"The provider cannot create the Cloud Connector client as there is a missing or empty value for the Cloud Connector Instance URL. "+
				"Set the Base URL value in the configuration or use the CC_INSTANCE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Cloud Connector Instance Username",
			"The provider cannot create the Cloud Connector client as there is a missing or empty value for the Cloud Connector Instance Username. "+
				"Set the username value in the configuration or use the CC_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Cloud Connector Instance Password",
			"The provider cannot create the Cloud Connector client as there is a missing or empty value for the Cloud Connector Instance Password. "+
				"Set the password value in the configuration or use the CC_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	u, err := url.Parse(instance_url)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Error while parsing Cloud Connector Instance URL",
			"The provider cannot create the Cloud Connector client as there is an error while parsing the provided Cloud Connector Instance URL.")
	}
	client := api.NewRestApiClient(c.httpClient, u, username, password)
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (c *cloudConnectorProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSubaccountsDataSource,
		NewSubaccountDataSource,
		NewSystemMappingsDataSource,
		NewSystemMappingDataSource,
		NewSystemMappingResourcesDataSource,
		NewSystemMappingResourceDataSource,
		NewDomainMappingsDataSource,
		NewSubaccountServiceChannelK8SDataSource,
		NewSubaccountServiceChannelsK8SDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (c *cloudConnectorProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubaccountResource,
		NewSystemMappingResource,
		NewSystemMappingResourceResource,
		NewDomainMappingResource,
		NewSubaccountServiceChannelK8SResource,
	}
}
