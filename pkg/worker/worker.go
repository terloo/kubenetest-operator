package worker

import (
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"
	"sync"

	"github.com/terloo/kubenetest-operator/pkg/meta"
	"github.com/terloo/kubenetest-operator/pkg/stats"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
)

type NetestWorker struct {
	lock        sync.Mutex
	workersChan map[*netip.Addr]chan *meta.NetestWork
	runningWork map[meta.NetestType]struct{}
	close       bool

	stats []*stats.NetestStats
}

func NewNetestWorker() *NetestWorker {
	workers := &NetestWorker{
		workersChan: make(map[*netip.Addr]chan *meta.NetestWork, 10),
		stats:       make([]*stats.NetestStats, 0),
	}
	return workers
}

func (w *NetestWorker) Work(ip *netip.Addr, work *meta.NetestWork) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if ip == nil {
		notReadyInfra := stats.NewInfraNetestStats(nil)
		notReadyInfra.Passed = false
		notReadyInfra.PodName = work.Value
		w.stats = append(w.stats, notReadyInfra)
		return
	}

	workerChan, ok := w.workersChan[ip]
	if !ok {
		workerChan = make(chan *meta.NetestWork, 1)
		go func() {
			defer runtime.HandleCrash()
			w.loopWork(ip, workerChan)
		}()
		w.workersChan[ip] = workerChan
	}

	workerChan <- work

}

func (w *NetestWorker) loopWork(ip *netip.Addr, workerChan <-chan *meta.NetestWork) {

	for work := range workerChan {

		switch work.Type {
		case meta.Ping:
			pingResp, err := TestPing(ip, work)
			if err != nil {
				klog.Error(err)
				continue
			}

			targetAddr, _ := netip.ParseAddr(pingResp.Addr)
			pingStats := stats.NewPingNetestStats(ip, &targetAddr)
			pingStats.Passed = pingResp.PkgSent == pingResp.PkgRecv
			pingStats.Metric = pingResp.AvgRtt

			klog.Info("ping result: ", pingStats)
			w.stats = append(w.stats, pingStats)
		case meta.Infra:
			infraResp, err := TestInfra(ip, work)
			if err != nil {
				klog.Error(err)
				continue
			}

			infraStats := stats.NewInfraNetestStats(ip)
			infraStats.Passed = infraResp.Bridgenf == "" && infraResp.Bridgenf6 == "" &&
				infraResp.Ipv4Forward == "" && infraResp.Ipv6DefaultForwarding == ""

			klog.Info("infra result: ", infraStats)
			w.stats = append(w.stats, infraStats)
		}

	}

}

func (w *NetestWorker) Close(name string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for _, workersChan := range w.workersChan {
		close(workersChan)
	}

	w.persistStat(name)
}

func (w *NetestWorker) persistStat(name string) {

	statFilePath := filepath.Join(os.TempDir(), name+"-stats.json")
	f, err := os.Create(statFilePath)
	if err != nil {
		klog.ErrorS(err, "persist stat error")
		return
	}
	b, err := json.Marshal(w.stats)
	if err != nil {
		klog.ErrorS(err, "persist stat error")
		return
	}

	klog.InfoS("success persist statis file", "name", name, "path", statFilePath)
	f.Write(b)
}
