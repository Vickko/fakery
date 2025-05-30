# 日志级别配置

本项目支持通过环境变量和配置文件设置日志级别，只有大于等于设置级别的日志才会输出到stdout。

## 支持的日志级别

| 级别 | 数值 | 描述 |
|------|------|------|
| DEBUG | 0 | 调试信息，最详细的日志 |
| INFO | 1 | 一般信息，默认级别 |
| WARN/WARNING | 2 | 警告信息 |
| ERROR | 3 | 错误信息 |
| FATAL | 4 | 致命错误，最高级别 |

## 配置方式

### 1. 环境变量配置（优先级最高）

```bash
# 设置日志级别为DEBUG
export LOG_LEVEL=debug
./bin/fakery

# 设置日志级别为WARN
export LOG_LEVEL=warn
./bin/fakery

# 设置日志级别为ERROR
export LOG_LEVEL=error
./bin/fakery
```

### 2. 配置文件配置

在 `configs/config.yaml` 或 `configs/config.docker.yaml` 中设置：

```yaml
log:
  level: "info"          # 日志级别: debug, info, warn, error, fatal
  enable_file_log: false # 是否启用文件日志（暂未实现）
  file_path: "./logs/app.log" # 日志文件路径（暂未实现）
```

## 优先级

1. 环境变量 `LOG_LEVEL`（最高优先级）
2. 配置文件中的 `log.level` 设置
3. 默认值 `info`

## 使用示例

### Docker 环境

```bash
# 使用环境变量设置日志级别
docker run -e LOG_LEVEL=debug your-app:latest

# 或在docker-compose.yml中设置
environment:
  - LOG_LEVEL=warn
```

### Kubernetes 环境

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fakery
spec:
  template:
    spec:
      containers:
      - name: fakery
        image: fakery:latest
        env:
        - name: LOG_LEVEL
          value: "debug"
```

## 测试脚本

运行测试脚本来验证不同日志级别的效果：

```bash
./scripts/test_log_levels.sh
```

## 注意事项

- 日志级别不区分大小写（debug/DEBUG/Debug 都有效）
- 只有大于等于设置级别的日志才会输出
- 所有日志都输出到 stdout，便于容器化部署时收集日志
- 如果设置了无效的日志级别，将使用默认值 `info` 