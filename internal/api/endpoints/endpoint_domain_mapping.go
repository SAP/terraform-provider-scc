package endpoints

import "fmt"

func GetDomainMappingEndpoint(regionHost, subaccount, internalDomain string) string {
	return fmt.Sprintf(GetDomainMappingBaseEndpoint(regionHost, subaccount)+"/%s", internalDomain)
}

func GetDomainMappingBaseEndpoint(regionHost, subaccount string) string {
	return fmt.Sprintf(GetSubaccountBaseEndpoint()+"/%s/%s/domainMappings", regionHost, subaccount)
}
