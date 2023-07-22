package main

import (
	"flag"
	"fmt"
	"github.com/DuanNengxin/wepass-pod/common"
	"github.com/DuanNengxin/wepass-pod/config"
	"github.com/DuanNengxin/wepass-pod/domain/repository"
	service2 "github.com/DuanNengxin/wepass-pod/domain/service"
	"github.com/DuanNengxin/wepass-pod/handler"
	"github.com/DuanNengxin/wepass-pod/plugin"
	pod "github.com/DuanNengxin/wepass-pod/proto"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	wrapperPrometheus "github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	ratelimit "github.com/asim/go-micro/plugins/wrapper/ratelimiter/uber/v3"
	opentracing2 "github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net"
	"net/http"
	"path/filepath"
)

var (
	migrate       bool
	mode          string
	k8sConfigPath *string
)

func init() {
	flag.BoolVar(&migrate, "migrate", false, "数据库迁移")
	flag.StringVar(&mode, "mode", "dev", "启动模式")
}

func main() {
	flag.Parse()
	config.ParseConfig()

	consulReg := consul.NewRegistry(
		registry.Addrs(
			[]string{fmt.Sprintf("%s:%s", config.Config().Consul.Host, config.Config().Consul.Port)}...,
		),
	)

	db := common.InitMysql(migrate)
	common.InitLogger(mode)
	tracer, i, err := common.NewTracer("wepass-pod", fmt.Sprintf("%s:%s", config.Config().Tracer.Host, config.Config().Tracer.Port))
	if err != nil {
		return
	}
	defer i.Close()
	opentracing.SetGlobalTracer(tracer)

	// 添加熔断器
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()

	// 启动监听程序
	go func() {
		//http://192.168.0.112:9092/turbine/turbine.stream
		//看板访问地址 http://127.0.0.1:9002/hystrix，url后面一定要带 /hystrix
		err = http.ListenAndServe(net.JoinHostPort("0.0.0.0", "9092"), hystrixStreamHandler)
		fmt.Println("333")
		if err != nil {
			fmt.Println(err)
		}
	}()

	go common.PrometheusBoot()

	if home := homedir.HomeDir(); home != "" {
		k8sConfigPath = flag.String("k8sConfigPath", filepath.Join(home, ".kube", "config"), "k8s config 在系统中的位置")
	} else {
		zap.S().Fatal("k8s config not found")
	}
	flag.Parse()

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", *k8sConfigPath)
	if err != nil {
		zap.S().Fatalf("create k8s config error %s", err.Error())
	}

	// 在集群中使用
	//k8sConfig, err := rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}

	// 创建k8s客户端
	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		zap.S().Fatal(err.Error())
	}

	srv := micro.NewService(
		micro.Name("wepass-pod"),
		micro.Version("v1.0"),
		micro.Registry(consulReg),
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())), // 服务端开启span
		micro.WrapClient(opentracing2.NewClientWrapper(opentracing.GlobalTracer())),   // 客户端添加，api方式访问
		micro.WrapHandler(ratelimit.NewHandlerWrapper(1000)),                          // 服务端限流
		micro.WrapClient(plugin.NewHystrixClientWrapper()),                            // 客户端熔断，只作为客户端的时候起作用
		micro.WrapHandler(wrapperPrometheus.NewHandlerWrapper()),                      // 添加prometheus监控
	)
	// 初始化服务
	srv.Init()

	podDataService := service2.NewPodDataService(repository.NewPodRepository(db), clientSet)
	// 创建服务句柄
	_ = pod.RegisterPodServiceHandler(srv.Server(), &handler.PodHandler{PodDataService: podDataService})

	if err := srv.Run(); err != nil {
		zap.S().Fatal(err)
	}

	select {}
}
