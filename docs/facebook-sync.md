# Facebook event sync

This guide explains how to sync Facebook events into Hugo content using this
module.

## User overview

From a user point of view, the sync does this:

1. Downloads your Facebook ICS feed.
2. Generates/updates markdown files in your events content directory.
3. Applies category mapping for going/interested/default responses.
4. Removes previously generated files before writing fresh output.
5. (exampleSite workflow only) deletes generated events older than 30 days.

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

- Trigger: manual (`workflow_dispatch`) and daily schedule
- Trigger: manual (`workflow_dispatch`) and weekly schedule
- Output: `exampleSite/content/events`
- Pruning: removes generated events older than 30 days

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
  --include-private=false \
  --category-going club \
  --category-interested other \
  --category-default other
```

### Key options

- `--category-going`: maps `PARTSTAT:ACCEPTED`
- `--category-interested`: maps `PARTSTAT:TENTATIVE`
- `--category-default`: fallback for other statuses
- `--include-private`: include `CLASS:PRIVATE` events when set to true
- `--exclude-organizer`: repeatable organizer exclusion filter
- `--cancelled`: optional file of event IDs to suppress
