# hugo-orangeru-events

Reusable Hugo module for interactive event calendars, event detail pages, and ICS feeds.

## What this module provides

- `layouts/partials/events/interactive-calendar.html`
- `layouts/events/list.html`
- `layouts/events/single.html`
- `layouts/events/list.calendar.ics`
- `layouts/events/single.calendar.ics`
- FullCalendar month/week/list views
- Category legend with filtering
- Hover/focus event preview popover with click-through navigation
- Optional mobile-aware subscribe button for an ICS feed

## Quick start

1. Add the module import in your site's `hugo.yaml`:

```yaml
module:
  imports:
    - path: github.com/orangeru-work/hugo-orangeru-events
```

2. Add the calendar output format in your site config:

```yaml
outputFormats:
  Calendar:
    protocol: "https://"
```

3. Either use an `events` content section, or set `type: events` on any section that should use the shared templates:

```yaml
---
title: Club Calendar
type: events
outputs:
  - html
  - calendar
cascade:
  type: events
  outputs:
    - html
    - calendar
---
```

4. If you want a custom list layout, you can still render the shared calendar partial directly:

```go-html-template
{{ partial "events/interactive-calendar.html" . }}
```

## Facebook ICS event generation (optional)

Want your events page to stay fresh without hand-editing markdown?  
This module includes a Facebook ICS sync flow that can generate event pages,
apply sensible category mapping, auto-link supported GPS coordinates, and run on
a schedule when you decide to turn it on.

For setup, secrets, feed URL instructions, and workflow options, see:
**[Facebook event sync guide](docs/facebook-sync.md)**.

## Supported front matter

Contained under `event`:

- `event.startdate` (required)
- `event.enddate`
- `event.category`
- `event.event_type` (legacy fallback when category is missing)
- `event.location`
- `event.external_url`
- `event.rsvp`
- `event.calendar_description` / `event.ics_description`

If `category` is missing, it defaults to `other`.

Nested event content is supported. Events can live directly under the section or
inside subfolders such as `content/events/2026/05/...`; the shared calendar and
ICS templates collect them recursively.

Generated Facebook event pages also include:

- `generated_by`
- `generated_at`

## Compatibility

- **Go / generator CLI:** developed and tested with Go `1.24.x`
- **Hugo sites:** validated in consuming sites running Hugo `0.147.x` and `0.148.x`
- **Hugo Modules:** required
- **Hugo extended:** not required by this module itself

If you use the Facebook sync workflow, match the workflow `go-version` to a
Go `1.24.x` runtime.

## v1 stability promise

For `v1.x`, the following are treated as stable public surface area:

- the partial path `layouts/partials/events/interactive-calendar.html`
- the `event.*` front matter contract used by the interactive calendar
- generated file markers `generated_by` and `generated_at`
- the Facebook sync CLI flags and their documented defaults
- the default pruning behavior (`--delete-generated-older-than-days=0` keeps generated files)

Changes to those contracts should require a new major version.

## Configuration (`params.eventsCalendar`)

Set these in `site.Params.eventsCalendar` (or page front matter `eventsCalendar` to override per page):

```yaml
params:
  eventsCalendar:
    showSubscribeButton: true
    subscribeFeedURL: "https://example.org/events/index.ics"
    categoryColors:
      race: "#fecaca"
      social: "#bfdbfe"
      training: "#bbf7d0"
```

Options:

- `showSubscribeButton` (bool, default `true`)
- `subscribeFeedURL` (string, default `<section rel permalink>index.ics`)
- `categoryColors` (map of category -> hex color)
- `calendarName` (string, default `site.Title`)
- `calendarDescription` (string, default `<calendarName> calendar feed`)
- `uidDomain` (string, default derived from `site.BaseURL`)
- `prodID` (string, default `-//<calendarName>//Calendar//EN`)

## Mobile subscribe behavior

When subscribe is enabled:

- It opens a subscribe options menu across platforms
- The menu includes Google Calendar, Apple/webcal, Outlook.com, and copy-to-clipboard for the feed URL
- Google Calendar uses a URL-encoded `webcal://` feed in `cid` for compatibility across clients

## Default month-view display

- In month view, event bullets and inline times are hidden by default
- Event titles are emphasized and allowed to wrap for readability

## External link rendering

- If `ICSDescription` includes the same URL as `external_url`, the duplicate raw URL is removed from the hover preview description
- Clicking a calendar event opens its event page, where the external link remains available

## exampleSite

A minimal runnable Hugo site is included under `exampleSite/`.

Run it locally:

```bash
cd exampleSite
hugo server
```
