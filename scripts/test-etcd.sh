#!/bin/bash

# 测试 etcd 服务注册功能

echo "=== 测试 etcd 服务注册功能 ==="

# 检查 etcd 是否运行
echo "1. 检查 etcd 服务状态..."
docker-compose -f deploy/docker-compose.yml ps etcd

# 等待 etcd 启动
echo "2. 等待 etcd 服务启动..."
sleep 10

# 检查 etcd 健康状态
echo "3. 检查 etcd 健康状态..."
docker exec fakery-etcd etcdctl endpoint health

# 启动 fakery 服务
echo "4. 启动 fakery 服务..."
docker-compose -f deploy/docker-compose.yml up -d fakery

# 等待服务启动
echo "5. 等待 fakery 服务启动..."
sleep 15

# 检查服务注册情况
echo "6. 检查服务注册情况..."
docker exec fakery-etcd etcdctl get --prefix "/fakery/services/"

# 检查服务健康状态
echo "7. 检查 fakery 服务健康状态..."
curl -f http://localhost:8000/health || echo "HTTP 健康检查失败"

echo "=== 测试完成 ===" 