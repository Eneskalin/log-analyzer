package models

type Rules struct {
	Id       string `json:"id"`
	LogType  string `json:"log_type"`
	Match    string `json:"match"`
	Severity string `json:"severity"`
}
