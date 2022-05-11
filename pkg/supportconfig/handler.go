package supportconfig

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type handler struct {
	g Generator
}

func NewHandler(g Generator) http.HandlerFunc {
	h := &handler{
		g: g,
	}
	return h.generateSupportConfig
}

const (
	tarContentType = "application/x-tar"
)

func (h *handler) generateSupportConfig(w http.ResponseWriter, r *http.Request) {
	archive, err := h.g.SupportConfig()
	if err != nil {
		h.handleErr(w, err)
		return
	}

	w.Header().Set("Content-Type", tarContentType)
	n, err := io.Copy(w, archive)
	if err != nil {
		h.handleErr(w, err)
		return
	}

	logrus.Debugf("wrote %v bytes in archive response", n)
}

func (h *handler) handleErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error": "%v"}`, err)
}
