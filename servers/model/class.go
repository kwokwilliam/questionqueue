package model

type Class struct {
	Code string   `json:"code"`
	Type []string `json:"type"`
}

func ValidateClass (code string) bool {
	return len(code) == 3 && code > "100" && code < "499"
}