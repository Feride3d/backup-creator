package model

import "time"

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiryTime  time.Time
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiryTime)
}
