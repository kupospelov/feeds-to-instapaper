feeds-to-instapaper(1)

# NAME

feeds-to-instapaper - upload feeds to Instapaper

# SYNOPSIS

*feeds-to-instapaper* [options...]

# DESCRIPTION

feeds-to-instapaper is a command-line tool that reads a list of RSS/Atom
feeds, finds new articles, and uploads them to Instapaper.

It keeps track of processed articles to avoid duplicates. The state is stored
in "$XDG_STATE_HOME/feeds-to-instapaper/added" (or
"~/.local/state/feeds-to-instapaper/added" if unset).

# OPTIONS

*-dry-run*
	Only fetch and print new feeds.

# CONFIGURATION

The configuration file is located at
"$XDG_CONFIG_HOME/feeds-to-instapaper/config.toml" (or
"~/.config/feeds-to-instapaper/config.toml" if unset).

Example:

```
[instapaper]
username = "your-username"
password = "your-password"

[feeds]
urls = [
  "https://example.com/feed.xml",
  "https://another.org/rss.xml"
]

[[hooks.new_article]]
spawn = ["notify-send", "New article on {{.Feed.Title}}", "{{.Title}}"]
```

The following sections are supported:

*[instapaper]*
	Instapaper credentials.

	*username* <string>
		Your Instapaper username. This is required.

	*password* <string>
		Your Instapaper password. This is required.

*[feeds]*
	Feeds to process.

	*urls* <array of strings>
		A list of RSS/Atom feed URLs to process.

*[hooks]*
	Hooks to run on events.

	*new_article* <array of tables>
		A list of hooks to run when a new article is found. Each hook is a
		table with a _spawn_ key.

		*spawn* <array of strings>
			A command and its arguments to execute. The arguments are
			interpreted as Go templates. See *TEMPLATES* for details.

# TEMPLATES

Hooks can use Go templates in their arguments. For full documentation on the
template syntax, see <https://pkg.go.dev/text/template>.

The following data is available in _new_article_ hooks:

*.Title*
	The article title.

*.Feed.Title*
	The feed title of the article.

*.Feed.Link*
	The feed link of the article.

# AUTHORS

Konstantin Pospelov <kupospelov@gmail.com>
