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
	// ErrAuthCommunication is used when http post request response from auth service could not be done
	ErrAuthCommunication = errors.New(http.StatusInternalServerError, "Falha ao comunicar com serviço.")
	// ErrParseAuthResponse is used when http post response from auth service could not be parsed
	ErrParseAuthResponse = errors.New(http.StatusInternalServerError, "Falha ao interpretar resposta de um de nossos serviços.")
	// ErrParseIncomeRequestdata is used when request data could not be parsed
	ErrParseIncomeRequestdata = errors.New(http.StatusInternalServerError, "Falha ao intepretar dados recebidos.")
)

// Midleware will hit iclinic auth service
func Midleware(createUserURL, deleteUserURL, subscriptionURL string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Error("[signup middleware] - Starting...")

			fields := log.Fields{
				"Create User URL": createUserURL,
				"Delete User URL": deleteUserURL,
				"Subscription URL": subscriptionURL,
			}
			log.WithFields(fields).Error("Signup urls...")

			incomingBodyBuffer, _ := ioutil.ReadAll(r.Body)

			res, err := DoRequest(http.MethodPost, createUserURL, incomingBodyBuffer)
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
					log.Error("Error when parsing auth service response.", err)
					errors.Handler(w, ErrParseAuthResponse)
					return
				}

				var postData interface{}
				errPostData := json.NewDecoder(bytes.NewBuffer(incomingBodyBuffer)).Decode(&postData)
				if errPostData != nil {
					log.Error("Error when parsing request data. ", errPostData)
					errors.Handler(w, ErrParseIncomeRequestdata)
					return
				}

				apiRequestData := postData.(map[string]interface{})
				apiRequestData["user"] = userData.(map[string]interface{})["id"]

				apiBufBody := &bytes.Buffer{}
				errAPIBufBody := json.NewEncoder(apiBufBody).Encode(apiRequestData)
				if errAPIBufBody != nil {
					log.Error("Error when buffering api request data.", err)
				}

				res, err := DoRequest(http.MethodPost, subscriptionURL, apiBufBody.Bytes())
				if err != nil {
					log.Error("Error when posting user data to api.", err)
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
						log.Error("Error when subscribing user to api.", err)
						errors.Handler(w, errors.New(res.StatusCode, "Falha ao efetivar cadastro."))
						return
					}

					deleteUserEndpoint := strings.Replace(deleteUserURL, "<id>", apiRequestData["user"].(string), 1)

					DoRequest(http.MethodDelete, deleteUserEndpoint, nil)

					render.JSON(w, res.StatusCode, data)
				}

			} else {
				var data interface{}
				err := json.NewDecoder(res.Body).Decode(&data)
				if err != nil {
					log.Error("Error when auth service error happened.", err)
					errors.Handler(w, errors.New(res.StatusCode, "Falha ao enviar dados de cadastro."))
					return
				}
				render.JSON(w, res.StatusCode, data)
			}
		})
	}
}

// DoRequest is used to make a http request
func DoRequest(method, url string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: time.Second * 60}
	res, err := client.Do(req)
	return res, err
}
