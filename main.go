package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicomo/EResourcesMetadataHub/controllers"
)

func main() {

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.HandleFunc("/", controllers.HomeHandler)
	router.HandleFunc("/upload", controllers.UploadHandler)
	router.HandleFunc("/{epackage}", controllers.EpackageHandler)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("%s not found\n", r.URL)))
	})
	http.ListenAndServe(":8080", router)

}
