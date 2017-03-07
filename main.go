package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicomo/abacaxi/config"
	"github.com/nicomo/abacaxi/controllers"
	"github.com/nicomo/abacaxi/middleware"
	"github.com/nicomo/abacaxi/session"
)

func main() {
	conf := config.GetConfig()
	session.Configure(conf.SessionConf)

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	router.Handle("/", middleware.DisallowAnon(http.HandlerFunc(controllers.HomeHandler)))
	router.Handle("/download/{filename:[\\w\\-\\.]+}", middleware.DisallowAnon(http.HandlerFunc(controllers.DownloadHandler)))
	router.Handle("/ebook/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.EbookHandler)))
	router.Handle("/ebook/delete/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.EbookDeleteHandler)))
	router.Handle("/search", middleware.DisallowAnon(http.HandlerFunc(controllers.SearchHandler)))
	router.Handle("/package/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceHandler)))
	router.Handle("/package/{targetservice}/{page:[0-9]+}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServicePageHandler)))
	router.Handle("/package/toggleactive/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceToggleActiveHandler)))
	router.Handle("/package/update/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceUpdateGetHandler))).Methods("GET")
	router.Handle("/package/update/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceUpdatePostHandler))).Methods("POST")
	router.Handle("/packagenew", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceNewGetHandler))).Methods("GET")
	router.Handle("/packagenew", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceNewPostHandler))).Methods("POST")
	router.Handle("/sudocgetrecord/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.GetRecordHandler)))
	router.Handle("/sudocgetrecords/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.GetRecordsTSHandler)))
	router.Handle("/sudoci2p/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.SudocI2PHandler)))
	router.Handle("/sudoci2p-ts-new/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.SudocI2PTSNewHandler)))
	router.Handle("/ebook/toggleacquired/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.EbookToggleAcquiredHandler)))
	router.Handle("/ebook/toggleactive/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.EbookToggleActiveHandler)))
	router.Handle("/upload", middleware.DisallowAnon(http.HandlerFunc(controllers.UploadGetHandler))).Methods("GET")
	router.Handle("/upload", middleware.DisallowAnon(http.HandlerFunc(controllers.UploadPostHandler))).Methods("POST")
	router.Handle("/users", middleware.DisallowAnon(http.HandlerFunc(controllers.UsersHandler)))
	// user login pages allowed for anon users
	router.Handle("/users/login", middleware.DisallowAuthed(http.HandlerFunc(controllers.UsersLoginGetHandler))).Methods("GET")
	router.Handle("/users/login", middleware.DisallowAuthed(http.HandlerFunc(controllers.UsersLoginPostHandler))).Methods("POST")
	router.Handle("/users/logout", middleware.DisallowAnon(http.HandlerFunc(controllers.UsersLogoutHandler)))
	router.Handle("/users/new", middleware.DisallowAnon(http.HandlerFunc(controllers.UsersNewGetHandler))).Methods("GET")
	router.Handle("/users/new", middleware.DisallowAnon(http.HandlerFunc(controllers.UsersNewPostHandler))).Methods("POST")
	router.Handle("/users/delete/{userID}", middleware.DisallowAnon(http.HandlerFunc(controllers.UsersDeleteHandler)))

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("%s not found\n", r.URL)))
	})
	http.ListenAndServe(":8080", router)

}
