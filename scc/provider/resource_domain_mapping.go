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

var _ resource.Resource = &DomainMappingResource{}

func NewDomainMappingResource() resource.Resource {
	return &DomainMappingResource{}
}

type DomainMappingResource struct {
	client *api.RestApiClient
}

type domainMappingResourceIdentityModel struct {
	Subaccount     types.String `tfsdk:"subaccount"`
	RegionHost     types.String `tfsdk:"region_host"`
	InternalDomain types.String `tfsdk:"internal_domain"`
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
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
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

func (rs *DomainMappingResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"subaccount": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"region_host": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"internal_domain": identityschema.StringAttribute{
				RequiredForImport: true,
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

	planBody := map[string]any{
		"virtualDomain":  plan.VirtualDomain.ValueString(),
		"internalDomain": plan.InternalDomain.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj.DomainMappings, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mappingRespObj, diags := GetDomainMapping(respObj, internalDomain)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := DomainMappingValueFrom(ctx, plan, *mappingRespObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := domainMappingResourceIdentityModel{
		Subaccount:     plan.Subaccount,
		RegionHost:     plan.RegionHost,
		InternalDomain: plan.InternalDomain,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
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

	diags = requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mappingRespObj, diags := GetDomainMapping(respObj, internalDomain)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := DomainMappingValueFrom(ctx, state, *mappingRespObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := domainMappingResourceIdentityModel{
		Subaccount:     state.Subaccount,
		RegionHost:     state.RegionHost,
		InternalDomain: state.InternalDomain,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
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

	if (state.RegionHost.ValueString() != regionHost) ||
		(state.Subaccount.ValueString() != subaccount) {
		resp.Diagnostics.AddError("Error updating the cloud connector domain mapping", "Failed to update the cloud connector domain mapping due to mismatched configuration values.")
		return
	}
	endpoint := endpoints.GetDomainMappingEndpoint(regionHost, subaccount, internalDomain)

	planBody := map[string]any{
		"virtualDomain":  plan.VirtualDomain.ValueString(),
		"internalDomain": plan.InternalDomain.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj.DomainMappings, "PUT", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint = endpoints.GetDomainMappingBaseEndpoint(regionHost, subaccount)

	diags = requestAndUnmarshal(r.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mappingRespObj, diags := GetDomainMapping(respObj, plan.InternalDomain.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := DomainMappingValueFrom(ctx, plan, *mappingRespObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := domainMappingResourceIdentityModel{
		Subaccount:     plan.Subaccount,
		RegionHost:     plan.RegionHost,
		InternalDomain: plan.InternalDomain,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
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

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := DomainMappingValueFrom(ctx, state, respObj)
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

func (rs *DomainMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
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

		return
	}

	var identity domainMappingResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), identity.Subaccount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), identity.RegionHost)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("internal_domain"), identity.InternalDomain)...)
}
