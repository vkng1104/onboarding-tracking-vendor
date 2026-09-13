package auth

type Coordinator struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type coordinatorWithPassword struct {
	Coordinator
	PasswordHash string
}
