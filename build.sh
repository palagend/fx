#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 默认配置
CROSS_COMPILE=0
VITE_APP_MODE="${VITE_APP_MODE:-frontend}"

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo
    echo "构建脚本 - 构建 Vue 前端和 Go 后端"
    echo
    echo "选项:"
    echo "  -m, --mode <mode>     指定运行模式: frontend (本地模式) 或 backend (后端模式)"
    echo "      --prod            生产环境构建 (等同于 --mode=backend)"
    echo "      --dev             开发环境构建 (等同于 --mode=frontend)"
    echo "  -x, --cross           启用交叉编译（生成多平台版本）"
    echo "  -h, --help            显示此帮助信息"
    echo
    echo "环境变量:"
    echo "  VITE_APP_MODE         默认运行模式，可被命令行参数覆盖"
    echo
    echo "示例:"
    echo "  $0                    # 默认构建（frontend 模式）"
    echo "  $0 --prod             # 生产环境构建"
    echo "  $0 --mode=backend     # 指定后端模式"
    echo "  $0 --prod --cross     # 生产环境 + 交叉编译"
    echo "  $0 -x -m backend      # 短选项形式"
}

# 使用 getopt 解析参数
# -o: 短选项
# --long: 长选项
# -n: 程序名（用于错误消息）
# --: 分隔选项和位置参数
PARSED=$(getopt -o xm:h \
    --long cross,mode:,prod,dev,help \
    -n "$0" -- "$@")

if [ $? != 0 ]; then
    echo "参数解析失败" >&2
    exit 1
fi

# 重新设置参数
eval set -- "$PARSED"

# 解析选项
while true; do
    case "$1" in
        -x|--cross)
            CROSS_COMPILE=1
            shift
            ;;
        -m|--mode)
            VITE_APP_MODE="$2"
            shift 2
            ;;
        --prod)
            VITE_APP_MODE="backend"
            shift
            ;;
        --dev)
            VITE_APP_MODE="frontend"
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        --)
            shift
            break
            ;;
        *)
            echo "未知选项: $1" >&2
            show_help >&2
            exit 1
            ;;
    esac
done

# 验证模式值
if [ "$VITE_APP_MODE" != "frontend" ] && [ "$VITE_APP_MODE" != "backend" ]; then
    echo "错误: VITE_APP_MODE 必须是 'frontend' 或 'backend'" >&2
    exit 1
fi

echo "=== 1. 构建 Vue 前端 ==="
echo "运行模式: $VITE_APP_MODE"

if [ ! -d "web" ]; then
    echo "错误: web 目录不存在" >&2
    exit 1
fi

if [ ! -f "web/package.json" ]; then
    echo "错误: web/package.json 不存在" >&2
    exit 1
fi

cd web

# 检测包管理器（优先使用 pnpm）
if command -v pnpm &> /dev/null; then
    PKG_MANAGER="pnpm"
    PKG_RUN="pnpm run"
elif command -v npm &> /dev/null; then
    PKG_MANAGER="npm"
    PKG_RUN="npm run"
else
    echo "错误: 未找到 pnpm 或 npm，请先安装 Node.js 包管理器" >&2
    exit 1
fi

if [ ! -d "node_modules" ] || [ ! -f "node_modules/.bin/vite" ]; then
    echo "安装依赖 (使用 $PKG_MANAGER)..."
    $PKG_MANAGER install
else
    echo "跳过依赖安装（node_modules 已存在）"
fi

echo "构建 Vue 项目 (VITE_APP_MODE=$VITE_APP_MODE)..."
VITE_APP_MODE="$VITE_APP_MODE" $PKG_RUN build

if [ ! -d "dist" ]; then
    echo "错误: Vue 构建失败，dist 目录不存在" >&2
    exit 1
fi

cd ..

echo "✓ Vue 构建完成"

echo ""
echo "=== 2. 构建 Go 后端 ==="

if [ ! -f "go.mod" ]; then
    echo "错误: go.mod 不存在" >&2
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
        echo "错误: Windows 版本编译失败" >&2
        exit 1
    fi

    echo "✓ Windows 64位版本编译完成"

    # 编译 Linux 版本
    echo "编译 Linux 64位版本..."
    GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64 .

    if [ ! -f "bin/app-linux-amd64" ]; then
        echo "错误: Linux 版本编译失败" >&2
        exit 1
    fi

    chmod +x bin/app-linux-amd64
    echo "✓ Linux 64位版本编译完成"

    # 如果是 macOS，也编译 Darwin 版本
    if [ "$(uname -s)" = "Darwin" ]; then
        echo "编译 macOS 64位版本..."
        GOOS=darwin GOARCH=amd64 go build -o bin/app-darwin-amd64 .

        if [ ! -f "bin/app-darwin-amd64" ]; then
            echo "错误: macOS 版本编译失败" >&2
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
        echo "错误: 本地编译失败" >&2
        exit 1
    fi

    chmod +x bin/app
    echo "✓ 本地版本编译完成"

    echo ""
    echo "=== 构建完成 ==="
    echo "输出文件："
    ls -lh bin/
fi
