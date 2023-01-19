package user

import "github.com/df-mc/dragonfly/server/player/form"

type Settings struct {
	SliderExploit bool
}

func DefaultSettings() Settings {
	return Settings{}
}

type settingsForm struct {
	SliderExploit form.Toggle
}

func NewSettings(u *User) form.Form {
	s := u.Settings()
	f := form.New(settingsForm{
		SliderExploit: form.NewToggle("Slider exploit", s.SliderExploit),
	}, "transfer")
	return f
}

func (s settingsForm) Submit(submitter form.Submitter) {
	u := submitter.(*User)
	st := u.Settings()
	st.SliderExploit = s.SliderExploit.Value()
	u.SetSettings(st)
}

func (u *User) SendSettingsForm() {
	u.SendForm(NewSettings(u))
}
