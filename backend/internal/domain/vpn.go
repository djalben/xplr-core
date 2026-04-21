package domain

type VPNClientTraffic struct {
	ProviderRef string `json:"providerRef"`
	Enabled     bool   `json:"enabled"`
	Upload      int64  `json:"upload"`
	Download    int64  `json:"download"`
}
