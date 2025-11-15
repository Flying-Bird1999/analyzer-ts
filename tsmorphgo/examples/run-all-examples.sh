#!/bin/bash

# TSMorphGo 示例批量运行脚本
# 运行所有示例并显示验证结果

set -e

echo "🚀 TSMorphGo 示例批量运行脚本"
echo "================================"
echo

# 定义示例列表
examples=(
    "basic_usage.go"
    "node_navigation.go"
    "parser_data.go"
    "comprehensive_verification.go"
    "path_aliases.go"
    "references.go"
)

# 定义示例名称
names=(
    "基础项目操作示例"
    "节点导航和类型收窄示例"
    "透传API验证示例"
    "综合API验证示例"
    "路径别名解析示例"
    "综合引用查找示例"
)

success_count=0
total_count=${#examples[@]}

echo "📊 开始运行 ${total_count} 个示例..."
echo "📝 注: references.go 包含了三个引用查找示例 (Hook函数、类型、工具函数)"
echo

# 运行每个示例
for i in "${!examples[@]}"; do
    example="${examples[$i]}"
    name="${names[$i]}"

    echo "🔍 运行示例 $((i+1))/${total_count}: ${name}"
    echo "文件: ${example}"
    echo "----------------------------------------"

    if go run -tags=examples "${example}"; then
        echo "✅ ${name} 运行成功"
        ((success_count++))
    else
        echo "❌ ${name} 运行失败"
    fi

    echo
    echo "========================================"
    echo
done

# 显示总结
echo "🎉 运行完成！"
echo "✅ 成功: ${success_count}/${total_count}"
echo "❌ 失败: $((total_count - success_count))/${total_count}"

if [ $success_count -eq $total_count ]; then
    echo "🎊 所有示例都运行成功！"
    exit 0
else
    echo "⚠️  有示例运行失败，请检查错误信息"
    exit 1
fi