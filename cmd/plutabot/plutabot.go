package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robustirc/bridge/robustsession"
	"gopkg.in/sorcix/irc.v2"
)

func logic() error {
	session, err := robustsession.Create("robustirc.net", "")
	if err != nil {
		return err
	}

	{
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		go func() {
			s := <-c
			log.Printf("Got signal: %v", s)
			session.Delete(fmt.Sprintf("signal %v", s))
		}()
	}

	postMessage := func(msg *irc.Message) error {
		log.Printf("-> %s", msg)
		return session.PostMessage(msg.String())
	}
	cmd := func(command string, params ...string) error {
		return postMessage(&irc.Message{
			Command: command,
			Params:  params,
		})
	}

	const desiredNick = "pluta"

	nick := desiredNick
	cmd(irc.NICK, nick)
	cmd(irc.USER, "pluta", "pluta", "pluta", "pluta")
	for rawmsg := range session.Messages {
		message := irc.ParseMessage(rawmsg)
		log.Printf("<- %s", message)
		switch message.Command {
		case irc.RPL_WELCOME: // logged in
			cmd(irc.JOIN, "#chaos-hd")

		case irc.ERR_NICKNAMEINUSE: // nickname already in use
			nick = nick + "_"
			cmd(irc.NICK, nick)

		case irc.PING:
			cmd(irc.PONG, message.Params...)

		case irc.NICK:
		case irc.QUIT:
			cmd(irc.NICK, desiredNick) // best effort
		}
	}
	return nil
}

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
