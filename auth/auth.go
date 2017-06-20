package auth

import (
	"errors"
	"gost/auth/cookies"
	"gost/auth/identity"
	"gost/security"
	"gost/util/encodeutil"
	"gost/util/hashutil"
	"gost/util/jsonutil"
	"net/http"
	"strings"
)

// The keys that are used in the request header to authorize the user
const (
	AuthorizationHeader = "Authorization"
	AuthorizationScheme = "GOST-TOKEN"
)

// Errors generated by the auth package
var (
	ErrInvalidScheme           = errors.New("The used authorization scheme is invalid or not supported")
	ErrInvalidGhostToken       = errors.New("The given token is expired or invalid")
	ErrInvalidUser             = errors.New("There is no application user with the given ID")
	ErrSessionExpired          = errors.New("The session was invalidated or has expired")
	ErrDeactivatedUser         = errors.New("The current user account is deactivated or inexistent")
	ErrInexistentClientDetails = errors.New("Missing client details. Cannot create authorization for anonymous client")
	ErrPasswordMismatch        = errors.New("The entered password is incorrect")

	errAnonymousUser = errors.New("The user has no identity")
)

// GenerateUserAuth generates a new gost-token, saves it in the database and returns it to the client
func GenerateUserAuth(email string, password string, clientDetails *cookies.Client) (string, error) {
	var user *identity.ApplicationUser
	var isUserExistent bool

	if user, isUserExistent = identity.IsUserEmailExistent(email); !isUserExistent {
		return ErrInvalidUser.Error(), ErrInvalidUser
	}

	if !hashutil.MatchHashString(user.Password, password) {
		return ErrPasswordMismatch.Error(), ErrPasswordMismatch
	}

	session, err := cookies.NewSession(user.ID, clientDetails)
	if err != nil {
		return err.Error(), err
	}

	err = session.Save()
	if err != nil {
		return err.Error(), err
	}

	ghostToken, err := generateGostToken(session)

	return ghostToken, err
}

// Authorize tries to authorize an existing gostToken
func Authorize(httpHeader http.Header) (*identity.Identity, error) {
	ghostToken, err := extractGostToken(httpHeader)
	if err != nil {
		if err == errAnonymousUser {
			return identity.NewAnonymous(), nil
		}

		return nil, err
	}

	encryptedToken, err := encodeutil.Decode([]byte(ghostToken))
	if err != nil {
		return nil, err
	}

	jsonToken, err := security.Decrypt(encryptedToken)
	if err != nil {
		return nil, err
	}

	cookie := new(cookies.Session)
	err = jsonutil.DeserializeJSON(jsonToken, cookie)
	if err != nil {
		return nil, err
	}

	session, err := cookies.GetSession(cookie.Token)
	if err != nil || session == nil {
		return nil, ErrSessionExpired
	}

	user, isUserActivated := identity.IsUserActivated(session.UserID)
	if !isUserActivated {
		return nil, ErrDeactivatedUser
	}

	go session.ResetToken()

	return identity.New(session, user), nil
}

func generateGostToken(session *cookies.Session) (string, error) {
	jsonToken, err := jsonutil.SerializeJSON(session)
	if err != nil {
		return err.Error(), err
	}

	encryptedToken, err := security.Encrypt(jsonToken)
	if err != nil {
		return err.Error(), err
	}

	ghostToken := encodeutil.Encode(encryptedToken)

	return string(ghostToken), nil
}

func extractGostToken(httpHeader http.Header) (string, error) {
	var gostToken string

	if gostToken = httpHeader.Get(AuthorizationHeader); len(gostToken) == 0 {
		return errAnonymousUser.Error(), errAnonymousUser
	}

	if !strings.Contains(gostToken, AuthorizationScheme) {
		return ErrInvalidScheme.Error(), ErrInvalidScheme
	}

	gostTokenValue := strings.TrimPrefix(gostToken, AuthorizationScheme)
	gostTokenValue = strings.TrimSpace(gostTokenValue)

	if len(gostTokenValue) == 0 {
		return ErrInvalidGhostToken.Error(), ErrInvalidGhostToken
	}

	return gostTokenValue, nil
}
