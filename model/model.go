package model

type Conn struct {
	Key        string `json:"key"`
	Hostname   string `json:"hostname"`
	User       string `json:"user"`
	Pass       string `json:"pass"`
	Desc       string `json:"desc"`
	RootEnable bool   `json:"root_enable"`
}
