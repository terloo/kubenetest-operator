package worker

import (
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"sync"

	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
)

type NetestWorkers struct {
	lock        sync.Mutex
	workersChan map[*netip.Addr]chan *NetestWork
}

func NewNetestWorkers() *NetestWorkers {
	workers := &NetestWorkers{
		workersChan: make(map[*netip.Addr]chan *NetestWork, 10),
	}
	return workers
}

func (w *NetestWorkers) Work(ip *netip.Addr, work *NetestWork) {
	w.lock.Lock()
	defer w.lock.Unlock()

	workerChan, ok := w.workersChan[ip]
	if !ok {
		workerChan = make(chan *NetestWork, 1)
		go func() {
			defer runtime.HandleCrash()
			w.loopWork(ip, workerChan)
		}()
		w.workersChan[ip] = workerChan
	}

	workerChan <- work

}

func (w *NetestWorkers) loopWork(ip *netip.Addr, workerChan <-chan *NetestWork) {

	for work := range workerChan {

		if work.Type == Ping {
			url := fmt.Sprintf("http://%s:8888/ping?addr=%s", ip.String(), work.Value)
			klog.Infof("request to pod %s url %s", ip.String(), url)
			resp, err := http.Get(url)
			if err != nil {
				klog.Error(err, "work ping fail")
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				klog.Error(err, "work ping fail")
			}
			respBody := make(map[string]interface{})
			json.Unmarshal(b, &respBody)
			klog.Info("resp: ", respBody)
		}

	}
}

const (
	Ping WorkType = "ping"
)

type WorkType string

type NetestWork struct {
	Type WorkType

	Value string
}
