package model

type Judge struct {
	User     string `json:"user"`
	Problem  string `json:"problem"`
	Language string `json:"language"`
	Code     string `json:"code"`
}
