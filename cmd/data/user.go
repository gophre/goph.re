package data

// User représente un utilisateur authentifié
type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	NickName  string `json:"nickname,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Provider  string `json:"provider"`
}
