package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/mmcdole/gofeed"
	"github.com/narrative/inthemix/engine"
)

var (
	token    string
	podcasts = map[string]string{
		"spinninsessions": "http://spinninsessions.spinninpodcasts.com/rss",
		"tritonia":        "http://tritonia.libsyn.com/rss",
		"hexagon":         "https://www.thisisdistorted.com/repository/xml/DonDiabloHexagonRadio1422895802.xml",
		"hardwell":        "http://podcast.djhardwell.com/podcast.xml",
		"tiesto":          "https://feed.pippa.io/public/shows/clublife-by-tiesto",
	}
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

	te, err := engine.Initialize(token)
	if err != nil {
		fmt.Println("Error initializing trigger engine: ", err)
		return
	}
	defer te.Close()
	addTriggers(te)
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

func addTriggers(te *engine.TriggerEngine) {
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
		talent := args[0]
		title, url, err := getLatest(talent)
		if err != nil {
			te.SendReply(fmt.Sprintf("Error getting latest episode: %v", err))
			return
		}
		te.SendReply(fmt.Sprintf("Get in the mix with %v: %v", title, url))
	})

	te.AddTrigger("talent", func(msg string) bool {
		return msg == "talent"
	}, func(_ []string) {
		djs := getDjs()
		te.SendReply("Current talent pool: " + strings.Join(djs, ", ") + ".")
	})
}

func getDjs() (djs []string) {
	for djName := range podcasts {
		djs = append(djs, djName)
	}
	return
}

func getLatest(dj string) (string, *url.URL, error) {

	dj = strings.ToLower(strings.TrimSpace(dj))
	fp := gofeed.NewParser()

	feedURL, ok := podcasts[dj]
	if !ok {
		return "", nil, errors.New("unknown talent")
	}

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return "", nil, err
	}

	latestItem := latestEpisodeFromFeed(feed)
	url, err := url.Parse(latestItem.Enclosures[0].URL)
	if err != nil {
		return "", nil, err
	}

	return latestItem.Title, url, nil
}

type byDate []*gofeed.Item

func (a byDate) Len() int {
	return len(a)
}
func (a byDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byDate) Less(i, j int) bool {
	return a[i].PublishedParsed.After(*(a[j].PublishedParsed))
}

func latestEpisodeFromFeed(feed *gofeed.Feed) *gofeed.Item {
	sort.Sort(byDate(feed.Items))
	return feed.Items[0]
}
