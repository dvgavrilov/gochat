package messaging

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/go-chi/jwtauth"
)

const (
	// SecWebSocketProtocol const
	SecWebSocketProtocol = "Sec-WebSocket-Protocol"
)

var jwtSecret = []byte("!!SECRET!!")
var tokenStart = "access_token"

func tokenFromWsRequest(r *http.Request) string {
	// there is no way to send authorization header,
	// so, we decided to use the Sec-WebSocket-Protocol header
	values := strings.Split(r.Header.Get(SecWebSocketProtocol), ",")
	if len(values) != 2 {
		return ""
	}

	if values[0] != tokenStart {
		return ""
	}

	return strings.TrimSpace(values[1])
}

func getUpgraderWebSocketHeader() http.Header {
	h := make(http.Header)
	h.Add(SecWebSocketProtocol, tokenStart)
	return h
}

func verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(ja, tokenFromWsRequest)(next)
	}
}

func authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			logrus.Error(err)
			http.Error(w, http.StatusText(401), 401)
			return
		}

		if token == nil || !token.Valid {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		next.ServeHTTP(w, r)
	})
}
