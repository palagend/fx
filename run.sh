#!/bin/bash

# 应用名称
APP_NAME="app"
APP_PATH="./bin/app"

# 检查应用是否存在
if [ ! -f "$APP_PATH" ]; then
    echo "错误: 找不到应用 $APP_PATH"
    echo "请先构建应用: ./build.sh"
    exit 1
fi

# 查找并杀死已存在的进程
echo "检查是否有正在运行的 $APP_NAME 进程..."
PID=$(pgrep -f "$APP_PATH")

if [ -n "$PID" ]; then
    echo "发现正在运行的进程 (PID: $PID)，正在停止..."
    kill -TERM "$PID" 2>/dev/null

    # 等待进程结束（最多5秒）
    for i in {1..5}; do
        if ! pgrep -f "$APP_PATH" > /dev/null; then
            echo "进程已停止"
            break
        fi
        sleep 1
    done

    # 如果进程仍在运行，强制杀死
    if pgrep -f "$APP_PATH" > /dev/null; then
        echo "强制终止进程..."
        pkill -9 -f "$APP_PATH"
    fi
else
    echo "没有发现正在运行的进程"
fi

# 启动应用
echo "启动 $APP_NAME..."
nohup "$APP_PATH" > app.log 2>&1 &

# 检查启动是否成功
sleep 2
NEW_PID=$(pgrep -f "$APP_PATH")

if [ -n "$NEW_PID" ]; then
    echo "$APP_NAME 启动成功 (PID: $NEW_PID)"
    echo "日志输出到 app.log"
    echo "访问地址: http://localhost:8080"
else
    echo "启动失败，请检查 app.log 获取错误信息"
    exit 1
fi
