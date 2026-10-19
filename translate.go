package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handleTranslate(w http.ResponseWriter, r *http.Request) {
	log.Println("Start function of Translate")

	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Flutterからのリクエストをデコード
	var clientReq TranslateRequest
	err := json.NewDecoder(r.Body).Decode(&clientReq)
	text := strings.TrimSpace(clientReq.Text)

	if err != nil || text == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// LLMへのプロンプト
	prompt := "次の英語のテキストを分析し、単語や句読点ごとに分割したJSON配列として出力してください。\n"
	prompt += "各要素は必ず以下のキーを持つオブジェクトにしてください。\n"
	prompt += `1. "showWord": 分割した元の文字列。\n`
	prompt += `2. "translation": その日本語訳（単語以外や句読点は空文字 ""）。\n`
	prompt += "コードブロックのバッククォート（```）や説明文は一切含めず、純粋なJSON配列のテキストのみを出力してください。\n"
	prompt += "入力例: I like apples.\n"
	prompt += `出力例: [{"showWord":"I","translation":"私は"},{"showWord":"like","translation":"好む"},{"showWord":"apples","translation":"りんご"},{"showWord":".","translation":""}]\n`
	prompt += fmt.Sprintf("入力英文: %s", text)

	// 共通関数でOllamaを呼び出し
	rawResponse, err := callOllama(prompt)
	if err != nil {
		log.Printf("Failed to process Ollama request: %v", err)
		http.Error(w, "Ollama execution failed", http.StatusInternalServerError)
		return
	}

	// 加工・二重チェック・フォールバック処理
	rawText := cleanJSONCodeBlock(rawResponse)

	// JSONパース確認
	var tokens []TokenResponse
	err = json.Unmarshal([]byte(rawText), &tokens)
	if err != nil || len(tokens) == 0 {
		log.Printf("[Warning] Failed to parse JSON from LLM. Error: %v. Raw text: %s", err, rawText)

		words := strings.Fields(clientReq.Text)
		tokens = make([]TokenResponse, len(words))
		for i, w := range words {
			tokens[i] = TokenResponse{
				ShowWord:    w,
				Translation: "",
			}
		}
	}

	// 最終レスポンスを作成
	validTokens := make([]TokenResponse, 0, len(tokens))
	for _, t := range tokens {
		if strings.TrimSpace(t.ShowWord) != "" {
			validTokens = append(validTokens, t)
		}
	}

	// Flutterへ結果を返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(validTokens)

	log.Println("Finish Translate")
}
