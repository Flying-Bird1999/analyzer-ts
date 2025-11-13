#!/bin/bash

echo "TSMorphGo Simple Demo Runner"
echo "============================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# 检查demo-react-app项目
if [ ! -d "demo-react-app" ]; then
    echo "Error: demo-react-app project not found"
    exit 1
fi

echo "Running TSMorphGo demo..."
echo "========================="

# 运行完整演示
go run -tags=examples main.go

echo ""
echo "Demo completed successfully!"