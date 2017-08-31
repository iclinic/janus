package icsignup

import (
	"encoding/json"
	"net/http"
	
	"github.com/hellofresh/janus/pkg/errors"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrAuthCommunication is used when http post request response from auth service could not be done
	ErrAuthCommunication = errors.New(http.StatusInternalServerError, "Could not communicate to auth service")
)

// Midleware will hit iclinic auth service
func Midleware(authURL, apiURL string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Error("[signup middleware] - Starting...")

			res, err := http.Post(authURL, "application/json", r.Body)
			if err != nil {
				log.Error("Fail when communicating to auth service ", err)
				errors.Handler(w, ErrAuthCommunication)
				return
			}

			if res.StatusCode == http.StatusCreated {
				log.Error("[signup middleware] - Created...")

			} else {
				var data interface{}
				err := json.NewDecoder(res.Body).Decode(&data)
				if err != nil {
					log.Error("Error when submiting post data to user service.", err)
					errors.Handler(w, errors.New(res.StatusCode, "Error when communicating to auth service"))
					return
				}

				log.Error("Error when submiting post data to user service @@@@@.", data)
			}

			// render.JSON(w, http.StatusCreated, data)
			// handler.ServeHTTP(w, r)
		})
	}
}
