package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"

	"github.com/terloo/kubenetest-operator/pkg/meta"
	"k8s.io/klog/v2"
)

type PingResponse struct {
	Addr    string `json:"addr,omitempty"`
	PkgSent int    `json:"pkgSent,omitempty"`
	PkgRecv int    `json:"pkgRecv,omitempty"`
	AvgRtt  int    `json:"avgRtt,omitempty"`
	MaxRtt  int    `json:"maxRtt,omitempty"`
	MinRtt  int    `json:"minRtt,omitempty"`
	Rtts    []int  `json:"rtts,omitempty"`
}

func TestPing(ip *netip.Addr, work *meta.NetestWork) (*PingResponse, error) {
	url := fmt.Sprintf("http://%s:8888/ping?addr=%s", ip.String(), work.Value)
	klog.Infof("request to pod %s url %s", ip.String(), url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respBody := new(PingResponse)
	json.Unmarshal(b, &respBody)
	return respBody, nil
}
