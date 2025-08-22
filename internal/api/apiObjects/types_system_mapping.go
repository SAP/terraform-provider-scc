package apiobjects

type SystemMappings struct {
	SystemMappings []SystemMapping `json:"system_mappings"`
}

type SystemMapping struct {
	VirtualHost           string             `json:"virtualHost"`
	VirtualPort           string             `json:"virtualPort"`
	InternalHost          string             `json:"localHost"`
	InternalPort          string             `json:"localPort"`
	CreationDate          string             `json:"creationDate"`
	Protocol              string             `json:"protocol"`
	BackendType           string             `json:"backendType"`
	AuthenticationMode    string             `json:"authenticationMode"`
	HostInHeader          string             `json:"hostInHeader"`
	Sid                   string             `json:"sid"`
	TotalResourcesCount   int64              `json:"totalResourcesCount"`
	EnabledResourcesCount int64              `json:"enabledResourcesCount"`
	Description           string             `json:"description"`
	SAPRouter             string             `json:"sapRouter"`
	SNCPartnerName        string             `json:"sncPartnerName"`
	AllowedClients        []string           `json:"allowedClients"`
	BlacklistedUsers      []BlacklistedUsers `json:"blacklistedUsers"`
}

type BlacklistedUsers struct {
	Client string `json:"client"`
	User   string `json:"user"`
}
