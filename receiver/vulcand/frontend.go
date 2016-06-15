package vulcand

type Frontend struct {
	Type      string           `json:"Type"`
	BackendId string           `json:"BackendId"`
	Route     string           `json:"Route"`
	Settings  FrontendSettings `json:"Settings"`
}

type FrontendSettings struct {
	TrustForwardHeader bool `json:"TrustForwardHeader"`
}
