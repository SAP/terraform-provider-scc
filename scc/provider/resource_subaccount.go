package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SubaccountResource{}

func NewSubaccountResource() resource.Resource {
	return &SubaccountResource{}
}

type SubaccountResource struct {
	client *api.RestApiClient
}

type subaccountResourceIdentityModel struct {
	Subaccount types.String `tfsdk:"subaccount"`
	RegionHost types.String `tfsdk:"region_host"`
}

func (r *SubaccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

func (r *SubaccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subaccount resource.
		
__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/subaccount>`,
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
			"cloud_user": schema.StringAttribute{
				MarkdownDescription: "User for the specified subaccount and region host.\n\n" +
					"**Required when creating the resource.**\n\n" +
					"This attribute is optional in the schema to support `terraform import`, " +
					"but must be provided during creation and certificate renewal operations.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cloud_password": schema.StringAttribute{
				MarkdownDescription: "Password for the cloud user.\n\n" +
					"**Required when creating the resource.**\n\n" +
					"This attribute is optional in the schema to support `terraform import`, " +
					"but must be provided during creation and certificate renewal operations.",
				Sensitive: true,
				Computed:  true,
				Optional:  true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"auto_renew_before_days": schema.Int64Attribute{
				MarkdownDescription: "Number of days before certificate expiration when the certificate should be renewed automatically. Minimum is 7 days, maximum is 45 days.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(14),
				Validators: []validator.Int64{
					int64validator.Between(7, 45),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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

func (rs *SubaccountResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
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

func (r *SubaccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubaccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubaccountConfig
	var respObj apiobjects.SubaccountResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.CloudUser.IsNull() || plan.CloudUser.IsUnknown() ||
		plan.CloudPassword.IsNull() || plan.CloudPassword.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing required credentials",
			"`cloud_user` and `cloud_password` must be provided when creating a subaccount.",
		)
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()

	endpoint := endpoints.GetSubaccountBaseEndpoint()

	planBody := map[string]any{
		"regionHost":    regionHost,
		"subaccount":    subaccount,
		"cloudUser":     plan.CloudUser.ValueString(),
		"cloudPassword": plan.CloudPassword.ValueString(),
		"description":   plan.Description.ValueString(),
		"locationID":    plan.LocationID.ValueString(),
		"displayName":   plan.DisplayName.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectedState := respObj.Tunnel.State == "Connected"

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

	responseModel, diags := SubaccountResourceValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.CloudUser = plan.CloudUser
	responseModel.CloudPassword = plan.CloudPassword
	responseModel.AutoRenewBeforeDays = plan.AutoRenewBeforeDays

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountResourceIdentityModel{
		Subaccount: plan.Subaccount,
		RegionHost: plan.RegionHost,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)

}

func (r *SubaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubaccountConfig
	var respObj apiobjects.SubaccountResource
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

	shouldRenew := shouldRenewCertificate(respObj.Tunnel.SubaccountCertificate.NotAfterTimeStamp, state.AutoRenewBeforeDays.ValueInt64())

	if shouldRenew {
		renewedRespObj, diags := r.renewCertificate(state, regionHost, subaccount)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() && renewedRespObj != nil {
			respObj = *renewedRespObj
			resp.Diagnostics.AddWarning(
				"Certificate Renewed",
				fmt.Sprintf("The subaccount certificate was automatically renewed because it was due to expire within %d days.", state.AutoRenewBeforeDays.ValueInt64()),
			)
		}
	}

	if respObj.Tunnel.State == "Connected" {
		// Trigger trust configuration sync for the subaccount without persisting to Terraform state
		diags = r.syncTrustConfiguration(regionHost, subaccount, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := SubaccountResourceValueFrom(ctx, state, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.CloudUser = state.CloudUser
	responseModel.CloudPassword = state.CloudPassword

	if state.AutoRenewBeforeDays.IsNull() {
		responseModel.AutoRenewBeforeDays = types.Int64Value(14)
	} else {
		responseModel.AutoRenewBeforeDays = state.AutoRenewBeforeDays
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity := subaccountResourceIdentityModel{
		Subaccount: state.Subaccount,
		RegionHost: state.RegionHost,
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (r *SubaccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SubaccountConfig
	var respObj apiobjects.SubaccountResource

	if diags := req.Plan.Get(ctx, &plan); appendAndCheckErrors(&resp.Diagnostics, diags) {
		return
	}

	if diags := req.State.Get(ctx, &state); appendAndCheckErrors(&resp.Diagnostics, diags) {
		return
	}

	diags := validateUpdateInputs(plan, state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	regionHost := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()
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

	if !plan.AutoRenewBeforeDays.IsNull() && !plan.AutoRenewBeforeDays.IsUnknown() {
		if shouldRenewCertificate(respObj.Tunnel.SubaccountCertificate.NotAfterTimeStamp, plan.AutoRenewBeforeDays.ValueInt64()) {
			renewedRespObj, diags := r.renewCertificate(plan, regionHost, subaccount)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() && renewedRespObj != nil {
				respObj = *renewedRespObj
				resp.Diagnostics.AddWarning(
					"Certificate Renewed",
					fmt.Sprintf("The subaccount certificate was automatically renewed during update because it was due to expire within %d days.", plan.AutoRenewBeforeDays.ValueInt64()),
				)
			}
		}
	}

	if respObj.Tunnel.State == "Connected" {
		// Trigger trust configuration sync for the subaccount without persisting to Terraform state
		diags = r.syncTrustConfiguration(regionHost, subaccount, &respObj)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	responseModel, diags := SubaccountResourceValueFrom(ctx, plan, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel.CloudUser = plan.CloudUser
	responseModel.CloudPassword = plan.CloudPassword
	responseModel.AutoRenewBeforeDays = plan.AutoRenewBeforeDays

	resp.Diagnostics.Append(resp.State.Set(ctx, responseModel)...)
}

func appendAndCheckErrors(diags *diag.Diagnostics, newDiags diag.Diagnostics) bool {
	*diags = append(*diags, newDiags...)
	return diags.HasError()
}

func validateUpdateInputs(plan, state SubaccountConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if plan.RegionHost.ValueString() != state.RegionHost.ValueString() ||
		plan.Subaccount.ValueString() != state.Subaccount.ValueString() {
		diags.AddError(
			"Update Failed",
			"failed to update the cloud connector subaccount due to mismatched configuration values",
		)
		return diags
	}
	return diags
}

func (r *SubaccountResource) syncTrustConfiguration(regionHost, subaccount string, respObj *apiobjects.SubaccountResource) diag.Diagnostics {
	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount) + "/trust"

	diags := requestAndUnmarshal(r.client, &respObj, "POST", endpoint, nil, false)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (r *SubaccountResource) updateTunnelState(plan SubaccountConfig, connectedState bool, endpoint string, respObj *apiobjects.SubaccountResource) diag.Diagnostics {
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

func shouldRenewCertificate(expiry, autoRenewBeforeDays int64) bool {
	expiryTime := time.Unix(expiry/1000, 0)
	renewalThreshold := time.Now().Add(time.Duration(autoRenewBeforeDays) * 24 * time.Hour)

	return expiryTime.Before(renewalThreshold)
}

func (r *SubaccountResource) renewCertificate(plan SubaccountConfig, regionHost, subaccount string) (*apiobjects.SubaccountResource, diag.Diagnostics) {
	var respObj apiobjects.SubaccountResource
	var diags diag.Diagnostics

	endpoint := endpoints.GetSubaccountEndpoint(regionHost, subaccount) + "/validity"

	reqBody := map[string]any{
		"user":     plan.CloudUser.ValueString(),
		"password": plan.CloudPassword.ValueString(),
	}

	diags = requestAndUnmarshal(r.client, &respObj, "POST", endpoint, reqBody, true)
	if diags.HasError() {
		return nil, diags
	}

	return &respObj, diags
}

func (r *SubaccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubaccountConfig
	var respObj apiobjects.SubaccountResource
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

func (rs *SubaccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	var identity subaccountResourceIdentityModel
	diags := resp.Identity.Get(ctx, &identity)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subaccount"), identity.Subaccount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_host"), identity.RegionHost)...)
}
