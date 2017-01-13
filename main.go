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
	router.HandleFunc("/ebook/{ebookId}", controllers.EbookHandler)
	router.HandleFunc("/search", controllers.SearchHandler)
	router.HandleFunc("/package/{targetservice}", controllers.TargetServiceHandler)
	router.HandleFunc("/packagenew", controllers.TargetServiceNewGetHandler).Methods("GET")
	router.HandleFunc("/packagenew", controllers.TargetServiceNewPostHandler).Methods("POST")
	router.HandleFunc("/upload", controllers.UploadHandler)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("%s not found\n", r.URL)))
	})
	http.ListenAndServe(":8080", router)

}
