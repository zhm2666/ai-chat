from openai import OpenAI
import os

# 从环境变量读取配置，或使用默认值
API_KEY = os.environ.get("OPENAI_API_KEY", "sk-hww5ISsBbQ8Q8HyraLFwM5D30tzMEmEAAT1e3qsoR4rhDvE9")
BASE_URL = os.environ.get("OPENAI_BASE_URL", "https://api.moonshot.cn/v1")
MODEL = os.environ.get("OPENAI_MODEL", "kimi-k2.6")

# 创建 OpenAI 客户端
client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

def chat_completion(messages, stream=True):
    """调用模型，返回流式响应"""
    response = client.chat.completions.create(
        model=MODEL,
        messages=messages,
        extra_body={
            "thinking": {"type": "disabled"}
        },
        stream=stream,
    )
    return response

if __name__ == "__main__":
    # 测试代码
    messages = [
        {
            "role": "system",
            "content": "你是 Kimi，由 Moonshot AI 提供的人工智能助手。",
        },
        {"role": "user", "content": "你好，1+1等于多少？"},
    ]
    
    print("开始测试...")
    response = chat_completion(messages)
    
    collected = []
    for idx, chunk in enumerate(response):
        delta = chunk.choices[0].delta
        if delta.content:
            collected.append(delta.content)
            print(f"[{idx}] {''.join(collected)}")
    
    print(f"\n完整回复: {''.join(collected)}")
