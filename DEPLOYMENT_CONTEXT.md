# New-API 部署上下文与操作手册

> 以后凡是涉及“合并 master、检查配置、部署服务器、升级版本、排查部署问题”的操作，先读取本文件，再继续执行。

## 1. 这份文件是干什么的

这份文件是当前项目的部署入口文档，目标是：

- 以后每次你只需要把这份文件发给 AI，就能快速恢复部署上下文
- 不需要再去 `docs/` 里翻多个文件
- 把“合并 master 后要检查什么、配置文件怎么核对、服务器怎么部署、部署后怎么验证”收敛到一个地方

这份文件优先解决的是“日常升级部署”和“合并上游后重新部署”，不是一次性迁移过程记录。

## 2. 当前生产架构

当前生产架构约定如下：

- 应用 `new-api` 独立部署
- PostgreSQL 独立部署
- Redis 独立部署
- 三者不放在同一个 Compose 里
- 应用基于当前仓库代码构建
- 后续预留应用节点扩容能力

推荐拓扑：

```text
reverse proxy / domain
        |
     new-api
      /    \
PostgreSQL  Redis
```

## 3. 当前服务器约定

当前新服务器：

- IP：`43.173.102.239`
- 用户：`root`
- SSH 端口：`22`

服务器目录约定：

```text
/srv/new-api/
  app/
  postgres/
  redis/
  backups/
```

目录职责：

- `/srv/new-api/app`：仓库代码
- `/srv/new-api/postgres`：PostgreSQL 独立 Compose
- `/srv/new-api/redis`：Redis 独立 Compose
- `/srv/new-api/backups`：数据库备份

Docker 网络约定：

- 外部网络名：`new-api-infra`

## 4. 仓库里哪些文件是部署核心

以后每次部署，重点看这几处：

- `DEPLOYMENT_CONTEXT.md`
- `deploy/production/app/docker-compose.yml`
- `deploy/production/app/.env.example`
- `deploy/production/postgres/docker-compose.yml`
- `deploy/production/postgres/.env.example`
- `deploy/production/redis/docker-compose.yml`
- `deploy/production/redis/.env.example`
- `docs/installation/production-deployment-baseline.md`

其中：

- 根目录这份 `DEPLOYMENT_CONTEXT.md` 是以后最先读的入口
- `docs/installation/production-deployment-baseline.md` 是更完整的长期基线
- `deploy/production/` 是生产模板

## 5. 合并 master 后必须先做什么

以后你修改代码，并且把 `master` 分支的新代码合并过来之后，不要直接部署，先做下面这些检查。

### 5.1 先看哪些文件变了

在本地仓库执行：

```bash
git diff <上次部署提交>..HEAD -- \
  .env.example \
  docker-compose.yml \
  Dockerfile \
  main.go \
  common/init.go \
  common/redis.go \
  model/main.go \
  router/main.go \
  README.zh_CN.md \
  deploy/production
```

重点看：

- 根目录 `.env.example` 有没有新增变量
- 根目录 `docker-compose.yml` 有没有新增服务、环境变量、挂载、健康检查
- `common/init.go` 有没有新增环境变量默认值
- `model/main.go` 有没有新增数据库迁移或数据库行为变化
- `router/main.go` 有没有影响 `FRONTEND_BASE_URL`、前端入口、主从行为的变化
- `deploy/production/` 模板是否需要同步调整

### 5.2 再搜代码里有没有新增环境变量读取

执行：

```bash
rg -n "os.Getenv\\(|os.LookupEnv\\(|GetEnvOrDefault\\(|GetEnvOrDefaultBool\\(|GetEnvOrDefaultString\\(" \
  main.go common model router controller service setting
```

看到新增项后，要判断：

1. 是不是数据库相关
2. 是不是 Redis 相关
3. 是不是会话/加密相关
4. 是不是节点角色/多节点相关
5. 是不是任务、缓存、超时、流式处理相关

如果是以上这些，默认按“可能影响生产部署”处理。

## 6. 哪些配置是必须稳定的

应用 `.env` 中，这些值必须稳定：

- `SQL_DSN`
- `REDIS_CONN_STRING`
- `SESSION_SECRET`
- `CRYPTO_SECRET`
- `NODE_TYPE`
- `NODE_NAME`
- `TZ`

建议显式保留的当前生产行为配置：

- `ERROR_LOG_ENABLED=true`
- `MEMORY_CACHE_ENABLED=true`
- `BATCH_UPDATE_ENABLED=true`
- `CHANNEL_UPDATE_FREQUENCY=30`
- `STREAMING_TIMEOUT=300`
- `RELAY_TIMEOUT=0`

## 7. 合并 master 后，什么时候要改服务器 `.env`

不是每次合并都要改服务器 `.env`，但碰到下面情况就要检查并可能修改：

### 必须检查的情况

- 新增数据库相关环境变量
- 新增 Redis 相关环境变量
- 新增 `SESSION_SECRET` / `CRYPTO_SECRET` 相关逻辑
- 新增 `NODE_TYPE` / `NODE_NAME` 相关逻辑
- 新增定时任务开关
- 新增缓存、超时、请求大小、流式处理相关参数

### 通常不急着改的情况

- 纯调试参数
- 纯监控参数
- 实验性开关
- 文档说明变化但默认值不影响当前生产行为

## 8. 当前生产模板怎么用

### 8.1 PostgreSQL

服务器目录：

```bash
/srv/new-api/postgres
```

启动：

```bash
cd /srv/new-api/postgres
docker compose --env-file .env up -d
```

### 8.2 Redis

服务器目录：

```bash
/srv/new-api/redis
```

启动：

```bash
cd /srv/new-api/redis
docker compose --env-file .env up -d
```

### 8.3 应用

服务器目录：

```bash
/srv/new-api/app
```

启动或升级：

```bash
cd /srv/new-api/app
docker compose \
  --env-file deploy/production/app/.env \
  -f deploy/production/app/docker-compose.yml \
  up -d --build
```

## 9. 当前生产端口约定

应用模板中这两个值控制对外监听：

- `APP_BIND_HOST`
- `APP_HOST_PORT`

当前逻辑：

- `APP_BIND_HOST=127.0.0.1`：只有服务器本机能访问，适合走 Nginx 反代
- `APP_BIND_HOST=0.0.0.0`：公网 IP 可直接访问

如果你发现“服务器本机能 curl，公网 IP 访问不了”，第一时间检查：

```bash
grep -n "APP_BIND_HOST\\|APP_HOST_PORT" /srv/new-api/app/deploy/production/app/.env
ss -lntp | grep ':3000'
```

## 10. 标准升级部署流程

以后每次标准部署按这个顺序做：

1. 读取 `DEPLOYMENT_CONTEXT.md`
2. 拉取或合并最新代码
3. 做“合并 master 后检查”
4. 判断是否需要更新服务器 `.env`
5. 判断是否需要同步 `deploy/production/postgres` / `redis` 模板
6. 在服务器更新 `/srv/new-api/app` 代码
7. 执行应用重建部署
8. 做部署后验证

## 11. 服务器部署后必须验证什么

### 11.1 Compose 状态

```bash
cd /srv/new-api/app
docker compose \
  --env-file deploy/production/app/.env \
  -f deploy/production/app/docker-compose.yml \
  ps
```

### 11.2 应用健康检查

如果是仅本机监听：

```bash
curl -fsS http://127.0.0.1:3000/api/status
```

如果是公网监听：

```bash
curl -fsS http://43.173.102.239:3000/api/status
```

### 11.3 应用日志

```bash
cd /srv/new-api/app
docker compose \
  --env-file deploy/production/app/.env \
  -f deploy/production/app/docker-compose.yml \
  logs --tail=200
```

### 11.4 数据库和 Redis

确认：

- 应用日志没有 PostgreSQL 连接或迁移失败
- 应用日志没有 Redis ping 失败
- 后台登录正常
- 用户、渠道、日志、令牌都能正常显示

## 12. 回滚前必须记得备份什么

每次升级前至少保留：

- 升级前 Git 提交号
- 升级前 PostgreSQL 备份
- 升级前应用 `.env`
- 升级前 Compose 文件副本

原因：

- 应用代码回滚容易
- 数据库结构一旦被新版本迁移，未必能无损直接退回旧版本

## 13. 后续问 AI 时建议怎么说

以后你想继续操作时，建议直接这样开头：

```text
先读取项目根目录的 DEPLOYMENT_CONTEXT.md，再根据里面的约定继续操作。
```

如果是合并 master 后部署，可以直接说：

```text
先读取 DEPLOYMENT_CONTEXT.md，然后帮我检查这次合并 master 后需要同步哪些配置，再部署到服务器。
```

如果是排查部署问题，可以直接说：

```text
先读取 DEPLOYMENT_CONTEXT.md，然后按里面的部署约定排查为什么这次服务器访问异常。
```

## 14. 这份文件和其他文档的关系

这份文件是入口，不是替代所有文档。

关系如下：

- `DEPLOYMENT_CONTEXT.md`：以后每次先读的根目录入口
- `docs/installation/production-deployment-baseline.md`：更完整的长期生产基线
- `deploy/production/`：实际生产模板

以后如果部署方式有重大变化，应该优先更新：

1. `DEPLOYMENT_CONTEXT.md`
2. `deploy/production/`
3. `docs/installation/production-deployment-baseline.md`
