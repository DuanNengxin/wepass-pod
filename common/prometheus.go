package common

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

func PrometheusBoot(port int) {
	http.Handle("/metrics", promhttp.Handler())
	// 启动web服务，监听8085端口
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
		if err != nil {
			zap.S().Fatal("ListenAndServe: ", err)
		}
	}()
}
