package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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

// RoutineTestHandler handles blabla
func RoutineTestHandler(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Client subscribed")

	myCounters := CountFeedback{
		CountDone:    0,
		CountDiscard: 0,
		CountCreate:  0,
		CountUpdate:  0,
	}

	for {
		time.Sleep(1 * time.Second)
		if myCounters.CountDone < 10 {
			if err != nil {
				fmt.Println(err)
				return
			}
			err = ws.WriteJSON(myCounters)
			if err != nil {
				fmt.Println(err)
				break
			}
			myCounters.CountDone += 2
		} else {
			fmt.Println("Client unsubscribed")
			err = ws.WriteJSON(myCounters)
			if err != nil {
				fmt.Println(err)
				break
			}
			ws.Close()
			break
		}
	}
}

func RoutineTestGetHandler(w http.ResponseWriter, r *http.Request) {
	views.RenderTmpl(w, "wsform", nil)
}

func RoutineTestPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k, v := range r.Form {
		fmt.Printf("k: %v / v: %v", k, v)
	}
}
