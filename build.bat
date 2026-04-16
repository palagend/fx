@echo off
setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

echo === 1. 构建 Vue 前端 ===

if not exist "web" (
    echo 错误: web 目录不存在
    exit /b 1
)

if not exist "web\package.json" (
    echo 错误: web\package.json 不存在
    exit /b 1
)

cd web

if not exist "node_modules" (
    echo 安装 npm 依赖...
    call npm install
) else (
    echo 跳过 npm 安装（node_modules 已存在）
)

echo 构建 Vue 项目...
call npm run build

if not exist "dist" (
    echo 错误: Vue 构建失败，dist 目录不存在
    exit /b 1
)

cd ..

echo.
echo ✓ Vue 构建完成
echo.

echo === 2. 构建 Go 后端 ===

if not exist "go.mod" (
    echo 错误: go.mod 不存在
    exit /b 1
)

echo 整理 Go 依赖...
call go mod tidy

echo 编译 Go 程序...
call go build -o bin/app.exe .

if not exist "bin\app.exe" (
    echo 错误: Go 编译失败，bin\app.exe 不存在
    exit /b 1
)

echo ✓ Go 编译完成
echo.

echo === 构建完成 ===
echo 输出文件：bin\app.exe
dir bin\app.exe
echo.
echo 运行程序：bin\app.exe

pause
