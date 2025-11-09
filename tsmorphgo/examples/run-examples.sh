#!/bin/bash

# =============================================================================
# TSMorphGo Examples 运行脚本
# =============================================================================
# 描述: 用于运行和管理 TSMorphGo 示例项目的Shell脚本
# 使用方法: ./run-examples.sh <command> [args...]
# 示例: ./run-examples.sh help, ./run-examples.sh all
# =============================================================================

# 脚本配置
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时报错

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目路径配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG_DIR="$SCRIPT_DIR/pkg"
PROJECT_ROOT="$SCRIPT_DIR"

# =============================================================================
# 工具函数
# =============================================================================

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo -e "${PURPLE}🔧 $1${NC}"
    echo -e "${PURPLE}================================${NC}"
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        print_error "命令 '$1' 未找到，请先安装"
        exit 1
    fi
}

# 检查示例文件是否存在
check_example_file() {
    local file_path="$1"
    if [[ ! -f "$file_path" ]]; then
        print_error "示例文件不存在: $file_path"
        return 1
    fi
    return 0
}

# 运行示例的通用函数
run_example() {
    local example_name="$1"
    local file_path="$2"
    local build_tag="$3"
    local description="$4"

    print_header "运行 $example_name"
    print_info "📁 功能: $description"
    print_info "📄 文件: $file_path"
    print_info "🏷️  构建标签: $build_tag"

    if check_example_file "$file_path"; then
        local dir_name=$(dirname "$file_path")
        local file_name=$(basename "$file_path")

        print_info "▶️  开始执行..."
        echo

        # 运行示例
        cd "$dir_name" && go run -tags "$build_tag" "$file_name"

        echo
        print_success "$example_name 运行完成！"
    fi
}

# =============================================================================
# 核心功能函数
# =============================================================================

# 显示帮助信息
show_help() {
    cat << EOF
${CYAN}🔧 TSMorphGo Examples 运行脚本${NC}

${YELLOW}📖 使用方法:${NC}
    $0 <命令> [参数...]

${YELLOW}🚀 快速开始:${NC}
    help        - 显示此帮助信息（默认）
    install     - 安装/更新项目依赖
    deps        - 检查项目依赖

${YELLOW}📦 批量运行示例:${NC}
    all         - 运行所有示例（推荐）
    test        - 运行项目测试

${YELLOW}🎯 单独运行示例:${NC}
    project-management      - 项目管理和内存文件系统示例
    node-navigation         - 节点导航和位置信息示例
    type-detection          - 类型检测和代码质量分析示例
    reference-finding       - 引用查找和符号系统示例
    specialized-apis        - 专用API示例

${YELLOW}🧹 维护命令:${NC}
    clean       - 清理编译和临时文件
    build       - 构建可执行文件
    fmt         - 格式化代码
    report      - 生成项目报告

${YELLOW}🔧 开发工具:${NC}
    check       - 检查环境配置
    status      - 显示项目状态

EOF
}

# 检查环境配置
check_environment() {
    print_header "检查环境配置"

    # 检查Go环境
    print_info "🔍 检查 Go 环境..."
    if check_command "go"; then
        local go_version=$(go version)
        print_success "Go 环境: $go_version"

        # 检查Go模块
        if [[ -f "$PROJECT_ROOT/go.mod" ]]; then
            print_success "Go 模块文件存在"
            local module_name=$(go list -m)
            print_info "模块名称: $module_name"
        else
            print_warning "未找到 go.mod 文件"
        fi
    fi

    # 检查项目文件
    print_info "🔍 检查项目文件..."

    local pkg_files=("$PKG_DIR/project-management.go" "$PKG_DIR/node-navigation.go" "$PKG_DIR/type-detection.go" "$PKG_DIR/reference-finding.go" "$PKG_DIR/specialized-apis.go")

    for file in "${pkg_files[@]}"; do
        if [[ -f "$file" ]]; then
            print_success "✓ $(basename "$file")"
        else
            print_error "✗ $(basename "$file") 不存在"
        fi
    done

    # 检查demo项目
    local demo_dir="$PROJECT_ROOT/demo-react-app"
    if [[ -d "$demo_dir" ]]; then
        local ts_files=$(find "$demo_dir" -name "*.ts" -o -name "*.tsx" | wc -l)
        print_success "✓ Demo React项目: $ts_files 个TypeScript文件"
    else
        print_error "✗ Demo React项目不存在"
    fi
}

# 安装依赖
install_dependencies() {
    print_header "安装项目依赖"

    check_command "go"

    print_info "📋 检查 Go 环境..."
    go version

    print_info "📋 下载依赖包..."
    if [[ -f "$PROJECT_ROOT/go.mod" ]]; then
        cd "$PROJECT_ROOT"
        go mod download
        go mod tidy
        print_success "依赖安装完成！"
    else
        print_error "未找到 go.mod 文件"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    print_header "检查项目依赖"

    check_command "go"

    print_info "📋 Go 版本信息:"
    go version

    if [[ -f "$PROJECT_ROOT/go.mod" ]]; then
        cd "$PROJECT_ROOT"
        print_info "📋 项目模块信息:"
        go list -m

        print_info "📋 依赖包版本:"
        go list -m all | grep -E "(tsmorphgo|typescript-go)" || print_warning "未找到特定依赖包"
    else
        print_warning "未找到 go.mod 文件"
    fi
}

# 运行所有示例
run_all_examples() {
    print_header "运行所有TSMorphGo示例"
    print_info "📋 执行计划:"
    print_info "  • 阶段1: 项目管理示例 - 项目创建和内存文件系统"
    print_info "  • 阶段2: 节点导航示例 - AST遍历和位置信息"
    print_info "  • 阶段3: 类型检测示例 - 类型分析和代码质量"
    print_info "  • 阶段4: 引用查找示例 - 符号系统和缓存优化"
    print_info "  • 阶段5: 专用API示例 - 特定语法结构分析"
    print_info "  • 总计: 5个综合示例"
    echo

    run_example "项目管理示例" "$PKG_DIR/project-management.go" "project_management" "项目创建、源文件管理、内存文件系统、动态文件创建"
    echo
    run_example "节点导航示例" "$PKG_DIR/node-navigation.go" "node_navigation" "节点遍历、祖先查找、位置计算、IDE集成"
    echo
    run_example "类型检测示例" "$PKG_DIR/type-detection.go" "type_detection" "类型识别、接口分析、代码质量、依赖关系"
    echo
    run_example "引用查找示例" "$PKG_DIR/reference-finding.go" "reference_finding" "引用查找、缓存优化、跳转定义、重命名安全"
    echo
    run_example "专用API示例" "$PKG_DIR/specialized-apis.go" "specialized_apis" "函数声明、调用表达式、属性访问、导入别名"

    print_success "🎉 所有示例运行完成！"
}

# 清理文件
clean_files() {
    print_header "清理编译和临时文件"

    print_info "🗑️ 清理Go编译产物..."
    find "$PROJECT_ROOT" -name "*.o" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.exe" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.out" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.test" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.prof" -delete 2>/dev/null || true

    print_info "🗑️ 清理临时文件..."
    find "$PROJECT_ROOT" -name "*.tmp" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name ".DS_Store" -delete 2>/dev/null || true

    print_info "🗑️ 清理IDE文件..."
    find "$PROJECT_ROOT" -name ".vscode" -type d -exec rm -rf {} + 2>/dev/null || true

    # 清理构建目录
    if [[ -d "$PROJECT_ROOT/bin" ]]; then
        rm -rf "$PROJECT_ROOT/bin"
        print_info "🗑️ 清理构建目录"
    fi

    print_success "清理完成！"
}

# 构建可执行文件
build_executables() {
    print_header "构建可执行文件"

    local bin_dir="$PROJECT_ROOT/bin"
    mkdir -p "$bin_dir"

    print_info "🏗️ 构建所有示例..."
    cd "$PKG_DIR"

    # 构建所有示例
    local examples=(
        "project-management:project_management"
        "node-navigation:node_navigation"
        "type-detection:type_detection"
        "reference-finding:reference_finding"
        "specialized-apis:specialized_apis"
    )

    for example in "${examples[@]}"; do
        local file="${example%:*}"
        local tag="${example#*:}"

        if [[ -f "$file.go" ]]; then
            go build -tags "$tag" -o "$bin_dir/$file" "$file.go"
            print_success "✓ $file"
        else
            print_warning "✗ $file.go 不存在"
        fi
    done

    print_success "构建完成！可执行文件位于 $bin_dir"
}

# 运行测试
run_tests() {
    print_header "运行项目测试"

    check_command "go"

    print_info "🧪 运行单元测试..."
    cd "$PROJECT_ROOT"
    go test ./... -v

    print_info "🏃 运行基准测试（如果存在）..."
    go test -bench=. ./... 2>/dev/null || print_warning "未找到基准测试"

    print_success "测试完成！"
}

# 格式化代码
format_code() {
    print_header "格式化代码"

    check_command "go"

    cd "$PROJECT_ROOT"
    go fmt ./...

    print_success "代码格式化完成！"
}

# 生成项目报告
generate_report() {
    print_header "生成项目报告"

    print_info "📊 项目统计:"

    # 统计Go文件
    local go_files=$(find "$PROJECT_ROOT" -name "*.go" | wc -l)
    local go_lines=$(find "$PROJECT_ROOT" -name "*.go" -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}' || echo "0")
    local project_size=$(du -sh "$PROJECT_ROOT" | cut -f1)

    echo "  • Go文件数量: $go_files"
    echo "  • 代码行数: $go_lines"
    echo "  • 项目大小: $project_size"

    # 统计TypeScript文件
    local ts_files=$(find "$PROJECT_ROOT/demo-react-app" -name "*.ts" -o -name "*.tsx" 2>/dev/null | wc -l)
    echo "  • TypeScript文件数量: $ts_files"

    print_info "📋 示例文件:"
    ls -la "$PKG_DIR"/*.go 2>/dev/null | awk '{print "  " $9 " (" $5 " bytes)"}' || print_warning "未找到示例文件"

    print_info "📋 架构信息:"
    echo "  • 新架构: 统一在 pkg/ 目录下管理所有示例"
    echo "  • 示例数量: 5个综合示例"
    echo "  • 覆盖范围: 基础和高级API的完整覆盖"
}

# 显示项目状态
show_status() {
    print_header "项目状态"

    # 基本信息
    print_info "📁 项目目录: $PROJECT_ROOT"
    print_info "🔧 脚本版本: 2.0.0"
    print_info "📅 最后更新: $(date)"
    print_info "🏗️ 架构: 统一pkg目录结构"

    # 文件统计
    local go_files=$(find "$PKG_DIR" -name "*.go" | wc -l)
    print_info "📄 Go示例文件: $go_files 个"

    # 目录状态
    print_info "📂 目录结构:"
    for dir in pkg demo-react-app; do
        if [[ -d "$PROJECT_ROOT/$dir" ]]; then
            print_success "  ✓ $dir"
        else
            print_error "  ✗ $dir (缺失)"
        fi
    done

    # 环境检查
    if command -v go &> /dev/null; then
        print_success "  ✓ Go 环境"
    else
        print_error "  ✗ Go 环境 (未安装)"
    fi

    # 示例文件状态
    print_info "📋 示例文件状态:"
    local examples=(
        "project-management"
        "node-navigation"
        "type-detection"
        "reference-finding"
        "specialized-apis"
    )

    for example in "${examples[@]}"; do
        if [[ -f "$PKG_DIR/$example.go" ]]; then
            print_success "  ✓ $example.go"
        else
            print_error "  ✗ $example.go (缺失)"
        fi
    done
}

# =============================================================================
# 主程序
# =============================================================================

# 主函数 - 处理命令行参数
main() {
    local command="${1:-help}"

    case "$command" in
        "help"|"-h"|"--help")
            show_help
            ;;
        "check")
            check_environment
            ;;
        "install")
            install_dependencies
            ;;
        "deps")
            check_dependencies
            ;;
        "all")
            run_all_examples
            ;;
        "test")
            run_tests
            ;;
        "clean")
            clean_files
            ;;
        "build")
            build_executables
            ;;
        "fmt")
            format_code
            ;;
        "report")
            generate_report
            ;;
        "status")
            show_status
            ;;
        "project-management")
            run_example "项目管理示例" "$PKG_DIR/project-management.go" "project_management" "项目创建、源文件管理、内存文件系统、动态文件创建"
            ;;
        "node-navigation")
            run_example "节点导航示例" "$PKG_DIR/node-navigation.go" "node_navigation" "节点遍历、祖先查找、位置计算、IDE集成"
            ;;
        "type-detection")
            run_example "类型检测示例" "$PKG_DIR/type-detection.go" "type_detection" "类型识别、接口分析、代码质量、依赖关系"
            ;;
        "reference-finding")
            run_example "引用查找示例" "$PKG_DIR/reference-finding.go" "reference_finding" "引用查找、缓存优化、跳转定义、重命名安全"
            ;;
        "specialized-apis")
            run_example "专用API示例" "$PKG_DIR/specialized-apis.go" "specialized_apis" "函数声明、调用表达式、属性访问、导入别名"
            ;;
        *)
            print_error "未知命令: $command"
            print_info "使用 '$0 help' 查看可用命令"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"