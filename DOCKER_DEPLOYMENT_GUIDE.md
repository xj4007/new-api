# Docker容器化部署完整指南

## 📋 配置文件说明

### 修改后的配置特点

#### ✅ docker-compose copy.yml
- **MySQL**: 端口映射 `13306:3306`，数据保存在 `./mysql_data`
- **Redis**: 端口映射 `26739:6379`，数据保存在 `./redis_data`
- **New-API**: 端口映射 `3010:3000`

#### ✅ .env copy.example
- **容器内通信**: `mysql:3306` 和 `redis:6379` (Docker网络)
- **外部访问**: `localhost:13306` 和 `localhost:26739`
- **数据持久化**: 本地目录映射

---

## 🚀 部署步骤

### 1️⃣ 准备配置文件

```bash
# 复制配置文件
cp "docker-compose copy.yml" docker-compose.yml
cp ".env copy.example" .env

# 检查配置（可选：修改密码）
nano .env
```

### 2️⃣ 构建前端

```bash
cd web
bun install
bun run build
cd ..
```

### 3️⃣ 启动服务

```bash
# 启动所有容器
docker-compose up -d

# 查看启动日志
docker-compose logs -f
```

### 4️⃣ 验证服务

```bash
# 检查容器状态
docker-compose ps

# 验证MySQL连接（外部访问）
mysql -h localhost -P 13306 -usupport -pXIANjian4 new-api -e "SHOW TABLES;"

# 验证Redis连接
docker exec redis redis-cli -a XIANjian4SANyun ping

# 访问应用
curl http://localhost:3010/api/status
```

---

## 📂 数据目录结构

启动后会在项目根目录生成：

```
new-api/
├── mysql_data/          # MySQL数据文件（可直接复制迁移）
│   ├── new-api/         # 数据库文件
│   ├── ibdata1
│   └── ...
├── redis_data/          # Redis持久化文件
│   └── dump.rdb
├── data/                # 应用数据
└── logs/                # 应用日志
```

---

## 🔌 连接信息汇总

### 容器内部访问（应用使用）

| 服务 | 地址 | 配置项 |
|------|------|--------|
| MySQL | `mysql:3306` | `SQL_DSN` |
| Redis | `redis:6379` | `REDIS_CONN_STRING` |

### 外部工具访问

| 服务 | 地址 | 用户名 | 密码 |
|------|------|--------|------|
| MySQL | `localhost:13306` | `support` | `XIANjian4` |
| MySQL (root) | `localhost:13306` | `root` | `XIANjian4` |
| Redis | `localhost:26739` | - | `XIANjian4SANyun` |
| New-API | `http://localhost:3010` | - | - |

### 使用Navicat/DBeaver连接MySQL

```
主机: localhost (或服务器IP)
端口: 13306
用户: support
密码: XIANjian4
数据库: new-api
```

### 使用Redis Desktop Manager

```
地址: localhost
端口: 26739
密码: XIANjian4SANyun
数据库索引: 1
```

---

## 🚚 服务器迁移步骤

### 方法1: 直接复制数据目录（推荐）

#### 旧服务器操作

```bash
# 停止服务
docker-compose down

# 打包所有数据
tar czf new-api-data-backup.tar.gz \
  mysql_data/ \
  redis_data/ \
  data/ \
  logs/ \
  docker-compose.yml \
  .env \
  Dockerfile

# 传输到新服务器
scp new-api-data-backup.tar.gz user@new-server:/opt/
```

#### 新服务器操作

```bash
# 解压
cd /opt
mkdir new-api
tar xzf new-api-data-backup.tar.gz -C new-api/
cd new-api

# 设置MySQL数据目录权限（重要！）
sudo chown -R 999:999 mysql_data/
sudo chmod -R 755 mysql_data/

# 设置Redis数据目录权限
sudo chown -R 999:999 redis_data/

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 验证数据
docker-compose exec mysql mysql -usupport -pXIANjian4 new-api -e "SELECT COUNT(*) FROM users;"
```

### 方法2: SQL导出导入

```bash
# 旧服务器导出
docker-compose exec mysql mysqldump -usupport -pXIANjian4 new-api > new-api-backup.sql

# 新服务器导入
docker-compose up -d mysql  # 先启动MySQL
sleep 10  # 等待MySQL完全启动
docker-compose exec -T mysql mysql -usupport -pXIANjian4 new-api < new-api-backup.sql

# 启动其他服务
docker-compose up -d
```

---

## 🛠️ 常用管理命令

### 服务管理

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart

# 重启单个服务
docker-compose restart new-api
docker-compose restart mysql

# 查看日志
docker-compose logs -f new-api
docker-compose logs --tail=100 mysql
```

### 数据库管理

```bash
# 进入MySQL容器
docker-compose exec mysql bash

# 直接执行SQL
docker-compose exec mysql mysql -usupport -pXIANjian4 new-api -e "SHOW TABLES;"

# 备份数据库
docker-compose exec mysql mysqldump -usupport -pXIANjian4 new-api > backup-$(date +%Y%m%d).sql

# 恢复数据库
docker-compose exec -T mysql mysql -usupport -pXIANjian4 new-api < backup.sql
```

### Redis管理

```bash
# 进入Redis容器
docker-compose exec redis redis-cli -a XIANjian4SANyun

# 查看Redis信息
docker-compose exec redis redis-cli -a XIANjian4SANyun INFO

# 查看所有键
docker-compose exec redis redis-cli -a XIANjian4SANyun KEYS '*'

# 清空当前数据库
docker-compose exec redis redis-cli -a XIANjian4SANyun FLUSHDB
```

### 容器管理

```bash
# 查看容器状态
docker-compose ps

# 查看容器资源占用
docker stats new-api mysql redis

# 清理未使用的资源
docker system prune -f

# 重新构建镜像
docker-compose build --no-cache
docker-compose up -d
```

---

## ⚠️ 重要注意事项

### 1. 首次启动MySQL可能较慢
MySQL首次启动需要初始化数据库，可能需要1-2分钟，查看日志：
```bash
docker-compose logs -f mysql
# 等待看到：ready for connections
```

### 2. 数据目录权限
MySQL容器使用UID 999，如果出现权限错误：
```bash
sudo chown -R 999:999 mysql_data/
sudo chown -R 999:999 redis_data/
```

### 3. 端口冲突
如果端口已被占用，修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "13307:3306"  # 改为其他端口
```

### 4. 防火墙配置
如果需要远程访问MySQL/Redis，需要开放端口：
```bash
# Ubuntu/Debian
sudo ufw allow 13306/tcp
sudo ufw allow 26739/tcp

# CentOS/RHEL
sudo firewall-cmd --add-port=13306/tcp --permanent
sudo firewall-cmd --add-port=26739/tcp --permanent
sudo firewall-cmd --reload
```

### 5. 生产环境安全建议
- ✅ 修改所有默认密码
- ✅ 修改 `SESSION_SECRET` 为复杂字符串
- ✅ 不要将MySQL/Redis端口暴露到公网
- ✅ 使用反向代理（Nginx）配置HTTPS
- ✅ 定期备份数据

---

## 📊 监控和维护

### 定时备份脚本

创建 `backup.sh`:
```bash
#!/bin/bash
BACKUP_DIR="/backup/new-api"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份MySQL
docker-compose exec -T mysql mysqldump -usupport -pXIANjian4 new-api | gzip > $BACKUP_DIR/mysql_$DATE.sql.gz

# 备份数据目录
tar czf $BACKUP_DIR/data_$DATE.tar.gz mysql_data/ redis_data/ data/

# 保留最近7天的备份
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_DIR"
```

添加到crontab:
```bash
chmod +x backup.sh
crontab -e
# 每天凌晨3点备份
0 3 * * * /opt/new-api/backup.sh >> /var/log/new-api-backup.log 2>&1
```

### 日志清理

```bash
# 清理应用日志
find logs/ -name "*.log" -mtime +30 -delete

# 清理Docker日志
echo '{"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}' | sudo tee /etc/docker/daemon.json
sudo systemctl restart docker
```

---

## 🔧 故障排查

### MySQL连接失败

```bash
# 检查MySQL是否运行
docker-compose ps mysql

# 查看MySQL日志
docker-compose logs mysql

# 测试连接
docker-compose exec mysql mysql -usupport -pXIANjian4 -e "SELECT 1;"
```

### Redis连接失败

```bash
# 检查Redis是否运行
docker-compose ps redis

# 测试连接
docker-compose exec redis redis-cli -a XIANjian4SANyun ping
```

### 应用无法访问

```bash
# 检查容器状态
docker-compose ps

# 查看应用日志
docker-compose logs --tail=100 new-api

# 检查健康状态
curl http://localhost:3010/api/status
```

### 数据丢失问题

确保数据目录映射正确：
```bash
# 检查挂载
docker inspect mysql | grep -A 10 Mounts
docker inspect redis | grep -A 10 Mounts

# 确认数据存在
ls -lh mysql_data/
ls -lh redis_data/
```

---

## 📞 技术支持

如遇到问题：
1. 查看日志：`docker-compose logs -f`
2. 检查配置：`cat .env`
3. 验证网络：`docker network inspect new-api_default`
4. 提交Issue：https://github.com/Calcium-Ion/new-api/issues
