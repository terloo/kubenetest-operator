package meta

type NetestType string

const (
	Infra NetestType = "infra"
	Ping  NetestType = "ping"
)

type NetestWork struct {
	Type    NetestType
	PodName string
	Value   string
}
