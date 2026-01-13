package token

import (
	"crypto/subtle"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TokenAuthProvide – provides constant-time Bearer token authentication
// using an internal pool of preconfigured tokens.
type TokenAuthProvide struct {
	forbidIfDisabled bool
	enable           bool
	token            []byte
	bearerRe         *regexp.Regexp
}

/*
NewTokenAuthProvide – creates a new TokenAuthProvide instance.

	otherwise an error is returned.
	If no tokens are provided, the function returns (nil, nil) meaning
	authentication is disabled.
*/
func NewTokenAuthProvide(forbidIfDisabled bool, envTokenKey string) (*TokenAuthProvide, error) {
	token := strings.TrimSpace(os.Getenv(envTokenKey))

	if token == "" {
		return &TokenAuthProvide{
			forbidIfDisabled: forbidIfDisabled,
		}, nil
	}

	if len(token)%16 != 0 {
		return nil, fmt.Errorf(
			"invalid token length: must be a multiple of 16",
		)
	}

	return &TokenAuthProvide{
		forbidIfDisabled: forbidIfDisabled,
		enable:           true,
		token:            []byte(token),
		bearerRe:         regexp.MustCompile(`^Bearer\s+(.+)$`),
	}, nil
}

/*
TestBearer – validates a Bearer token extracted from the Authorization header.

	Comparison is performed in constant time using subtle.ConstantTimeCompare to prevent timing attacks.
	If the provider is nil, authentication is considered disabled and always returns true.
*/
func (tap *TokenAuthProvide) TestBearer(bearer string) bool {
	if !tap.enable {
		return !tap.forbidIfDisabled
	}

	mch := tap.bearerRe.FindSubmatch([]byte(bearer))
	if len(mch) != 2 {
		return false
	}

	token := mch[1]

	if len(token) == len(tap.token) {
		return subtle.ConstantTimeCompare(token, tap.token) == 1
	}

	return false
}

func (tap *TokenAuthProvide) Enabled() bool {
	return tap.enable
}
