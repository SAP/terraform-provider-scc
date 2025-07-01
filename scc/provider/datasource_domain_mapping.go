package provider

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/validation/uuidvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &DomainMappingDataSource{}

func NewDomainMappingDataSource() datasource.DataSource {
	return &DomainMappingDataSource{}
}

type DomainMappingDataSource struct {
	client *api.RestApiClient
}

func (d *DomainMappingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_mapping"
}

func (d *DomainMappingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Domain Mapping Data Source.

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
				Computed:            true,
			},
			"internal_domain": schema.StringAttribute{
				MarkdownDescription: "Domain used on the on-premise side.",
				Required:            true,
			},
		},
	}
}

func (d *DomainMappingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DomainMappingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainMappingConfig
	var respObj apiobjects.DomainMappings
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regionHost := data.RegionHost.ValueString()
	subaccount := data.Subaccount.ValueString()
	internalDomain := data.InternalDomain.ValueString()

	endpoint := fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/domainMappings", regionHost, subaccount)

	err := requestAndUnmarshal(d.client, &respObj.DomainMappings, "GET", endpoint, nil, true)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchDomainMappingsFailed, err.Error())
		return
	}

	mappingRespObj, err := GetDomainMapping(respObj, internalDomain)
	if err != nil {
		resp.Diagnostics.AddError(errMsgFetchDomainMappingFailed, fmt.Sprintf("%s", err))
		return
	}

	responseModel, err := DomainMappingValueFrom(ctx, data, *mappingRespObj)
	if err != nil {
		resp.Diagnostics.AddError(errMsgMapDomainMappingFailed, fmt.Sprintf("%s", err))
		return
	}
	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
