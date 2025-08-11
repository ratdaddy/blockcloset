package httpapi

import "net/http"

func NewStdlibRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path != "/" {
			w.Header().Set("Location", r.URL.Path)
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.NotFound(w, r)
	})
	return mux
}
