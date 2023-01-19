package user

import "golang.org/x/oauth2"

type TokenSource struct {
	token    *oauth2.Token
	username string
}

func NewTokenSource(token *oauth2.Token, username string) *TokenSource {
	return &TokenSource{
		token:    token,
		username: username,
	}
}

// Token ...
func (t *TokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}

// Username ...
func (t *TokenSource) Username() string {
	return t.username
}
