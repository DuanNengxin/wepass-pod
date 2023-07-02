package common

import (
	"github.com/afex/hystrix-go/hystrix"
	"net"
	"net/http"
)

func InitHystrix() {
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go func() {
		// 启动监听程序， 看板url:http://localhost:9002/hystrix
		// 输入框搜索：http://{机器ip}:81/turbine/turbine.stream
		err := http.ListenAndServe(net.JoinHostPort("", "81"), hystrixStreamHandler)
		if err != nil {

		}
	}()
}
