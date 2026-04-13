package files

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/shaolim/wa-integration/pkg/storage"
	"github.com/shaolim/wa-integration/web"
)

var filesTmpl = template.Must(template.New("files").Parse(web.FilesHTML))

type Lister interface {
	ListObjects(ctx context.Context) ([]storage.Object, error)
}

type Handler struct {
	lister Lister
}

func NewHandler(lister Lister) *Handler {
	return &Handler{lister: lister}
}

func (h *Handler) ServeFiles(rw http.ResponseWriter, r *http.Request) {
	objects, err := h.lister.ListObjects(r.Context())
	if err != nil {
		slog.Error("files: list objects", slog.Any("error", err))
		http.Error(rw, "Failed to list files", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	filesTmpl.Execute(rw, map[string]any{"Files": objects})
}
