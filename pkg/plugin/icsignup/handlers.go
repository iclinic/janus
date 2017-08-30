package icsignup

import (
	"net/http"

	"github.com/hellofresh/janus/pkg/render"

	log "github.com/sirupsen/logrus"
)

type Handler struct {}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Request to create a new account..")

		render.JSON(w, http.StatusCreated, nil)
	}
}