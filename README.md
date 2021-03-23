# 项目中使用的工具集
## 1. grpc包 grpc服务端和客户端封装

1. 启动grpc server

```goland
			
server = grpc.NewServer({端口号})

server.RegisterService({proto文件描述},{实例化proto实现类})

server.Start()

```

2. 服务注册发现

``` goland
服务注册发现基于etcd 调用microsoft包
```