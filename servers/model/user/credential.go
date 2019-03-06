package model

type Credential struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CredentialUpdate struct {
	Email           string `json:"email"`
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	NewPasswordConf string `json:"new_password_conf"`
}

