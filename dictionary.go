package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handleDictionary(w http.ResponseWriter, r *http.Request) {
	log.Println("Start function of Dictionary search")

	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Flutterからのリクエストをデコード
	var clientReq DictionaryRequest
	err := json.NewDecoder(r.Body).Decode(&clientReq)
	word := strings.TrimSpace(clientReq.Word)

	if err != nil || strings.TrimSpace(clientReq.Word) == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// LLMへのプロンプト設定
	prompt := "指定された英単語の日本語訳と、補足メモ（意味やカテゴリなど）をいくつか提案し、JSON配列として出力してください。\n"
	prompt += "各要素は必ず以下のキーを持つオブジェクトにしてください。\n"
	prompt += `1. "translation": 日本語での訳語。\n`
	prompt += `2. "memo": その訳語が使われる主な文脈、ニュアンス、または補足説明（品詞やカテゴリなど）。特になければ空文字 "" にしてください。\n`
	prompt += "コードブロックのバッククォート（```）や説明文は一切含めず、純粋なJSON配列のテキストのみを出力してください。\n"
	prompt += "入力例: apple\n"
	prompt += `出力例: [{"translation":"りんご","memo":"果物"},{"translation":"アップル","memo":"IT企業名"}]\n`
	prompt += fmt.Sprintf("入力単語: %s", word)

	// 共通関数でOllamaを呼び出し
	rawResponse, err := callOllama(prompt)
	if err != nil {
		log.Printf("Failed to process Ollama request: %v", err)
		http.Error(w, "Ollama execution failed", http.StatusInternalServerError)
		return
	}

	// 加工・クレンジング・フォールバック処理
	rawText := cleanJSONCodeBlock(rawResponse)

	// JSONパース確認
	var dicts []DictionaryResponse
	err = json.Unmarshal([]byte(rawText), &dicts)
	if err != nil || len(dicts) == 0 {
		log.Printf("[Warning] Failed to parse JSON from LLM for word '%s'. Error: %v. Raw text: %s", word, err, rawText)

		// パース失敗時のフォールバック
		dicts = []DictionaryResponse{}
	}

	// 最終レスポンスを作成
	validDicts := make([]DictionaryResponse, 0, len(dicts))
	for _, d := range dicts {
		if strings.TrimSpace(d.Translation) != "" {
			validDicts = append(validDicts, DictionaryResponse{
				Translation: strings.TrimSpace(d.Translation),
				Memo:        strings.TrimSpace(d.Memo),
			})
		}
	}

	// Flutterへ結果を返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(validDicts)

	log.Println("Finish Dictionary search")
}
