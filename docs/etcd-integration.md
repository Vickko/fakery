# Etcd 集成文档

本文档介绍了 Fakery 项目中 etcd 的集成和使用方法。

## 功能概述

Fakery 项目集成了 etcd 作为服务注册和配置管理中心，提供以下功能：

1. **服务注册与发现**：自动注册 HTTP 和 gRPC 服务到 etcd
2. **配置管理**：支持动态配置更新和监听
3. **健康检查**：定期续租确保服务可用性

## 配置说明

### etcd 配置项

在 `configs/config.yaml` 中配置 etcd 相关参数：

```yaml
etcd:
  endpoints:
    - "etcd:2379"  # etcd 服务地址
  timeout: 5s      # 连接超时时间
  
  # 服务注册配置
  registry:
    namespace: "/fakery/services"  # 服务注册命名空间
    ttl: 30s                      # 租约时间
    service:
      name: "fakery"              # 服务名称
      version: "v1.0.0"           # 服务版本
      metadata:                   # 服务元数据
        weight: 100
        region: "default"
  
  # 配置管理
  config:
    namespace: "/fakery/config"   # 配置命名空间
    watch_timeout: 10s           # 监听超时时间
```

## 部署说明

### 使用 Docker Compose

1. 启动 etcd 服务：
```bash
docker-compose -f deploy/docker-compose.yml up -d etcd
```

2. 等待 etcd 健康检查通过：
```bash
docker-compose -f deploy/docker-compose.yml ps etcd
```

3. 启动 Fakery 服务：
```bash
docker-compose -f deploy/docker-compose.yml up -d fakery
```

### 验证部署

运行测试脚本验证 etcd 集成：
```bash
./scripts/test-etcd.sh
```

## 服务注册

### 自动注册

Fakery 服务启动时会自动注册以下信息到 etcd：

- **HTTP 服务**：`/fakery/services/fakery/http://IP:8000`
- **gRPC 服务**：`/fakery/services/fakery/grpc://IP:9000`

### 注册信息格式

```json
{
  "name": "fakery",
  "version": "v1.0.0",
  "addr": "http://172.18.0.3:8000",
  "metadata": {
    "weight": "100",
    "region": "default"
  }
}
```

### 手动查看注册信息

```bash
# 查看所有注册的服务
docker exec fakery-etcd etcdctl get --prefix "/fakery/services/"

# 查看特定服务
docker exec fakery-etcd etcdctl get "/fakery/services/fakery/http://172.18.0.3:8000"
```

## 配置管理

### 设置配置

```bash
# 设置配置
docker exec fakery-etcd etcdctl put "/fakery/config/app.json" '{"debug": true, "log_level": "info"}'
```

### 监听配置变化

Fakery 服务会自动监听配置变化，当配置更新时会触发回调函数。

### 获取配置

```bash
# 获取配置
docker exec fakery-etcd etcdctl get "/fakery/config/app.json"
```

## 故障排除

### 常见问题

1. **etcd 连接失败**
   - 检查 etcd 服务是否正常运行
   - 验证网络连接和端口配置
   - 查看 etcd 日志：`docker logs fakery-etcd`

2. **服务注册失败**
   - 检查 etcd 配置是否正确
   - 验证服务权限和网络访问
   - 查看应用日志：`docker logs fakery-app`

3. **配置监听不生效**
   - 确认配置路径正确
   - 检查监听回调函数是否正确实现
   - 验证 etcd 事件是否正常触发

### 日志查看

```bash
# 查看 etcd 日志
docker logs fakery-etcd

# 查看 Fakery 应用日志
docker logs fakery-app

# 实时查看日志
docker logs -f fakery-app
```

## API 接口

### 健康检查

- **HTTP**：`GET http://localhost:8000/health`
- **gRPC**：使用 gRPC 健康检查协议

### 服务发现

通过 etcd 客户端可以发现已注册的服务：

```go
// 示例：发现服务
resp, err := etcdClient.Get(ctx, "/fakery/services/", clientv3.WithPrefix())
for _, kv := range resp.Kvs {
    var serviceInfo ServiceInfo
    json.Unmarshal(kv.Value, &serviceInfo)
    fmt.Printf("发现服务: %s at %s\n", serviceInfo.Name, serviceInfo.Addr)
}
```

## 最佳实践

1. **服务命名**：使用有意义的服务名称和版本号
2. **元数据管理**：合理设置服务元数据，便于负载均衡和路由
3. **配置分层**：按环境和模块组织配置结构
4. **监控告警**：监控 etcd 集群状态和服务注册情况
5. **备份恢复**：定期备份 etcd 数据

## 扩展功能

### 负载均衡

可以基于服务注册信息实现客户端负载均衡：

```go
// 示例：简单的轮询负载均衡
func (lb *LoadBalancer) GetNextService() (*ServiceInfo, error) {
    services := lb.getHealthyServices()
    if len(services) == 0 {
        return nil, errors.New("no healthy services available")
    }
    
    index := atomic.AddUint64(&lb.counter, 1) % uint64(len(services))
    return services[index], nil
}
```

### 配置热更新

实现配置热更新功能：

```go
// 示例：配置热更新
func (app *App) watchConfig() {
    app.etcdClient.WatchConfig(context.Background(), "app.json", func(data []byte) {
        var config AppConfig
        if err := json.Unmarshal(data, &config); err == nil {
            app.updateConfig(config)
            app.logger.Info("配置已更新")
        }
    })
}
``` 