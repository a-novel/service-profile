package handlers

import (
	"net/http"

	"github.com/a-novel-kit/golib/httpf"
)

// ownerErrors maps a request that acts for no user to 403: one whose claims name none, such as an
// anonymous session's (serviceauthentication.ErrAnonymousCaller), or one that carries no claims.
var ownerErrors = httpf.ErrMap{nil: http.StatusForbidden}
