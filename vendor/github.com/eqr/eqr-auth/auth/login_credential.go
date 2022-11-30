package auth

type LoginCredential struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}
