# 指定基础镜像为 Ubuntu 20.04
FROM golang:latest

# 设置工作目录
WORKDIR /app

# 更新系统并安装 Git
RUN apt-get update && apt-get install -y git

RUN git clone https://github.com/adlternative/tinygithub.git && \
    cd tinygithub && \
    go install

CMD [ "/go/bin/tinygithub", "server"]
