FROM quay.io/0voice/python:3.10-alpine
ADD ./curl-amd64 /usr/bin/curl
RUN chmod +x /usr/bin/curl
WORKDIR /app
ENV PORT 3002

ADD ./tokenizer.py /app/tokenizer.py
ADD ./requirements.txt /app/requirements.txt
RUN pip install -i  https://pypi.tuna.tsinghua.edu.cn/simple --trusted-host pypi.tuna.tsinghua.edu.cn  --upgrade pip
RUN pip install --root-user-action=ignore -i https://pypi.tuna.tsinghua.edu.cn/simple --trusted-host pypi.tuna.tsinghua.edu.cn  -r requirements.txt

HEALTHCHECK --interval=10s --timeout=1s --retries=3 \
CMD curl -f -X POST -H "Content-Type: application/json" -d "{\"role\":\"user\",\"content\":\"hello\"}" http://localhost:${PORT}/tokenizer/gpt-3.5-turbo-0301 || exit 1
CMD ["sh", "-c","nuxt --port ${PORT} --module tokenizer.py --workers 2"]