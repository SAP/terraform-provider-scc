package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.Resource = &DomainMappingResource{}

func NewDomainMappingResource() resource.Resource {
	return &DomainMappingResource{}
}

type DomainMappingResource struct {
	client *api.RestApiClient
}

func (r *DomainMappingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_mapping"
}

func (r *DomainMappingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Domain Mapping Resource.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/domain-mappings>`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "Region Host Name.",
				Required:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The ID of the subaccount.",
				Required:            true,
			},
			"virtual_domain": schema.StringAttribute{
				MarkdownDescription: "Domain used on the cloud side.",
				Required:            true,
			},
			"internal_domain": schema.StringAttribute{
				MarkdownDescription: "Domain used on the on-premise side.",
				Required:            true,
			},
		},
	}
}

func (r *DomainMappingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *DomainMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainMappingConfig
	var respObj apiobjects.DomainMappings
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
	internalDomain := plan.InternalDomain.ValueString()
	endpoint := endpoints.GetDomainMappingBaseEndpoint(regionHost, subaccount)

	planBody := map[string]string{
		"virtualDomain":  plan.VirtualDomain.ValueString(),
		"internalDomain": plan.InternalDomain.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj.DomainMappings, "POST", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgAddDomainMappingFailed, err.Error())
		return
	}

	err = requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchDomainMappingsFailed, err.Error())
		return
	}

	mappingRespObj, err := GetDomainMapping(respObj, internalDomain)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchDomainMappingFailed, fmt.Sprintf("%s", err))
		return
	}

	responseModel, err := DomainMappingValueFrom(ctx, plan, *mappingRespObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapDomainMappingFailed, fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DomainMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainMappingConfig
	var respObj apiobjects.DomainMappings
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	internalDomain := state.InternalDomain.ValueString()
	endpoint := endpoints.GetDomainMappingBaseEndpoint(regionHost, subaccount)

	err := requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchDomainMappingFailed, err.Error())
		return
	}

	mappingRespObj, err := GetDomainMapping(respObj, internalDomain)
	if err != nil {
		resp.Diagnostics.AddError("error getting Domain Mapping", fmt.Sprintf("%s", err))
		return
	}

	responseModel, err := DomainMappingValueFrom(ctx, state, *mappingRespObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping domain mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DomainMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DomainMappingConfig
	var respObj apiobjects.DomainMappings

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
	internalDomain := state.InternalDomain.ValueString()

	if (plan.RegionHost.ValueString() != regionHost) ||
		(plan.Subaccount.ValueString() != subaccount) {
		resp.Diagnostics.AddError("error updating the cloud connector domain mapping.", "Failed to update the cloud connector domain mapping due to mismatched configuration values.")
		return
	}
	endpoint := endpoints.GetDomainMappingEndpoint(regionHost, subaccount, internalDomain)

	planBody := map[string]string{
		"virtualDomain":  plan.VirtualDomain.ValueString(),
		"internalDomain": plan.InternalDomain.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj.DomainMappings, "PUT", endpoint, planBody, false)
	if err != nil {
		resp.Diagnostics.AddError("error updating the cloud connector domain mapping", err.Error())
		return
	}

	endpoint = endpoints.GetDomainMappingBaseEndpoint(regionHost, subaccount)

	err = requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchDomainMappingsFailed, err.Error())
		return
	}

	mappingRespObj, err := GetDomainMapping(respObj, plan.InternalDomain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("error getting Domain Mapping", fmt.Sprintf("%s", err))
		return
	}

	responseModel, err := DomainMappingValueFrom(ctx, plan, *mappingRespObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping domain mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DomainMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainMappingConfig
	var respObj apiobjects.DomainMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	internalDomain := state.InternalDomain.ValueString()
	endpoint := endpoints.GetDomainMappingEndpoint(regionHost, subaccount, internalDomain)

	err := requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError("error deleting the domain mapping", err.Error())
		return
	}

	responseModel, err := DomainMappingValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping domain mapping value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (rs *DomainMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: region_host, subaccount, internal_domain. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("internal_domain"), idParts[2])...)
}
