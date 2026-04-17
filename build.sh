#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 默认启用交叉编译
CROSS_COMPILE=1

# 解析命令行参数
while getopts "x" opt; do
    case $opt in
        x)
            CROSS_COMPILE=1
            ;;
        *)
            echo "用法: $0 [-x]"
            echo "  -x  启用交叉编译（默认已启用）"
            exit 1
            ;;
    esac
done

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

if [ $CROSS_COMPILE -eq 1 ]; then
    echo ""
    echo "=== 3. 交叉编译 ==="

    # 编译 Windows 版本
    echo "编译 Windows 64位版本..."
    GOOS=windows GOARCH=amd64 go build -o bin/app-windows-amd64.exe .

    if [ ! -f "bin/app-windows-amd64.exe" ]; then
        echo "错误: Windows 版本编译失败"
        exit 1
    fi

    echo "✓ Windows 64位版本编译完成"

    # 编译 Linux 版本
    echo "编译 Linux 64位版本..."
    GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64 .

    if [ ! -f "bin/app-linux-amd64" ]; then
        echo "错误: Linux 版本编译失败"
        exit 1
    fi

    chmod +x bin/app-linux-amd64
    echo "✓ Linux 64位版本编译完成"

    # 如果是 macOS，也编译 Darwin 版本
    if [ "$(uname -s)" = "Darwin" ]; then
        echo "编译 macOS 64位版本..."
        GOOS=darwin GOARCH=amd64 go build -o bin/app-darwin-amd64 .

        if [ ! -f "bin/app-darwin-amd64" ]; then
            echo "错误: macOS 版本编译失败"
            exit 1
        fi

        chmod +x bin/app-darwin-amd64
        echo "✓ macOS 64位版本编译完成"
    fi

    echo ""
    echo "=== 构建完成 ==="
    echo "输出文件："
    ls -lh bin/
else
    # 本地编译
    echo "编译本地版本..."
    go build -o bin/app .

    if [ ! -f "bin/app" ]; then
        echo "错误: 本地编译失败"
        exit 1
    fi

    chmod +x bin/app
    echo "✓ 本地版本编译完成"

    echo ""
    echo "=== 构建完成 ==="
    echo "输出文件："
    ls -lh bin/
fi
