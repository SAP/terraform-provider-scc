package apiobjects

type DomainMapping struct{
	VirtualDomain string `json:"virtualDomain"`
	InternalDomain string `json:"internalDomain"`
}

type DomainMappings struct{
	DomainMappings []DomainMapping `json:"domain_mappings"`
}