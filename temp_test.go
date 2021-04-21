package dgo_test

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/sam1677/dgo"
)

type (
	S = *discordgo.Session
	M = *discordgo.MessageCreate
)

//go:embed token.txt
var token string

func TestAll(t *testing.T) {
	h, err := dgo.New("!", token)
	if err != nil {
		t.Error(err)
		return
	}

	h.AddHandler("say", &dgo.Handler{
		func(s S, m M) {
			cmd, argc, argv := h.CmdArgs(m)
			fmt.Println(cmd, argc, argv)
			if argc == 0 {
				h.DefaultHandler(s, m)
				return
			}

			s.ChannelMessageSend(m.ChannelID, strings.Join(argv, " "))
		},
		"!say <text>",
		"Say something",
	})

	err = h.Open()
	if err != nil {
		t.Error(err)
		return
	}

	inte := make(chan os.Signal, 1)

	signal.Notify(inte, os.Interrupt, syscall.SIGTERM)
	<-inte
}
