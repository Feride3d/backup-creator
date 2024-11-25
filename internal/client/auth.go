package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Feride3d/backup-creator/internal/model"
)

type AuthClient struct {
	authURL      string
	clientID     string
	clientSecret string
}

func NewAuthClient(authURL, clientID, clientSecret string) *AuthClient {
	return &AuthClient{authURL, clientID, clientSecret}
}

func (a *AuthClient) GetAccessToken() (model.Token, error) {
	if _, err := url.ParseRequestURI(a.authURL); err != nil {
		return model.Token{}, fmt.Errorf("invalid auth URL: %v", err)
	}
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     a.clientID,
		"client_secret": a.clientSecret,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to marshal payload: %v", err)
	}
	resp, err := http.Post(a.authURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to send token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiError struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&apiError)
		return model.Token{}, fmt.Errorf("failed to get access token: %s - %s", apiError.Error, apiError.ErrorDescription)
	}

	var accessToken model.Token
	if err := json.NewDecoder(resp.Body).Decode(&accessToken); err != nil {
		return model.Token{}, fmt.Errorf("failed to decode token response: %v", err)
	}

	accessToken.ExpiryTime = time.Now().Add(time.Duration(accessToken.ExpiresIn) * time.Second)

	return accessToken, nil
}
