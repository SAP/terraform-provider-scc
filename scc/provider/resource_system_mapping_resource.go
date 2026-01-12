package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SystemMappingResourceResource{}

func NewSystemMappingResourceResource() resource.Resource {
	return &SystemMappingResourceResource{}
}

type SystemMappingResourceResource struct {
	client *api.RestApiClient
}

type systemMappingResourceResourceIdentityModel struct {
	Subaccount  types.String `tfsdk:"subaccount"`
	RegionHost  types.String `tfsdk:"region_host"`
	VirtualHost types.String `tfsdk:"virtual_host"`
	VirtualPort types.String `tfsdk:"virtual_port"`
	URLPath     types.String `tfsdk:"url_path"`
}

func (r *SystemMappingResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping_resource"
}

func (r *SystemMappingResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector System Mapping Resource Resource.
				
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/system-mapping-resources>`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "Region Host Name.",
				Required:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The ID of the subaccount.",
				Required:            true,
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
			},
			"virtual_host": schema.StringAttribute{
				MarkdownDescription: "Virtual host used on the cloud side.",
				Required:            true,
			},
			"virtual_port": schema.StringAttribute{
				MarkdownDescription: "Virtual port used on the cloud side.",
				Required:            true,
			},
			"url_path": schema.StringAttribute{
				MarkdownDescription: "The resource itself, which, depending on the owning system mapping, is either a URL path (or the leading section of it), or a RFC function name.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Boolean flag indicating whether the resource is enabled.",
				Computed:            true,
				Optional:            true,
			},
			"path_only": schema.BoolAttribute{
				MarkdownDescription: `Boolean flag determining whether access is granted only if the requested resource is an exact match.
				
__UI Equivalent:__ *Access Policy*

- true → *Path Only (Sub-Paths Are Excluded)*
- false → *Path And All Sub-Paths*`,
				Computed: true,
				Optional: true,
			},
			"websocket_upgrade_allowed": schema.BoolAttribute{
				MarkdownDescription: "Boolean flag indicating whether websocket upgrade is allowed. This property is of relevance only if the owning system mapping employs protocol HTTP or HTTPS.",
				Computed:            true,
				Optional:            true,
			},
			"creation_date": schema.StringAttribute{
				MarkdownDescription: "Date of creation of system mapping resource.",
				Computed:            true,
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the system mapping resource.",
				Computed:            true,
				Optional:            true,
			},
		},
	}
}

func (rs *SystemMappingResourceResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"subaccount": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"region_host": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"virtual_host": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"virtual_port": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"url_path": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *SystemMappingResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SystemMappingResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	virtualHost := plan.VirtualHost.ValueString()
	virtualPort := plan.VirtualPort.ValueString()
	resourceID := CreateEncodedResourceID(plan.URLPath.ValueString())
	endpoint := endpoints.GetSystemMappingResourceBaseEndpoint(regionHost, subaccount, virtualHost, virtualPort)

	planBody := map[string]any{
		"id":                      plan.URLPath.ValueString(),
		"enabled":                 fmt.Sprintf("%t", plan.Enabled.ValueBool()),
		"exactMatchOnly":          fmt.Sprintf("%t", plan.PathOnly.ValueBool()),
		"websocketUpgradeAllowed": fmt.Sprintf("%t", plan.WebsocketUpgradeAllowed.ValueBool()),
		"description":             plan.Description.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint = endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resourceID)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, planBody, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingResourceValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := systemMappingResourceResourceIdentityModel{
		Subaccount:  plan.Subaccount,
		RegionHost:  plan.RegionHost,
		VirtualHost: plan.VirtualHost,
		VirtualPort: plan.VirtualPort,
		URLPath:     plan.URLPath,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SystemMappingResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	virtualHost := state.VirtualHost.ValueString()
	virtualPort := state.VirtualPort.ValueString()
	resourceID := CreateEncodedResourceID(state.URLPath.ValueString())
	endpoint := endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resourceID)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingResourceValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := systemMappingResourceResourceIdentityModel{
		Subaccount:  state.Subaccount,
		RegionHost:  state.RegionHost,
		VirtualHost: state.VirtualHost,
		VirtualPort: state.VirtualPort,
		URLPath:     state.URLPath,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SystemMappingResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
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

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	virtualHost := plan.VirtualHost.ValueString()
	virtualPort := plan.VirtualPort.ValueString()
	resourceID := CreateEncodedResourceID(plan.URLPath.ValueString())

	if (state.RegionHost.ValueString() != regionHost) ||
		(state.Subaccount.ValueString() != subaccount) ||
		(state.VirtualHost.ValueString() != virtualHost) ||
		(state.VirtualPort.ValueString() != virtualPort) {
		resp.Diagnostics.AddError("Error updating the cloud connector system mapping resource", "Failed to update the cloud connector system mapping resource due to mismatched configuration values.")
		return
	}
	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/systemMappings/%s:%s/resources/%s", regionHost, subaccount, virtualHost, virtualPort, resourceID)

	planBody := map[string]any{
		"enabled":                 fmt.Sprintf("%t", plan.Enabled.ValueBool()),
		"exactMatchOnly":          fmt.Sprintf("%t", plan.PathOnly.ValueBool()),
		"websocketUpgradeAllowed": fmt.Sprintf("%t", plan.WebsocketUpgradeAllowed.ValueBool()),
		"description":             plan.Description.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint = endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resourceID)

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, planBody, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingResourceValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := systemMappingResourceResourceIdentityModel{
		Subaccount:  plan.Subaccount,
		RegionHost:  plan.RegionHost,
		VirtualHost: plan.VirtualHost,
		VirtualPort: plan.VirtualPort,
		URLPath:     plan.URLPath,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SystemMappingResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemMappingResourceConfig
	var respObj apiobjects.SystemMappingResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	virtualHost := state.VirtualHost.ValueString()
	virtualPort := state.VirtualPort.ValueString()
	resourceID := CreateEncodedResourceID(state.URLPath.ValueString())
	endpoint := endpoints.GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resourceID)

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := SystemMappingResourceValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (rs *SystemMappingResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		idParts := strings.Split(req.ID, ",")

		if len(idParts) != 5 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: region_host, subaccount, virtual_host, virtual_port, url_path. Got: %q", req.ID),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_host"), idParts[2])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_port"), idParts[3])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("url_path"), idParts[4])...)

		return

	}

	var identity systemMappingResourceResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), identity.RegionHost)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), identity.Subaccount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_host"), identity.VirtualHost)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_port"), identity.VirtualPort)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("url_path"), identity.URLPath)...)
}
