package prometh

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
)

func Push(pushGateAddr string, delay float64, net string) {
	jobName := fmt.Sprintf("filscan_bell_finalheight_delay")
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: jobName})
	gauge.Set(delay)
	err := push.New(pushGateAddr, jobName).Grouping("module", "filscan").Grouping("net", net).Collector(gauge).Push()
	if err != nil {
		log.Printf("push prometheus %s failed:%s", pushGateAddr, err)
	}

}
