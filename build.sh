#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== 1. 构建 Vue 前端 ==="

if [ ! -d "web" ]; then
    echo "错误: web 目录不存在"
    exit 1
fi

if [ ! -f "web/package.json" ]; then
    echo "错误: web/package.json 不存在"
    exit 1
fi

cd web

if [ ! -d "node_modules" ]; then
    echo "安装 npm 依赖..."
    npm install
else
    echo "跳过 npm 安装（node_modules 已存在）"
fi

echo "构建 Vue 项目..."
npm run build

if [ ! -d "dist" ]; then
    echo "错误: Vue 构建失败，dist 目录不存在"
    exit 1
fi

cd ..

echo "✓ Vue 构建完成"

echo ""
echo "=== 2. 构建 Go 后端 ==="

if [ ! -f "go.mod" ]; then
    echo "错误: go.mod 不存在"
    exit 1
fi

echo "整理 Go 依赖..."
go mod tidy

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        ARCH="amd64"
        ;;
esac

echo "当前系统架构: $ARCH"

# 交叉编译 Windows 版本
echo ""
echo "=== 3. 交叉编译 Windows 版本 ==="

echo "编译 Windows 64位版本..."
GOOS=windows GOARCH=amd64 go build -o bin/app-windows-amd64.exe .

if [ ! -f "bin/app-windows-amd64.exe" ]; then
    echo "错误: Windows 版本编译失败"
    exit 1
fi

echo "✓ Windows 64位版本编译完成"

# 如果是 Linux 系统，也编译 Linux 版本
if [ "$(uname -s)" = "Linux" ]; then
    echo "编译 Linux 64位版本..."
    GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64 .
    
    if [ ! -f "bin/app-linux-amd64" ]; then
        echo "错误: Linux 版本编译失败"
        exit 1
    fi
    
    chmod +x bin/app-linux-amd64
    echo "✓ Linux 64位版本编译完成"
    
    echo ""
    echo "=== 构建完成 ==="
    echo "输出文件："
    ls -lh bin/
else
    echo ""
    echo "=== 构建完成 ==="
    echo "输出文件："
    ls -lh bin/
    echo ""
    echo "运行 Windows 程序：bin/app-windows-amd64.exe"
fi
