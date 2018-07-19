#基于Ubuntu18.04
FROM ubuntu:18.04
MAINTAINER Docker Newbee <newbee@docker.com>

#安装git gcc openssl
RUN apt-get update
RUN apt-get install -y git gcc make wget

#下载go安装包并且解压至/usr/local
RUN cd /root/ && wget https://storage.googleapis.com/golang/go1.10.3.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.10.3.linux-amd64.tar.gz

#设置环境变量
ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV PATH /usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

#下载代码打包生成执行文件
RUN cd /root/ && mkdir go && cd go && mkdir src && cd src && mkdir github.com && cd github.com && mkdir ecoball && cd ecoball \
   && git clone https://github.com/ecoball/go-ecoball.git && cd go-ecoball/ && make

EXPOSE 80

STOPSIGNAL SIGTERM
