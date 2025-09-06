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

type Processor struct {
	parser     Parser
	instapaper Instapaper
	state      *state.State
}

func New(parser Parser, instapaper Instapaper, state *state.State) *Processor {
	return &Processor{
		parser:     parser,
		instapaper: instapaper,
		state:      state,
	}
}

func (p *Processor) ProcessFeeds(feedURLs []string) error {
	itemsChan := make(chan *gofeed.Item)
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
					itemsChan <- item
				}
			}
		}(feedURL)
	}
	go func() {
		wg.Wait()
		close(itemsChan)
	}()

	items := make([]*gofeed.Item, 0)
	for item := range itemsChan {
		items = append(items, item)
	}
	slices.SortFunc(items, func(a, b *gofeed.Item) int {
		var atime time.Time
		if a.PublishedParsed != nil {
			atime = *a.PublishedParsed
		}
		var btime time.Time
		if b.PublishedParsed != nil {
			btime = *b.PublishedParsed
		}
		return atime.Compare(btime)
	})

	for _, item := range items {
		log.Printf("Adding to Instapaper: %s", item.Title)
		err := p.instapaper.Add(item.Link, item.Title)
		if err != nil {
			log.Printf("Error adding link %s: %v", item.Link, err)
			continue
		}

		p.state.Append(item.Link)
	}

	return nil
}
