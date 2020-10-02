package main

import "net/http"

func filterMiddleware() http.Handler {

	filter := timeoutHandler()

	return filter
}
