# 项目文档：CLI 工具参数说明

本文档详细说明了命令行工具中 `--name` 和 `--greet` 选项的用法、示例及预期输出。

## 1. 选项说明

### `--name`
- **描述**：指定要问候的目标名称。
- **类型**：字符串 (String)
- **是否必需**：是
- **默认值**：无
- **示例值**：`Alice`, `World`, `Developer`

### `--greet`
- **描述**：指定问候语的类型或前缀。
- **类型**：字符串 (String)
- **是否必需**：否
- **默认值**：`Hello`
- **可选值**：`Hello`, `Hi`, `Greetings`, `Welcome`

## 2. 使用示例

### 示例 1：基本用法
使用默认问候语问候指定名称。

**命令：**
```bash
python main.py --name Alice
```

**预期输出：**
```text
Hello Alice
```

### 示例 2：自定义问候语
同时指定名称和自定义问候语。

**命令：**
```bash
python main.py --name Bob --greet Hi
```

**预期输出：**
```text
Hi Bob
```

### 示例 3：复杂问候语
使用非默认的问候语前缀。

**命令：**
```bash
python main.py --name "Team Lead" --greet Greetings
```

**预期输出：**
```text
Greetings Team Lead
```

## 3. 错误处理

如果未提供必需的 `--name` 参数，程序将抛出错误并提示用法。

**命令：**
```bash
python main.py --greet Hi
```

**预期输出：**
```text
Error: The following argument is required: --name
Usage: main.py --name <NAME> [--greet <GREETING>]
```
---
