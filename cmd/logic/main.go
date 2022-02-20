package main

import (
	"gim/config"
	"gim/internal/logic/api"
	"gim/internal/logic/app"
	"gim/internal/logic/proxy"
	"gim/pkg/db"
	"gim/pkg/interceptor"
	"gim/pkg/logger"
	"gim/pkg/pb"
	"gim/pkg/rpc"
	"gim/pkg/urlwhitelist"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func initProxy() {
	proxy.MessageProxy = app.MessageApp
	proxy.DeviceProxy = app.DeviceApp
}

func main() {
	logger.Init()
	db.InitMysql(config.Logic.MySQL)
	db.InitRedis(config.Logic.RedisIP, config.Logic.RedisPassword)

	// 初始化APP代理
	initProxy()

	// 初始化RpcClient
	rpc.InitConnectIntClient(config.RPCAddr.ConnectRPCAddr)
	rpc.InitBusinessIntClient(config.RPCAddr.BusinessRPCAddr)

	// 这里有个鉴权的过程，应该是设备的鉴权，登录鉴权在SignIn处理，每收到一个请求都会验证设备是否符合要求
	server := grpc.NewServer(grpc.UnaryInterceptor(interceptor.NewInterceptor("logic_interceptor", urlwhitelist.Logic)))

	// 监听服务关闭信号，服务平滑重启
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)
		s := <-c
		logger.Logger.Info("server stop", zap.Any("signal", s))
		server.GracefulStop()
	}()

	// server中注册内部Logic和外部Logic
	pb.RegisterLogicIntServer(server, &api.LogicIntServer{})
	pb.RegisterLogicExtServer(server, &api.LogicExtServer{})
	listen, err := net.Listen("tcp", config.Logic.RPCListenAddr)
	if err != nil {
		panic(err)
	}

	logger.Logger.Info("rpc服务已经开启")
	err = server.Serve(listen)
	if err != nil {
		logger.Logger.Error("serve error", zap.Error(err))
	}
}
