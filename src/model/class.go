package model

type Class struct {
	Code string   `bson:"code"`
	Type []string `bson:"type"`
}

func ValidateClass (code string) bool {
	return len(code) == 3 && code > "100" && code < "499"
}