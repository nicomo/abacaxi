package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nicomo/abacaxi/logger"
	"github.com/nicomo/abacaxi/views"
)

type CountFeedback struct {
	CountDone    int
	CountDiscard int
	CountCreate  int
	CountUpdate  int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*
##
##
POST to Handler #1
Handler #1 :
  -- creates a channel
  -- returns to handler #2 passing the channel through a context
  -- parses file
Handler #2 :
  -- receives the channel through the context
  -- polls the channel to get the data and display it
##
##
*/

// RoutineTestHandler handles blabla
func RoutineTestHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Println("in RoutineTestHandler")

	d := make(map[string]interface{})

	// grab channel from context
	ch, err := fromContextCountFeedback(r.Context())
	if err != nil {
		logger.Error.Println(err)
	}
	logger.Debug.Println(ch)
	for elem := range ch {
		logger.Debug.Println("++++ %v", elem)
		d["countFeedback"] = elem
		views.RenderTmpl(w, "wsform2", d)
	}
}

func FormGetHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Println("in FormGetHandler")
	views.RenderTmpl(w, "wsform", nil)
}

func FormPostHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Println("in FormPostHandler")

	r.ParseForm()
	// create the channel
	ch := make(chan CountFeedback)
	defer close(ch)
	// get a context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	// create the counter
	myCounters := CountFeedback{
		CountDone:    0,
		CountDiscard: 0,
		CountCreate:  0,
		CountUpdate:  0,
	}
	// insert channel in context
	ctx = newContextCountFeedback(ctx, ch)

	// initial redirect to http, with context
	go RoutineTestHandler(w, r.WithContext(ctx))

	// do something with the values received from the form
	/*
		for k, v := range r.Form {
			fmt.Printf("k: %v / v: %v", k, v)
		}
	*/
	// count and update the channel
	for {
		logger.Debug.Println("in FormPostHandler for loop")
		time.Sleep(1 * time.Second)
		if myCounters.CountDone < 10 {
			// send mycounters in the channel
			myCounters.CountDone += 2
			ch <- myCounters
		} else {
			break
		}
	}
}
