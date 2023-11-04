//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bit-fever/core/auth/role"
	"github.com/bit-fever/core/req"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

//=============================================================================

type OidcController struct {
	authority string
	client    *http.Client
	context   *context.Context
	verifier  *oidc.IDTokenVerifier
}

//=============================================================================

type userToken struct {
	JTI      string `json:"jti,omitempty"`
	SID      string `json:"sid,omitempty"`
	Name     string `json:"given_name,omitempty"`
	Surname  string `json:"family_name,omitempty"`
	Username string `json:"preferred_username,omitempty"`
	Email    string `json:"email,omitempty"`

	ResourceAccess map[string]json.RawMessage `json:"resource_access,omitempty"`
}

//-----------------------------------------------------------------------------

type realmRoles struct {
	Roles []role.Role `json:"roles,omitempty"`
}

//=============================================================================

func NewOidcController(authority string, client *http.Client) (*OidcController, error) {
	ccontext      := oidc.ClientContext(context.Background(), client)
	provider, err := oidc.NewProvider(ccontext, authority)
	if err != nil {
		return nil, errors.New("Authorisation failed while getting the provider: "+ err.Error())
	}

	oidcConfig := &oidc.Config{
		SkipClientIDCheck: true,
	}
	verifier := provider.Verifier(oidcConfig)

	return &OidcController{
		authority: authority,
		client   : client,
		context  : &ccontext,
		verifier : verifier,
	}, nil
}

//=============================================================================

func (oc *OidcController) Secure(h func(c *gin.Context, us *UserSession), roles []role.Role) func(c *gin.Context) {
	return func(c *gin.Context) {
		rawAccessToken := c.Request.Header.Get("Authorization")
		tokens := strings.Split(rawAccessToken, " ")
		if len(tokens) != 2 {
			req.ReturnUnauthorizedError(c, "Authorisation failed due to a bad header")
			return
		}

		idToken, err := oc.verifier.Verify(*oc.context, tokens[1])
		if err != nil {
			req.ReturnUnauthorizedError(c, "Authorisation failed while verifying the token: "+ err.Error())
			return
		}

		var ut userToken
		if err := idToken.Claims(&ut); err != nil {
			req.ReturnUnauthorizedError(c, "Authorization failed while getting claims: "+ err.Error())
			return
		}

		us := buildUserSession(&ut, idToken)

		if ! us.IsUserInRole(roles) {
			req.ReturnForbiddenError(c, "User not allowed to access this API")
			return
		}

		h(c, us)
	}
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func buildUserSession(ut *userToken, it *oidc.IDToken) *UserSession {
	return &UserSession{
		SessionID: ut.SID,
		Username : ut.Username,
		Name     : ut.Name,
		Surname  : ut.Surname,
		Email    : ut.Email,
		IssuedAt : it.IssuedAt,
		Expiry   : it.Expiry,
		Roles    : buildRoleMap(ut),
	}
}

//=============================================================================

func buildRoleMap(ut *userToken) map[role.Role]any {
	userRoles := map[role.Role]any{}

	for k,v := range ut.ResourceAccess {
		if k != "account" {
			realmRoles := realmRoles{}
			err := json.Unmarshal(v, &realmRoles)

			if err == nil {
				for _, r := range realmRoles.Roles {
					userRoles[r] = nil
				}
			}
		}
	}

	return userRoles
}

//=============================================================================