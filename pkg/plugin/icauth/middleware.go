package icauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hellofresh/janus/pkg/errors"
	"github.com/hellofresh/janus/pkg/proxy"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrAuthorizationFieldNotFound is used when the http Authorization header is missing from the request
	ErrAuthorizationFieldNotFound = errors.New(http.StatusBadRequest, "Authorization Header inválido.")
	// ErrBearerMalformed is used when the Bearer string in the Authorization header is not found or is malformed
	ErrBearerMalformed = errors.New(http.StatusBadRequest, "Bearer token mal formado.")
	// ErrAccessTokenNotAuthorized is used when the access token is not found on the storage
	ErrAccessTokenNotAuthorized = errors.New(http.StatusUnauthorized, "Token não autorizado.")
	// ErrVerifyToken is used when a http get request to the identity service could not be made
	ErrVerifyToken = errors.New(http.StatusInternalServerError, "Falha ao comunicar com serviço.")
	// ErrParseServiceResponse is used when http get request response from identity service could not be parsed
	ErrParseServiceResponse = errors.New(http.StatusInternalServerError, "Falha ao interpretar resposta do serviço.")
)

// Token JWT returned from identity service
type Token struct {
	Token string `json:"token"`
}

// Midleware will hit iclinic auth service
func Midleware(url string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeaderValue := r.Header.Get("Authorization")

			parts := strings.Split(authHeaderValue, " ")
			if len(parts) < 2 {
				errors.Handler(w, ErrAuthorizationFieldNotFound)
				return
			}

			if strings.ToLower(parts[0]) != "bearer" {
				errors.Handler(w, ErrBearerMalformed)
				return
			}

			jwtToken := parts[1]

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Warn("Could not make a request to our services...", jwtToken, url)
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

			client := &http.Client{Timeout: time.Minute * 3}
			resp, err := client.Do(req)
			if err != nil {
				log.Error("Error when verifying your token ", err, " at ", url)
				errors.Handler(w, ErrVerifyToken)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				var token Token
				err := json.NewDecoder(resp.Body).Decode(&token)
				if err != nil {
					log.Error("Error when parsing identity service response ", err)
					errors.Handler(w, ErrParseServiceResponse)
					return
				}

				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))

				ctx := context.WithValue(r.Context(), "token", token.Token)

				handler.ServeHTTP(w, r.WithContext(ctx))
			} else {
				handler.ServeHTTP(w, r)
			}
		})
	}
}

// OutMiddleware will
func OutMiddleware(url string) proxy.OutLink {
	return func(req *http.Request, res *http.Response) (*http.Response, error) {
		value := req.Context().Value("token")
		if value != nil {
			res.Header.Add("X-iClinic-Token", value.(string))
		}
		return res, nil
	}
}
