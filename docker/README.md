# Docker 安装与基本使用

这份文档用于学习 Docker 的安装、核心概念和基础命令。后续学习 MySQL、Redis、Nginx 等软件时，可以先在这里完成 Docker 环境准备，再到对应软件目录查看具体容器运行方式。

官方参考文档：

- Docker Engine Ubuntu 安装文档：https://docs.docker.com/engine/install/ubuntu/
- Docker Engine Debian 安装文档：https://docs.docker.com/engine/install/debian/
- Docker Engine Linux 安装入口：https://docs.docker.com/engine/install/

## 1. Docker 是什么

Docker 可以把应用和运行环境打包成镜像，再用镜像启动容器。容器和宿主机共享操作系统内核，但有独立的文件系统、进程空间和网络配置。

几个基础概念：

- 镜像：应用和依赖环境的模板，例如 `mysql:8.4`、`nginx:latest`。
- 容器：镜像运行起来后的实例，可以启动、停止、删除。
- 仓库：存放镜像的地方，例如 Docker Hub。
- 端口映射：把容器内端口暴露到宿主机端口。
- 数据卷：把容器数据保存到宿主机，避免容器删除后数据丢失。

Docker 适合学习环境的原因：

- 可以快速安装不同版本的软件。
- 删除容器后容易重来。
- 不容易污染宿主机系统目录。
- 多个软件可以用独立容器隔离运行。

## 2. 在 Ubuntu/Debian 安装 Docker Engine

下面步骤适用于 Ubuntu/Debian。其他 Linux 发行版请参考 Docker 官方 Linux 安装入口。

### 2.1 确认系统版本

```bash
cat /etc/os-release
```

如果输出中能看到 `Ubuntu` 或 `Debian`，可以继续使用下面的步骤。

### 2.2 卸载可能冲突的旧包

Linux 发行版的软件源里可能有 `docker.io`、`docker-compose` 等非 Docker 官方包。为了避免冲突，先执行：

```bash
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do
  sudo apt remove -y "$pkg"
done
```

如果提示某些包没有安装，是正常情况。

### 2.3 安装基础依赖

```bash
sudo apt update
sudo apt install -y ca-certificates curl
```

创建 APT keyrings 目录：

```bash
sudo install -m 0755 -d /etc/apt/keyrings
```

### 2.4 添加 Docker 官方 GPG key

Ubuntu 执行：

```bash
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
```

Debian 执行：

```bash
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
```

### 2.5 添加 Docker 官方 APT 源

Ubuntu 执行：

```bash
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
```

Debian 执行：

```bash
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/debian
Suites: $(. /etc/os-release && echo "$VERSION_CODENAME")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
```

更新软件源：

```bash
sudo apt update
```

### 2.6 安装 Docker Engine

```bash
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

安装完成后检查 Docker 服务：

```bash
sudo systemctl status docker
```

如果 Docker 没有运行，手动启动：

```bash
sudo systemctl start docker
```

设置开机自启：

```bash
sudo systemctl enable docker
```

### 2.7 验证 Docker 可用

```bash
sudo docker --version
sudo docker run hello-world
```

如果 `hello-world` 能正常输出说明，Docker 基础环境可用。

第一次运行 `hello-world` 时，如果本地没有镜像，Docker 会先输出：

```text
Unable to find image 'hello-world:latest' locally
```

这不是错误，只表示本地还没有这个镜像，Docker 会继续尝试从 Docker Hub 下载。只有后面出现 `i/o timeout`、`connection refused`、`TLS handshake timeout` 等信息时，才说明访问 Docker Hub 失败。

### 2.8 可选：配置 Docker Hub 镜像加速器

国内网络访问 Docker Hub 可能超时。如果出现类似错误：

```text
failed to resolve reference "docker.io/library/hello-world:latest"
dial tcp ...:443: i/o timeout
```

可以给 Docker 配置 registry mirror。编辑 Docker daemon 配置文件：

```bash
sudo mkdir -p /etc/docker
sudo nano /etc/docker/daemon.json
```

写入以下内容，把示例地址替换成你当前可用的镜像加速器地址：

```json
{
  "registry-mirrors": [
    "https://你的镜像加速器地址"
  ]
}
```

保存后重启 Docker：

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

检查配置是否生效：

```bash
sudo docker info
```

在输出中找到 `Registry Mirrors`，如果能看到你配置的地址，说明 Docker 已读取配置。

再次验证：

```bash
sudo docker run hello-world
```

镜像加速器地址经常会变更、限流或停止服务。建议优先使用你云厂商账号提供的专属镜像加速器，或者学校、公司网络管理员提供的地址。

### 2.9 可选：允许当前用户不加 sudo 使用 Docker

默认情况下，普通用户直接运行 `docker` 可能会遇到权限错误。如果这是个人学习机器，可以把当前用户加入 `docker` 用户组：

```bash
sudo usermod -aG docker "$USER"
```

然后退出当前终端并重新登录，或重启系统。重新登录后验证：

```bash
docker run hello-world
```

注意：加入 `docker` 用户组的用户基本等同于拥有较高的系统权限。只建议在个人学习环境中这样做；如果是多人服务器，应继续使用 `sudo docker ...` 或咨询管理员。

## 3. Docker 基本使用

### 3.1 镜像命令

```bash
# 拉取镜像
docker pull nginx:latest

# 查看本机镜像
docker images

# 删除镜像
docker rmi nginx:latest
```

如果没有配置免 `sudo`，把命令写成 `sudo docker ...`。

### 3.2 容器命令

```bash
# 启动一个 Nginx 容器
docker run --name web-demo -p 8080:80 -d nginx:latest

# 查看正在运行的容器
docker ps

# 查看所有容器，包括已停止的容器
docker ps -a

# 查看容器日志
docker logs web-demo

# 停止容器
docker stop web-demo

# 启动已停止的容器
docker start web-demo

# 删除容器，删除前需要先停止
docker rm web-demo
```

启动 Nginx 后，可以在浏览器访问：

```text
http://127.0.0.1:8080
```

### 3.3 进入容器执行命令

```bash
docker exec -it web-demo bash
```

如果镜像里没有 `bash`，可以尝试：

```bash
docker exec -it web-demo sh
```

退出容器 shell：

```bash
exit
```

### 3.4 端口映射

容器默认有自己的网络空间。`-p 宿主机端口:容器端口` 可以把容器服务暴露到宿主机。

示例：

```bash
docker run --name web-demo -p 8080:80 -d nginx:latest
```

含义：

- 容器内 Nginx 监听 `80` 端口。
- 宿主机通过 `8080` 端口访问容器。
- 访问地址是 `http://127.0.0.1:8080`。

### 3.5 数据卷和目录挂载

容器删除后，容器内部文件通常也会随之消失。需要保留数据时，应使用数据卷或目录挂载。

目录挂载示例：

```bash
mkdir -p ~/docker-learning/nginx-html
echo "hello docker" > ~/docker-learning/nginx-html/index.html

docker run --name web-demo \
  -p 8080:80 \
  -v ~/docker-learning/nginx-html:/usr/share/nginx/html:ro \
  -d nginx:latest
```

参数说明：

- `-v 宿主机目录:容器目录`：把宿主机目录挂载到容器内。
- `:ro`：只读挂载，容器不能修改宿主机目录内容。

## 4. 常见清理命令

```bash
# 删除已停止的容器
docker container prune

# 删除未被使用的镜像
docker image prune

# 查看 Docker 占用空间
docker system df

# 清理未使用的容器、网络、镜像和构建缓存
docker system prune
```

`docker system prune` 会删除未使用资源。执行前确认没有需要保留的容器或镜像。

## 5. 建议的学习顺序

1. 理解镜像和容器的区别。
2. 掌握 `docker pull`、`docker run`、`docker ps`、`docker logs`、`docker exec`、`docker stop`、`docker rm`。
3. 理解端口映射，例如 `-p 8080:80`。
4. 理解目录挂载和数据持久化，例如 `-v 本机目录:容器目录`。
5. 再学习 Docker Compose，用一个配置文件管理多个容器。
