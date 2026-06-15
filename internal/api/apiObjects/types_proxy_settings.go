package apiobjects

type ProxySettings struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}
