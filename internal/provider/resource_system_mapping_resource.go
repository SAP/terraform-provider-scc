package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-cloudconnector/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.Resource = &SystemMappingResourceResource{}
var _ resource.ResourceWithImportState = &SystemMappingResourceResource{}

func NewSystemMappingResourceResource() resource.Resource {
	return &SystemMappingResourceResource{}
}

type SystemMappingResourceResource struct {
	client *api.RestApiClient
}

func (r *SystemMappingResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_mapping_resource"
}

func (r *SystemMappingResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cloud Connector System Mapping Resource Resource",
		Attributes: map[string]schema.Attribute{
			"credentials": schema.SingleNestedAttribute{
				MarkdownDescription: "Input parameters required to configure the subaccount connected to cloud connector.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"region_host": schema.StringAttribute{
						MarkdownDescription: "Region Host Name.",
						Required:            true,
					},
					"subaccount": schema.StringAttribute{
						MarkdownDescription: "The ID of the subaccount.",
						Required:            true,
					},
					"virtual_host": schema.StringAttribute{
						MarkdownDescription: "Virtual host used on the cloud side.",
						Required:            true,
					},
					"virtual_port": schema.StringAttribute{
						MarkdownDescription: "Virtual port used on the cloud side.",
						Required:            true,
					},
				},
			},
			"system_mapping_resource": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The resource itself, which, depending on the owning system mapping, is either a URL path (or the leading section of it), or a RFC function name.",
						Required:            true,
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Boolean flag indicating whether the resource is enabled.",
						Computed:            true,
						Optional:            true,
					},
					"exact_match_only": schema.BoolAttribute{
						MarkdownDescription: "Boolean flag determining whether access is granted only if the requested resource is an exact match.",
						Computed:            true,
						Optional:            true,
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
	var plan SystemMappingResourceDataSourceData
	var respObj apiobjects.SystemMappingResourceDataSource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := plan.Credentials.RegionHost.ValueString()
	subaccount := plan.Credentials.Subaccount.ValueString()
	virtual_host := plan.Credentials.VirtualHost.ValueString()
	virtual_port := plan.Credentials.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(plan.SystemMappingResource.ID.ValueString())
	endpoint := endpoints.GetSystemMappingResourceBaseEndpoint(region_host, subaccount, virtual_host, virtual_port)

	planBody := map[string]string{
		"id":                      plan.SystemMappingResource.ID.ValueString(),
		"enabled":                 fmt.Sprintf("%t", plan.SystemMappingResource.Enabled.ValueBool()),
		"exactMatchOnly":          fmt.Sprintf("%t", plan.SystemMappingResource.ExactMatchOnly.ValueBool()),
		"websocketUpgradeAllowed": fmt.Sprintf("%t", plan.SystemMappingResource.WebsocketUpgradeAllowed.ValueBool()),
		"description":             plan.SystemMappingResource.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj.SystemMappingResource, "POST", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error creating the cloud connector system mapping resource", err.Error())
		return
	}

	endpoint = endpoints.GetSystemMappingResourceEndpoint(region_host, subaccount, virtual_host, virtual_port, resource_id)

	err = requestAndUnmarshal(r.client, &respObj.SystemMappingResource, "GET", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mapping resource", err.Error())
		return
	}

	responseModel, err := SystemMappingResourceFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping resource value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *SystemMappingResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemMappingResourceDataSourceData
	var respObj apiobjects.SystemMappingResourceDataSource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.Credentials.RegionHost.ValueString()
	subaccount := state.Credentials.Subaccount.ValueString()
	virtual_host := state.Credentials.VirtualHost.ValueString()
	virtual_port := state.Credentials.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(state.SystemMappingResource.ID.ValueString())
	endpoint := endpoints.GetSystemMappingResourceEndpoint(region_host, subaccount, virtual_host, virtual_port, resource_id)

	err := requestAndUnmarshal(r.client, &respObj.SystemMappingResource, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mapping resource", err.Error())
		return
	}

	responseModel, err := SystemMappingResourceFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping resource value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SystemMappingResourceDataSourceData
	var respObj apiobjects.SystemMappingResourceDataSource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := plan.Credentials.RegionHost.ValueString()
	subaccount := plan.Credentials.Subaccount.ValueString()
	virtual_host := plan.Credentials.VirtualHost.ValueString()
	virtual_port := plan.Credentials.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(plan.SystemMappingResource.ID.ValueString())
	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/systemMappings/%s:%s/resources/%s", region_host, subaccount, virtual_host, virtual_port, resource_id)

	planBody := map[string]string{
		"enabled":                 fmt.Sprintf("%t", plan.SystemMappingResource.Enabled.ValueBool()),
		"exactMatchOnly":          fmt.Sprintf("%t", plan.SystemMappingResource.ExactMatchOnly.ValueBool()),
		"websocketUpgradeAllowed": fmt.Sprintf("%t", plan.SystemMappingResource.WebsocketUpgradeAllowed.ValueBool()),
		"description":             plan.SystemMappingResource.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj.SystemMappingResource, "PUT", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error updating the cloud connector system mapping resource", err.Error())
		return
	}

	endpoint = endpoints.GetSystemMappingResourceEndpoint(region_host, subaccount, virtual_host, virtual_port, resource_id)

	err = requestAndUnmarshal(r.client, &respObj.SystemMappingResource, "GET", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector system mapping resource", err.Error())
		return
	}

	responseModel, err := SystemMappingResourceFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping resource value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemMappingResourceDataSourceData
	var respObj apiobjects.SystemMappingResourceDataSource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.Credentials.RegionHost.ValueString()
	subaccount := state.Credentials.Subaccount.ValueString()
	virtual_host := state.Credentials.VirtualHost.ValueString()
	virtual_port := state.Credentials.VirtualPort.ValueString()
	resource_id := CreateEncodedResourceID(state.SystemMappingResource.ID.ValueString())
	endpoint := endpoints.GetSystemMappingResourceEndpoint(region_host, subaccount, virtual_host, virtual_port, resource_id)

	err := requestAndUnmarshal(r.client, &respObj.SystemMappingResource, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError("error deleting the system mapping resource", err.Error())
		return
	}

	responseModel, err := SystemMappingResourceFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping system mapping resource value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SystemMappingResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
