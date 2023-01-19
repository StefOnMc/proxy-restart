package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/mymaqc/proxy/user"
)

type Settings struct {
	SliderExploit form.Toggle
}

func NewSettings(u *user.User) form.Form {
	s := u.Settings()
	f := form.New(Settings{
		SliderExploit: form.NewToggle("Slider exploit", s.SliderExploit),
	}, "transfer")
	return f
}
func (s Settings) Submit(submitter form.Submitter) {
	u := submitter.(*user.User)
	st := u.Settings()
	st.SliderExploit = s.SliderExploit.Value()
	u.SetSettings(st)
}
