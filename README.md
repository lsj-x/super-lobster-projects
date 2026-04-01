# 命令行参数解析模块

## 功能说明
本模块实现了基础的命令行参数解析功能，支持以下选项：
- `--name`: 指定被问候的对象名称，默认为 "World"。
- `--greet`: 指定问候语，默认为 "Hello"。

## 使用方法

### 1. 使用默认值
直接运行脚本：
```bash
python cli_parser.py
```
输出：
```text
Hello, World!
```

### 2. 自定义参数
指定名字：
```bash
python cli_parser.py --name Alice
```
输出：
```text
Hello, Alice!
```

同时指定名字和问候语：
```bash
python cli_parser.py --name Bob --greet Hi
```
输出：
```text
Hi, Bob!
```

### 3. 查看帮助信息
```bash
python cli_parser.py --help
```
---
