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

func TestInfra(ip *netip.Addr, work *meta.NetestWork) (*InfraResponse, error) {
	url := fmt.Sprintf("http://%s:8888/infra", ip.String())
	klog.Infof("request to pod %s url %s", ip.String(), url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respBody := new(InfraResponse)
	json.Unmarshal(b, &respBody)
	return respBody, nil
}

type InfraResponse struct {
	Bridgenf              string `json:"bridgenf,omitempty"`
	Bridgenf6             string `json:"bridgenf6,omitempty"`
	Ipv4Forward           string `json:"ipv4Forward,omitempty"`
	Ipv6DefaultForwarding string `json:"ipv6DefaultForwarding,omitempty"`
}
