import argparse
import sys

def parse_arguments():
    """
    解析命令行参数，支持 --name 和 --greet 选项。
    如果未提供参数，则使用默认值。
    """
    parser = argparse.ArgumentParser(
        description="一个简单的命令行参数解析示例",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    # 定义 --name 参数，默认值为 "World"
    parser.add_argument(
        "--name",
        type=str,
        default="World",
        help="要问候的名字 (默认: World)"
    )

    # 定义 --greet 参数，默认值为 "Hello"
    parser.add_argument(
        "--greet",
        type=str,
        default="Hello",
        help="问候语 (默认: Hello)"
    )

    # 解析参数
    args = parser.parse_args()
    return args

def main():
    args = parse_arguments()
    print(f"{args.greet}, {args.name}!")

if __name__ == "__main__":
    main()
