package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/controllers"
)

func main() {

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.HandleFunc("/", controllers.HomeHandler)
	router.HandleFunc("/download/{param:[\\w\\-\\.]+}", controllers.DownloadHandler)
	router.HandleFunc("/ebook/{ebookID}", controllers.EbookHandler)
	router.HandleFunc("/ebook/delete/{ebookID}", controllers.EbookDeleteHandler)
	router.HandleFunc("/search", controllers.SearchHandler)
	router.HandleFunc("/package/{targetservice}", controllers.TargetServiceHandler)
	router.HandleFunc("/packagenew", controllers.TargetServiceNewGetHandler).Methods("GET")
	router.HandleFunc("/packagenew", controllers.TargetServiceNewPostHandler).Methods("POST")
	router.HandleFunc("/sudocgetrecord/{ebookID}", controllers.GetRecordHandler)
	router.HandleFunc("/sudocgetrecords/{targetservice}", controllers.GetRecordsTSHandler)
	router.HandleFunc("/sudoci2p/{ebookID}", controllers.SudocI2PHandler)
	router.HandleFunc("/sudoci2p-ts-new/{targetservice}", controllers.SudocI2PTSNewHandler)
	router.HandleFunc("/upload", controllers.UploadHandler)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("%s not found\n", r.URL)))
	})
	http.ListenAndServe(":8080", router)

}
