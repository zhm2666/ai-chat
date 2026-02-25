# tokenizer 

## 镜像构建
``` 
docker build -t 192.168.239.161:5000/tokenizer:1.0.0 .
```

## 服务启动
``` 
docker service create --name tokenizer \
-p 3002:3002 --replicas 3 \
--with-registry-auth \
192.168.239.161:5000/tokenizer:1.0.0
```