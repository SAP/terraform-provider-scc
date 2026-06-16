package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SubaccountABAPServiceChannelResource{}

func NewSubaccountABAPServiceChannelResource() resource.Resource {
	return &SubaccountABAPServiceChannelResource{}
}

type SubaccountABAPServiceChannelResource struct {
	Client *api.RestApiClient
}

type subaccountABAPServiceChannelResourceIdentityModel struct {
	Subaccount types.String `tfsdk:"subaccount"`
	RegionHost types.String `tfsdk:"region_host"`
	Type       types.String `tfsdk:"type"`
	ID         types.Int64  `tfsdk:"id"`
}

func (r *SubaccountABAPServiceChannelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_abap_service_channel"
}

func (r *SubaccountABAPServiceChannelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount ABAP Service Channel Resource.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

__Operational notes:__
* The SCC API serializes mutations on service channels within the same subaccount using an internal lock.
  Creating multiple ABAP service channels in parallel will fail with a ` + "`ConcurrentModificationException`" + ` (HTTP 400)
  because concurrent requests contend on that lock. Use ` + "`-parallelism=1`" + ` or add explicit ` + "`depends_on`" + `
  between channel resources to serialize creation.
* Channel creation and activation are two separate API calls. If activation fails (e.g. HTTP 500 because SCC cannot
  resolve or reach ` + "`abap_cloud_tenant_host`" + `), the channel already exists in SCC in a **disabled** state.
  The provider saves this partial state so Terraform tracks the resource — no ` + "`terraform import`" + ` is needed.
  Fix the DNS/connectivity issue and re-run ` + "`terraform apply`" + ` to enable the channel.
* Use ` + "`enabled = false`" + ` as the safe default until SCC host DNS/connectivity to the ABAP tenant host is verified.

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/subaccount-service-channels>`,
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
			"snc_encrypted": schema.BoolAttribute{
				MarkdownDescription: "Boolean flag indicating whether the channel is encrypted using SNC (Secure Network Connection).",
				Required:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"abap_cloud_tenant_host": schema.StringAttribute{
				MarkdownDescription: "Host name to access the Host of ABAP Cloud Tenant.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"instance_number": schema.Int64Attribute{
				MarkdownDescription: "Local Instance number under which the ABAP Cloud system is reachable for the client systems.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(00, 99),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Unique identifier for the subaccount service channel (a positive integer number, starting with 1). This identifier is unique across all types of service channels.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of Subaccount Service Channel.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port of the subaccount service channel for the ABAP Cloud System. The port numbers result from the following pattern: `33<LocalInstanceNumber>`, for activated SNC (Secure Network Connection) `48<LocalInstanceNumber>`.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Boolean flag indicating whether the channel is enabled and therefore should be open. " +
					"Defaults to `false`. Setting `enabled = false` is the recommended safe default until SCC host DNS/connectivity " +
					"to `abap_cloud_tenant_host` is verified — activation is a separate API call and will fail with HTTP 500 if SCC " +
					"cannot reach the ABAP tenant host. When activation fails the channel is left in a disabled state in SCC; " +
					"the provider saves this state so no `terraform import` is needed. Fix connectivity and re-apply to enable.",
				Optional: true,
				Computed: true,
			},
			"connections": schema.Int64Attribute{
				MarkdownDescription: "Maximal number of open connections.",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment or short description. This property is not supplied if no comment was provided.",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.SingleNestedAttribute{
				MarkdownDescription: "Current connection state; this property is only available if the channel is enabled.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"connected": schema.BoolAttribute{
						MarkdownDescription: "A Boolean flag indicating whether the channel is connected.",
						Computed:            true,
					},
					"opened_connections": schema.Int64Attribute{
						MarkdownDescription: "The number of open, possibly idle connections.",
						Computed:            true,
					},
					"connected_since_time_stamp": schema.Int64Attribute{
						MarkdownDescription: "The time stamp, a UTC long number, for the first time the channel was opened/connected.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (rs *SubaccountABAPServiceChannelResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"subaccount": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"region_host": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"type": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"id": identityschema.Int64Attribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *SubaccountABAPServiceChannelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *SubaccountABAPServiceChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannels
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()

	var serviceChannelType string
	if plan.SNCEncrypted.ValueBool() {
		serviceChannelType = "ABAPCloudSNC"
	} else {
		serviceChannelType = "ABAPCloud"
	}
	endpoint := endpoints.GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, serviceChannelType)

	planBody := map[string]any{
		"abapCloudTenantHost": plan.ABAPCloudTenantHost.ValueString(),
		"instanceNumber":      fmt.Sprintf("%d", plan.InstanceNumber.ValueInt64()),
		"connections":         fmt.Sprintf("%d", plan.Connections.ValueInt64()),
		"comment":             plan.Comment.ValueString(),
	}

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj.SubaccountABAPServiceChannels, "POST", endpoint, planBody, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj.SubaccountABAPServiceChannels, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceChannelRespObj, diags := r.getSubaccountABAPServiceChannel(respObj, plan.ABAPCloudTenantHost.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := serviceChannelRespObj.ID

	if !plan.Enabled.IsNull() {
		endpoint = endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType, id)
		enableDiags := r.enableSubaccountABAPServiceChannel(plan, *serviceChannelRespObj, endpoint+"/state")
		if enableDiags.HasError() {
			// The channel was created but enabling failed (e.g. HTTP 500 when SCC
			// cannot reach the ABAP tenant host). Save the created-but-disabled state
			// so Terraform tracks the resource and the user can fix connectivity and
			// re-apply (or destroy) without having to run `terraform import` first.
			partialModel, partialDiags := model.SubaccountABAPServiceChannelValueFrom(ctx, plan, *serviceChannelRespObj)
			if !partialDiags.HasError() {
				_ = resp.State.Set(ctx, partialModel)
				_ = resp.Identity.Set(ctx, subaccountABAPServiceChannelResourceIdentityModel{
					Subaccount: plan.Subaccount,
					RegionHost: plan.RegionHost,
					ID:         partialModel.ID,
					Type:       types.StringValue(serviceChannelType),
				})
			}
			resp.Diagnostics.Append(enableDiags...)
			return
		}

		diags = helpers.RequestAndUnmarshal(r.Client, &serviceChannelRespObj, "GET", endpoint, nil, true)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := model.SubaccountABAPServiceChannelValueFrom(ctx, plan, *serviceChannelRespObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.SNCEncrypted = plan.SNCEncrypted

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountABAPServiceChannelResourceIdentityModel{
		Subaccount: plan.Subaccount,
		RegionHost: plan.RegionHost,
		ID:         responseModel.ID,
		Type:       types.StringValue(serviceChannelType),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubaccountABAPServiceChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	id := state.ID.ValueInt64()

	var serviceChannelType string
	if state.SNCEncrypted.ValueBool() {
		serviceChannelType = "ABAPCloudSNC"
	} else {
		serviceChannelType = "ABAPCloud"
	}
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType, id)

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := model.SubaccountABAPServiceChannelValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.SNCEncrypted = state.SNCEncrypted

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountABAPServiceChannelResourceIdentityModel{
		Subaccount: state.Subaccount,
		RegionHost: state.RegionHost,
		ID:         state.ID,
		Type:       types.StringValue(serviceChannelType),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubaccountABAPServiceChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state model.SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannel

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
	id := state.ID.ValueInt64()

	if (state.RegionHost.ValueString() != regionHost) ||
		(state.Subaccount.ValueString() != subaccount) {
		resp.Diagnostics.AddError("Error updating the cloud connector subaccount ABAP service channel", "Failed to update the cloud connector ABAP service channel due to mismatched configuration values.")
		return
	}
	// Update Service Channel
	var serviceChannelType string
	if state.SNCEncrypted.ValueBool() {
		serviceChannelType = "ABAPCloudSNC"
	} else {
		serviceChannelType = "ABAPCloud"
	}
	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType, id)
	diags = r.updateSubaccountABAPServiceChannel(plan, respObj, endpoint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Enable/Disable Service Channel
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		diags = r.enableSubaccountABAPServiceChannel(plan, respObj, endpoint+"/state")
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	endpoint = endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType, id)

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := model.SubaccountABAPServiceChannelValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.SNCEncrypted = plan.SNCEncrypted

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountABAPServiceChannelResourceIdentityModel{
		Subaccount: state.Subaccount,
		RegionHost: state.RegionHost,
		ID:         state.ID,
		Type:       types.StringValue(serviceChannelType),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubaccountABAPServiceChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.SubaccountABAPServiceChannelConfig
	var respObj apiobjects.SubaccountABAPServiceChannel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	id := state.ID.ValueInt64()

	var serviceChannelType string
	if state.SNCEncrypted.ValueBool() {
		serviceChannelType = "ABAPCloudSNC"
	} else {
		serviceChannelType = "ABAPCloud"
	}

	endpoint := endpoints.GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType, id)

	diags = helpers.RequestAndUnmarshal(r.Client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *SubaccountABAPServiceChannelResource) getSubaccountABAPServiceChannel(serviceChannels apiobjects.SubaccountABAPServiceChannels, targetABAPCloudTenantHost string) (*apiobjects.SubaccountABAPServiceChannel, diag.Diagnostics) {
	var diags diag.Diagnostics
	for _, channel := range serviceChannels.SubaccountABAPServiceChannels {
		if channel.ABAPCloudTenantHost == targetABAPCloudTenantHost {
			return &channel, nil
		}
	}
	diags.AddError("Subaccount ABAP Service Channel Not Found", "The specified subaccount ABAP service channel with the given ABAP Cloud Tenant Host was not found.")
	return nil, diags
}

func (r *SubaccountABAPServiceChannelResource) enableSubaccountABAPServiceChannel(plan model.SubaccountABAPServiceChannelConfig, respObj apiobjects.SubaccountABAPServiceChannel, endpoint string) diag.Diagnostics {
	planBody := map[string]any{
		"enabled": fmt.Sprintf("%t", plan.Enabled.ValueBool()),
	}

	diags := helpers.RequestAndUnmarshal(r.Client, &respObj, "PUT", endpoint, planBody, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (r *SubaccountABAPServiceChannelResource) updateSubaccountABAPServiceChannel(plan model.SubaccountABAPServiceChannelConfig, respObj apiobjects.SubaccountABAPServiceChannel, endpoint string) diag.Diagnostics {
	planBody := map[string]any{
		"abapCloudTenantHost": plan.ABAPCloudTenantHost.ValueString(),
		"instanceNumber":      fmt.Sprintf("%d", plan.InstanceNumber.ValueInt64()),
		"connections":         fmt.Sprintf("%d", plan.Connections.ValueInt64()),
		"comment":             plan.Comment.ValueString(),
	}

	diags := helpers.RequestAndUnmarshal(r.Client, &respObj, "PUT", endpoint, planBody, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (rs *SubaccountABAPServiceChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		idParts := strings.Split(req.ID, ",")

		if len(idParts) != 4 ||
			idParts[0] == "" ||
			idParts[1] == "" ||
			idParts[2] == "" ||
			idParts[3] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: region_host, subaccount, type, id. Got: %q", req.ID),
			)
			return
		}

		regionHost := strings.TrimSpace(idParts[0])
		subaccount := strings.TrimSpace(idParts[1])
		serviceChannelType := strings.TrimSpace(idParts[2])
		idStr := strings.TrimSpace(idParts[3])

		var sncEncrypted bool
		switch serviceChannelType {
		case "ABAPCloudSNC":
			sncEncrypted = true
		case "ABAPCloud":
			sncEncrypted = false
		default:
			resp.Diagnostics.AddError(
				"Invalid Service Channel Type",
				fmt.Sprintf("The 'type' part of the import identifier must be either 'ABAPCloud' or 'ABAPCloudSNC'. Got: %q", serviceChannelType),
			)
			return
		}

		intID, err := strconv.Atoi(idStr)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid ID Format",
				fmt.Sprintf("The 'id' part must be an integer. Got: %q", idStr),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), regionHost)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), subaccount)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), serviceChannelType)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snc_encrypted"), sncEncrypted)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), intID)...)

		return
	}

	var identity subaccountABAPServiceChannelResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), identity.Subaccount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), identity.RegionHost)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), identity.Type)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("snc_encrypted"), identity.Type.ValueString() == "ABAPCloudSNC")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), identity.ID)...)

}
