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
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SubaccountUsingAuthResource{}

func NewSubaccountUsingAuthResource() resource.Resource {
	return &SubaccountUsingAuthResource{}
}

type SubaccountUsingAuthResource struct {
	client *api.RestApiClient
}

type subaccountUsingAuthResourceIdentityModel struct {
	Subaccount types.String `tfsdk:"subaccount"`
	RegionHost types.String `tfsdk:"region_host"`
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
				MarkdownDescription: `Subaccount authentication data, used instead of providing cloud_user, cloud_password, subaccount and region_host (as of version 2.17.0).
				This attribute is defined as **optional** in the schema to support terraform import and updates to existing resources.
				However, it is **required during resource creation** when using authentication-based onboarding.
				This value must be downloaded from the subaccount and used within **5 minutes**, as it expires shortly after generation. 
				It is used only during **resource creation**.  

**Note:**  
- This value **will be persisted** in the Terraform state file. It is the user's responsibility to keep the state file secure.  
- Updating this value will **force resource recreation**.
- This value is **not required** for updating optional attributes such as location_id, display_name, description or tunnel settings.`,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
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
			"connected": schema.BoolAttribute{
				MarkdownDescription: `Specifies whether the subaccount should be connected to the Cloud Connector.

- **true** → attempts to establish a tunnel connection.
- **false** → disconnects the subaccount from the Cloud Connector.

The value is persisted in state based on what you configure (not overwritten by runtime status).  
The actual tunnel status is reported by the Cloud Connector and may differ:

- *Connected* → tunnel established successfully.
- *Disconnected* → tunnel was intentionally or unintentionally closed.
- *ConnectFailure* → tunnel could not be established (e.g., invalid credentials, network issues).  

**Important:**  
In case of *ConnectFailure*, the provider will issue a warning but will **not reset** the value of connected.  
To recover, set connected = false, apply, and then set it back to true to retry the connection.`,
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"tunnel": schema.SingleNestedAttribute{
				MarkdownDescription: "Details of connection tunnel used by the subaccount.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						MarkdownDescription: "State of the tunnel. Possible values are: \n" +
							getFormattedValueAsTableRow("state", "description") +
							getFormattedValueAsTableRow("---", "---") +
							getFormattedValueAsTableRow("`Connected`", "The tunnel is active and functioning properly.") +
							getFormattedValueAsTableRow("`ConnectFailure`", "The tunnel failed to establish a connection due to an issue.") +
							getFormattedValueAsTableRow("`Disconnected`", "The tunnel was previously connected but is now intentionally or unintentionally disconnected."),
						Computed: true,
					},
					"connected_since": schema.StringAttribute{
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
							"valid_to": schema.StringAttribute{
								MarkdownDescription: "Timestamp of the end of the validity period.",
								Computed:            true,
							},
							"valid_from": schema.StringAttribute{
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

func (rs *SubaccountUsingAuthResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"subaccount": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"region_host": identityschema.StringAttribute{
				RequiredForImport: true,
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

	if plan.AuthenticationData.IsNull() || plan.AuthenticationData.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing required credentials",
			"`authentication_data` must be provided when creating a subaccount.",
		)
		return
	}

	endpoint := endpoints.GetSubaccountBaseEndpoint()

	planBody := map[string]any{
		"authenticationData": plan.AuthenticationData.ValueString(),
		"description":        plan.Description.ValueString(),
		"locationID":         plan.LocationID.ValueString(),
		"displayName":        plan.DisplayName.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectedState := respObj.Tunnel.State == "Connected"
	regionHost := respObj.RegionHost
	subaccount := respObj.Subaccount

	if !plan.Connected.IsNull() && !plan.Connected.IsUnknown() {
		endpoint = endpoints.GetSubaccountEndpoint(regionHost, subaccount)
		diags = r.updateTunnelState(plan, connectedState, endpoint, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if respObj.Tunnel.State == "ConnectFailure" {
		resp.Diagnostics.AddWarning(
			"Tunnel connection failed",
			"The subaccount was created/updated successfully, but the tunnel could not be established (state=ConnectFailure). "+
				"You can retry by toggling 'connected' from false to true.",
		)
	}

	if respObj.Tunnel.State == "Connected" {
		// Trigger trust configuration sync for the subaccount without persisting to Terraform state
		diags = r.syncTrustConfiguration(regionHost, subaccount, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.AuthenticationData = plan.AuthenticationData

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountUsingAuthResourceIdentityModel{
		Subaccount: responseModel.Subaccount,
		RegionHost: responseModel.RegionHost,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
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

	diags = requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if respObj.Tunnel.State == "Connected" {
		// Trigger trust configuration sync for the subaccount without persisting to Terraform state
		diags = r.syncTrustConfiguration(regionHost, subaccount, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.AuthenticationData = state.AuthenticationData

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountUsingAuthResourceIdentityModel{
		Subaccount: responseModel.Subaccount,
		RegionHost: responseModel.RegionHost,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
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

	diags := validateAuthDataInputs(plan, state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	regionHost := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount)

	updateBody := map[string]any{
		"locationID":  plan.LocationID.ValueString(),
		"displayName": plan.DisplayName.ValueString(),
		"description": plan.Description.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, updateBody, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectedState := respObj.Tunnel.State == "Connected"

	if !plan.Connected.IsNull() && !plan.Connected.IsUnknown() {
		diags = r.updateTunnelState(plan, connectedState, endpoint, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if respObj.Tunnel.State == "ConnectFailure" {
		resp.Diagnostics.AddWarning(
			"Tunnel connection failed",
			"The subaccount was created/updated successfully, but the tunnel could not be established (state=ConnectFailure). "+
				"You can retry by toggling 'connected' from false to true.",
		)
	}

	if respObj.Tunnel.State == "Connected" {
		// Trigger trust configuration sync for the subaccount without persisting to Terraform state
		diags = r.syncTrustConfiguration(regionHost, subaccount, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := SubaccountUsingAuthResourceValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	responseModel.AuthenticationData = plan.AuthenticationData

	resp.Diagnostics.Append(resp.State.Set(ctx, responseModel)...)

	identity := subaccountUsingAuthResourceIdentityModel{
		Subaccount: state.Subaccount,
		RegionHost: state.RegionHost,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func appendAndCheckErrorsCopy(diags *diag.Diagnostics, newDiags diag.Diagnostics) bool {
	*diags = append(*diags, newDiags...)
	return diags.HasError()
}

func (r *SubaccountUsingAuthResource) syncTrustConfiguration(regionHost, subaccount string, respObj *apiobjects.SubaccountUsingAuthResource) diag.Diagnostics {
	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount) + "/trust"

	diags := requestAndUnmarshal(r.client, &respObj, "POST", endpoint, nil, false)
	if diags.HasError() {
		return diags
	}

	return diags
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

	diags = requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountUsingAuthResource) updateTunnelState(plan SubaccountUsingAuthConfig, connectedState bool, endpoint string, respObj *apiobjects.SubaccountUsingAuthResource) diag.Diagnostics {
	var diags diag.Diagnostics
	// Check if the desired state is different from the current state
	desiredState := plan.Connected.ValueBool()
	if desiredState == connectedState {
		return diags
	}
	// Update the tunnel state
	patch := map[string]any{"connected": desiredState}

	diags = requestAndUnmarshal(r.client, respObj, "PUT", endpoint+"/state", patch, false)
	if diags.HasError() {
		return diags
	}

	// Re-fetch to update tunnel state
	diags = requestAndUnmarshal(r.client, respObj, "GET", endpoint, nil, true)
	if diags.HasError() {
		return diags
	}

	return diags
}

func validateAuthDataInputs(plan, state SubaccountUsingAuthConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if plan.AuthenticationData.ValueString() != state.AuthenticationData.ValueString() {
		diags.AddError(
			"Update Failed",
			"failed to update the cloud connector subaccount due to mismatched configuration values",
		)
		return diags
	}
	return diags
}

func (rs *SubaccountUsingAuthResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
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

		return
	}

	var identity subaccountUsingAuthResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), identity.Subaccount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), identity.RegionHost)...)
}
