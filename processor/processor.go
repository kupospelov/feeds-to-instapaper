package processor

import (
	"log"
	"slices"
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
	feeds := make([]*gofeed.Feed, 0, len(feedURLs))
	for _, feedURL := range feedURLs {
		feed, err := p.parser.ParseURL(feedURL)
		if err != nil {
			log.Printf("Error parsing feed %s: %v", feedURL, err)
			continue
		}

		feeds = append(feeds, feed)
	}

	items := make([]*gofeed.Item, 0)
	for _, feed := range feeds {
		for _, item := range feed.Items {
			if p.state.IsProcessed(item.Link) {
				continue
			}

			p.state.MarkProcessed(item.Link)
			items = append(items, item)
		}
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
