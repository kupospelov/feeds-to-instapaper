GO ?= go
SCDOC ?= scdoc

all: feeds-to-instapaper doc/feeds-to-instapaper.1

feeds-to-instapaper:
	$(GO) build

doc/feeds-to-instapaper.1: doc/feeds-to-instapaper.1.scd
	$(SCDOC) <doc/feeds-to-instapaper.1.scd >doc/feeds-to-instapaper.1

.PHONY: feeds-to-instapaper
