package model

type User struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

type Problem struct {
	Problem string `json:"problem"`
}

type Judge struct {
	Problem  string `json:"problem"`
	Language string `json:"language"`
	Code     string `json:"code"`
}
