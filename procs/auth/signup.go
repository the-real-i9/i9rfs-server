package authprocs

type AuthSignup struct{}

func (aus AuthSignup) RequestNewAccount(email string, reply *string) error {

	return nil
}

func (aus AuthSignup) VerifyEmail(args struct {
	Token string
	Code  int
}, reply *string) error {

	return nil
}

func (aus AuthSignup) RegisterUser(args struct {
	Token    string
	UserInfo map[string]any
}, reply *string) error {

	return nil
}
