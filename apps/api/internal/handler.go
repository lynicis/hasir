package internal

import "net/http"

type GlobalHandler interface {
	RegisterRoutes() (string, http.Handler)
}
