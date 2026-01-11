# 変数定義（修正しやすくするため）
DOCKER_COMPOSE = docker compose
EXEC_SERVER = $(DOCKER_COMPOSE) exec server
EXEC_DB = $(DOCKER_COMPOSE) exec db

# フォニーターゲット：コマンドとして実行してくれるおまじない
.PHONY: up down restart logs ps tidy build test db-shell help

# デフォルトのコマンド（make とだけ打った時に実行される）
help:
	@echo "利用可能なコマンド:"
	@echo "  make up      - コンテナをバックグラウンドで起動"
	@echo "  make down    - コンテナを停止・削除"
	@echo "  make restart - コンテナを再起動"
	@echo "  make logs    - サーバーのログをリアルタイム表示"
	@echo "  make tidy    - Goの依存関係を整理 (container内)"
	@echo "  make db      - PostgreSQLのシェルに入る"

# --- コンテナ操作 ---
up:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

restart:
	$(DOCKER_COMPOSE) restart

logs:
	$(DOCKER_COMPOSE) logs -f server

ps:
	$(DOCKER_COMPOSE) ps

# --- Go 開発 ---
tidy:
	$(EXEC_SERVER) go mod tidy

test:
	$(EXEC_SERVER) go test ./...

build:
	$(EXEC_SERVER) go build -v ./...

# --- データベース ---
db:
	$(EXEC_DB) psql -U postgres -d chat_db