package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type BackupActionConfig struct {
	Password   types.String `tfsdk:"password"`
	PasswordWO types.String `tfsdk:"password_wo"`
}
