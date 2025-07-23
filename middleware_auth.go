package main

import (
	"fmt"
	"net/http"

	"github.com/mohamedkhalaf47/rssAggregator/internal/auth"
	"github.com/mohamedkhalaf47/rssAggregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			respondWithError(w, 403, fmt.Sprintf("Authentication error: %v", err))
			return
		}
		user, err := apiCfg.DB.GetUserByApiKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, 404, fmt.Sprintf("User not found: %v", err))
			return
		}
		handler(w, r, user)
	}
}
