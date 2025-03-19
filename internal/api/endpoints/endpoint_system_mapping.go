package endpoints

import "fmt"

func GetSystemMappingBaseEndpoint (regionHost, subaccount string) string{
	return fmt.Sprintf("/api/v1/configuration/subaccounts/%s/%s/systemMappings", regionHost, subaccount)
}

func GetSystemMappingEndpoint (regionHost, subaccount, virtualHost, virtualPort string) string{
	return fmt.Sprintf(GetSystemMappingBaseEndpoint(regionHost, subaccount)+ "/%s:%s", virtualHost,virtualPort)
}