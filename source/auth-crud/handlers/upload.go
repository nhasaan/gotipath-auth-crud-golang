package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "/uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "failed to prepare upload dir", http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(header.Filename)
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(uploadDir, name)

	out, err := os.Create(path)
	if err != nil {
		http.Error(w, "failed to store file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"path":"%s"}`, path)))
}
