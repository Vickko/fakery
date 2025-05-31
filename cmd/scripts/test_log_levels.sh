#!/bin/bash

# 测试日志级别功能的脚本

echo "测试日志级别功能..."

# 编译程序
echo "编译程序..."
cd $(dirname $0)/..
go build -o bin/fakery cmd/fakery/main.go cmd/fakery/wire_gen.go

if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

echo "编译成功"

# 测试不同的日志级别
echo ""
echo "1. 测试默认日志级别 (info):"
timeout 5s ./bin/fakery -conf ./configs/config.yaml &
sleep 2
kill $! 2>/dev/null

echo ""
echo "2. 测试 DEBUG 级别:"
LOG_LEVEL=debug timeout 5s ./bin/fakery -conf ./configs/config.yaml &
sleep 2
kill $! 2>/dev/null

echo ""
echo "3. 测试 WARN 级别:"
LOG_LEVEL=warn timeout 5s ./bin/fakery -conf ./configs/config.yaml &
sleep 2
kill $! 2>/dev/null

echo ""
echo "4. 测试 ERROR 级别:"
LOG_LEVEL=error timeout 5s ./bin/fakery -conf ./configs/config.yaml &
sleep 2
kill $! 2>/dev/null

echo ""
echo "测试完成！观察上面的输出，可以看到不同日志级别输出的差异。" 