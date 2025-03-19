package endpoints

import "fmt"

func GetSystemMappingResourceBaseEndpoint (regionHost, subaccount, virtualHost, virtualPort string) string{
	return GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort) + "/resources"
}

func GetSystemMappingResourceEndpoint (regionHost, subaccount, virtualHost, virtualPort, resourceID string) string{
	return fmt.Sprintf(GetSystemMappingResourceBaseEndpoint(regionHost, subaccount, virtualHost, virtualPort)+ "/%s", resourceID)
}