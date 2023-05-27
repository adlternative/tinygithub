# 指定基础镜像为 Ubuntu 20.04
FROM golang:latest

# 设置工作目录
WORKDIR /tinygithub

# 更新系统并安装 Git
RUN apt-get update && apt-get install -y git

COPY . .

RUN go install

ENTRYPOINT ["/go/bin/tinygithub"]
CMD [ "server"]
