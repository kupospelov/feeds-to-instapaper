package processor

import (
	"log"
	"slices"
	"sync"
	"time"

	"github.com/kupospelov/feeds-to-instapaper/state"
	"github.com/mmcdole/gofeed"
)

type Parser interface {
	ParseURL(feedURL string) (*gofeed.Feed, error)
}

type Instapaper interface {
	Add(link, title string) error
}

type Hooks interface {
	NewArticle(feed *gofeed.Feed, item *gofeed.Item)
}

type Processor struct {
	parser     Parser
	instapaper Instapaper
	hooks      Hooks
	state      *state.State
}

func New(parser Parser, instapaper Instapaper, hooks Hooks, state *state.State) *Processor {
	return &Processor{
		parser:     parser,
		instapaper: instapaper,
		hooks:      hooks,
		state:      state,
	}
}

func (p *Processor) ProcessFeeds(feedURLs []string) error {
	type feedItem struct {
		feed *gofeed.Feed
		item *gofeed.Item
	}
	itemsChan := make(chan feedItem)
	var wg sync.WaitGroup

	for _, feedURL := range feedURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			feed, err := p.parser.ParseURL(url)
			if err != nil {
				log.Printf("Error parsing feed %s: %v", url, err)
				return
			}

			for _, item := range feed.Items {
				if p.state.MarkProcessed(item.Link) {
					itemsChan <- feedItem{feed, item}
				}
			}
		}(feedURL)
	}
	go func() {
		wg.Wait()
		close(itemsChan)
	}()

	feedItems := make([]feedItem, 0)
	for item := range itemsChan {
		feedItems = append(feedItems, item)
	}
	slices.SortFunc(feedItems, func(a, b feedItem) int {
		var atime time.Time
		if a.item.PublishedParsed != nil {
			atime = *a.item.PublishedParsed
		}
		var btime time.Time
		if b.item.PublishedParsed != nil {
			btime = *b.item.PublishedParsed
		}
		return atime.Compare(btime)
	})

	for _, fi := range feedItems {
		log.Printf("Adding to Instapaper: %s", fi.item.Title)
		err := p.instapaper.Add(fi.item.Link, fi.item.Title)
		if err != nil {
			log.Printf("Error adding link %s: %v", fi.item.Link, err)
			continue
		}

		p.hooks.NewArticle(fi.feed, fi.item)
		p.state.Append(fi.item.Link)
	}

	return nil
}
