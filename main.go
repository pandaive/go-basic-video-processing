package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// channel to send notifications
var notification chan string

//map to store pending notifications
var pending map[string]string

func main() {
	pending = make(map[string]string)
	notification = make(chan string)
	go notificationSender()

	r := mux.NewRouter()
	r.HandleFunc("/", uploadImage).Methods("PUT")
	r.HandleFunc("/", getNotification).Methods("GET")
	http.Handle("/", r)

	http.ListenAndServe("localhost:8080", nil)
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	err := r.ParseMultipartForm(10000)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("File too large!")))
		return
	}

	fmt.Fprintf(w, "Processing request\n")
	timestamp := fmt.Sprint(time.Now().Unix())
	filepath := "./frames/" + timestamp
	os.MkdirAll(filepath, os.ModePerm)

	file, err := os.Create(filepath + "/file")
	defer file.Close()
	if err != nil {
		panic(err)
	}

	video, err := r.MultipartForm.File["file"][0].Open()
	defer video.Close()
	n, err := io.Copy(file, video)
	if err != nil {
		panic(err)
	}

	callbackURL := r.URL.Query().Get("callback")
	log.Printf("Received " + timestamp + "for " + callbackURL)
	w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\n", n)))

	go processVideo(timestamp)
	w.Write([]byte(fmt.Sprintf("Video is being processed.\n")))
	pending[timestamp] = callbackURL

}

func getNotification(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got notification!")
}
