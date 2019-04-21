## 简介
docker为应用软件提供一个完整的独立的运行环境，比物理机与虚拟机更加轻量。

- 官网: <https://www.docker.com/>
- 文档: <https://docs.docker.com/>
- 源码: <https://github.com/docker/docker-ce>
- 入门: <https://yeasy.gitbooks.io/docker_practice>

配置镜像加速器

MAC->Preferences->Daemon->Register Mirrors->"https://registry.docker-cn.com"

### 基本命令
下载镜像，启动容器。
```
$ docker pull busybox:latest
$ docker run -it busybox
#
```

挂载目录，启动容器。
```
$ docker run -it -v ~/share:/home/share busybox
```

### 镜像管理 docker image

- 查看: docker image ls
- 删除: docker image rm
- 清理: docker image prune

### 容器管理 docker container
- 查看: docker container ls
- 查看: docker container ls -a
- 清理: docker container prune

### 启动容器 docker run
- 交互式启动: docker run -it busybox

- 守护式启动: docker run -dt busybox
  - 交互式连接: docker exec -it *container* sh
  - 一次性执行: docker exec *container* ls
  - 停止容器: docker stop *container*

### 制作镜像

- 交互式: docker commit *container* *repos:tag*
- 脚本式: docker build *deploy_path*

```
$ mkdir image && cd image
$ vi Dockerfile
FROM debian
RUN apt-get update\
      && apt-get install python \
      && apt-get install git
$ docker build .
```

