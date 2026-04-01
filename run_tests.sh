#!/bin/bash

# 设置脚本在遇到错误时退出
set -e

# 打印开始信息
echo "=== Running pytest tests ==="

# 运行 pytest，启用详细输出、生成 JUnit XML 报告、并显示覆盖率（如果安装了 pytest-cov）
pytest --verbose --tb=short --junitxml=test_results.xml ${1:-}

# 检查 pytest 是否成功执行
if [ $? -eq 0 ]; then
    echo ""
    echo "=== All tests passed! ==="
else
    echo ""
    echo "=== Some tests failed! ==="
    exit 1
fi

# 如果安装了 pytest-cov，显示覆盖率报告
if command -v pytest-cov &> /dev/null; then
    echo ""
    echo "=== Coverage Report ==="
    pytest --cov=. --cov-report=term-missing
fi

# 输出测试报告文件位置
echo ""
echo "=== Test report saved to: test_results.xml ==="
---
