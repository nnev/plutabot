package main

import (
	"log"
	"strings"

	"github.com/robustirc/bridge/robustsession"
)

func logic() error {
	session, err := robustsession.Create("robustirc.net", "")
	if err != nil {
		return err
	}
	session.PostMessage("NICK pluta")
	session.PostMessage("USER pluta pluta pluta pluta")
	session.PostMessage("JOIN #chaos-hd")
	for msg := range session.Messages {
		log.Printf("<- %s", msg)
		if strings.HasPrefix(msg, "PING ") {
			session.PostMessage("PONG " + strings.TrimPrefix(msg, "PING "))
			log.Printf("-> pong")
		}
	}
	return nil
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
