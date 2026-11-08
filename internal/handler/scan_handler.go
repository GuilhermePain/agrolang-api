package handler

import (
	"context"
	"log"
	"net/http"
)

// ScanHandler triggers an on-demand risk scan (RF-02, RF-03). Any logged-in
// user may trigger it, but the scan itself is global: it processes every
// property, exactly like the background loop. Mount it behind RequireSession.
type ScanHandler struct {
	run func(ctx context.Context) error
}

func NewScanHandler(run func(ctx context.Context) error) *ScanHandler {
	return &ScanHandler{run: run}
}

func (h *ScanHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /scan/run", h.runScan)
}

func (h *ScanHandler) runScan(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUser(w, r); !ok {
		return
	}

	if err := h.run(r.Context()); err != nil {
		log.Printf("scan run error: %v", err)
		writeError(w, http.StatusInternalServerError, "scan failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "scan_completed"})
}
