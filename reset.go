package main

import "net/http"

func (cfg *apiConfig) handlerResetHits(_ http.ResponseWriter, _ *http.Request) {
	cfg.fileServerHits.Store(0)
}
