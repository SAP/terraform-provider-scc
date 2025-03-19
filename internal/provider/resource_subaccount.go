package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-cloudconnector/internal/api"
	apiobjects "github.com/SAP/terraform-provider-cloudconnector/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-cloudconnector/internal/api/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.Resource = &SubaccountResource{}

func NewSubaccountResource() resource.Resource {
	return &SubaccountResource{}
}

type SubaccountResource struct {
	client *api.RestApiClient
}

func (r *SubaccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

func (r *SubaccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cloud Connector Subaccount resource",
		Attributes: map[string]schema.Attribute{
			"region_host": schema.StringAttribute{
				MarkdownDescription: "Region Host Name.",
				Required:            true,
			},
			"subaccount": schema.StringAttribute{
				MarkdownDescription: "The ID of the subaccount.",
				Required:            true,
			},
			"cloud_user": schema.StringAttribute{
				MarkdownDescription: "User for the specified subaccount and region host.",
				Required:            true,
			},
			"cloud_password": schema.StringAttribute{
				MarkdownDescription: "Password for the cloud user.",
				Sensitive:           true,
				Required:            true,
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
				MarkdownDescription: "Array of connection tunnels used by the subaccount.",
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
					"connected_since_time_stamp": schema.Int64Attribute{
						MarkdownDescription: "Timestamp of the start of the connection.",
						Computed:            true,
					},
					"connections": schema.Int64Attribute{
						MarkdownDescription: "Number of subaccount connections.",
						Computed:            true,
					},
					// "service_channels": schema.ListNestedAttribute{
					// 	MarkdownDescription: "Type and state of the service channels used.",
					// 	Computed:            true,
					// 	NestedObject: schema.NestedAttributeObject{
					// 		Attributes: map[string]schema.Attribute{
					// 			"type": schema.StringAttribute{
					// 				MarkdownDescription: "Type of Subaccount Service Channel.",
					// 				Computed:            true,
					// 			},
					// 			"state": schema.StringAttribute{
					// 				MarkdownDescription: "Current connection state.",
					// 				Computed:            true,
					// 			},
					// 			"details": schema.StringAttribute{
					// 				MarkdownDescription: "Details about the Subaccount Service Channel.",
					// 				Computed:            true,
					// 			},
					// 			"comment": schema.StringAttribute{
					// 				MarkdownDescription: "Comment or short description.",
					// 				Computed:            true,
					// 			},
					// 		},
					// 	},
					// },
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
					"user": schema.StringAttribute{
						MarkdownDescription: "User for the specified region host and subaccount.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (r *SubaccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

	endpoint := endpoints.GetSubaccountBaseEndpoint()

	planBody := map[string]string{
		"regionHost":    plan.RegionHost.ValueString(),
		"subaccount":    plan.Subaccount.ValueString(),
		"cloudUser":     plan.CloudUser.ValueString(),
		"cloudPassword": plan.CloudPassword.ValueString(),
		"description":   plan.Description.ValueString(),
		"locationID":    plan.LocationID.ValueString(),
		"displayName":   plan.DisplayName.ValueString(),
	}

	err := requestAndUnmarshal(r.client, &respObj, "POST", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError("error creating the cloud connector subaccount.", err.Error())
		return
	}

	responseModel, err := SubaccountResourceValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *SubaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubaccountConfig
	var respObj apiobjects.SubaccountResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()
	endpoint := endpoints.GetSubaccountEndpoint(region_host, subaccount)

	err := requestAndUnmarshal(r.client, &respObj, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError("error fetching the cloud connector subaccount", err.Error())
		return
	}

	responseModel, err := SubaccountResourceValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SubaccountConfig
	var respObj apiobjects.SubaccountResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	region_host := plan.RegionHost.ValueString()
	subaccount := plan.Subaccount.ValueString()

	planBody := map[string]string{
		"locationID":  plan.LocationID.ValueString(),
		"displayName": plan.DisplayName.ValueString(),
		"description": plan.Description.ValueString(),
	}

	endpoint := endpoints.GetSubaccountEndpoint(region_host, subaccount)

	err := requestAndUnmarshal(r.client, &respObj, "PUT", endpoint, planBody, true)
	if err != nil {
		resp.Diagnostics.AddError("error updating the cloud connector subaccount.", err.Error())
		return
	}

	responseModel, err := SubaccountResourceValueFrom(ctx, plan, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount value", fmt.Sprintf("%s", err))
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubaccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubaccountConfig
	var respObj apiobjects.SubaccountResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// endpoint:= "/api/v1/configuration/subaccounts/%s/%s"
	region_host := state.RegionHost.ValueString()
	subaccount := state.Subaccount.ValueString()

	endpoint := endpoints.GetSubaccountEndpoint(region_host, subaccount)

	err := requestAndUnmarshal(r.client, &respObj, "DELETE", endpoint, nil, false)
	if err != nil {
		resp.Diagnostics.AddError("error deleting the subaccount", err.Error())
		return
	}

	responseModel, err := SubaccountResourceValueFrom(ctx, state, respObj)
	if err != nil {
		resp.Diagnostics.AddError("error mapping subaccount value", fmt.Sprintf("%s", err))
		return
	}

	diags = resp.State.Set(ctx, responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
