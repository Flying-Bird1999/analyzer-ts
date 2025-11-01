#!/bin/bash

# TSMorphGo API 验证测试脚本

echo "🚀 TSMorphGo API 验证测试"
echo "=================================================="

# 检查 demo-react-app 是否存在
if [ ! -d "demo-react-app" ]; then
    echo "❌ 错误: demo-react-app 目录不存在"
    exit 1
fi

# 检查 api-examples 目录是否存在
if [ ! -d "api-examples" ]; then
    echo "❌ 错误: api-examples 目录不存在"
    exit 1
fi

# 进入 api-examples 目录
cd api-examples

# 定义测试函数
run_test() {
    local test_name=$1
    local test_file=$2
    local project_path=$3
    local tag=$4 # New parameter for the build tag

    echo ""
    echo "🧪 运行测试: $test_name"
    echo "------------------------------"

    if go run -tags "$tag" "$test_file" "$project_path"; then
        echo "✅ $test_name 测试通过"
    else
        echo "❌ $test_name 测试失败"
        return 1
    fi
}

# 运行所有测试
echo "开始运行所有 API 验证测试..."

tests=(
    "基础分析:01-basic-analysis.go:../demo-react-app:example01"
    "符号分析:02-symbol-analysis.go:../demo-react-app:example02"
    "接口扫描:03-interface-scan.go:../demo-react-app:example03"
    "依赖检查:04-dependency-check.go:../demo-react-app:example04"
    "节点导航:05-node-navigation.go:../demo-react-app:example05"
    "表达式分析:06-expression-analysis.go:../demo-react-app:example06"
    "类型检查:07-type-checking.go:../demo-react-app:example07"
    "LSP服务:08-lsp-service.go:../demo-react-app:example08"
    "高级符号:09-advanced-symbols.go:../demo-react-app:example09"
    "QuickInfo底层能力验证:10-quickinfo-test-working.go:../demo-react-app:example10"
)

failed_tests=0

for test in "${tests[@]}"; do
    IFS=':' read -r name file path tag <<< "$test" # Read the new tag parameter

    if ! run_test "$name" "$file" "$path" "$tag"; then # Pass the tag to run_test
        ((failed_tests++))
    fi

    echo ""
    sleep 1  # 添加间隔，避免输出混乱
done

# 输出测试结果摘要
echo "=================================================="
if [ $failed_tests -eq 0 ]; then
    echo "🎉 所有测试通过！TSMorphGo API 功能正常"
else
    echo "❌ 发现 $failed_tests 个测试失败"
    exit 1
fi

# 显示生成的文件
echo ""
echo "📁 生成的文件:"
if [ -f "interfaces.json" ]; then
    echo "  - interfaces.json (接口扫描结果)"
fi
if [ -f "api.json" ]; then
    echo "  - api.json (API 文档 JSON)"
fi
if [ -d "docs" ]; then
    echo "  - docs/ (文档目录)"
}

echo ""
echo "✅ 测试完成！"