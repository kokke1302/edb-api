package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Ollamaへプロンプトを送信し、LLMの生テキストレスポンスを取得する共通関数
func callOllama(prompt string) (string, error) {
	// LLM指定
	modelName := os.Getenv("OLLAMA_MODEL")
	if modelName == "" {
		modelName = "qwen2.5-coder:7b"
	}

	// リクエストの作成
	ollamaReq := OllamaRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: false,
	}

	// JSON化
	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return "", err
	}

	// Ollamaサーバーへリクエスト送信
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://ollama:11434"
	}

	// Ollamaサーバーからのレスポンス
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(ollamaHost+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("ollama returned bad status: %d", resp.StatusCode)
		return "", err
	}

	var ollamaResp OllamaResponse
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}

// マークダウンコードブロック (```json ... ```) が含まれている場合の抽出・クレンジング
func cleanJSONCodeBlock(rawText string) string {
	rawText = strings.TrimSpace(rawText)
	jsonCodeBlockRegex := regexp.MustCompile("(?s)```(?:json)?" + `\s*(.*?)\s*` + "```")

	matches := jsonCodeBlockRegex.FindStringSubmatch(rawText)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return rawText
}
