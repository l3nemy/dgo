package dgo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrCommandExists = errors.New("command exists. If you want to override it, use AddHandlerOverride instead")
	ErrCommandEmpty  = errors.New("command is empty")
)

type (
	MessageHandler func(*discordgo.Session, *discordgo.MessageCreate)

	Helper struct {
		prefix         string
		Session        *discordgo.Session
		handlerList    map[string]*Handler
		DefaultHandler MessageHandler
	}

	Handler struct {
		Handler     MessageHandler
		Usage       string
		Description string
	}
)

func New(prefix string, botToken string) (h *Helper, err error) {
	s, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return
	}

	return NewHelperFromSession(prefix, s), nil
}

func NewHelperFromSession(prefix string, s *discordgo.Session) (h *Helper) {
	h = new(Helper)
	h.prefix = prefix
	h.Session = s
	h.handlerList = make(map[string]*Handler)
	h.DefaultHandler = func(s *discordgo.Session, mc *discordgo.MessageCreate) {
		fields := []*discordgo.MessageEmbedField{}
		for cmd, handler := range h.handlerList {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  h.prefix + cmd,
				Value: fmt.Sprintf("%s (Usage: %s)", handler.Description, handler.Usage),
			})
		}

		s.ChannelMessageSendEmbed(
			mc.ChannelID,
			&discordgo.MessageEmbed{
				Title:  "Help",
				Fields: fields,
			},
		)
	}
	s.AddHandler(h.handlerProxy)
	return
}

func (h *Helper) Open() error {
	return h.Session.Open()
}

func (h *Helper) SetPrefix(prefix string) {
	h.prefix = prefix
}

func (h *Helper) MustAddHandlers(handlers map[string]*Handler) map[string]func() {
	return h.mustAddHandlers(handlers, true)
}

func (h *Helper) AddHandlers(handlers map[string]*Handler) (map[string]func(), error) {
	return h.addHandlers(handlers, false)
}

func (h *Helper) AddHandlersOverride(handlers map[string]*Handler) map[string]func() {
	return h.mustAddHandlers(handlers, false)
}

func (h *Helper) addHandlers(handlers map[string]*Handler, override bool) (map[string]func(), error) {
	ret := map[string]func(){}

	for command, handler := range handlers {
		d, err := h.addHandler(command, handler, override)
		if err != nil {
			return nil, err
		}
		ret[command] = d
	}
	return ret, nil
}

func (h *Helper) mustAddHandlers(handlers map[string]*Handler, override bool) map[string]func() {
	ret, err := h.addHandlers(handlers, override)
	if err != nil {
		panic(err)
	}
	return ret
}

func (h *Helper) MustAddHandler(command string, handler *Handler) func() {
	return h.mustAddHandler(command, handler, false)
}

func (h *Helper) AddHandler(command string, handler *Handler) (func(), error) {
	return h.addHandler(command, handler, false)
}

func (h *Helper) AddHandlerOverride(command string, handler *Handler) func() {
	return h.mustAddHandler(command, handler, true)
}

func (h *Helper) CmdArgs(m *discordgo.MessageCreate) (cmd string, argc int, argv []string) {
	return h.cmdArgs(m.Content)
}

func (h *Helper) mustAddHandler(command string, handler *Handler, override bool) func() {
	ret, err := h.addHandler(command, handler, override)
	if err != nil {
		panic(err)
	}
	return ret
}

func (h *Helper) addHandler(command string, handler *Handler, override bool) (func(), error) {
	if command == "" {
		return nil, ErrCommandEmpty
	}

	if _, ex := h.handlerList[command]; ex && !override {
		return nil, ErrCommandExists
	}

	h.handlerList[command] = handler

	return func() {
		h.deleteHandler(command)
	}, nil
}

func (h *Helper) deleteHandler(command string) {
	delete(h.handlerList, command)
}

func (h *Helper) handlerProxy(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !strings.HasPrefix(m.Content, h.prefix) || m.Author.Bot {
		return
	}

	cmd, _, _ := h.CmdArgs(m)
	if len(cmd) == 0 {
		return
	}

	for command, handler := range h.handlerList {
		if command == cmd {
			handler.Handler(s, m)
			return
		}
	}
	h.DefaultHandler(s, m)
}

func (h *Helper) cmdArgs(content string) (cmd string, argc int, argv []string) {
	c := strings.Split(
		strings.TrimSpace(
			content[len(h.prefix):],
		),
		" ",
	)

	cmd = c[0]
	if len(cmd) == 0 {
		return
	}

	argv = c[1:]
	argc = len(argv)
	return
}
