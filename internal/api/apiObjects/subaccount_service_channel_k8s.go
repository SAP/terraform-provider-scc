package apiobjects

type SubaccountServiceChannelK8S struct{
	K8SCluster string `json:"k8sCluster"`
	K8SService string `json:"k8sService"`
	ID int64 `json:"id"`
	Type string `json:"type"`
	Port int64 `json:"port"`
	Enabled bool `json:"enabled"`
	Connections int64 `json:"connections"`
	Comment string `json:"comment"`
	State SubaccountServiceChannelK8SState `json:"state"`
}

type SubaccountServiceChannelK8SState struct{
	Connected bool `json:"connected"`
	OpenedConnections int64 `json:"openedConnections"`
	ConnectedSinceTimeStamp int64 `json:"connectedSinceTimeStamp"`
}

type SubaccountServiceChannelsK8S struct{
	SubaccountServiceChannelsK8S []SubaccountServiceChannelK8S `json:"service_channels_k8s"`
}

