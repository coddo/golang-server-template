package authapi

import (
	"errors"
	"gost/api"
	"gost/auth"
	"gost/auth/cookies"
	"gost/filter"
	"gost/util"
	"log"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

// Errors generated by then Auth endpoint
var (
	ErrPasswordMatch     = errors.New("The passwords do not match")
	ErrTokenNotSpecified = errors.New("The session token was not specified")
)

// AuthAPI defines the API endpoint for user authorization
type AuthAPI int

// AuthModel is a binding model used for receiving the authentication data
type AuthModel struct {
	AppUserID            string          `json:"appUserID"`
	Password             string          `json:"password"`
	PasswordConfirmation string          `json:"passwordConfirmation"`
	ClientDetails        *cookies.Client `json:"clientDetails"`
}

// GetAllSessions retrieves all the sessions for a certain user account
func (a *AuthAPI) GetAllSessions(params *api.Request) api.Response {
	if !params.Identity.IsAuthorized() {
		return api.Unauthorized()
	}

	userID, found, err := filter.GetIDFromParams("token", params.Form)
	if !found {
		return api.BadRequest(api.ErrIDParamNotSpecified)
	}
	if err != nil {
		return api.InternalServerError(err)
	}

	userSessions, err := cookies.GetUserSessions(userID)
	if err != nil {
		return api.InternalServerError(err)
	}

	return api.JSONResponse(http.StatusOK, userSessions)
}

// CreateSession creates a new session for an existing user account
func (a *AuthAPI) CreateSession(params *api.Request) api.Response {
	model := &AuthModel{}

	err := util.DeserializeJSON(params.Body, model)
	if err != nil {
		return api.BadRequest(err)
	}

	if model.Password != model.PasswordConfirmation {
		return api.BadRequest(ErrPasswordMatch)
	}

	if !bson.IsObjectIdHex(model.AppUserID) {
		return api.BadRequest(api.ErrInvalidIDParam)
	}

	token, err := auth.GenerateUserAuth(bson.ObjectIdHex(model.AppUserID), model.ClientDetails)
	if err != nil {
		return api.BadRequest(err)
	}

	log.Println("TOKEN:", token)
	return api.PlainTextResponse(http.StatusOK, token)
}

// KillSession deletes a session for an existing user account based on
// the session token
func (a *AuthAPI) KillSession(params *api.Request) api.Response {
	if !params.Identity.IsAuthorized() {
		return api.Unauthorized()
	}

	sessionToken, found := filter.GetStringValueFromParams("token", params.Form)
	if !found || len(sessionToken) == 0 {
		return api.BadRequest(ErrTokenNotSpecified)
	}

	session, err := cookies.GetSession(sessionToken)
	if err != nil {
		return api.InternalServerError(err)
	}

	err = session.Delete()
	if err != nil {
		return api.InternalServerError(err)
	}

	return api.StatusResponse(http.StatusOK)
}
