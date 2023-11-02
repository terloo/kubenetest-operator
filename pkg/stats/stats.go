package stats

import (
	"net/netip"

	"github.com/terloo/kubenetest-operator/pkg/meta"
)

type NetestStats struct {
	NetestType meta.NetestType
	PodName    string
	SourceAddr *netip.Addr
	TargetAddr *netip.Addr
	Passed     bool
	Metric     int
}

func NewInfraNetestStats(source *netip.Addr) *NetestStats {
	stats := &NetestStats{
		NetestType: meta.Infra,
		SourceAddr: source,
	}
	return stats
}

func NewPingNetestStats(source, target *netip.Addr) *NetestStats {
	stats := &NetestStats{
		NetestType: meta.Ping,
		SourceAddr: source,
		TargetAddr: target,
	}
	return stats
}
