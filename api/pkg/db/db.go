package db

type Note struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	BuyTarget  string `json:"buyTarget"`
	SellTarget string `json:"sellTarget"`
	Notes      string `json:"notes"`
	Output     string `json:"output"`
}