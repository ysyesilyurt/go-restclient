package restclient

import (
	"net/http"
)

type BasicAuthenticator struct {
	Username, Password string
}

func NewBasicAuthenticator(username, password string) Authenticator {
	return &BasicAuthenticator{
		Username: username,
		Password: password,
	}
}

func (ba BasicAuthenticator) Apply(request *http.Request) error {
	request.SetBasicAuth(ba.Username, ba.Password)
	return nil
}
