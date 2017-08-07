package core

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type trigger struct {
	name    string
	matches func(string) bool
	action  func([]string)
}

type TriggerSystem struct {
	s        *discordgo.Session
	m        *discordgo.MessageCreate
	triggers []trigger
}

//
func Initialize(token string) (*TriggerSystem, error) {
	te := new(TriggerSystem)
	s, err := discordgo.New("Bot " + token)
	te.s = s
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return nil, err
	}

	s.AddHandler(messageCreate(te))

	return te, nil
}

func (te *TriggerSystem) Close() {
	te.s.Close()
}

func (te *TriggerSystem) Open() error {
	return te.s.Open()
}

func (te *TriggerSystem) AddTrigger(name string, matches func(string) bool, action func([]string)) {
	te.triggers = append(te.triggers, trigger{name: name, matches: matches, action: action})
}

func (te *TriggerSystem) SendReply(message string) {
	te.s.ChannelMessageSend(te.m.ChannelID, message)
}

func (te *TriggerSystem) RunTriggerReader(s *discordgo.Session, m *discordgo.MessageCreate) {
	te.s = s
	te.m = m

	message := te.m.Content
	split := strings.Split(message, " ")

	for _, trigger := range te.triggers {
		if trigger.matches(strings.TrimPrefix(split[0], "!")) {
			trigger.action(split[1:])
		}
	}
}

func (te *TriggerSystem) GetAvailableCommands() (result []string) {
	for _, trigger := range te.triggers {
		result = append(result, trigger.name)
	}
	return
}

func messageCreate(te *TriggerSystem) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Content[0] != '!' {
			return
		}
		te.RunTriggerReader(s, m)
	}
}
