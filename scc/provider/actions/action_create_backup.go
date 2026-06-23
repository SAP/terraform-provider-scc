package actions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/terraform-provider-scc/internal/api"
	"github.com/SAP/terraform-provider-scc/internal/api/endpoints"
	"github.com/SAP/terraform-provider-scc/scc/provider/helpers"
	"github.com/SAP/terraform-provider-scc/scc/provider/model"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
)

type CreateBackupAction struct {
	Client *api.RestApiClient
}

var _ action.Action = &CreateBackupAction{}

func NewCreateBackupAction() action.Action {
	return &CreateBackupAction{}
}

func (a *CreateBackupAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_create_backup"
}

func (a *CreateBackupAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a backup of the cloud connector instance.",
		Attributes: map[string]schema.Attribute{
			"password": schema.StringAttribute{
				MarkdownDescription: "The password used to protect the backup file. Action schema attributes cannot be marked as sensitive, so this value may be visible in Terraform configuration and logs. Use a strong password and handle it securely.",
				Required:            true,
			},
		},
	}
}

func (a *CreateBackupAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *CreateBackupAction) InvokeWithPlan(ctx context.Context, plan model.BackupActionConfig, resp *action.InvokeResponse) {
	if plan.Password.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Password",
			"A non-empty password is required to create a backup.",
		)
		return
	}

	endpoint := endpoints.GetBackupEndpoint()
	planBody := map[string]any{
		"password": plan.Password.ValueString(),
	}

	helpers.SafeProgress(resp, "Initiating backup creation...")
	backupResponse, diags := helpers.SendRequestFunc(a.Client, planBody, endpoint, helpers.ActionCreateRequest)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if backupResponse == nil || backupResponse.Body == nil {
		resp.Diagnostics.AddError("Invalid API Response", "Backup response body is nil")
		return
	}

	defer func() {
		if err := backupResponse.Body.Close(); err != nil {
			// log but don’t fail action
			resp.Diagnostics.AddWarning(
				"Failed to close response body",
				err.Error(),
			)
		}
	}()

	backupBytes, err := io.ReadAll(backupResponse.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read Backup Response",
			fmt.Sprintf("An error occurred while reading the backup response body: %v", err),
		)
		return
	}

	contentType := backupResponse.Header.Get("Content-Type")
	if contentType != "application/zip" {
		resp.Diagnostics.AddError(
			"Unexpected Content-Type",
			fmt.Sprintf("Expected 'application/zip', but got '%s'. The backup response may be invalid.", contentType),
		)
		return
	}

	if len(backupBytes) == 0 {
		resp.Diagnostics.AddError(
			"Empty Backup Response",
			"The API response did not contain a valid backup archive.",
		)
		return
	}

	fileName := fmt.Sprintf("scc_backup_%s.zip", time.Now().UTC().Format("20060102_150405"))

	filePath := filepath.Join(".", fileName)

	err = os.WriteFile(filePath, backupBytes, 0644)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Write Backup to File",
			fmt.Sprintf("An error occurred while writing the backup to file: %v", err),
		)
		return
	}

	helpers.SafeProgress(resp, "Backup generated successfully")
	helpers.SafeProgress(resp, fmt.Sprintf("Backup saved to %s", filePath))
}

func (a *CreateBackupAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var plan model.BackupActionConfig
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	a.InvokeWithPlan(ctx, plan, resp)
}
