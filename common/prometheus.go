package common

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

func PrometheusBoot() {
	http.Handle("/metrics", promhttp.Handler())
	// 启动web服务，监听8085端口
	go func() {
		err := http.ListenAndServe("127.0.0.1:8085", nil)
		if err != nil {
			zap.S().Fatal("ListenAndServe: ", err)
		}
	}()
}
