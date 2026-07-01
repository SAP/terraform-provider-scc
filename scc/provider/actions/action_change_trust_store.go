package actions

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
)

type ChangeTrustStoreAction struct {
	Client *api.RestApiClient
}

var _ action.Action = &ChangeTrustStoreAction{}

func NewChangeTrustStoreAction() action.Action {
	return &ChangeTrustStoreAction{}
}

func (a *ChangeTrustStoreAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_change_trust_store"
}

func (a *ChangeTrustStoreAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Changing Back-End Trust Store, determining trust based on a list of allowed back-ends, each represented by its respective X.509 certificate, is considered more secure.
		
~> **Caution:** We discourage trusting all backends and recommend that you trust only specific backends represented by their respective certificates.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Associate Administrator
	
__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/truststore-ca-certificates#change-truststore-configuration-(master-only)>`,
		Attributes: map[string]schema.Attribute{
			"trust_all_backends": schema.BoolAttribute{
				MarkdownDescription: "Flag (boolean) indicating whether all backends are trusted (true), or only the backends represented by certificates are trusted (false)",
				Required:            true,
			},
		},
	}
}

func (a *ChangeTrustStoreAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.RestApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *api.RestApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.Client = client
}

func (a *ChangeTrustStoreAction) InvokeWithPlan(ctx context.Context, plan model.BackendTrustStoreActionConfig, resp *action.InvokeResponse) {
	if plan.TrustAllBackends.IsUnknown() || plan.TrustAllBackends.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Action Plan",
			"The action plan is invalid. Please ensure that the `trust_all_backends` attribute is set correctly.",
		)
		return
	}

	endpoint := endpoints.GetBackendTrustStoreBaseEndpoint()
	planBody := map[string]any{
		"trustAllBackends": plan.TrustAllBackends.ValueBool(),
	}

	helpers.SafeProgress(resp, "Updating backend trust configuration...")
	backendResponse, diags := helpers.SendRequestFunc(a.Client, planBody, endpoint, helpers.ActionPatchRequest)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	defer func() {
		if err := backendResponse.Body.Close(); err != nil {
			resp.Diagnostics.AddError(
				"Failed to Close Response Body",
				fmt.Sprintf("failed to close response body: %v", err),
			)
		}
	}()

	if backendResponse.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(backendResponse.Body)
		closeErr := backendResponse.Body.Close()
		if err != nil {
			resp.Diagnostics.AddError("Failed to Read Response Body", fmt.Sprintf("failed to read response body: %v", err))
			return
		}

		if closeErr != nil {
			resp.Diagnostics.AddError("Failed to Close Response Body", fmt.Sprintf("failed to close response body: %v", closeErr))
			return
		}

		resp.Diagnostics.AddError(
			"Failed to Change Backend Trust Configuration",
			fmt.Sprintf(
				"Expected HTTP %d but received %d.\nResponse: %s",
				http.StatusNoContent,
				backendResponse.StatusCode,
				string(body),
			),
		)
		return
	}

	helpers.SafeProgress(resp, "Backend trust configuration changed successfully.")
}

func (a *ChangeTrustStoreAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var plan model.BackendTrustStoreActionConfig
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	a.InvokeWithPlan(ctx, plan, resp)
}
