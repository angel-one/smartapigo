package smartapigo

import (
	"net/http"
)

// UserSession represents the response after a successful authentication.
type UserSession struct {
	UserProfile
	UserSessionTokens
}

// UserSessionTokens represents response after renew access token.
type UserSessionTokens struct {
	AccessToken  string `json:"jwtToken"`
	RefreshToken string `json:"refreshToken"`
	FeedToken    string `json:"feedToken"`
}

// UserProfile represents a user's personal and financial profile.
type UserProfile struct {
	ClientCode    string   `json:"clientcode"`
	UserName      string   `json:"name"`
	Email         string   `json:"email"`
	Phone         string   `json:"mobileno"`
	Broker        string   `json:"broker"`
	Products      []string `json:"products"`
	LastLoginTime string   `json:"lastlogintime"`
	Exchanges     []string `json:"exchanges"`
}

// GenerateSession gets a user session details in exchange of username and password.
// Access token is automatically set if the session is retrieved successfully.
// Do the token exchange with the `requestToken` obtained after the login flow,
// and retrieve the `accessToken` required for all subsequent requests. The
// response contains not just the `accessToken`, but metadata for the user who has authenticated.
//totp used is required for 2 factor authentication
func (c *Client) GenerateSession(totp string) (UserSession, error) {

	// construct url values
	params := make(map[string]interface{})
	params["clientcode"] = c.clientCode
	params["password"] = c.password
	params["totp"] = totp

	var session UserSession
	err := c.doEnvelope(http.MethodPost, URILogin, params, nil, &session)
	// Set accessToken on successful session retrieve
	if err == nil && session.AccessToken != "" {
		c.SetAccessToken(session.AccessToken)
	}
	return session, err
}

// RenewAccessToken renews expired access token using valid refresh token.
func (c *Client) RenewAccessToken(refreshToken string) (UserSessionTokens, error) {

	params := map[string]interface{}{}
	params["refreshToken"] = refreshToken

	var session UserSessionTokens
	err := c.doEnvelope(http.MethodPost, URIUserSessionRenew, params, nil, &session, true)

	// Set accessToken on successful session retrieve
	if err == nil && session.AccessToken != "" {
		c.SetAccessToken(session.AccessToken)
	}

	return session, err
}

// GetUserProfile gets user profile.
func (c *Client) GetUserProfile() (UserProfile, error) {
	var userProfile UserProfile
	err := c.doEnvelope(http.MethodGet, URIUserProfile, nil, nil, &userProfile, true)
	return userProfile, err
}

// Logout from User Session.
func (c *Client) Logout() (bool, error) {
	var status bool
	params := map[string]interface{}{}
	params["clientcode"] = c.clientCode
	err := c.doEnvelope(http.MethodPost, URILogout, params, nil, nil, true)
	if err == nil {
		status = true
	}
	return status, err
}
