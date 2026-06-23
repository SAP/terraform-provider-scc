package actions

import "github.com/hashicorp/terraform-plugin-framework/action"

func All() []func() action.Action {
	return []func() action.Action{
		NewGenerateCSRAction,
		NewCreateBackupAction,
	}
}
