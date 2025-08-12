package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &SubaccountUsingAuthResource{}

func NewSubaccountUsingAuthResource() resource.Resource {
	return &SubaccountUsingAuthResource{}
}

type SubaccountUsingAuthResource struct {
	client *api.RestApiClient
}

func (r *SubaccountUsingAuthResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount_using_auth"
}

func (r *SubaccountUsingAuthResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount resource using Authentication Data.
		
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/subaccount>`,
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "Region Host Name.",
				Computed:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The ID of the subaccount.",
				Computed:            true,
				Validators: []validator.String{
					uuidvalidator.ValidUUID(),
				},
			},
			"authentication_data": schema.StringAttribute{
				MarkdownDescription: `Subaccount authentication data, used instead of cloud_user, cloud_password, subaccount and region_host (as of version 2.17.0).
This value must be downloaded from the subaccount and used within **5 minutes**, as it expires shortly after generation. It is used only during **resource creation** and 
is **not required** for updating optional attributes such as location_id, display_name, description or tunnel.  

**Note:**  
- This value **will be persisted** in the Terraform state file. It is the user's responsibility to keep the state file secure.  
- If this value is updated, **the resource will be recreated**.`,
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"location_id": schema.StringAttribute{
				MarkdownDescription: "Location identifier for the Cloud Connector instance.",
				Computed:            true,
				Optional:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name of the subaccount.",
				Computed:            true,
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the subaccount.",
				Computed:            true,
				Optional:            true,
			},
			"tunnel": schema.SingleNestedAttribute{
				MarkdownDescription: "Details of connection tunnel used by the subaccount.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						MarkdownDescription: "State of the tunnel. Possible values are: \n" +
							getFormattedValueAsTableRow("state", "description") +
							getFormattedValueAsTableRow("---", "---") +
							getFormattedValueAsTableRow("`Connected`", "The tunnel is active and functioning properly.") +
							getFormattedValueAsTableRow("`ConnectFailure`", "The tunnel failed to establish a connection due to an issue.") +
							getFormattedValueAsTableRow("`Disconnected`", "The tunnel was previously connected but is now intentionally or unintentionally disconnected."),
						Optional: true,
						Computed: true,
					},
					"connected_since_time_stamp": schema.Int64Attribute{
						MarkdownDescription: "Timestamp of the start of the connection.",
						Computed:            true,
					},
					"connections": schema.Int64Attribute{
						MarkdownDescription: "Number of subaccount connections.",
						Computed:            true,
					},
					"subaccount_certificate": schema.SingleNestedAttribute{
						MarkdownDescription: "Information on the subaccount certificate such as validity period, issuer and subject DN.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"not_after_time_stamp": schema.Int64Attribute{
								MarkdownDescription: "Timestamp of the end of the validity period.",
								Computed:            true,
							},
							"not_before_time_stamp": schema.Int64Attribute{
								MarkdownDescription: "Timestamp of the beginning of the validity period.",
								Computed:            true,
							},
							"subject_dn": schema.StringAttribute{
								MarkdownDescription: "The subject distinguished name.",
								Computed:            true,
							},
							"issuer": schema.StringAttribute{
								MarkdownDescription: "Certificate authority (CA) that issued this certificate.",
								Computed:            true,
							},
							"serial_number": schema.StringAttribute{
								MarkdownDescription: "Unique identifier for the certificate, typically assigned by the CA.",
								Computed:            true,
							},
						},
					},
					"application_connections": schema.ListNestedAttribute{
						MarkdownDescription: "Array of connections to application instances. Each connection provides information about a specific application instance accessible through the cloud connector.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"connection_count": schema.Int64Attribute{
									MarkdownDescription: "Number of active connections to the specified application instance.",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the connected application instance.",
									Computed:            true,
								},
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of the connected application instance.",
									Computed:            true,
								},
							},
						},
					},
					"service_channels": schema.ListNestedAttribute{
						MarkdownDescription: "Type and state of the service channels used (types: HANA database, Virtual Machine or RFC)",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of the service channel (e.g., HANA, VM, or RFC).",
									Computed:            true,
								},
								"state": schema.StringAttribute{
									MarkdownDescription: "Current operational state of the service channel.",
									Computed:            true,
								},
								"details": schema.StringAttribute{
									MarkdownDescription: "Technical details about the service channel.",
									Computed:            true,
								},
								"comment": schema.StringAttribute{
									MarkdownDescription: "Optional user-provided comment or annotation regarding the service channel.",
									Computed:            true,
								},
							},
						},
					},
					"user": schema.StringAttribute{
						MarkdownDescription: "User for the specified region host and subaccount.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (r *SubaccountUsingAuthResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubaccountUsingAuthResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubaccountUsingAuthConfig
	var respObj apiobjects.SubaccountUsingAuthResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSubaccountBaseEndpoint()

	planBody := map[string]string{
		"authenticationData": plan.AuthenticationData.ValueString(),
		"description":        plan.Description.ValueString(),
		"locationID":         plan.LocationID.ValueString(),
		"displayName":        plan.DisplayName.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgAddSubaccountFailed, err.Error())
		return
	}

	// Trigger trust configuration sync for the subaccount without persisting to Terraform state
	regionHost := respObj.RegionHost
	subaccount := respObj.Subaccount

	endpoint = endpoints.GetSubaccountEndpoint(regionHost, subaccount) + "/trust"

	err = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgAddSubaccountFailed, err.Error())
		return
	}

	responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, plan, respObj)
	if diags.HasError() {
		resp.Diagnostics.AddError(errMsgMapSubaccountFailed, fmt.Sprintf("%s", diags))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountUsingAuthResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubaccountUsingAuthConfig
	var respObj apiobjects.SubaccountUsingAuthResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount)

	err := requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchSubaccountFailed, err.Error())
		return
	}

	// Trigger trust configuration sync for the subaccount without persisting to Terraform state
	endpoint = endpoint + "/trust"

	err = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgAddSubaccountFailed, err.Error())
		return
	}

	responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, state, respObj)
	if diags.HasError() {
		resp.Diagnostics.AddError(errMsgMapSubaccountFailed, fmt.Sprintf("%s", diags))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountUsingAuthResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SubaccountUsingAuthConfig
	var respObj apiobjects.SubaccountUsingAuthResource

	if diags := req.Plan.Get(ctx, &plan); appendAndCheckErrorsCopy(&resp.Diagnostics, diags) {
		return
	}

	if diags := req.State.Get(ctx, &state); appendAndCheckErrorsCopy(&resp.Diagnostics, diags) {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount)

	planBody := map[string]string{
		"locationID":  plan.LocationID.ValueString(),
		"displayName": plan.DisplayName.ValueString(),
		"description": plan.Description.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgUpdateSubaccountFailed, err.Error())
		return
	}

	if shouldUpdateTunnelCopy(plan) {
		if err := r.updateTunnelState(ctx, plan, state, endpoint, &respObj, &resp.Diagnostics); err != nil {
			return
		}
	}

	// Trigger trust configuration sync for the subaccount without persisting to Terraform state
	endpoint = endpoint + "/trust"

	err = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgAddSubaccountFailed, err.Error())
		return
	}

	if responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, plan, respObj); diags.HasError() {
		resp.Diagnostics.AddError(errMsgMapSubaccountFailed, fmt.Sprintf("%s", diags))
	} else {
		resp.Diagnostics.Append(resp.State.Set(ctx, responseModel)...)
	}
}

func appendAndCheckErrorsCopy(diags *diag.Diagnostics, newDiags diag.Diagnostics) bool {
	*diags = append(*diags, newDiags...)
	return diags.HasError()
}

func shouldUpdateTunnelCopy(plan SubaccountUsingAuthConfig) bool {
	return !plan.Tunnel.IsNull() && !plan.Tunnel.IsUnknown()
}

func (r *SubaccountUsingAuthResource) updateTunnelState(ctx context.Context, plan, state SubaccountUsingAuthConfig, endpoint string, respObj *apiobjects.SubaccountUsingAuthResource, diagnostics *diag.Diagnostics) error {
	var planTunnel, stateTunnel SubaccountTunnelData

	if diags := state.Tunnel.As(ctx, &stateTunnel, basetypes.ObjectAsOptions{}); appendAndCheckErrors(diagnostics, diags) {
		return fmt.Errorf("error reading state tunnel")
	}
	if diags := plan.Tunnel.As(ctx, &planTunnel, basetypes.ObjectAsOptions{}); appendAndCheckErrors(diagnostics, diags) {
		return fmt.Errorf("error reading plan tunnel")
	}

	desiredState := planTunnel.State.ValueString()
	if desiredState == stateTunnel.State.ValueString() {
		return nil
	}

	connected := desiredState != "Disconnected"
	patch := map[string]string{"connected": fmt.Sprintf("%t", connected)}

	if err := requestAndUnmarshal(r.client, respObj, "PUT", endpoint+"/state", patch, false); err != nil {
		diagnostics.AddError(errMsgUpdateSubaccountFailed, err.Error())
		return err
	}

	// Re-fetch to update tunnel state
	if err := requestAndUnmarshal(r.client, respObj, "GET", endpoint, nil, true); err != nil {
		diagnostics.AddError(errMsgUpdateSubaccountFailed, err.Error())
		return err
	}

	return nil
}

func (r *SubaccountUsingAuthResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubaccountUsingAuthConfig
	var respObj apiobjects.SubaccountUsingAuthResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()

	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount)

	err := requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError(errMsgDeleteSubaccountFailed, err.Error())
		return
	}

	responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, state, respObj)
	if diags.HasError() {
		resp.Diagnostics.AddError(errMsgMapSubaccountFailed, fmt.Sprintf("%s", diags))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (rs *SubaccountUsingAuthResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: region_host, subaccount. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), idParts[1])...)
}
