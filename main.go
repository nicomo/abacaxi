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
	// get config params
	conf := config.GetConfig()

	// create a session store
	session.StoreCreate(conf.SessionStoreKey)

	// create a router & all toures
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// home page
	router.Handle("/", http.HandlerFunc(controllers.HomeHandler))

	// all inner pages subject to authentication
	router.Handle("/download/{filename:[\\w\\-\\.]+}", middleware.DisallowAnon(http.HandlerFunc(controllers.DownloadHandler)))
	router.Handle("/record/{recordID}", middleware.DisallowAnon(http.HandlerFunc(controllers.RecordHandler)))
	router.Handle("/record/delete/{ebookID}", middleware.DisallowAnon(http.HandlerFunc(controllers.RecordDeleteHandler)))
	router.Handle("/record/toggleacquired/{recordID}", middleware.DisallowAnon(http.HandlerFunc(controllers.RecordToggleAcquiredHandler)))
	router.Handle("/record/toggleactive/{recordID}", middleware.DisallowAnon(http.HandlerFunc(controllers.RecordToggleActiveHandler)))
	router.Handle("/package/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceHandler)))
	router.Handle("/package/{targetservice}/{page:[0-9]+}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServicePageHandler)))
	router.Handle("/package/toggleactive/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceToggleActiveHandler)))
	router.Handle("/package/update/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceUpdateGetHandler))).Methods("GET")
	router.Handle("/package/update/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceUpdatePostHandler))).Methods("POST")
	router.Handle("/packagenew", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceNewGetHandler))).Methods("GET")
	router.Handle("/packagenew", middleware.DisallowAnon(http.HandlerFunc(controllers.TargetServiceNewPostHandler))).Methods("POST")
	router.Handle("/search", middleware.DisallowAnon(http.HandlerFunc(controllers.SearchHandler)))
	router.Handle("/sudocgetrecord/{recordID}", middleware.DisallowAnon(http.HandlerFunc(controllers.GetRecordHandler)))
	router.Handle("/sudocgetrecords/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.GetRecordsTSHandler)))
	router.Handle("/sudoci2p/{recordID}", middleware.DisallowAnon(http.HandlerFunc(controllers.SudocI2PHandler)))
	router.Handle("/sudoci2p-ts-new/{targetservice}", middleware.DisallowAnon(http.HandlerFunc(controllers.SudocI2PTSNewHandler)))
	router.Handle("/upload", middleware.DisallowAnon(http.HandlerFunc(controllers.UploadGetHandler))).Methods("GET")
	router.Handle("/upload", middleware.DisallowAnon(http.HandlerFunc(controllers.UploadPostHandler))).Methods("POST")
	router.Handle("/users", middleware.DisallowAnon(http.HandlerFunc(controllers.UsersHandler)))
	router.Handle("/users/delete/{userID}", middleware.DisallowAnon(http.HandlerFunc(controllers.UserDeleteHandler)))
	// user login pages allowed for anon users only
	router.Handle("/users/login", middleware.DisallowAuthed(http.HandlerFunc(controllers.UserLoginGetHandler))).Methods("GET")
	router.Handle("/users/login", middleware.DisallowAuthed(http.HandlerFunc(controllers.UserLoginPostHandler))).Methods("POST")
	router.Handle("/users/logout", middleware.DisallowAnon(http.HandlerFunc(controllers.UserLogoutHandler)))
	router.Handle("/users/new", middleware.DisallowAnon(http.HandlerFunc(controllers.UserNewGetHandler))).Methods("GET")
	router.Handle("/users/new", middleware.DisallowAnon(http.HandlerFunc(controllers.UserNewPostHandler))).Methods("POST")

	// 404
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("%s not found\n", r.URL)))
	})

	// serve
	http.ListenAndServe(":8080", router)

}
