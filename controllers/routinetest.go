package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nicomo/abacaxi/session"
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
	// Get session
	sess := session.Instance(r)

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

			// TODO : close the connection +
			// pass on counter Report as flash message +
			// redirect to ts

			fmt.Println("Client unsubscribed")

			reportMssg := "Parsed " +
				strconv.Itoa(myCounters.CountDone) + "records.\n" +
				"Discarded " +
				strconv.Itoa(myCounters.CountDiscard) + "records.\n" +
				"Created " +
				strconv.Itoa(myCounters.CountCreate) + "records.\n" +
				"Updated " +
				strconv.Itoa(myCounters.CountUpdate)

			sess.AddFlash(reportMssg)
			sess.Save(r, w)
			ws.Close()
			break
		}
	}
}
