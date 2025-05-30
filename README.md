# Fakery - Kratos 微服务项目

基于 Kratos 框架构建的微服务项目，集成了 etcd 作为服务注册和配置管理中心。

## 功能特性

- 🚀 基于 Kratos v2 框架
- 📡 etcd 服务注册与发现
- ⚙️ 动态配置管理
- 🐳 Docker 容器化部署
- 🔄 自动服务健康检查
- �� gRPC 和 HTTP 双协议支持
- 🎯 多环境配置文件支持

## 快速开始

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- Make

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd fakery
```

2. **安装依赖**
```bash
make init
```

3. **生成代码**
```bash
make all
```

4. **本地运行**
```bash
# 本地调试时会自动使用 configs/config.yaml
go run cmd/fakery/main.go -conf configs/
```

### Docker 部署

#### 使用 docker-compose (推荐)

```bash
# 启动完整的服务栈（包括 MySQL、Redis、etcd）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f fakery
```

#### 传统 Docker 部署

1. **启动 etcd 服务**
```bash
docker-compose -f deploy/docker-compose.yml up -d etcd
```

2. **启动应用服务**
```bash
docker-compose -f deploy/docker-compose.yml up -d fakery
```

3. **验证部署**
```bash
# 运行测试脚本
./scripts/test-etcd.sh

# 或手动检查
curl http://localhost:8000/health
```

## 配置管理

### 多环境配置

项目支持根据运行环境自动选择不同的配置文件：

- **本地开发环境**: 使用 `configs/config.yaml`
  - 数据库连接：`127.0.0.1:3306`
  - Redis 连接：`127.0.0.1:6379`

- **Docker 环境**: 使用 `configs/config.docker.yaml`
  - 数据库连接：`mysql:3306` (Docker 服务名)
  - Redis 连接：`redis:6379` (Docker 服务名)

### 环境变量控制

通过 `ENVIRONMENT` 环境变量控制配置文件选择：

```bash
# 强制使用 Docker 配置
export ENVIRONMENT=docker
go run cmd/fakery/main.go -conf configs/

# 使用默认本地配置
unset ENVIRONMENT
go run cmd/fakery/main.go -conf configs/
```

### 手动指定配置文件

```bash
# 指定特定配置文件
go run cmd/fakery/main.go -conf configs/config.yaml

# 指定配置目录（会根据环境自动选择文件）
go run cmd/fakery/main.go -conf configs/
```

## etcd 集成

### 服务注册

应用启动时会自动注册以下服务到 etcd：

- **HTTP 服务**：`http://IP:8000`
- **gRPC 服务**：`grpc://IP:9000`

### 配置管理

支持动态配置更新和监听，配置存储在 etcd 的 `/fakery/config/` 命名空间下。

### 查看注册信息

```bash
# 查看所有注册的服务
docker exec fakery-etcd etcdctl get --prefix "/fakery/services/"

# 设置配置
docker exec fakery-etcd etcdctl put "/fakery/config/app.json" '{"debug": true}'

# 获取配置
docker exec fakery-etcd etcdctl get "/fakery/config/app.json"
```

详细文档请参考：[etcd 集成文档](docs/etcd-integration.md)

## 项目结构

```
fakery/
├── api/                    # API 定义文件
├── cmd/fakery/            # 应用入口
├── configs/               # 配置文件
├── deploy/                # 部署文件
├── docs/                  # 文档
├── internal/              # 内部代码
│   ├── biz/              # 业务逻辑
│   ├── conf/             # 配置结构
│   ├── data/             # 数据层
│   ├── server/           # 服务器
│   └── service/          # 服务实现
├── scripts/              # 脚本文件
└── third_party/          # 第三方依赖
```

## 开发指南

### 添加新的 API

1. **定义 proto 文件**
```bash
kratos proto add api/helloworld/v1/greeter.proto
```

2. **生成代码**
```bash
make api
```

3. **实现服务**
```bash
kratos proto server api/helloworld/v1/greeter.proto -t internal/service
```

### 配置管理

配置文件位于 `configs/config.yaml`，支持以下配置：

- **server**: HTTP/gRPC 服务器配置
- **data**: 数据库和缓存配置
- **etcd**: 服务注册和配置管理

### Wire 依赖注入

使用 Wire 进行依赖注入：

```bash
cd cmd/fakery
wire
```

## API 文档

### HTTP 接口

- **健康检查**: `GET /health`
- **Swagger 文档**: `GET /swagger/`

### gRPC 接口

- **端口**: 9000
- **健康检查**: 使用 gRPC 健康检查协议

## 监控和日志

### 查看日志

```bash
# 应用日志
docker logs -f fakery-app

# etcd 日志
docker logs -f fakery-etcd
```

### 健康检查

```bash
# HTTP 健康检查
curl http://localhost:8000/health

# etcd 健康检查
docker exec fakery-etcd etcdctl endpoint health
```

## 故障排除

### 常见问题

1. **etcd 连接失败**
   - 检查 etcd 服务状态
   - 验证网络配置
   - 查看错误日志

2. **服务注册失败**
   - 确认 etcd 配置正确
   - 检查服务权限
   - 验证网络连通性

3. **配置更新不生效**
   - 检查配置路径
   - 验证监听回调
   - 查看 etcd 事件

详细故障排除指南请参考：[etcd 集成文档](docs/etcd-integration.md)

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

