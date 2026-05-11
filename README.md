# hugo-orangeru-events

Reusable Hugo module for an interactive events calendar partial.

## What this module provides

- `layouts/partials/events/interactive-calendar.html`
- FullCalendar month/week/list views
- Category legend with filtering
- Event detail preview panel
- Optional mobile-aware subscribe button for an ICS feed

## Quick start

1. Add the module import in your site's `hugo.yaml`:

```yaml
module:
  imports:
    - path: github.com/orangeru-work/hugo-orangeru-events
```

2. Render it from your events list template:

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

Generated Facebook event pages also include:

- `generated_by`
- `generated_at`

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

## Mobile subscribe behavior

When subscribe is enabled:

- It opens a subscribe options menu across platforms
- The menu includes Google Calendar, Apple/webcal, Outlook.com, and copy-to-clipboard for the feed URL
- Google Calendar uses a URL-encoded `webcal://` feed in `cid` for compatibility across clients

## Default month-view display

- In month view, event bullets and inline times are hidden by default
- Event titles are emphasized and allowed to wrap for readability

## External link rendering

- If `ICSDescription` includes the same URL as `external_url`, the duplicate raw URL is removed from the details description
- The event details panel keeps a single labeled link: **External event page**

## exampleSite

A minimal runnable Hugo site is included under `exampleSite/`.

Run it locally:

```bash
cd exampleSite
hugo server
```
