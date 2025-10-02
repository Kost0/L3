package repository

type Notify struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Text   string `json:"text"`
	SendAt string `json:"send_at"`
	Email  string `json:"email"`
	TGUser string `json:"tg_user"`
}
