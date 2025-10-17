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

func (i *instapaper) assertAddedItems(t *testing.T, expected []addedItem) {
	if got, want := len(i.addedItems), len(expected); got != want {
		t.Errorf("len(addedItems)=%d, want=%d", got, want)
	}
	for j := range len(expected) {
		if got, want := i.addedItems[j].link, expected[j].link; got != want {
			t.Errorf("addedItems[%d].link=%s, want=%s", j, got, want)
		}
		if got, want := i.addedItems[j].title, expected[j].title; got != want {
			t.Errorf("addedItems[%d].title=%s, want=%s", j, got, want)
		}
	}
}

type hooks struct {
	newArticles []newArticle
}

type newArticle struct {
	feedTitle string
	itemTitle string
}

func (h *hooks) NewArticle(feed *gofeed.Feed, item *gofeed.Item) {
	h.newArticles = append(h.newArticles, newArticle{feedTitle: feed.Title, itemTitle: item.Title})
}

func (h *hooks) assertNewArticles(t *testing.T, expected []newArticle) {
	if got, want := len(h.newArticles), len(expected); got != want {
		t.Errorf("len(newArticles)=%d, want=%d", got, want)
	}
	for i := range len(expected) {
		if got, want := h.newArticles[i].feedTitle, expected[i].feedTitle; got != want {
			t.Errorf("newArticles[%d].feedTitle=%s, want=%s", i, got, want)
		}
		if got, want := h.newArticles[i].itemTitle, expected[i].itemTitle; got != want {
			t.Errorf("newArticles[%d].itemTitle=%s, want=%s", i, got, want)
		}
	}
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

func assertNewStateItems(t *testing.T, s *state.State, expected []string) {
	if got, want := len(s.NewItems), len(expected); got != want {
		t.Errorf("len(newItems)=%d, want=%d", got, want)
	}
	for i := range len(expected) {
		if got, want := s.NewItems[i], expected[i]; got != want {
			t.Errorf("newItems[%d]=%s, want=%s", i, got, want)
		}
	}
}

func TestSuccess(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Title: "Feed 1 Title",
			Link:  "http://example.com",
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
	testHooks := &hooks{}
	processor := New(testParser, testInstapaper, testHooks, testState, false)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	testInstapaper.assertAddedItems(t, []addedItem{
		{"http://example.com/1", "Article 1"},
		{"http://example.com/2", "Article 2"},
		{"http://example.com/3", "Article 3"},
	})
	testHooks.assertNewArticles(t, []newArticle{
		{"Feed 1 Title", "Article 1"},
		{"Feed 1 Title", "Article 2"},
		{"Feed 1 Title", "Article 3"},
	})
	assertNewStateItems(t, testState, []string{
		"http://example.com/1",
		"http://example.com/2",
		"http://example.com/3",
	})
}

func TestSkipProcessed(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Title: "Feed 1 Title",
			Link:  "http://example.com",
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
	testHooks := &hooks{}
	processor := New(testParser, testInstapaper, testHooks, testState, false)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	testInstapaper.assertAddedItems(t, []addedItem{
		{"http://example.com/2", "New article"},
	})
	testHooks.assertNewArticles(t, []newArticle{
		{"Feed 1 Title", "New article"},
	})
	assertNewStateItems(t, testState, []string{
		"http://example.com/2",
	})
}

func TestParserError(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Title: "Feed 1 Title",
			Link:  "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Article 1"},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{}
	testState := state.EmptyWithPath("test")
	testHooks := &hooks{}
	processor := New(testParser, testInstapaper, testHooks, testState, false)

	err := processor.ProcessFeeds([]string{"http://example.com", "http://error.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	testInstapaper.assertAddedItems(t, []addedItem{
		{"http://example.com/1", "Article 1"},
	})
	testHooks.assertNewArticles(t, []newArticle{
		{"Feed 1 Title", "Article 1"},
	})
	assertNewStateItems(t, testState, []string{
		"http://example.com/1",
	})
}

func TestInstapaperError(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Title: "Feed 1 Title",
			Link:  "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Article 1"},
				{Link: "http://example.com/2", Title: "Article 2"},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{err: map[string]error{"http://example.com/1": errors.New("API error")}}
	testState := state.EmptyWithPath("test")
	testHooks := &hooks{}
	processor := New(testParser, testInstapaper, testHooks, testState, false)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	testInstapaper.assertAddedItems(t, []addedItem{
		{"http://example.com/2", "Article 2"},
	})
	testHooks.assertNewArticles(t, []newArticle{
		{"Feed 1 Title", "Article 2"},
	})
	assertNewStateItems(t, testState, []string{
		"http://example.com/2",
	})
}

func TestDryRun(t *testing.T) {
	feeds := []*gofeed.Feed{
		{
			Title: "Feed 1 Title",
			Link:  "http://example.com",
			Items: []*gofeed.Item{
				{Link: "http://example.com/1", Title: "Article 1"},
				{Link: "http://example.com/2", Title: "Article 2"},
			},
		},
	}
	testParser := createParser(feeds)
	testInstapaper := &instapaper{}
	testState := state.EmptyWithPath("test")
	testState.MarkProcessed("http://example.com/1")
	testHooks := &hooks{}
	processor := New(testParser, testInstapaper, testHooks, testState, true)

	err := processor.ProcessFeeds([]string{"http://example.com"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	testInstapaper.assertAddedItems(t, []addedItem{})
	testHooks.assertNewArticles(t, []newArticle{})
	assertNewStateItems(t, testState, []string{})
	if testState.MarkProcessed("http://example.com/2") {
		t.Errorf("Expected 'http://example.com/2' to have been marked as processed")
	}
}
