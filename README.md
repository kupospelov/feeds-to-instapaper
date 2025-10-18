# feeds-to-instapaper

An application that checks RSS, Atom, or JSON feeds and adds new articles to Instapaper.

## Installation

```bash
go install github.com/kupospelov/feeds-to-instapaper@latest
```

## Configuration

Create a configuration file at `~/.config/feeds-to-instapaper/config.toml`:

```toml
[instapaper]
username = "your-instapaper-username"
password = "your-instapaper-password"

[feeds]
urls = [
    "https://example.com/feed.xml",
    "https://another-site.com/atom",
]
```

## Building

Dependencies:
* Go
* make
* scdoc (optional, for man pages)

Run `make`.

## Usage

Run `feeds-to-instapaper`.

You can use cron or systemd timers to schedule the runs. Check out the [examples](https://github.com/kupospelov/feeds-to-instapaper/tree/main/doc/systemd).

The application writes a state file at `~/.local/state/feeds-to-instapaper/added` to keep track of previously processed articles.
