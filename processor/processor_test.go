package processor

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kupospelov/feeds-to-instapaper/state"
	"github.com/mmcdole/gofeed"
)

type parser struct {
	feeds map[string]*gofeed.Feed
}

func (p *parser) ParseURL(feedURL string) (*gofeed.Feed, error) {
	feed, ok := p.feeds[feedURL]
	if ok {
		return feed, nil
	}

	return nil, fmt.Errorf("not found")
}

type instapaper struct {
	addedItems []addedItem
	err        map[string]error
}

type addedItem struct {
	link  string
	title string
}

func (i *instapaper) Add(link, title string) error {
	err := i.err[link]
	if err != nil {
		return err
	}

	i.addedItems = append(i.addedItems, addedItem{link: link, title: title})
	return nil
}

func parseTime(t string) *time.Time {
	timestamp, _ := time.Parse(time.Kitchen, t)
	return &timestamp
}

func createParser(feeds []*gofeed.Feed) Parser {
	p := &parser{
		feeds: make(map[string]*gofeed.Feed),
	}
	for _, feed := range feeds {
		p.feeds[feed.Link] = feed
	}
	return p
}

func TestSuccess(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Link: "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Article 1", PublishedParsed: parseTime("3:00PM")},
				{Link: "http://example.com/3", Title: "Article 3"},
				{Link: "http://example.com/2", Title: "Article 2", PublishedParsed: parseTime("3:02PM")},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{}
	testState := state.EmptyWithPath("test")
	processor := New(testParser, testInstapaper, testState)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if got, want := len(testInstapaper.addedItems), 3; got != want {
		t.Errorf("len(addedItems)=%d, want=%d", got, want)
	}
	if got, want := len(testState.NewItems), 3; got != want {
		t.Errorf("len(newItems)=%d, want=%d", got, want)
	}
	expectedItems := []struct {
		link, title string
	}{
		{"http://example.com/1", "Article 1"},
		{"http://example.com/2", "Article 2"},
		{"http://example.com/3", "Article 3"},
	}
	for i := range 3 {
		if got, want := testInstapaper.addedItems[i].link, expectedItems[i].link; got != want {
			t.Errorf("addedItems[%d].link=%s, want=%s", i, got, want)
		}
		if got, want := testInstapaper.addedItems[i].title, expectedItems[i].title; got != want {
			t.Errorf("addedItems[%d].title=%s, want=%s", i, got, want)
		}
		if got, want := testState.NewItems[i], expectedItems[i].link; got != want {
			t.Errorf("addedItems[%d].title=%s, want=%s", i, got, want)
		}
	}
}

func TestSkipProcessed(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Link: "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Already processed"},
				{Link: "http://example.com/2", Title: "New article"},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{}
	testState := state.EmptyWithPath("test")
	testState.MarkProcessed("http://example.com/1")
	processor := New(testParser, testInstapaper, testState)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if got, want := len(testInstapaper.addedItems), 1; got != want {
		t.Errorf("len(addedItems)=%d, want=%d", got, want)
	}
	if got, want := testInstapaper.addedItems[0].link, "http://example.com/2"; got != want {
		t.Errorf("addedItems[0].link=%s, want=%s", got, want)
	}
	if got, want := len(testState.NewItems), 1; got != want {
		t.Errorf("len(newItems)=%d, want=%d", got, want)
	}
	if got, want := testState.NewItems[0], "http://example.com/2"; got != want {
		t.Errorf("addedItems[0].link=%s, want=%s", got, want)
	}
}

func TestParserError(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Link: "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Article 1"},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{}
	testState := state.EmptyWithPath("test")
	processor := New(testParser, testInstapaper, testState)

	err := processor.ProcessFeeds([]string{"http://example.com", "http://error.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if got, want := len(testInstapaper.addedItems), 1; got != want {
		t.Errorf("len(addedItems)=%d, want=%d", got, want)
	}
	if got, want := testInstapaper.addedItems[0].link, "http://example.com/1"; got != want {
		t.Errorf("addedItems[0].link=%s, want=%s", got, want)
	}
	if got, want := len(testState.NewItems), 1; got != want {
		t.Errorf("len(newItems)=%d, want=%d", got, want)
	}
	if got, want := testState.NewItems[0], "http://example.com/1"; got != want {
		t.Errorf("addedItems[0].link=%s, want=%s", got, want)
	}
}

func TestInstapaperError(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Link: "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Article 1"},
				{Link: "http://example.com/2", Title: "Article 2"},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{err: map[string]error{"http://example.com/1": errors.New("API error")}}
	testState := state.EmptyWithPath("test")
	processor := New(testParser, testInstapaper, testState)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if got, want := len(testInstapaper.addedItems), 1; got != want {
		t.Errorf("len(addedItems)=%d, want=%d", got, want)
	}
	if got, want := testInstapaper.addedItems[0].link, "http://example.com/2"; got != want {
		t.Errorf("addedItems[0].link=%s, want=%s", got, want)
	}
	if got, want := len(testState.NewItems), 1; got != want {
		t.Errorf("len(newItems)=%d, want=%d", got, want)
	}
	if got, want := testState.NewItems[0], "http://example.com/2"; got != want {
		t.Errorf("addedItems[0].link=%s, want=%s", got, want)
	}
}
