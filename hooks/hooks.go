package hooks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/kupospelov/feeds-to-instapaper/config"
	"github.com/mmcdole/gofeed"
)

var (
	ErrNoSpawnCommand  = errors.New("no spawn command configured")
	ErrInvalidTemplate = errors.New("failed to parse template")
)

type Spawn struct {
	command      string
	argTemplates []*template.Template
}

type Hooks struct {
	newArticle []Spawn
}

func New(hooks config.Hooks) (*Hooks, error) {
	var h Hooks
	for _, hook := range hooks.NewArticle {
		if len(hook.Spawn) == 0 {
			return nil, fmt.Errorf("%w: new_article", ErrNoSpawnCommand)
		}

		spawn := Spawn{
			command: hook.Spawn[0],
		}
		for _, argTemplate := range hook.Spawn[1:] {
			t, err := template.New("arg").Parse(argTemplate)
			if err != nil {
				return nil, fmt.Errorf("%w %s: %v", ErrInvalidTemplate, argTemplate, err)
			}
			spawn.argTemplates = append(spawn.argTemplates, t)
		}

		h.newArticle = append(h.newArticle, spawn)
	}
	return &h, nil
}

func (h *Hooks) NewArticle(feed *gofeed.Feed, item *gofeed.Item) {
	type FeedTemplate struct {
		Title string
		Link  string
	}
	type ItemTemplate struct {
		Title string
		Feed  FeedTemplate
	}
	templateData := ItemTemplate{
		Title: item.Title,
		Feed: FeedTemplate{
			Title: feed.Title,
			Link:  strings.TrimSuffix(feed.Link, "/")},
	}
	for _, hook := range h.newArticle {
		args := make([]string, len(hook.argTemplates))
		for i, t := range hook.argTemplates {
			var buf bytes.Buffer
			err := t.Execute(&buf, templateData)
			if err != nil {
				log.Printf("failed to execute template: %v", err)
				continue
			}
			args[i] = buf.String()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, hook.command, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("error running hook command %s: %v.\n%s", hook.command, err, out)
		}
	}
}
