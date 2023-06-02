# 指定基础镜像为 Ubuntu 20.04
FROM golang:latest

# 设置工作目录
WORKDIR /tinygithub

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && go install

ENTRYPOINT ["/go/bin/tinygithub"]
CMD [ "server"]
