package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(
					w, r,
					fmt.Errorf("authorization header is missing"))
				return
			}
			app.logger.Infoln("Header:", authHeader)
			// parse it -> get base64
			// header value will be something like "Basic YWRtaW46YWRtaW4=". Symbols it is Base64 encoding.
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(
					w, r,
					fmt.Errorf("authorization header is malformed"))
				return
			}

			// decode it
			//YWRtaW46YWRtaW4= becomes admin:admin
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(
					w, r,
					err)
				return
			}

			// check credentials
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass
			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicErrorResponse(
					w, r,
					fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
