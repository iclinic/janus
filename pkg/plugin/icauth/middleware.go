package icauth

import (
	// "io/ioutil"
	"fmt"
	"encoding/json"
	"context"
	"net/http"
	"strings"
	"time"
	
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrAuthorizationFieldNotFound is used when the http Authorization header is missing from the request
	ErrAuthorizationFieldNotFound = errors.New(http.StatusBadRequest, "Authorization field missing")
	// ErrBearerMalformed is used when the Bearer string in the Authorization header is not found or is malformed
	ErrBearerMalformed = errors.New(http.StatusBadRequest, "Bearer token malformed")
	// ErrAccessTokenNotAuthorized is used when the access token is not found on the storage
	ErrAccessTokenNotAuthorized = errors.New(http.StatusUnauthorized, "access token not authorized")
	// ErrPartnerFieldNotFound is used when the http X-iClinic-Partner is missing from the requrest
	ErrPartnerFieldNotFound = errors.New(http.StatusBadRequest, "X-iClinic-Partner field missing")
	// ErrVerifyToken is used when a http get request to the identity service could not be made
	ErrVerifyToken = errors.New(http.StatusInternalServerError, "Could not communicate to our identity service")
	// ErrParseServiceResponse is used when http get request response from identity service could not be parsed
	ErrParseServiceResponse = errors.New(http.StatusInternalServerError, "Could not parse service response")
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

			partner := r.Header.Get("X-iClinic-Partner")
			if partner == "" {
				errors.Handler(w, ErrPartnerFieldNotFound)
				return
			}

			jwtToken := parts[1]

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Warn("Could not make a request to our services...", jwtToken, url)
				return
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
			req.Header.Add("X-iClinic-Partner", partner)

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
			log.WithField("token", value.(string)).Info("Here is the token JWT....")

			var data interface{}
			err := json.NewDecoder(res.Body).Decode(&data)
			if err != nil {
				log.Error("Error when parsing identity service response ", err)
				return res, nil
			}
			log.Infof("DAATAA: %v", data)

		}
		return res, nil
	}
}