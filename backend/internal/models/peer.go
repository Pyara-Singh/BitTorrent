package models

type Peer struct {
	IP   string `json:"ip"`
	Port uint16 `json:"port"`
}
