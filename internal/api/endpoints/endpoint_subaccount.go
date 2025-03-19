package endpoints

import "fmt"

func GetSubaccountEndpoint(regionHost, subaccount string) string {
	return fmt.Sprintf(GetSubaccountBaseEndpoint()+"/%s/%s", regionHost, subaccount)
}

func GetSubaccountBaseEndpoint() string {
	return "/api/v1/configuration/subaccounts"
}