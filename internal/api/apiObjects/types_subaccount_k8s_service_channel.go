package apiobjects

type SubaccountK8SServiceChannel struct {
	K8SCluster  string                           `json:"k8sCluster"`
	K8SService  string                           `json:"k8sService"`
	ID          int64                            `json:"id"`
	Type        string                           `json:"type"`
	Port        int64                            `json:"port"`
	Enabled     bool                             `json:"enabled"`
	Connections int64                            `json:"connections"`
	Comment     string                           `json:"comment"`
	State       SubaccountK8SServiceChannelState `json:"state"`
}

type SubaccountK8SServiceChannelState struct {
	Connected               bool  `json:"connected"`
	OpenedConnections       int64 `json:"openedConnections"`
	ConnectedSinceTimeStamp int64 `json:"connectedSinceTimeStamp"`
}

type SubaccountK8SServiceChannels struct {
	SubaccountK8SServiceChannels []SubaccountK8SServiceChannel `json:"service_channels_k8s"`
}
