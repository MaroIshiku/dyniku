package server

import (
	"net/http"

	"github.com/MaroIshiku/dyniku/internal/models"
)

func (h *handlers) index(w http.ResponseWriter, _ *http.Request) {
	var htmlData models.HTMLData
	for _, record := range h.db.SelectAll() {
		row := record.HTML(h.timeNow())
		htmlData.Rows = append(htmlData.Rows, row)
	}
	err := h.indexTemplate.ExecuteTemplate(w, "index.html", htmlData)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed generating webpage: "+err.Error())
	}
}

func (h *handlers) setupPage(w http.ResponseWriter, r *http.Request) {
	authFile, err := h.auth.load()
	if err == nil && authFile.Admin != nil && authFile.SetupCompleted {
		if _, ok := h.currentUser(r); ok {
			http.Redirect(w, r, "./", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "login", http.StatusSeeOther)
		return
	}
	h.index(w, r)
}
