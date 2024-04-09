package authprocs

type Auth struct{}

func (au Auth) Login(args struct {
	EmailOrUsername string
	Password        string
}, reply *string) error {

	return nil
}

func (au Auth) GetUser(token string, reply *string) error {

	return nil
}
