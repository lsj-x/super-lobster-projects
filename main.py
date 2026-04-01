import sys
import json

def generate_greeting(params: dict) -> str:
    """
    根据解析后的参数动态生成问候语。
    支持参数：name (姓名), language (语言), time_of_day (时间段)
    """
    name = params.get("name", "朋友")
    language = params.get("language", "zh")
    time_of_day = params.get("time_of_day", "day")

    greetings = {
        "zh": {
            "morning": f"早上好，{name}！愿你今天充满活力。",
            "afternoon": f"下午好，{name}！工作/学习顺利吗？",
            "evening": f"晚上好，{name}！好好休息吧。",
            "day": f"你好，{name}！很高兴见到你。"
        },
        "en": {
            "morning": f"Good morning, {name}! Have a wonderful day.",
            "afternoon": f"Good afternoon, {name}! How is your day going?",
            "evening": f"Good evening, {name}! Time to relax.",
            "day": f"Hello, {name}! Nice to meet you."
        }
    }

    # 默认 fallback 到中文通用问候
    if language not in greetings:
        language = "zh"
    
    if time_of_day not in greetings[language]:
        time_of_day = "day"

    return greetings[language][time_of_day]

def main():
    # 模拟解析后的参数（实际场景中可能来自 CLI 参数、配置文件或 API 请求）
    # 示例输入：{"name": "张三", "language": "zh", "time_of_day": "morning"}
    try:
        # 尝试从标准输入读取 JSON 参数，如果没有则使用默认值
        input_data = sys.stdin.read().strip()
        if input_data:
            params = json.loads(input_data)
        else:
            params = {
                "name": "用户",
                "language": "zh",
                "time_of_day": "day"
            }
        
        greeting = generate_greeting(params)
        print(greeting)
        
    except json.JSONDecodeError:
        print("错误：输入参数格式无效，请提供有效的 JSON 数据。", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"发生未知错误: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
