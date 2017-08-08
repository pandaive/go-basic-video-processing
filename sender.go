package main

import (
	"log"
	"net/http"
)

func notificationSender() {
	for timestamp := range notification {
		go func(timestamp string) {
			log.Printf("Sending notification..")
			_, err := http.Get(pending[timestamp])
			if err != nil {
				log.Printf("Couldn't reach host " + pending[timestamp])
				panic(err)
			}
		}(timestamp)
	}

}
