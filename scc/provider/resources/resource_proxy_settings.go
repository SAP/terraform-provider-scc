package resources

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProxySettingsResource{}

func NewProxySettingsResource() resource.Resource {
	return &ProxySettingsResource{}
}

type ProxySettingsResource struct {
	Client *api.RestApiClient
}

type proxySettingsResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *ProxySettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_settings"
}

func (r *ProxySettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Proxy Settings Resource.

__Tips:__
* You must be assigned to the following roles:
	* Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/proxy-settings>`,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The name of the proxy host.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the proxy host.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The username for the proxy authentication.",
				Optional:            true,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for the proxy authentication.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("user")),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the proxy settings resource. Used for import and identity purposes. The value is always `proxy-settings`.",
				Computed:            true,
			},
		},
	}
}

func (rs *ProxySettingsResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *ProxySettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *ProxySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.createOrUpdate(ctx, req.Plan, &resp.Diagnostics, &resp.State, &resp.Identity)
}

func (r *ProxySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.ProxySettingsResourceConfig
	var respObj apiobjects.ProxySettings
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetProxySettingsEndpoint()

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := model.ProxySettingsResourceValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := proxySettingsResourceIdentityModel{
		ID: types.StringValue("proxy-settings"),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *ProxySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.createOrUpdate(ctx, req.Plan, &resp.Diagnostics, &resp.State, &resp.Identity)
}

func (r *ProxySettingsResource) createOrUpdate(ctx context.Context, requestPlan tfsdk.Plan, responseDiagnostics *diag.Diagnostics, responseState *tfsdk.State, responseIdentity **tfsdk.ResourceIdentity) {
	var plan model.ProxySettingsResourceConfig
	var respObj apiobjects.ProxySettings
	diags := requestPlan.Get(ctx, &plan)
	responseDiagnostics.Append(diags...)
	if responseDiagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetProxySettingsEndpoint()

	planBody := map[string]any{
		"host": plan.Host.ValueString(),
		"port": plan.Port.ValueInt64(),
	}

	if !plan.User.IsNull() && !plan.User.IsUnknown() {
		planBody["user"] = plan.User.ValueString()
	}

	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		planBody["password"] = plan.Password.ValueString()
	}

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "PUT", endpoint, planBody, false)
	responseDiagnostics.Append(diags...)
	if responseDiagnostics.HasError() {
		return
	}

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoint, nil, true)
	responseDiagnostics.Append(diags...)
	if responseDiagnostics.HasError() {
		return
	}

	responseModel, diags := model.ProxySettingsResourceValueFrom(ctx, plan, respObj)
	responseDiagnostics.Append(diags...)
	if responseDiagnostics.HasError() {
		return
	}

	diags = responseState.Set(ctx, responseModel)
	responseDiagnostics.Append(diags...)
	if responseDiagnostics.HasError() {
		return
	}

	identity := proxySettingsResourceIdentityModel{
		ID: types.StringValue("proxy-settings"),
	}

	diags = (*responseIdentity).Set(ctx, identity)
	responseDiagnostics.Append(diags...)
	if responseDiagnostics.HasError() {
		return
	}
}

func (r *ProxySettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.ProxySettingsResourceConfig
	var respObj apiobjects.ProxySettings
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetProxySettingsEndpoint()

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (rs *ProxySettingsResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if req.ID != "proxy-settings" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Expected import identifier \"proxy-settings\". Got: %q",
				req.ID,
			),
		)
		return
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), "proxy-settings")...,
	)
}
