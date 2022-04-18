package supportconfig

import "net/http"

type handler struct {
	g Generator
}

func NewHandler(g Generator) http.HandlerFunc {
	h := &handler{
		g: g,
	}
	return h.generateSupportConfig
}

func (h *handler) generateSupportConfig(w http.ResponseWriter, r *http.Request) {}
