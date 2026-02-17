package main

import (
	"net/http"
)

type healthResp struct {
	Status  string `json:"status"`
	Env     string `json:"env"`
	Version string `json:"version"`
}

// Healthcheck godoc
//
//	@Summary		Health check
//	@Description	Health check of an app
//	@Tags			ops
//	@Produce		json
//	@Success		200	{object}	main.envelopeSuccess{data=main.healthResp}
//	@Failure		500	{object}	main.envelopeErr
//	@Security		ApiKeyAuth
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := healthResp{
		Status:  "ok",
		Env:     app.config.env,
		Version: version}
	if ok := app.jsonResponse(w, http.StatusOK, resp); ok != nil {
		app.internalServerError(w, r, ok)
	}
}
