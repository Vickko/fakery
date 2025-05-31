#!/bin/bash

# 测试配置文件选择功能

echo "=== 配置文件选择测试 ==="
echo

# 编译项目
echo "1. 编译项目..."
go build -o bin/fakery ./cmd/fakery
if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi
echo "✅ 编译成功"
echo

# 测试本地环境配置
echo "2. 测试本地环境配置选择..."
timeout 3s ./bin/fakery -conf configs/ > /tmp/local_config_test.log 2>&1
if grep -q "config.yaml" /tmp/local_config_test.log; then
    echo "✅ 本地环境正确使用 config.yaml"
else
    echo "❌ 本地环境配置文件选择错误"
    echo "日志内容："
    cat /tmp/local_config_test.log
fi
echo

# 测试Docker环境配置
echo "3. 测试Docker环境配置选择..."
timeout 3s env ENVIRONMENT=docker ./bin/fakery -conf configs/ > /tmp/docker_config_test.log 2>&1
if grep -q "config.docker.yaml" /tmp/docker_config_test.log; then
    echo "✅ Docker环境正确使用 config.docker.yaml"
else
    echo "❌ Docker环境配置文件选择错误"
    echo "日志内容："
    cat /tmp/docker_config_test.log
fi
echo

# 测试手动指定配置文件
echo "4. 测试手动指定配置文件..."
timeout 3s ./bin/fakery -conf configs/config.yaml > /tmp/manual_config_test.log 2>&1
if grep -q "config.yaml" /tmp/manual_config_test.log; then
    echo "✅ 手动指定配置文件正常工作"
else
    echo "❌ 手动指定配置文件失败"
    echo "日志内容："
    cat /tmp/manual_config_test.log
fi
echo

echo "=== 测试完成 ==="

# 清理临时文件
rm -f /tmp/*_config_test.log 