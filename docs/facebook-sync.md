# Facebook event sync

This guide explains how to sync Facebook events into Hugo content using this
module.

## User overview

From a user point of view, the sync does this:

1. Downloads your Facebook ICS feed.
2. Generates/updates markdown files in your events content directory.
3. Applies category mapping for going/interested/default responses.
4. Skips events marked `PARTSTAT:DECLINED`.
5. Writes event-specific values under an `event:` front matter block.
6. Adds flat generation markers: `generated_by` and `generated_at`.
7. Converts supported GPS coordinate text to Google Maps links.
8. Leaves feature images unset (no generated `feature-img` field).
9. Keeps previously generated files by default.
10. Preserves unchanged generated files instead of rewriting them just to refresh `generated_at`.
11. Can optionally delete generated events older than a configured number of days.
12. (exampleSite workflow only) opts into deleting generated events older than 30 days.

The module expects event metadata in the `event` block (for example,
`event.startdate`, `event.category`, `event.location`).

## 1. Get your Facebook ICS feed URL

Use Facebook's calendar export flow while signed in:

1. Open Facebook Events calendar/export page in your browser.
2. Find the **Upcoming Events** iCalendar link.
3. Copy the full URL.

It should look like:

`https://www.facebook.com/events/ical/upcoming/?uid=<uid>&key=<key>`

## 2. Create the repository secret

In the repository where the workflow runs, create:

- **Secret name:** `FACEBOOK_EVENTS_ICS_URL`
- **Secret value:** the full ICS URL from step 1

The workflows also support a repository variable with the same name, but secret
is preferred.

## 3. Choose how to run sync

### A) Sync this module's example site

Use `.github/workflows/example-site-events.yml` in this repo.

- Trigger: manual (`workflow_dispatch`) and weekly schedule
- Output: `exampleSite/content/events`
- Pruning: opts into removing generated events older than 30 days

### B) Sync a consuming site repository

Use `contrib/workflows/facebook-events-sync.yml` as a template.

1. Copy it into your site repo at `.github/workflows/facebook-events-sync.yml`
2. Keep manual-only, or uncomment `push`/`schedule` to enable automatic runs
3. Ensure `FACEBOOK_EVENTS_ICS_URL` is configured in that repo

## CLI usage

```bash
go run ./cmd/facebook-events \
  --ics /tmp/events.ics \
  --output site/content/events \
  --cancelled cancelled-events.txt \
  --delete-generated-older-than-days 30 \
  --include-private=false \
  --category-going club \
  --category-interested other \
  --category-default other
```

### Key options

- `--category-going`: maps `PARTSTAT:ACCEPTED`
- `--category-interested`: maps `PARTSTAT:TENTATIVE`
- `--category-default`: fallback for non-declined statuses not matched above
- `--include-private`: include `CLASS:PRIVATE` events when set to true
- `--exclude-organizer`: repeatable organizer exclusion filter
- `--cancelled`: optional file of event IDs to suppress
- `--delete-generated-older-than-days`: optional cleanup window for generated files; `0` keeps generated files indefinitely

By default, the sync does **not** delete previously generated event markdown files.
Set `--delete-generated-older-than-days` only if you want age-based pruning.
When pruning is enabled, the module only deletes files marked with
`generated_by: facebook-events` and a parseable `generated_at` timestamp.

For `v1.x`, the sync treats the generated markers, CLI flags, and documented
default pruning behavior as stable public contract. Any breaking change to that
surface should require a new major version.

## GPS coordinate handling

The sync normalizes common coordinate formats found in Facebook descriptions and
turns them into Google Maps links.

Supported patterns:

- `34.545° N, 112.468° W`
- `34.545, -112.468`

Behavior:

- In generated body content, coordinates are converted to clickable markdown
  links.
- In `event.ics_description`, coordinates are preserved as plain text plus a
  Google Maps URL.
- `event.location` is copied from the ICS feed as-is.
