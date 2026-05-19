#!/usr/bin/env python3
"""
Python HTTP 服务，使用 Flask
"""

import os
import json
from flask import Flask, request, Response
from openai import OpenAI

API_KEY = os.environ.get("OPENAI_API_KEY", "sk-hww5ISsBbQ8Q8HyraLFwM5D30tzMEmEAAT1e3qsoR4rhDvE9")
BASE_URL = os.environ.get("OPENAI_BASE_URL", "https://api.moonshot.cn/v1")
MODEL = os.environ.get("OPENAI_MODEL", "kimi-k2.6")
PORT = int(os.environ.get("PYTHON_SERVER_PORT", "8081"))

openai_client = OpenAI(api_key=API_KEY, base_url=BASE_URL)

app = Flask(__name__)


@app.route('/api/chat-process', methods=['POST'])
def chat_process():
    try:
        body = request.json
        messages = body.get("messages", [])
        stream = body.get("stream", True)

        print(f"[Flask] 收到请求: model={MODEL}, stream={stream}, messages_count={len(messages)}")

        response = openai_client.chat.completions.create(
            model=MODEL,
            messages=messages,
            extra_body={"thinking": {"type": "disabled"}},
            stream=stream,
        )

        if stream:
            def generate():
                try:
                    for chunk in response:
                        data = chunk.model_dump_json()
                        yield f"data: {data}\n\n"
                    yield "data: [DONE]\n\n"
                except Exception as e:
                    print(f"[Flask] 流式生成错误: {e}")

            return Response(
                generate(),
                mimetype='text/event-stream',
                headers={
                    'Cache-Control': 'no-cache',
                    'Connection': 'keep-alive',
                }
            )
        else:
            return json.loads(response.model_dump_json())

    except Exception as e:
        print(f"[Flask] 错误: {e}")
        return {"error": str(e)}, 500


if __name__ == "__main__":
    print(f"[Flask] 启动在端口 {PORT}")
    app.run(host="0.0.0.0", port=PORT, threaded=True)
