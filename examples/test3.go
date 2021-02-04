package main

import (
	"log"
	"time"
)

func beat(c_out chan string) {
	log.Print("Beat: Starting loop")
	for i := 0; i < 10; i++ {
		log.Print("Beat: repeating loop")
		time.Sleep(1 * time.Second)
		c_out <- "BIP - " + time.Now().Format("2006-01-02 15:04:05.000")
	}
	close(c_out)

}

func p(c_in chan string) {
	var v string
	var more bool

	more = true
	log.Print("p: Starting loop")
	for more {
		log.Print("p: repeating loop")
		v, more = <-c_in
		log.Print(v)
	}
	log.Print("p: chan closed")

}

func main() {
	c := make(chan string)

	go beat(c)
	go p(c)

	time.Sleep(125 * time.Second)
	log.Print("ENDING PROGRAM")
	return
}
