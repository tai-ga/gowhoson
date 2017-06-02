package gowhoson

const (
	CLIENT_CONFIG = "client.json"
	SERVER_CONFIG = "gowhoson.json"
)

type ClientConfig struct {
	Mode   string
	Server string
}

type ServerConfig struct {
	TCP string
	UDP string
}
