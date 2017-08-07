package core

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"

	"github.com/mmcdole/gofeed"
)

type FeedInfo struct {
	talent talentPool
}

type feed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type talentPool []feed

func FillPool(configFilePath string) (*talentPool, error) {
	configFile, err := ioutil.ReadFile("talentpool.json")
	if err != nil {
		return nil, err
	}

	var pool talentPool
	json.Unmarshal(configFile, &pool)

	return &pool, nil
}

func (pool *talentPool) GetDjs() (djs []string) {
	for _, feed := range *pool {
		djs = append(djs, feed.Name)
	}
	return
}

func (pool *talentPool) findDj(dj string) (string, bool) {
	dj = strings.ToLower(strings.TrimSpace(dj))
	for _, feed := range *pool {
		if feed.Name == dj {
			return feed.URL, true
		}
	}
	return "", false
}

func (pool *talentPool) GetLatest(dj string) (string, *url.URL, error) {

	dj = strings.ToLower(strings.TrimSpace(dj))
	fp := gofeed.NewParser()

	feedURL, ok := pool.findDj(dj)
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
