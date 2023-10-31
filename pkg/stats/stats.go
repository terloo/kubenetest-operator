package stats

import (
	"net/netip"

	"github.com/terloo/kubenetest-operator/pkg/meta"
)

type NetestStats struct {
	NetestType meta.NetestType
	SourceAddr netip.Addr
	TargetAddr netip.Addr
	Passed     bool
	Metric     int
}

func NewPingNetestStats(source, target netip.Addr) *NetestStats {
	stats := &NetestStats{
		NetestType: meta.Ping,
		SourceAddr: source,
		TargetAddr: target,
	}
	return stats
}
