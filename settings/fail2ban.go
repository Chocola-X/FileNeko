package settings

import "fmt"

// Fail2Ban limits failed password checks from one connection source, shared
// between JSON login and WebDAV. Durations are expressed in whole minutes.
type Fail2Ban struct {
	Enabled       bool `json:"enabled"`
	MaxAttempts   uint `json:"maxAttempts"`
	WindowMinutes uint `json:"windowMinutes"`
	BanMinutes    uint `json:"banMinutes"`
}

func DefaultFail2Ban() Fail2Ban {
	return Fail2Ban{Enabled: true, MaxAttempts: 5, WindowMinutes: 5, BanMinutes: 5}
}

func (s *Settings) PasswordBanPolicy() Fail2Ban {
	if s.Fail2Ban == nil {
		return DefaultFail2Ban()
	}
	return *s.Fail2Ban
}

func (p Fail2Ban) Validate() error {
	if p.MaxAttempts < 1 || p.MaxAttempts > 100 ||
		p.WindowMinutes < 1 || p.WindowMinutes > 1440 ||
		p.BanMinutes < 1 || p.BanMinutes > 10080 {
		return fmt.Errorf("fail2ban: maxAttempts must be 1–100, windowMinutes 1–1440, banMinutes 1–10080")
	}
	return nil
}
