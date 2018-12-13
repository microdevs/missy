package url

import (
	"net/http"

	"github.com/go-chi/chi"
)

func Parameter(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
