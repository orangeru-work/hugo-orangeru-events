# hugo-orangeru-events

Reusable Hugo module for interactive event calendars across orangeru sites.

## Included partials

- `partials/events/interactive-calendar.html`

## Usage

Add the module import in your site's `hugo.yaml`:

```yaml
module:
  imports:
    - path: github.com/orangeru-work/hugo-orangeru-events
```

Then render on an events list page template:

```go-html-template
{{ partial "events/interactive-calendar.html" . }}
```

This partial supports event front matter using either:

- `startdate` / `enddate`
- `startDate` / `endDate`

Optional fields:

- `category` (preferred)
- `event_type` (fallback for older content)
- `patr` (legacy fallback: `true` => `club`, `false` => `other`)
- `location`
- `external_url`
- `rsvp`
- `ICSDescription`

## Category + legend behavior

- Events are grouped and filtered by `category`
- The calendar renders a color-coded legend with category counts
- Clicking legend chips filters the calendar
- If no category is set, events default to `other`

## Mobile subscribe button

- The calendar includes a **Subscribe to calendar** button for the section feed at `events/index.ics`
- On iPhone/iPad, it opens the feed with the `webcal://` scheme
- On Android, it opens Google Calendar subscribe flow with the same ICS feed URL
