# 翻訳できる単語帳 バックエンド (edb-api)

[翻訳できる単語帳](https://github.com/kokke1302/edb)（Flutterアプリ）から呼び出される、翻訳・辞書用のAPIサーバーです。
HTTPリクエストをローカルLLM（Ollama）に問い合わせ、結果をアプリが扱いやすいJSONに整形して返します。

```
Flutterアプリ → Go APIサーバー → Ollama（ローカルLLM）
```

## 技術スタック

- 言語: Go（標準ライブラリ: fmt, log, strings, bytes, os, time, regexp, encoding/json, net/http）
- LLM実行環境: Ollama（qwen2.5-coder:7b）
- コンテナ: Docker, Docker Compose

## 機能

- `POST /api/translate`: 英文を単語・句読点ごとに分割し、それぞれの日本語訳を返す
- `POST /api/dictionary`: 英単語の日本語訳と補足メモを複数候補で返す
- LLMの出力が不安定でも動き続ける設計
  - コードブロック（```` ```json ````）が付いていても除去してパース
  - JSONとして解釈できない場合は、`/api/translate` は空白区切りの単語を訳なしで返却、`/api/dictionary` は空配列を返却
  - 空文字の要素を除外してから返却

## 使い方

本アプリのAPIは、各自のPCでセルフホストすることを想定しています。

【起動】

```bash
git clone https://github.com/kokke1302/edb-api.git # ローカルへクローン
cd edb-api
docker compose build --no-cache # コンテナのビルド
docker compose up -d # コンテナの起動
```

【モデルのダウンロード（初回のみ）】

```bash
docker exec -it ollama ollama pull \<モデル\> # 開発中は qwen2.5-coder:7b を使用
```

【動作検証】

```bash
# ollamaの応答確認
curl http://localhost:11434/api/generate -d '{ 
  "model": "qwen2.5-coder:7b",
  "prompt": "「I watched a bird」を辞書で引けるように、`/`で分割して。思考（thinking）は不要。",
  "stream": false
}'

# 文章の翻訳
curl -X POST http://localhost:8080/api/translate \
  -H "Content-Type: application/json" \
  -d '{"text": "I watched a crow."}'

# 単語の辞書検索
curl -X POST http://localhost:8080/api/dictionary \
  -H "Content-Type: application/json" \
  -d '{"word": "crow"}'
```

【アプリから使う】

ネイティブ版の場合、同じネットワーク内であれば、drawer→設定→API URL にて `http://<PCのIPアドレス>:8080` を入力し保存すれば接続できます。

Web版（GitHub Pages）はHTTPSのため、`http://` のAPIへは接続できません。Cloudflare Tunnelなどのトンネルサービスで、ローカルのAPIをHTTPS化してください。（本APIには認証・レート制限・CORS設定がありません。Web版から呼び出す場合は、Go側にCORSヘッダーの追加が必要です。）

## API仕様

### POST /api/translate

リクエスト

```json
{ "text": "I like apples." }
```

レスポンス

```json
[
  { "showWord": "I", "translation": "私は" },
  { "showWord": "like", "translation": "好む" },
  { "showWord": "apples", "translation": "りんご" },
  { "showWord": ".", "translation": "" }
]
```

### POST /api/dictionary

リクエスト

```json
{ "word": "apple" }
```

レスポンス

```json
[
  { "translation": "りんご", "memo": "果物" },
  { "translation": "アップル", "memo": "IT企業名" }
]
```

## 設定

環境変数は `docker-compose.yml` で設定します。

| 変数 | 既定値 | 内容 |
| --- | --- | --- |
| `OLLAMA_HOST` | `http://ollama:11434` | OllamaのURL |
| `OLLAMA_MODEL` | `qwen2.5-coder:7b` | 使用するモデル名 |
| `PORT` | `8080` | APIサーバーのポート |

GPUを使う場合は、`docker-compose.yml` のコメントアウトされた `deploy` ブロックを有効にしてください（NVIDIA Container Toolkitが必要です）。

## ファイル構成

```
.
├── main.go          # エントリーポイント・ルーティング
├── translate.go     # /api/translate のハンドラ
├── dictionary.go    # /api/dictionary のハンドラ
├── ollama.go        # Ollama呼び出し・出力のクレンジング
├── struct.go        # リクエスト/レスポンスの構造体
├── Dockerfile
└── docker-compose.yml
```

## 備考

### GitHub Actions

7Bクラスのモデルは、ダウンロードで4~5GB程度あり、GitHub Actionsの無料ランナー（GPUなし）では実行が重く、現実的ではありません。
そのため、翻訳・辞書の動作確認は、ローカルのDocker環境で行う方針です。
Dockerイメージの公開もしていません。Ollamaとモデルが別々に必要で、イメージ単体では動かないためです。

### 適切な言語モデルの選び方

qwen3系統も試しましたが、qwen2.5-coderの方がリクエスト成功率が高かったです。
JSONを指定している関係上、Chat出力より穴埋め的思考の方が重要だと考えられます。
Coder系のモデルをおすすめします。

なお、プロンプトにてJSON出力を強制すると出力が安定しません。
出力を諦めてしまったり、Goが処理の終了と判定してプログラムを進めてしまう現象が見られます。

## 関連リポジトリ

- [edb（フロントエンド）](https://github.com/kokke1302/edb): Flutter製のアプリ本体

## ライセンス

MIT [LICENSE](./LICENSE)
