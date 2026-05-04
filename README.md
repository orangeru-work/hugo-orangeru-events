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

## Supported front matter

Required:

- `startdate` or `startDate`

Optional:

- `enddate` or `endDate`
- `category` (preferred)
- `event_type` (legacy string fallback)
- `location`
- `external_url`
- `rsvp`
- `calendar_description` or `ics_description` (`ICSDescription` also supported as legacy fallback)

If `category` is missing, it defaults to `other`.

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

- On iPhone/iPad, it opens `webcal://...`
- On Android, it opens Google Calendar subscribe flow with the same ICS feed URL
- On desktop, it opens a subscribe menu (Google Calendar, Outlook.com, Apple Calendar/webcal, direct ICS download)

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
