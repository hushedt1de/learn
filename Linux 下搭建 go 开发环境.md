# Linux 下搭建 Go 开发环境

本文档适用于 Linux 下安装 Go 开发环境，并特别处理国内网络环境下 Go 安装包下载和 Go module 依赖拉取较慢的问题。

截至 2026-06-19，Go 官方稳定版为 `go1.26.4`。如果后续版本更新，可以到 Go 官方下载页确认最新版：

- https://go.dev/dl/

## 1. 确认系统架构

先确认当前 Linux 系统架构：

```bash
uname -m
```

常见对应关系：

```text
x86_64  -> linux-amd64
aarch64 -> linux-arm64
```

下文默认以 `linux-amd64` 为例。如果你的机器是 ARM64，把文件名中的 `amd64` 改成 `arm64`。

## 2. 下载 Go 安装包

优先使用官方源：

```bash
wget https://go.dev/dl/go1.26.4.linux-amd64.tar.gz
```

如果官方源下载慢，可以使用国内镜像，例如阿里云镜像：

```bash
wget https://mirrors.aliyun.com/golang/go1.26.4.linux-amd64.tar.gz
```

ARM64 机器使用：

```bash
wget https://mirrors.aliyun.com/golang/go1.26.4.linux-arm64.tar.gz
```

## 3. 校验安装包

官方 `go1.26.4.linux-amd64.tar.gz` 的 SHA256 为：

```text
1153d3d50e0ac764b447adfe05c2bcf08e889d42a02e0fe0259bd47f6733ad7f
```

执行校验：

```bash
sha256sum go1.26.4.linux-amd64.tar.gz
```

输出值一致再继续安装。

## 4. 删除旧版本并安装

不要直接覆盖已有 `/usr/local/go`，先删除旧目录：

```bash
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz
```

如果你下载的是 ARM64 包，对应改成：

```bash
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.4.linux-arm64.tar.gz
```

## 5. 配置环境变量

如果使用 Bash：

```bash
nano ~/.bashrc
```

如果使用 Zsh：

```bash
nano ~/.zshrc
```

加入以下内容：

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

Bash 立即生效：

```bash
source ~/.bashrc
```

Zsh 立即生效：

```bash
source ~/.zshrc
```

## 6. 配置 Go Module 镜像源

国内环境建议配置 Go module 代理：

```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

验证配置：

```bash
go env GOPROXY
```

预期输出：

```text
https://goproxy.cn,direct
```

如果你需要拉取公司内网或私有 Git 仓库中的模块，建议配置 `GOPRIVATE`，避免私有模块路径走公共代理和校验数据库：

```bash
go env -w GOPRIVATE=git.example.com,*.corp.example.com
```

请把示例域名替换成你的真实私有 Git 域名。

## 7. 验证安装

检查 Go 版本：

```bash
go version
```

预期类似：

```text
go version go1.26.4 linux/amd64
```

创建一个简单项目测试：

```bash
mkdir -p ~/code/go-hello
cd ~/code/go-hello
go mod init example.com/hello
cat > main.go <<'EOF'
package main

import "fmt"

func main() {
	fmt.Println("hello go")
}
EOF
go run .
```

预期输出：

```text
hello go
```

## 8. 安装常用开发工具

安装 Go 语言服务器 `gopls`：

```bash
go install golang.org/x/tools/gopls@latest
```

安装调试器 `dlv`：

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

验证：

```bash
gopls version
dlv version
```

## 9. 常见问题

### go: command not found

检查 Go 是否已解压：

```bash
ls /usr/local/go/bin/go
```

检查 `PATH`：

```bash
echo $PATH
```

重新加载 shell 配置：

```bash
source ~/.bashrc
```

如果使用 Zsh：

```bash
source ~/.zshrc
```

### 下载依赖很慢

重新配置 Go module 代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go clean -modcache
go mod tidy
```

### 私有仓库拉取失败

配置 `GOPRIVATE`：

```bash
go env -w GOPRIVATE=你的私有Git域名
```

如果私有仓库使用 SSH，可以配置 Git URL 替换：

```bash
git config --global url."ssh://git@你的私有Git域名/".insteadOf "https://你的私有Git域名/"
```

## 10. 资料来源

- Go 官方安装文档：https://go.dev/doc/install
- Go 官方下载页：https://go.dev/dl/
- Go 官方 module proxy 说明：https://proxy.golang.org/
- Goproxy.cn 使用说明：https://goproxy.cn/
- 阿里云 Golang 镜像：https://mirrors.aliyun.com/golang/
