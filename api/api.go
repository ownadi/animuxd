package api

import (
	"animuxd/xdcc"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type requestFilePayload struct {
	BotNick       string
	PackageNumber int
	FileName      string
}

func NewRouter(engine xdcc.XDCCEngine) http.Handler {
	router := httprouter.New()

	createDownload := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var payload requestFilePayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if payload.BotNick == "" || payload.PackageNumber == 0 || payload.FileName == "" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		requestPromise := engine.RequestFile(payload.BotNick, payload.PackageNumber, payload.FileName)
		<-requestPromise

		w.WriteHeader(http.StatusCreated)
	}

	indexDownloads := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := engine.DownloadsJSON(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	router.POST("/downloads", createDownload)
	router.GET("/downloads", indexDownloads)

	return router
}
