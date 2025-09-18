# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

New API 是一个基于Go语言开发的新一代大模型网关与AI资产管理系统，在 One API 基础上进行了二次开发。项目采用前后端分离架构：
- 后端：Go + Gin + GORM
- 前端：React + Vite + Semi UI
- 数据库支持：SQLite（默认）、MySQL、PostgreSQL
- 缓存支持：Redis、内存缓存

## 常用开发命令

### 环境准备（零基础部署指南）

由于用户已安装Docker Compose，推荐使用Docker部署：

```bash
# 1. 使用Docker Compose部署（推荐）
# 确保docker-compose.yml文件存在并已配置
docker-compose up -d

# 2. 查看服务状态
docker-compose ps
docker-compose logs new-api

# 3. 停止服务
docker-compose down

# 4. 重新构建并启动（代码更新后）
docker-compose up -d --build
```

### 本地开发命令

```bash
# 启动完整开发环境（前端+后端）
make all

# 仅构建前端
make build-frontend

# 仅启动后端开发服务器
make start-backend

# 手动启动后端（用于调试）
go run main.go

# 前端开发（需要进入web目录）
cd web
bun install    # 安装依赖
bun run dev    # 启动开发服务器
bun run build  # 构建生产版本

# 前端代码格式化和检查
cd web
bun run lint      # 检查代码格式
bun run lint:fix  # 自动修复格式问题
bun run eslint    # ESLint检查
bun run eslint:fix # 自动修复ESLint问题
```

### Docker相关命令

```bash
# 构建Docker镜像
docker build -t new-api .

# 运行单个容器（使用SQLite）
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v /home/ubuntu/data/new-api:/data \
  calciumion/new-api:latest

# 运行单个容器（使用MySQL）
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" \
  -e TZ=Asia/Shanghai \
  -v /home/ubuntu/data/new-api:/data \
  calciumion/new-api:latest
```

### 系统服务管理（Linux部署）

```bash
# 配置systemd服务（需要修改路径）
sudo cp one-api.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable one-api
sudo systemctl start one-api
sudo systemctl status one-api

# 查看服务日志
sudo journalctl -u one-api -f
```

### 数据库相关

```bash
# Go模块管理
go mod download  # 下载依赖
go mod tidy      # 清理依赖
go mod vendor    # 创建vendor目录

# 数据库迁移会自动进行，无需手动执行
# 项目启动时会自动初始化数据库结构
```

## 项目架构

### 目录结构说明

```
├── main.go                 # 应用入口点，初始化所有组件
├── common/                 # 通用工具和配置
│   ├── constants.go        # 常量定义
│   ├── database.go         # 数据库配置
│   ├── redis.go           # Redis配置
│   └── init.go            # 初始化逻辑
├── controller/            # HTTP控制器层
│   ├── channel.go         # 渠道管理
│   ├── user.go           # 用户管理
│   ├── token.go          # 令牌管理
│   └── relay.go          # 请求中继
├── model/                # 数据模型层
│   ├── channel.go        # 渠道模型
│   ├── user.go          # 用户模型
│   ├── token.go         # 令牌模型
│   └── main.go          # 数据库初始化
├── router/               # 路由配置
│   ├── api-router.go    # API路由
│   ├── relay-router.go  # 中继路由
│   └── web-router.go    # Web路由
├── middleware/           # 中间件
├── service/             # 业务服务层
├── relay/               # 请求中继逻辑
├── setting/             # 设置管理
├── web/                 # React前端应用
│   ├── package.json     # 前端依赖配置
│   ├── src/            # 源代码
│   └── dist/           # 构建输出
└── docker-compose.yml   # Docker部署配置
```

### 核心架构组件

1. **应用启动流程**（main.go:35-207）：
   - 加载环境变量和配置
   - 初始化数据库连接
   - 启动Redis缓存
   - 初始化HTTP服务器
   - 启动后台任务

2. **数据库架构**：
   - 主数据库：用户、渠道、令牌等核心数据
   - 日志数据库：请求日志和统计数据
   - 支持SQLite、MySQL、PostgreSQL

3. **缓存系统**：
   - Redis缓存：渠道信息、用户状态
   - 内存缓存：高频访问数据
   - 自动同步机制

4. **请求中继系统**：
   - 多渠道负载均衡
   - 自动重试机制
   - 渠道健康检查

## 重要配置文件

### 环境变量配置（.env）

基础配置：
```bash
PORT=3000                                    # 服务端口
SQL_DSN=root:123456@tcp(mysql:3306)/new-api # 数据库连接
REDIS_CONN_STRING=redis://redis             # Redis连接
TZ=Asia/Shanghai                             # 时区设置
SESSION_SECRET=your-secret-key               # 会话密钥（多机部署必需）
```

高级配置：
```bash
STREAMING_TIMEOUT=300          # 流模式超时时间（秒）
GENERATE_DEFAULT_TOKEN=false   # 新用户是否生成默认令牌
MEMORY_CACHE_ENABLED=true      # 启用内存缓存
SYNC_FREQUENCY=60              # 缓存同步频率（秒）
CHANNEL_UPDATE_FREQUENCY=30    # 渠道更新频率（秒）
BATCH_UPDATE_ENABLED=true      # 批量更新功能
UPDATE_TASK=true               # 更新异步任务
DIFY_DEBUG=true                # Dify调试输出
```

### Docker Compose配置要点

docker-compose.yml包含三个服务：
- new-api: 主应用服务，端口3000
- mysql: 数据库服务，内部通信
- redis: 缓存服务，内部通信

关键挂载点：
- `./data:/data`: 数据持久化
- `./logs:/app/logs`: 日志文件

## 开发指南

### 添加新功能的流程

1. **数据模型**：在model/目录添加新的数据结构
2. **控制器**：在controller/目录实现业务逻辑
3. **路由**：在router/目录配置API路由
4. **前端**：在web/src目录实现前端界面

### 常见开发任务

1. **添加新的AI提供商支持**：
   - 在relay/目录添加适配器
   - 更新controller/channel.go
   - 修改前端渠道管理界面

2. **修改数据库结构**：
   - 更新model/目录中的结构体
   - GORM会自动处理迁移（AutoMigrate）

3. **添加新的API端点**：
   - 在controller/目录添加处理函数
   - 在router/目录配置路由
   - 更新前端API调用

### 调试和故障排除

1. **查看应用日志**：
   ```bash
   docker-compose logs -f new-api
   ```

2. **查看数据库连接**：
   ```bash
   docker-compose exec mysql mysql -uroot -p123456 -e "SHOW DATABASES;"
   ```

3. **Redis缓存检查**：
   ```bash
   docker-compose exec redis redis-cli ping
   ```

4. **健康检查**：
   - 访问 http://localhost:3000/api/status
   - 检查各项服务状态

### 多机部署注意事项

必须设置的环境变量：
- `SESSION_SECRET`：确保多机登录状态一致
- `CRYPTO_SECRET`：共用Redis时数据加密
- `NODE_TYPE=slave`：从节点配置

## 前端开发说明

前端基于React + Vite构建，使用Semi UI组件库：

### 技术栈
- React 18 + React Router
- Vite + TypeScript
- Semi UI + Tailwind CSS  
- Axios + i18next

### 开发环境
```bash
cd web
bun install          # 安装依赖（比npm更快）
bun run dev          # 启动开发服务器（端口默认5173）
bun run build        # 构建生产版本
```

### 代码规范
```bash
bun run lint         # 代码格式检查
bun run lint:fix     # 自动修复格式问题
bun run eslint       # ESLint检查
bun run eslint:fix   # 自动修复ESLint问题
```

## 部署生产环境

### 使用Docker Compose（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/Calcium-Ion/new-api.git
cd new-api

# 2. 根据需要修改docker-compose.yml中的配置
# 特别注意修改MySQL密码、数据库名等

# 3. 启动服务
docker-compose up -d

# 4. 查看服务状态
docker-compose ps

# 5. 访问应用
# http://your-server-ip:3000
```

### 使用预构建镜像

```bash
docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -v $PWD/data:/data \
  -v $PWD/logs:/app/logs \
  calciumion/new-api:latest
```

## 关键特性

1. **多模型支持**：OpenAI、Claude、Gemini等主流模型
2. **中继功能**：自动负载均衡和故障转移
3. **用户管理**：令牌管理、配额控制、权限分组
4. **数据统计**：使用量统计、费用计算、数据看板
5. **多种登录**：邮箱、GitHub、LinuxDO、Telegram等
6. **国际化**：多语言界面支持
7. **缓存优化**：提示缓存计费、渠道缓存

## 疑难解答

### 常见问题

1. **数据库连接失败**：检查SQL_DSN环境变量格式
2. **Redis连接问题**：确认REDIS_CONN_STRING配置
3. **前端访问404**：确保前端构建文件存在于web/dist目录
4. **令牌验证失败**：检查SESSION_SECRET配置
5. **多机部署问题**：确保所有实例使用相同的SESSION_SECRET和CRYPTO_SECRET

### 性能优化

1. **启用缓存**：设置Redis和内存缓存
2. **批量更新**：启用BATCH_UPDATE_ENABLED
3. **数据库优化**：调整连接池参数
4. **负载均衡**：使用多渠道分流

该项目结构清晰，采用标准的Go web应用架构，适合快速开发和部署AI网关服务。


请先备份redis的数据再进行构建，防止清空

🚀 现在请在服务器上重新构建：

  # 清理之前的构建缓存
  docker-compose down
  docker system prune -f

  # 重新构建（会使用新的代理配置）
  docker-compose build --no-cache

  # 或者直接构建并启动
  docker-compose up -d --build


🚀 现在请重新启动服务：
  # 停止现有容器
  docker-compose down

  # 重新启动
  docker-compose up -d

  # 查看日志
  docker-compose logs -f new-api

  🌐 访问地址：
  - 新的访问地址: http://您的服务器IP:3010
  - 健康检查: http://您的服务器IP:3010/api/status











主从架构的真实工作方式

  主节点 (NODE_TYPE=master)

  - ✅ 处理API请求：完整的API服务功能
  - ✅ 数据库管理：负责数据库迁移和结构更新
  - ✅ 后台任务调度：执行批量任务更新
  - ✅ 完整功能：所有功能都启用

  从节点 (NODE_TYPE=slave)

  - ✅ 处理API请求：完整的API服务功能
  - ❌ 跳过数据库迁移：不执行数据库结构变更
  - ❌ 跳过后台任务：不执行批量更新任务
  - ✅ 专注请求处理：减少资源消耗

  不同服务器部署配置

  主服务器配置 (.env)

  # 主节点 - 完整功能
  NODE_TYPE=master

  # 数据库和Redis连接
  SQL_DSN=support:XIANjian4@tcp(43.154.19.173:13306)/new-api
  REDIS_CONN_STRING=redis://:XIANjian4SANyun@43.154.19.173:26739/1

  # 会话密钥（所有节点必须相同）
  SESSION_SECRET=XIANjian4

  # 其他配置
  PORT=3000
  TZ=Asia/Shanghai
  ERROR_LOG_ENABLED=true
  STREAMING_TIMEOUT=300
  MEMORY_CACHE_ENABLED=true
  SYNC_FREQUENCY=60
  CHANNEL_UPDATE_FREQUENCY=30
  BATCH_UPDATE_ENABLED=true
  UPDATE_TASK=true  # 主节点执行后台任务

  从服务器配置 (.env)

  # 从节点 - 专注API处理
  NODE_TYPE=slave

  # 使用相同的数据库和Redis（关键！）
  SQL_DSN=support:XIANjian4@tcp(43.154.19.173:13306)/new-api
  REDIS_CONN_STRING=redis://:XIANjian4SANyun@43.154.19.173:26739/1

  # 会话密钥（必须与主节点完全相同）
  SESSION_SECRET=XIANjian4

  # 从节点可以设置前端URL
  FRONTEND_BASE_URL=https://new-api.xhm.gd.cn

  # 其他配置
  PORT=3000
  TZ=Asia/Shanghai
  ERROR_LOG_ENABLED=true
  STREAMING_TIMEOUT=300
  MEMORY_CACHE_ENABLED=true
  SYNC_FREQUENCY=60
  CHANNEL_UPDATE_FREQUENCY=30
  BATCH_UPDATE_ENABLED=true
  # UPDATE_TASK=false  # 从节点不执行后台任务，减少资源消耗

  负载均衡配置示例

  在负载均衡器（如Nginx）中配置：

  upstream new_api_backend {
      server 主服务器IP:3000 weight=1;
      server 从服务器IP1:3000 weight=1;
      server 从服务器IP2:3000 weight=1;
  }

  server {
      listen 80;
      server_name new-api.xhm.gd.cn;

      location / {
          proxy_pass http://new_api_backend;
          proxy_set_header Host $host;
          proxy_set_header X-Real-IP $remote_addr;
      }
  }

  关键优势

  1. 高可用性：主服务器故障时，从服务器继续提供API服务
  2. 性能分工：主服务器处理重任务，从服务器专注API响应
  3. 水平扩展：可以添加更多从服务器分担负载
  4. 数据一致性：共享数据库和Redis确保状态同步



===============================================================


  因为买的供应商服务器2/4G内存都不能编译前端只能本地编译完成后丢上服务器部署, 修改了Dockerfile改成直接读取本地文件，需要在本地先编译执行下面的命令就好:
  构建前端：
  cd web
  bun install
  bun run build