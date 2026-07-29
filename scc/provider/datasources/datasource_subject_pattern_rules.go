package datasources

import (
	"context"
	"fmt"

	"github.com/SAP/terraform-provider-scc/internal/api"
	apiobjects "github.com/SAP/terraform-provider-scc/internal/api/apiObjects"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &SubjectPatternRulesDataSource{}

func NewSubjectPatternRulesDataSource() datasource.DataSource {
	return &SubjectPatternRulesDataSource{}
}

type SubjectPatternRulesDataSource struct {
	Client *api.RestApiClient
}

func (d *SubjectPatternRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subject_pattern_rules"
}

func (d *SubjectPatternRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Cloud Connector Subject Pattern Rules Data Source.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Associate Administrator
	* Subaccount Administrator 
	* Display
	* Support
	* Monitoring

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/proxy-settings>`,
		Attributes: map[string]schema.Attribute{
			"subject_pattern_rules": schema.ListNestedAttribute{
				MarkdownDescription: "List of subject pattern rules.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the subject pattern rule.",
							Computed:            true,
						},
						"condition": schema.SingleNestedAttribute{
							MarkdownDescription: "Condition of the subject pattern rule.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"variable": schema.StringAttribute{
									MarkdownDescription: "Variable of the condition to be evaluated.",
									Computed:            true,
								},
								"operator": schema.StringAttribute{
									MarkdownDescription: "Operator of the condition to be used.",
									Computed:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Value of the condition to be evaluated against.",
									Computed:            true,
								},
							},
						},
						"subject_pattern": schema.SingleNestedAttribute{
							MarkdownDescription: "Subject pattern of the subject pattern rule.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"cn": schema.StringAttribute{
									MarkdownDescription: "Common Name (CN) of the subject pattern.",
									Computed:            true,
								},
								"email": schema.StringAttribute{
									MarkdownDescription: "Email (EMAIL) of the subject pattern.",
									Computed:            true,
								},
								"l": schema.StringAttribute{
									MarkdownDescription: "Locality (L) of the subject pattern.",
									Computed:            true,
								},
								"ou": schema.StringAttribute{
									MarkdownDescription: "Organization Unit (OU) of the subject pattern.",
									Computed:            true,
								},
								"o": schema.StringAttribute{
									MarkdownDescription: "Organization (O) of the subject pattern.",
									Computed:            true,
								},
								"st": schema.StringAttribute{
									MarkdownDescription: "State (ST) of the subject pattern.",
									Computed:            true,
								},
								"c": schema.StringAttribute{
									MarkdownDescription: "Country (C) of the subject pattern.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *SubjectPatternRulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.Client = client
}

func (d *SubjectPatternRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data model.SubjectPatternRulesDataSourceConfig
	var respObj apiobjects.SubjectPatternRules
	diags := req.Config.Get(ctx, &data)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := endpoints.GetSubjectPatternRulesBaseEndpoint()

	diags = helpers.RequestAndUnmarshal(d.Client, &respObj, "GET", endpoint, nil, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, diags := model.SubjectPatternRulesValueFrom(ctx, data, respObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
