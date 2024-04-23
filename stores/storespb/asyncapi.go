package storespb

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var asyncAPI embed.FS

func RegisterAsyncAPI(mux *chi.Mux) error {
	const specRoot = "/stores-asyncapi/"

	//mount the swagger specification
	mux.Mount(specRoot, http.StripPrefix(specRoot, http.FileServer(http.FS(asyncAPI))))
	return nil

}
