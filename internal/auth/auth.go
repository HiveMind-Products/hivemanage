package auth

import (
	"errors"
	"os"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// TODO: Move this to the authservice package
// we're not currently using this btw, soooooo whatever
func NewGithubConfig() *oauth2.Config {
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	config := &oauth2.Config{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  github.Endpoint.AuthURL,
			TokenURL: github.Endpoint.TokenURL,
		},
	}

	return config
}

// NewDiscordConfig builds the Discord OAuth2 config from environment variables.
// DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET and DISCORD_REDIRECT_URI must be set.
func NewDiscordConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("DISCORD_REDIRECT_URI"),
		Scopes:       []string{"identify", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}

func CurrentOrgId(c echo.Context) (string, error) {
	org_id, ok := c.Get(OrgIDContextKey).(string)
	if !ok {
		return "", errors.New("org not found in context")
	}

	return org_id, nil
}
