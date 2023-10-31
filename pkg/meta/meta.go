package meta

const (
	Ping NetestType = "ping"
)

type NetestType string

type NetestWork struct {
	Type  NetestType
	Value string
}
