#!/bin/bash
# TalkAboutIt 构建脚本
# 提供跨平台编译和本地构建能力

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/bin"
BINARY_NAME="talkaboutit"

# 默认参数
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
CGO_ENABLED="${CGO_ENABLED:-0}"

function print_help() {
    echo "TalkAboutIt 构建脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  build       编译二进制（默认）"
    echo "  build-all   跨平台编译（linux/amd64, darwin/amd64, darwin/arm64）"
    echo "  test        运行单元测试"
    echo "  integration 运行集成测试"
    echo "  eval        运行评测框架"
    echo "  clean       清理构建产物"
    echo "  help        显示本帮助信息"
    echo ""
    echo "环境变量:"
    echo "  GOOS        目标操作系统（默认: 当前系统）"
    echo "  GOARCH      目标架构（默认: 当前架构）"
    echo "  CGO_ENABLED 是否启用 CGO（默认: 0）"
}

function build_binary() {
    local os="$1"
    local arch="$2"
    local suffix=""

    if [ "${os}" = "windows" ]; then
        suffix=".exe"
    fi

    local output="${BUILD_DIR}/${BINARY_NAME}-${os}-${arch}${suffix}"
    echo "==> 编译 ${os}/${arch} -> ${output}"

    mkdir -p "${BUILD_DIR}"
    GOOS="${os}" GOARCH="${arch}" CGO_ENABLED="${CGO_ENABLED}" \
        go build -ldflags="-s -w" -o "${output}" "${PROJECT_ROOT}/cmd/server"
}

function cmd_build() {
    echo "==> 开始构建..."
    build_binary "${GOOS}" "${GOARCH}"
    echo "==> 构建完成"
}

function cmd_build_all() {
    echo "==> 开始跨平台构建..."
    build_binary "linux" "amd64"
    build_binary "darwin" "amd64"
    build_binary "darwin" "arm64"
    echo "==> 跨平台构建完成"
}

function cmd_test() {
    echo "==> 运行单元测试..."
    cd "${PROJECT_ROOT}"
    go test ./... -v -count=1
}

function cmd_integration() {
    echo "==> 运行集成测试..."
    cd "${PROJECT_ROOT}"
    go test -tags=integration ./test/ -v -count=1
}

function cmd_eval() {
    echo "==> 运行评测框架..."
    cd "${PROJECT_ROOT}/test/eval"
    go test -tags=eval ./... -v -count=1
}

function cmd_clean() {
    echo "==> 清理构建产物..."
    rm -rf "${BUILD_DIR}"
    rm -f "${PROJECT_ROOT}/test/eval/baseline.json"
    echo "==> 清理完成"
}

# 主入口
case "${1:-build}" in
    build)
        cmd_build
        ;;
    build-all)
        cmd_build_all
        ;;
    test)
        cmd_test
        ;;
    integration)
        cmd_integration
        ;;
    eval)
        cmd_eval
        ;;
    clean)
        cmd_clean
        ;;
    help|--help|-h)
        print_help
        ;;
    *)
        echo "未知命令: $1"
        print_help
        exit 1
        ;;
esac
