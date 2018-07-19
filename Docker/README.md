
此Dockerfile基于Ubuntu 18.04
```
Ubuntu安装docker：
14.04 版本系统中已经自带了 Docker 包，可以直接安装。
$ sudo apt-get update
$ sudo apt-get install -y docker.io
$ sudo ln -sf /usr/bin/docker.io /usr/local/bin/docker
$ sudo sed -i '$acomplete -F _docker docker' /etc/bash_completion.d/docker.io

要安装最新的 Docker 版本，首先需要安装 apt-transport-https 支持，之后通过添加源来安装。
$ sudo apt-get install apt-transport-https
$ sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9
$ sudo bash -c "echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"
$ sudo apt-get update
$ sudo apt-get install lxc-docker


如果是较14.04 之前版本
低版本的 Ubuntu 系统，需要先更新内核。
$ sudo apt-get update
$ sudo apt-get install linux-image-generic-lts-raring linux-headers-generic-lts-raring
$ sudo reboot
然后重复上面的步骤即可。
安装之后启动 Docker 服务。
$ sudo service docker start

CentOS 系列安装 Docker
Docker 支持 CentOS6 及以后的版本。

对于 CentOS6，可以使用 EPEL 库安装 Docker，命令如下
$ sudo yum install http://mirrors.yun-idc.com/epel/6/i386/epel-release-6-8.noarch.rpm
$ sudo yum install docker-io

CentOS7 系统 CentOS-Extras 库中已带 Docker，可以直接安装：
$ sudo yum install docker
安装之后启动 Docker 服务，并让它随系统启动自动加载。
$ sudo service docker start
$ sudo chkconfig docker on

安装并启动Docker完毕

下载的Dockerfile目录下，执行Dockerfile来构建镜像文件
sudo docker build -t="node:3.6" . （双引号里的内容为镜像文件名称，可自行命名）

查看构建Docker镜像文件
docker images

启动容器
通过镜像文件启动进入容器
sudo docker run -t -i node:3.6 /bin/bash  （node:3.6位镜像文件名称）
进入容器
go-ecoball位于root/go/src/github.com/ecoball/go-ecoball/路径下，这时候就可以在容器里启动go-ecoball服务

守护态运行容器
sudo docker run -t -i -d node:3.6 /bin/bash
查看运行的容器的信息，包括容器id等
docker ps
操作守护态容器
sudo docker attach 容器id

退出容器
exit
删除容器
docker rm 容器id
删除镜像文件
docker rmi 镜像id (先删除依赖这个镜像的所有容器)
