# MySQL 8.4 LTS 学习环境安装指南

这份文档用于搭建一个适合学习 MySQL 的本地环境。文档提供两种安装方式：

- Docker 安装：推荐给初学者作为第一套学习环境，版本固定、清理方便、不容易污染系统。
- Ubuntu/Debian 原生安装：更贴近服务器部署方式，适合想理解 Linux 服务管理、配置文件和系统目录的人。

本文以 MySQL 8.4 LTS 为默认版本。官方参考文档：

- MySQL 8.4 Linux 安装文档：https://dev.mysql.com/doc/refman/8.4/en/linux-installation.html
- MySQL 8.4 Docker 部署文档：https://dev.mysql.com/doc/refman/8.4/en/docker-mysql-getting-started.html
- Docker Hub MySQL 官方镜像：https://hub.docker.com/_/mysql

## 1. Docker 和 Linux 原生安装有什么区别

Docker 可以理解为把 MySQL 和它运行需要的环境打包到一个独立容器里。你在宿主机上用 Docker 命令启动、停止、删除这个容器。

Linux 原生安装则是把 MySQL 直接安装到系统里，由系统服务管理器 `systemd` 管理，配置、日志和数据目录也直接落在系统目录中。

| 对比项 | Docker 安装 | Linux 原生安装 |
| --- | --- | --- |
| 学习成本 | 需要先学几个 Docker 命令 | 需要理解 Linux 包管理和系统服务 |
| 版本控制 | 很容易固定到 `8.4` | 受软件源影响更大 |
| 清理重装 | 删除容器和数据卷即可 | 需要清理包、服务、配置和数据目录 |
| 系统污染 | 较少 | 较多 |
| 贴近生产服务器 | 适合容器化部署场景 | 适合传统服务器部署场景 |
| 推荐用途 | 本地学习、实验、反复重装 | 学习 Linux 服务管理和真实部署 |

如果你只是想先开始学习 SQL、表、索引、事务，建议先用 Docker。Docker 并不会让 MySQL 学习更复杂，主要多掌握几条容器命令。

## 2. 方式一：使用 Docker 安装 MySQL 8.4 LTS

继续之前，确认 Docker 已经可用：

```bash
docker --version
docker run hello-world
```

如果你的系统尚未配置免 `sudo` 使用 Docker，可以把本文中的 `docker ...` 命令改成 `sudo docker ...`。

### 2.1 拉取 MySQL 镜像

```bash
docker pull mysql:8.4
```

检查镜像：

```bash
docker images mysql
```

### 2.2 创建数据目录

建议把 MySQL 数据放到一个明确的目录中，方便后续清理和观察。

```bash
mkdir -p ~/mysql-learning/mysql-data
```

### 2.3 启动 MySQL 容器

下面的命令会创建一个名为 `mysql84` 的容器：

```bash
docker run --name mysql84 \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=Root_pass_123 \
  -v ~/mysql-learning/mysql-data:/var/lib/mysql \
  -d mysql:8.4
```

参数说明：

- `--name mysql84`：容器名。
- `-p 3306:3306`：把本机 `3306` 端口映射到容器内 MySQL 的 `3306` 端口。
- `-e MYSQL_ROOT_PASSWORD=Root_pass_123`：设置 MySQL `root` 用户密码。
- `-v ~/mysql-learning/mysql-data:/var/lib/mysql`：把数据保存到本机目录。
- `-d mysql:8.4`：后台运行 MySQL 8.4 镜像。

学习环境可以使用上面的密码。真实项目不要把密码直接写在命令、脚本或公开文档中。

### 2.4 查看启动状态

```bash
docker ps
docker logs mysql84
```

第一次启动时 MySQL 会初始化数据目录，可能需要等待几十秒。日志中出现 MySQL 准备接受连接的信息后，再继续连接。

### 2.5 连接 MySQL

方式一：进入容器内部连接。

```bash
docker exec -it mysql84 mysql -uroot -p
```

输入密码：

```text
Root_pass_123
```

方式二：如果本机已安装 MySQL 客户端，可以从宿主机连接。

```bash
mysql -h127.0.0.1 -P3306 -uroot -p
```

### 2.6 验证数据库可用

登录 MySQL 后执行：

```sql
SELECT VERSION();

CREATE DATABASE learn_mysql;

USE learn_mysql;

CREATE TABLE students (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(50) NOT NULL,
  age INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO students (name, age) VALUES
  ('Alice', 18),
  ('Bob', 20);

SELECT * FROM students;
```

能看到两条学生数据，就说明 MySQL 已经可以正常使用。

退出 MySQL：

```sql
exit;
```

### 2.7 停止、启动和删除容器

停止容器：

```bash
docker stop mysql84
```

再次启动：

```bash
docker start mysql84
```

删除容器：

```bash
docker stop mysql84
docker rm mysql84
```

注意：上面的删除只删除容器，不删除 `~/mysql-learning/mysql-data` 中的数据。只要数据目录还在，用同样的挂载路径重新创建容器后，数据仍然存在。

如果你确认要删除学习数据：

```bash
rm -rf ~/mysql-learning/mysql-data
```

删除数据目录是不可恢复操作，执行前确认里面没有需要保留的数据。

## 3. 方式二：在 Ubuntu/Debian 原生安装 MySQL

这一节适用于 Ubuntu/Debian 系统。不同发行版的软件源中 MySQL 版本可能不同，如果必须固定到 MySQL 8.4 LTS，优先参考 MySQL 官方 APT Repository。

### 3.1 更新软件源

```bash
sudo apt update
```

### 3.2 安装 MySQL Server

```bash
sudo apt install mysql-server
```

查看已安装版本：

```bash
mysql --version
```

如果版本不是你想要的 MySQL 8.4 LTS，请改用 MySQL 官方 APT Repository 方式安装。官方入口：

```text
https://dev.mysql.com/doc/refman/8.4/en/linux-installation.html
```

### 3.3 查看服务状态

```bash
sudo systemctl status mysql
```

常用服务命令：

```bash
# 启动 MySQL
sudo systemctl start mysql

# 停止 MySQL
sudo systemctl stop mysql

# 重启 MySQL
sudo systemctl restart mysql

# 设置开机自启
sudo systemctl enable mysql
```

### 3.4 初始化安全配置

Ubuntu/Debian 安装完成后，可以执行：

```bash
sudo mysql_secure_installation
```

这个命令通常会引导你完成以下设置：

- 设置或调整 root 密码策略。
- 移除匿名用户。
- 禁止远程 root 登录。
- 删除测试数据库。
- 刷新权限表。

学习环境可以按提示选择相对简单的配置；如果机器暴露在公网，应该使用更严格的密码和访问控制。

### 3.5 登录 MySQL

在部分 Ubuntu/Debian 环境中，`root` 用户默认使用系统认证插件，可能需要用 `sudo` 登录：

```bash
sudo mysql
```

如果已经配置了 MySQL root 密码，也可以尝试：

```bash
mysql -uroot -p
```

### 3.6 验证数据库可用

登录 MySQL 后执行：

```sql
SELECT VERSION();

CREATE DATABASE learn_mysql;

USE learn_mysql;

CREATE TABLE students (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(50) NOT NULL,
  age INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO students (name, age) VALUES
  ('Alice', 18),
  ('Bob', 20);

SELECT * FROM students;
```

能看到两条学生数据，就说明 MySQL 原生安装环境可用。

退出 MySQL：

```sql
exit;
```

### 3.7 常见目录

Ubuntu/Debian 原生安装后，常见目录通常如下：

```text
/etc/mysql/          配置文件目录
/var/lib/mysql/      数据目录
/var/log/mysql/      日志目录
```

不要直接手动删除或修改数据目录中的文件。需要备份、恢复或迁移时，优先使用 `mysqldump`、`mysql` 客户端或官方工具。

### 3.8 卸载 MySQL

仅卸载软件包：

```bash
sudo apt remove mysql-server
```

卸载并清理配置：

```bash
sudo apt purge mysql-server
sudo apt autoremove
```

如果还要删除数据目录，需要先确认数据不再需要：

```bash
sudo rm -rf /var/lib/mysql
```

删除数据目录是不可恢复操作，不要在生产环境中直接执行。

## 4. 常见问题与排查

### 4.1 Docker 容器启动后立刻退出

查看日志：

```bash
docker logs mysql84
```

常见原因：

- 没有设置 `MYSQL_ROOT_PASSWORD`。
- 本机目录权限导致 MySQL 无法写入数据。
- 本机 `3306` 端口已被其他 MySQL 占用。

检查端口占用：

```bash
sudo lsof -i :3306
```

如果端口已被占用，可以把宿主机端口改成 `3307`：

```bash
docker run --name mysql84 \
  -p 3307:3306 \
  -e MYSQL_ROOT_PASSWORD=Root_pass_123 \
  -v ~/mysql-learning/mysql-data:/var/lib/mysql \
  -d mysql:8.4
```

连接时使用：

```bash
mysql -h127.0.0.1 -P3307 -uroot -p
```

### 4.2 Docker 中修改 root 密码不生效

MySQL 镜像只会在首次初始化空数据目录时读取 `MYSQL_ROOT_PASSWORD`。如果 `~/mysql-learning/mysql-data` 已经有旧数据，重新运行容器时这个环境变量不会覆盖旧密码。

解决方式：

- 记得旧密码：登录后用 SQL 修改密码。
- 不需要旧数据：删除数据目录后重新创建容器。

### 4.3 原生安装后无法用 `mysql -uroot -p` 登录

Ubuntu/Debian 可能默认让 MySQL `root` 使用系统认证。先尝试：

```bash
sudo mysql
```

进入后可以再创建一个学习用户：

```sql
CREATE USER 'learn'@'localhost' IDENTIFIED BY 'Learn_pass_123';
GRANT ALL PRIVILEGES ON *.* TO 'learn'@'localhost';
FLUSH PRIVILEGES;
```

之后使用：

```bash
mysql -ulearn -p
```

### 4.4 忘记当前装的是哪个版本

```bash
mysql --version
```

登录 MySQL 后也可以执行：

```sql
SELECT VERSION();
```

## 5. 建议的学习顺序

安装完成后，可以按这个顺序继续学习：

1. 数据库、表、行、列的基本概念。
2. `CREATE DATABASE`、`CREATE TABLE`、`INSERT`、`SELECT`。
3. `WHERE`、`ORDER BY`、`LIMIT`、聚合函数。
4. 主键、唯一索引、普通索引。
5. 表关联：`JOIN`。
6. 事务：`START TRANSACTION`、`COMMIT`、`ROLLBACK`。
7. 备份和恢复：`mysqldump`。

建议先用 Docker 环境大胆实验。需要重来时，停止并删除容器，再清理数据目录即可重新开始。
