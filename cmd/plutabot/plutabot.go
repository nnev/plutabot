package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robustirc/bridge/robustsession"
	"gopkg.in/sorcix/irc.v2"
)

func logic(fifo string) error {
	if fifo != "" {
		if err := syscall.Mkfifo(fifo, 0664); err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}

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

	if fifo != "" {
		go func() {
			for {
				b, err := ioutil.ReadFile(fifo)
				if err != nil {
					log.Fatal(err)
				}
				text := strings.TrimSpace(string(b))
				if text == "" {
					continue
				}
				for _, line := range strings.Split(text, "\n") {
					cmd(irc.PRIVMSG, "#chaos-hd", line)
				}
			}
		}()
	}

	go func() {
		for range time.Tick(1 * time.Minute) {
			cmd(irc.PING, "plutabot")
		}
	}()

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
			if nick != desiredNick {
				nick = desiredNick
				cmd(irc.NICK, desiredNick) // best effort
			}
		}
	}
	return nil
}

func main() {
	fifo := flag.String("fifo", "", "path to message FIFO")
	flag.Parse()
	if err := logic(*fifo); err != nil {
		log.Fatal(err)
	}
}
