package web

import (
	"encoding/json"
	"log"
	"net/http"

	"bifrost/pkg/models"
)

func StartServer(port string) {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("pkg/web/static")))

	mux.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.GetMockReport())
	})

	log.Println("Server running on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}