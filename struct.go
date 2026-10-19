package main

// 文章解析 API (/api/translate) 用 -------------------------------------

// Flutterからのリクエスト
type TranslateRequest struct {
	Text string `json:"text"`
}

// Flutterへ返すトークン構造体
type TokenResponse struct {
	ShowWord    string `json:"showWord"`
	Translation string `json:"translation"`
}

// 単語辞書 API (/api/dictionary) 用 -------------------------------------

// Flutterからのリクエスト
type DictionaryRequest struct {
	Word string `json:"word"`
}

// Flutterへ返す辞書レスポンス構造体
type DictionaryResponse struct {
	Translation string `json:"translation"`
	Memo        string `json:"memo"`
}

// Ollama 共通構造体 --------------------------------------------------

// Ollamaへのリクエスト
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	// Format string `json:"format,omitempty"`
}

// Ollamaからのレスポンス
type OllamaResponse struct {
	Response string `json:"response"`
}
