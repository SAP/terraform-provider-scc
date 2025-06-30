package provider

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	InstanceURL       types.String `tfsdk:"instance_url"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	CaCertificate     types.String `tfsdk:"ca_certificate"`
	ClientCertificate types.String `tfsdk:"client_certificate"`
	ClientKey         types.String `tfsdk:"client_key"`
}

func (c *cloudConnectorProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scc"
}

func (c *cloudConnectorProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Terraform Provider for SAP Cloud Connector allows users to manage and configure SAP Cloud Connector instances within SAP BTP (Business Technology Platform). It enables automation of connectivity between SAP BTP subaccounts and on-premise systems using Terraform.",
		Attributes: map[string]schema.Attribute{
			"instance_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the Cloud Connector instance. This can also be sourced from the `SCC_INSTANCE_URL` environment variable.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^https?://`), "must be a valid URL starting with http:// or https://"),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username used for Basic Authentication with the Cloud Connector instance. This can also be sourced from the `SCC_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password used for Basic Authentication with the Cloud Connector instance. This can also be sourced from the `SCC_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"ca_certificate": schema.StringAttribute{
				MarkdownDescription: "Contents of a PEM-encoded CA certificate used to verify the Cloud Connector server. Use `file(\"path/to/ca.pem\")` in the provider block to load from a file. This can also be sourced from the `SCC_CA_CERTIFICATE` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_certificate": schema.StringAttribute{
				MarkdownDescription: "Contents of a PEM-encoded client certificate used for mutual TLS authentication. Use `file(\"path/to/cert.pem\")` in the provider block to load from a file. This can also be sourced from the `SCC_CLIENT_CERTIFICATE` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_key": schema.StringAttribute{
				MarkdownDescription: "Contents of a PEM-encoded client private key used for mutual TLS authentication. Use `file(\"path/to/key.pem\")` in the provider block to load from a file. This can also be sourced from the `SCC_CLIENT_KEY` environment variable.",
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

	if resp.Diagnostics.HasError() {
		return
	}

	instance_url := getNonEmptyAttribute(config.InstanceURL, "SCC_INSTANCE_URL")
	username := getNonEmptyAttribute(config.Username, "SCC_USERNAME")
	password := getNonEmptyAttribute(config.Password, "SCC_PASSWORD")
	ca_certificate := getNonEmptyAttribute(config.CaCertificate, "SCC_CA_CERTIFICATE")
	client_certificate := getNonEmptyAttribute(config.ClientCertificate, "SCC_CLIENT_CERTIFICATE")
	client_key := getNonEmptyAttribute(config.ClientKey, "SCC_CLIENT_KEY")

	// Validate required values
	if instance_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Missing Cloud Connector Instance URL",
			"The provider cannot create the Cloud Connector client because the Cloud Connector Instance URL is empty.",
		)
		return
	}

	basicAuth := username != "" && password != ""
	certAuth := client_certificate != "" && client_key != ""

	if ca_certificate != "" {
		if err := validatePEM(ca_certificate); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("ca_certificate"),
				"Invalid CA Certificate",
				"The provided CA certificate is not a valid PEM-encoded block.",
			)
			return
		}
	}

	if client_certificate != "" {
		if err := validatePEM(client_certificate); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_certificate"),
				"Invalid Client Certificate",
				"The provided client certificate is not a valid PEM-encoded block.",
			)
			return
		}
	}

	if client_key != "" {
		if err := validatePEM(client_key); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("client_key"),
				"Invalid Client Key",
				"The provided client key is not a valid PEM-encoded block.",
			)
			return
		}
	}

	if !basicAuth && !certAuth {
		resp.Diagnostics.AddError(
			"Missing Authentication Details",
			"Either a username/password or a client certificate/key must be provided for authentication.",
		)
		return
	}
	if basicAuth && certAuth {
		resp.Diagnostics.AddError(
			"Conflicting Authentication Details",
			"Both Basic Authentication and Certificate-based Authentication were provided. Only one can be used.",
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the base URL
	parsedURL, err := url.Parse(instance_url)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Invalid Cloud Connector Instance URL",
			fmt.Sprintf("Failed to parse the provided Cloud Connector Instance URL: %s. Error: %v", instance_url, err),
		)
		return
	}

	// Convert CA certificate to []byte only if provided
	caCertBytes := []byte(ca_certificate)
	clientCertBytes := []byte(client_certificate)
	clientKeyBytes := []byte(client_key)

	client, err := api.NewRestApiClient(c.httpClient, parsedURL, username, password, caCertBytes, clientCertBytes, clientKeyBytes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Creation Failed",
			fmt.Sprintf("Failed to create Cloud Connector client: %v", err),
		)
		return
	}

	if err := testProviderConnection(client); err != nil {
		resp.Diagnostics.AddError(
			"Cloud Connector Authentication Failed",
			fmt.Sprintf("Authentication or connectivity check failed: %v", err),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func getNonEmptyAttribute(attr types.String, envVar string) string {
	if !attr.IsNull() && attr.ValueString() != "" {
		return attr.ValueString()
	}
	return os.Getenv(envVar)
}

func testProviderConnection(client *api.RestApiClient) error {
	resp, err := client.GetRequest("/api/v1/connector/version")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	if cerr := resp.Body.Close(); cerr != nil {
		return fmt.Errorf("failed to close response body: %w", cerr)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication rejected with status: %s", resp.Status)
	}

	return nil
}

func validatePEM(data string) error {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return errors.New("data is not a valid PEM block")
	}
	return nil
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
		NewDomainMappingDataSource,
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
