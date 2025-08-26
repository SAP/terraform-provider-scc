package provider

import (
	"context"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
				MarkdownDescription: "The username used for Basic Authentication with the Cloud Connector instance. This can also be sourced from the `SCC_USERNAME` environment variable (useful when storing and retrieving secrets from secure stores).",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password used for Basic Authentication with the Cloud Connector instance. This can also be sourced from the `SCC_PASSWORD` environment variable (useful when storing and retrieving secrets from secure stores).",
				Optional:            true,
				Sensitive:           true,
			},
			"ca_certificate": schema.StringAttribute{
				MarkdownDescription: `Contents of a PEM-encoded CA certificate used to verify the **UI Certificate** of the Cloud Connector server.
Use **file(\"path/to/ca_cert.pem\")** in the provider block to load from a file. This can also be sourced from the **SCC_CA_CERTIFICATE** environment variable (useful when storing and retrieving secrets from secure stores).

**Note:** 
- This should match the UI certificate presented by the Cloud Connector.
- If the certificate chain involves intermediate certificates, ensure they are included to complete the trust chain.`,
				Optional:  true,
				Sensitive: true,
			},
			"client_certificate": schema.StringAttribute{
				MarkdownDescription: `Contents of a PEM-encoded **client certificate** used for **mutual TLS (mTLS) authentication** with the Cloud Connector.
Use **file(\"path/to/client_cert.pem\")** in the provider block to load from a file. This can also be sourced from the **SCC_CLIENT_CERTIFICATE** environment variable (useful when storing and retrieving secrets from secure stores).

**Note:** 
- This must be the client certificate associated with the private key provided in client_key attribute.
- If the certificate chain includes intermediate certificates, ensure they are included in the PEM file (with the client cert first, followed by intermediates) to complete the trust chain.`,
				Optional:  true,
				Sensitive: true,
			},
			"client_key": schema.StringAttribute{
				MarkdownDescription: `Contents of a PEM-encoded **client private key** used for **mutual TLS (mTLS) authentication** with the Cloud Connector.
Use **file(\"path/to/client_key.pem\")** in the provider block to load from a file. This can also be sourced from the **SCC_CLIENT_KEY** environment variable (useful when storing and retrieving secrets from secure stores).

**Note:**
- This key must match the client certificate provided in client_certificate attribute.`,
				Optional:  true,
				Sensitive: true,
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

	instanceURL, username, password, caCertificate, clientCertificate, clientKey := resolveAttributes(config)

	// Validate values from config
	if !validateConfig(instanceURL, username, password, caCertificate, clientCertificate, clientKey, resp) {
		return
	}

	// Parse Instance URL
	parsedURL := parseInstanceURL(instanceURL, resp)
	if parsedURL == nil {
		return
	}
	// Create Client
	client := createClient(c.httpClient, parsedURL, username, password, caCertificate, clientCertificate, clientKey, resp)
	if client == nil {
		return
	}
	// Test Provider Connection
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

func resolveAttributes(config cloudConnectorProviderData) (string, string, string, string, string, string) {
	return getNonEmptyAttribute(config.InstanceURL, "SCC_INSTANCE_URL"),
		getNonEmptyAttribute(config.Username, "SCC_USERNAME"),
		getNonEmptyAttribute(config.Password, "SCC_PASSWORD"),
		getNonEmptyAttribute(config.CaCertificate, "SCC_CA_CERTIFICATE"),
		getNonEmptyAttribute(config.ClientCertificate, "SCC_CLIENT_CERTIFICATE"),
		getNonEmptyAttribute(config.ClientKey, "SCC_CLIENT_KEY")
}

func getNonEmptyAttribute(attr types.String, envVar string) string {
	if !attr.IsNull() && attr.ValueString() != "" {
		return attr.ValueString()
	}
	return os.Getenv(envVar)
}

func validateConfig(instanceURL, username, password, caCertificate, clientCertificate, clientKey string, resp *provider.ConfigureResponse) bool {
	if instanceURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Missing Cloud Connector Instance URL",
			"The provider cannot create the Cloud Connector client because the Cloud Connector Instance URL is empty.",
		)
		return false
	}

	if caCertificate != "" && !validatePEMBlock(caCertificate, "ca_certificate", "CA Certificate", resp) {
		return false
	}
	if clientCertificate != "" && !validatePEMBlock(clientCertificate, "client_certificate", "Client Certificate", resp) {
		return false
	}
	if clientKey != "" && !validatePEMBlock(clientKey, "client_key", "Client Key", resp) {
		return false
	}

	basicAuth := username != "" && password != ""
	certAuth := clientCertificate != "" && clientKey != ""

	switch {
	case !basicAuth && !certAuth:
		resp.Diagnostics.AddError(
			"Missing Authentication Details",
			"Either a username/password or a client certificate/key must be provided for authentication.",
		)
		return false
	case basicAuth && certAuth:
		resp.Diagnostics.AddError(
			"Conflicting Authentication Details",
			"Both Basic Authentication and Certificate-based Authentication were provided. Only one can be used.",
		)
		return false
	}

	return true
}
func parseInstanceURL(instanceURL string, resp *provider.ConfigureResponse) *url.URL {
	parsedURL, err := url.Parse(instanceURL)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("instance_url"),
			"Invalid Cloud Connector Instance URL",
			fmt.Sprintf("Failed to parse the provided Cloud Connector Instance URL: %s. Error: %v", instanceURL, err),
		)
		return nil
	}
	return parsedURL
}
func createClient(httpClient *http.Client, parsedURL *url.URL, username, password, caCertificate, clientCertificate, clientKey string, resp *provider.ConfigureResponse) *api.RestApiClient {
	client, diags := api.NewRestApiClient(
		httpClient,
		parsedURL,
		username,
		password,
		[]byte(caCertificate),
		[]byte(clientCertificate),
		[]byte(clientKey),
	)
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"Client Creation Failed",
			fmt.Sprintf("Failed to create Cloud Connector client: %v", diags),
		)
		return nil
	}
	return client
}
func testProviderConnection(client *api.RestApiClient) diag.Diagnostics {
	resp, diags := client.GetRequest("/api/v1/connector/version")
	if diags.HasError() {
		return diags
	}
	if cerr := resp.Body.Close(); cerr != nil {
		diags.AddError("Response Body Close Error", fmt.Sprintf("Failed to close response body: %v", cerr))
		return diags
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		diags.AddError("Authentication Failed", "The provided authentication details were rejected by the Cloud Connector instance. Please verify your credentials and try again.")
		return diags
	}

	return diags
}

func validatePEMBlock(pemString, attribute, title string, resp *provider.ConfigureResponse) bool {
	if err := validatePEM(pemString); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root(attribute),
			fmt.Sprintf("Invalid %s", title),
			fmt.Sprintf("The provided %s is not a valid PEM-encoded block.", title),
		)
		return false
	}
	return true
}

func validatePEM(data string) diag.Diagnostics {
	var diags diag.Diagnostics
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		diags.AddError("Invalid PEM Block", "data is not a valid PEM block")
		return diags
	}
	return diags
}

// DataSources defines the data sources implemented in the provider.
func (c *cloudConnectorProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSubaccountsDataSource,
		NewSubaccountConfigurationDataSource,
		NewSystemMappingsDataSource,
		NewSystemMappingDataSource,
		NewSystemMappingResourcesDataSource,
		NewSystemMappingResourceDataSource,
		NewDomainMappingsDataSource,
		NewDomainMappingDataSource,
		NewSubaccountK8SServiceChannelDataSource,
		NewSubaccountK8SServiceChannelsDataSource,
		NewSubaccountABAPServiceChannelDataSource,
		NewSubaccountABAPServiceChannelsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (c *cloudConnectorProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubaccountResource,
		NewSubaccountUsingAuthResource,
		NewSystemMappingResource,
		NewSystemMappingResourceResource,
		NewDomainMappingResource,
		NewSubaccountK8SServiceChannelResource,
		NewSubaccountABAPServiceChannelResource,
	}
}
