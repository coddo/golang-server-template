package apifilter

import (
	"gost/models"
)

func CheckUserSessionIntegrity(userSession *models.UserSession) bool {
	switch {
	case len(userSession.ApplicationUser.Id) == 0:
		return false
	case len(userSession.Token) == 0:
		return false
	}

	return true
}
