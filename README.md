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

- `event_type`
- `location`
- `external_url`
- `rsvp`
- `ICSDescription`
