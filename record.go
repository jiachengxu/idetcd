package idetcd

type Record struct {
	Ipv4 string `json:"ipv4,omitempty"`
	Ipv6 string `json:"ipv6,omitempty"`
	Port string `json:"port,omitempty"`
}
