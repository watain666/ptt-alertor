# Ptt-Alertor

<img align="right" src="https://raw.githubusercontent.com/watain666/ptt-alertor/master/logo.jpg">

[![Build Status](https://github.com/watain666/ptt-alertor/actions/workflows/main.yml/badge.svg)](https://github.com/watain666/ptt-alertor/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/watain666/ptt-alertor/branch/master/graph/badge.svg)](https://codecov.io/gh/watain666/ptt-alertor)
[![Go Report Card](https://goreportcard.com/badge/github.com/watain666/ptt-alertor)](https://goreportcard.com/report/github.com/watain666/ptt-alertor)
[![Code Climate](https://api.codeclimate.com/v1/badges/f7047295fce56a0465dc/maintainability)](https://codeclimate.com/github/watain666/ptt-alertor/maintainability)
[![StackShare](https://img.shields.io/badge/tech-stack-0690fa.svg?style=flat)](https://stackshare.io/watain666/ptt-alertor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Docker Compose 自架教學

完整圖文步驟請看 [Docker Compose 自架教學](https://ptt-alertor.tiaui.co/docs#self-hosting)。

### 1. 取得專案與設定檔

先安裝 Docker Engine、Compose v2 與 Git；不需要在主機安裝 Go。

```bash
git clone https://github.com/watain666/ptt-alertor.git
cd ptt-alertor
cp .env.example .env
```

已有 `.env` 時直接編輯，避免覆蓋。即使只做本機測試也需要這個檔案，
因為 `dynamodb-init` 會讀取它。

### 2. 建立 Telegram Bot 並設定 HTTPS

在官方 [@BotFather](https://t.me/BotFather) 傳送 `/newbot`，取得自己機器人的 Token。
編輯 `.env`，替換以下範例值：

```dotenv
APP_HOST=https://ptt.example.com
APP_WS_HOST=wss://ptt.example.com/ws
TELEGRAM_TOKEN=YOUR_BOT_TOKEN
AUTH_USER=admin
AUTH_PW=YOUR_LONG_RANDOM_PASSWORD
```

將 HTTPS 反向代理或 Cloudflare Tunnel 指向 app 的 `9090` 埠，保留
`/telegram/` 與 `/ws` 路徑並支援 WebSocket。只有 localhost 無法接收
Telegram webhook；僅試看頁面時，可保留範例中的 localhost 設定與空白 Token。
同一個 Bot 只能使用一個 webhook，請使用專屬 Bot 並妥善保管 Token。

資料庫使用 Compose 內的 Redis 與 DynamoDB Local，保留 `.env.example`
中的資料庫設定即可，不需要 AWS 雲端帳號。

### 3. 啟動與驗證

```bash
docker compose config --quiet
docker compose up --build -d
docker compose ps -a
docker compose logs --tail=80 app dynamodb-init
```

- 本機頁面：<http://localhost:9090>；文件：<http://localhost:9090/docs>。
- `dynamodb-init` 完成建表後顯示 `Exited (0)` 是正常的。
- 設定有效的 Token 與 HTTPS 網址後，app 紀錄應出現 `Telegram Bot Sets Webhook Success`。
- 在**自己的機器人**點選「開始」，傳送 `新增 Lifeismoney 咖啡`，再傳送 `清單` 確認。
- 網頁中的 Telegram 連結預設屬於本站 Bot；對外提供自架首頁時，請在
  `public/telegram.html` 與 `public/tpls/header.tpl` 替換為自己的機器人帳號，再建置 app。
- 對外僅需 HTTPS 入口。6060 是診斷埠、8000 是資料庫埠，請移除這兩個
  主機映射或限制為本機存取；Redis 維持 Docker 內部網路。

### 4. 更新、暫停與備份

先備份資料並保存自己的程式修改，再更新 app：

```bash
git pull --ff-only
docker compose build app
docker compose up -d --no-deps app
```

修改 `.env` 後也需要 `docker compose up -d --no-deps app`；`restart` 不會套用新環境變數。

暫停服務使用 `docker compose stop`；恢復使用 `docker compose start`。
**目前 DynamoDB Local 沒有永久 volume，資料仍在容器內。** 不要用
`docker compose down` 或重建資料庫來更新網頁，否則會失去看板與文章資料。
Redis 的訂閱資料位於 `redis-data` volume；`down -v` 也會將此 volume 刪除。
長期運作前，請另行規劃 DynamoDB 的持久化儲存。

一致性備份範例（會短暫暫停 app 與資料庫；請使用新的備份目錄）：

```bash
mkdir -m 700 backup
docker compose stop app dynamodb-local
docker compose exec -T redis redis-cli SAVE
docker compose cp redis:/data/dump.rdb backup/redis-dump.rdb
docker compose cp dynamodb-local:/home/dynamodblocal/shared-local-instance.db backup/dynamodb.db
docker compose start dynamodb-local
docker compose start app
```

備份也應安全保存 `.env`。備份檔可能包含訂閱資料，請勿上傳至 Git 或公開服務。

參考：[Docker Compose](https://docs.docker.com/compose/gettingstarted/)、
[Telegram Bot](https://core.telegram.org/bots/tutorial)、
[DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html)。

## Frontend preview

To preview the landing page, documentation, and rankings without starting the
bot, background jobs, Redis, or DynamoDB:

```bash
go run ./cmd/preview
```

Open http://127.0.0.1:9091. Templates are reloaded on each request; refresh after
editing. The preview intentionally has no live notification count or ranking data.
The production application continues to serve the same templates on port 9090.

```bash
go test ./cmd/preview
```

## API

### Board

- GET /boards

- GET /boards/[board name]/articles

- GET /boards/[board name]/articles/[article code]

### Keyword

- GET /keyword/boards

### Author

- GET /author/boards

### PushSum

- GET /pushsum/boards

### Articles

- GET /articles

### User (Auth)

- GET /users

- GET /users/[account]

- POST /users

```json
{
  "profile": {
    "account": "sample",
    "email": "sample@mail.com"
  },
  "subscribes": [
    {
      "board": "gossiping",
      "keywords": ["問卦", "爆卦", "公告"]
    },
    {
      "board": "lol",
      "keywords": ["閒聊"]
    }
  ]
}
```

- PUT /users/[account]

```json
{
  "profile": {
    "account": "sample",
    "email": "sample@mail.com"
  },
  "subscribes": []
}
```

## Credits

### Real Life

Rose Li, Aries Huang, Scott Kao, Amy Li

### Ptt

DMM, oas, bestpika, Zero0910, lucky0509, wbreeze, chang0206, lindo0130, hungys, gyman7788, tooilxui, myamyakoko, whkuo, papago89, timeline, Kamikiri

### Facebook

Mr.clu, Woqeker
