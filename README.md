# go-chat-app
【外部】           【サーバー内部 (internal)】             【外部】
  ブラウザ  <------>  (1) API 層 (Handler)  <------>  (2) Service 層
    (UI)              (ポート 8080)              (ビジネスロジック)
                            |                         |
                            |                         |
                            v                         v
                   (4) WebSocket 層 (ws)     (3) Repo 層 (Repository)
                    (リアルタイム通知)           (SQL実行 / DB操作)
                            |                         |
                            |                         |
                            +-------------------------+
                                        |
                                        v
                               【データベース】
                                 PostgreSQL