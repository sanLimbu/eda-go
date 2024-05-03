package rest

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var swaggerUI embed.FS

func RegisterSwagger(mux *chi.Mux) error {
	const specRoot = "/stores-spec/"

	// mount the swagger specifications
	mux.Mount(specRoot, http.StripPrefix(specRoot, http.FileServer(http.FS(swaggerUI))))

	return nil
}
