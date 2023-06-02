# 指定基础镜像为 Ubuntu 20.04
FROM golang:latest

# 设置工作目录
WORKDIR /tinygithub

ENV PATH=$PATH:/usr/local/go/bin:/go/bin

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && go install

ENTRYPOINT ["tinygithub"]
CMD [ "server"]
