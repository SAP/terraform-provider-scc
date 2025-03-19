package endpoints

import "fmt"

func GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, serviceChannelType string) string{
	return fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/channels/%s", regionHost, subaccount, serviceChannelType)
}

func GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType string, id int64) string{
	return fmt.Sprintf(GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, serviceChannelType) + "/%d", id)
}



