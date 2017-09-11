package icsignup

import (
	"strings"
	"time"
	"io/ioutil"
	"bytes"
	"encoding/json"
	"net/http"
	
	"github.com/hellofresh/janus/pkg/render"
	"github.com/hellofresh/janus/pkg/errors"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrPartnerFieldNotFound is used when the http X-iClinic-Partner is missing from the requrest
	ErrPartnerFieldNotFound = errors.New(http.StatusBadRequest, "X-iClinic-Partner field missing")
	// ErrAuthCommunication is used when http post request response from auth service could not be done
	ErrAuthCommunication = errors.New(http.StatusInternalServerError, "Could not communicate to auth service")
)


// Midleware will hit iclinic auth service
func Midleware(createUserURL, deleteUserURL, subscriptionURL string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Error("[signup middleware] - Starting...")

			partner := r.Header.Get("X-iClinic-Partner")
			if partner == "" {
				errors.Handler(w, ErrPartnerFieldNotFound)
				return
			}

			incomingBodyBuffer, _ := ioutil.ReadAll(r.Body)

			res, err := DoRequest(http.MethodPost, createUserURL, partner, incomingBodyBuffer)
			if err != nil {
				log.Error("Fail when communicating to auth service ", err)
				errors.Handler(w, ErrAuthCommunication)
				return
			}
			defer res.Body.Close()

			if res.StatusCode == http.StatusCreated {
				var userData interface{}
				err := json.NewDecoder(res.Body).Decode(&userData)
				if err != nil {
					log.Error("Error parsing auth data.", err)
					errors.Handler(w, errors.New(http.StatusInternalServerError, "Error when communicating to auth service"))
					return
				}

				var postData interface{}
				errPostData := json.NewDecoder(bytes.NewBuffer(incomingBodyBuffer)).Decode(&postData)
				if errPostData != nil {
					log.Error("Error parsing api data. ", errPostData)
					errors.Handler(w, errors.New(http.StatusInternalServerError, "Error when communicating to auth service"))
					return
				}

				apiData := postData.(map[string]interface{})
				apiData["user"] = userData.(map[string]interface{})["id"]

				apiBufBody := &bytes.Buffer{}
				errAPIBufBody := json.NewEncoder(apiBufBody).Encode(apiData)
				if errAPIBufBody != nil {
					log.Error("Error when auth service error happened.", err)
				}

				res, err := DoRequest(http.MethodPost, subscriptionURL, "", apiBufBody.Bytes())
				if err != nil {
					log.Error("Error when auth service error happened.", err)
				}
				defer res.Body.Close()

				if res.StatusCode == http.StatusCreated {
					render.JSON(w, http.StatusCreated, map[string]bool{
						"created": true,
					})
				} else {
					// error when posting subscription data
					var data interface{}
					err := json.NewDecoder(res.Body).Decode(&data)
					if err != nil {
						log.Error("Error when auth service error happened.", err)
						errors.Handler(w, errors.New(res.StatusCode, "Error when communicating to auth service"))
						return
					}

					deleteUserEndpoint := strings.Replace(deleteUserURL, "<id>", apiData["user"].(string), 1)

					DoRequest(http.MethodDelete, deleteUserEndpoint, partner, nil)

					render.JSON(w, res.StatusCode, data)
				}

			} else {
				var data interface{}
				err := json.NewDecoder(res.Body).Decode(&data)
				if err != nil {
					log.Error("Error when auth service error happened.", err)
					errors.Handler(w, errors.New(res.StatusCode, "Error when communicating to auth service"))
					return
				}
				render.JSON(w, res.StatusCode, data)
			}
		})
	}
}

// DoRequest is used to make a http request
func DoRequest(method, url, partner string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-iClinic-Partner", partner)

	client := &http.Client{Timeout: time.Second * 60}
	res, err := client.Do(req)
	return res, err
}
