package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/narrative/inthemix/core"
)

var (
	token string
)

func init() {
	flag.StringVar(&token, "token", "", "Bot Token")
	flag.Parse()
}

func main() {

	if token == "" {
		fmt.Println("Please provide discord bot token.")
		flag.Usage()
		return
	}

	pool, err := core.FillPool("talentpool.json")
	if err != nil {
		fmt.Println("Unable to read talent pool file: ", err)
	}

	te, err := core.Initialize(token)
	if err != nil {
		fmt.Println("Error initializing trigger engine: ", err)
		return
	}
	defer te.Close()

	te.AddTrigger("help", func(msg string) bool {
		return msg == "help"
	}, func(args []string) {
		if len(args) == 0 {
			availableCommands := te.GetAvailableCommands()
			for i := range availableCommands {
				availableCommands[i] = "!" + availableCommands[i]
			}
			te.SendReply("Available commands:" + strings.Join(availableCommands, ", "))
			return
		}
	})

	te.AddTrigger("latest", func(msg string) bool {
		return msg == "latest"
	}, func(args []string) {
		if len(args) == 0 {
			te.SendReply("Please specify a talent. Use !talent to view a list of available talent.")
			return
		}
		artistName := args[0]
		title, url, err := pool.GetLatest(artistName)
		if err != nil {
			te.SendReply("Unknown talent. Use !talent to view a list of available talent.")
			return
		}
		te.SendReply(fmt.Sprintf("Get in the mix with %v: %v", title, url))
	})

	te.AddTrigger("talent", func(msg string) bool {
		return msg == "talent"
	}, func(_ []string) {
		djs := pool.GetDjs()
		te.SendReply("Current talent pool: " + strings.Join(djs, ", ") + ".")
	})

	err = te.Open()

	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}

	fmt.Println("InTheMix is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
