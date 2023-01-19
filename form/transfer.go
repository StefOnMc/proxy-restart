package form

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/mymaqc/proxy/user"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var (
	tokenSlice []string
	tokenMap   = map[string]*user.TokenSource{}
	/*secret     = "MYGIGADICKBRORESTARTILIVEYOURDICKBROWTF"
	tokenURL   = "http://45.158.77.93/api/accounts?auth=" + secret*/
)

/* func init() {
	var tokens map[string]map[string]interface{}
	resp, err := http.Get(tokenURL)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		fmt.Println(err)
	}
	for name, data := range tokens {
		var token *oauth2.Token
		d, _ := json.Marshal(data["token"])
		err = json.Unmarshal(d, &token)
		if err != nil {
			return
		}
		tokenMap[name] = user.NewTokenSource(token, name)
		tokenSlice = append(tokenSlice, name)
	}
} */

func init() {
	src := tokenSource()
	tokenSlice = append(tokenSlice, "toncompte")
	token, _ := src.Token()
	tokenMap["toncompte"] = user.NewTokenSource(token, "toncompte")
}

func tokenSource() oauth2.TokenSource {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	token := new(oauth2.Token)
	tokenData, err := ioutil.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token, err = auth.RequestLiveToken()
		check(err)
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		token, err = auth.RequestLiveToken()
		check(err)
		src = auth.RefreshTokenSource(token)
	}
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c

		tok, _ := src.Token()
		b, _ := json.Marshal(tok)
		_ = ioutil.WriteFile("token.tok", b, 0644)
		os.Exit(0)
	}()
	return src
}

type transfer struct {
	Token   form.Dropdown
	Address form.Input
	Port    form.Input
}

func NewTransfer() form.Form {
	f := form.New(transfer{
		Token:   form.NewDropdown("Account", tokenSlice, 0),
		Address: form.NewInput("Address", "", "address"),
		Port:    form.NewInput("Port", "19132", "port"),
	}, "transfer")
	return f
}

func (t transfer) Submit(submitter form.Submitter) {
	u := submitter.(*user.User)
	tk, ok := tokenMap[tokenSlice[t.Token.Value()]]
	if !ok {
		return
	}
	u.Transfer(fmt.Sprintf("%s:%s", t.Address.Value(), t.Port.Value()), tk)
}
