# UI Bakery source export

## What's inside

- [Export](#Export) - what it should look like
- [Snapshot](#Snapshot) - what's available
- [Download](#Download) - what we have

## Intro

Modules, developed with UI Bakery, don't follow any best (ar any at all) 
practices of modern software development. Fair enough, UI Bakery is a "no 
code" solution. Still, some kind of control of software development would
be welcomed.

UI Bakery provides Git integration, but only with a higher subscription tier.
Without that, we have exports and snapshots.

## Export

UI Bakery provides export of the current project, which comes in the form of
a downloadable ZIP archive. Information inside is organized by the application
pages (with hirearchy where applicable). Items not specific to any page are
exported on the root level.

Exports must be created manually, by a user. However, produced ZIP archive can
be readily uploaded into another instance of UI Bakery (providing that versions
don't differ too much).

## Snapshot

Snapshots are available in the table `project_snapshot`. Use the similar query to
limit the scope:

```sql
select
    id
    , model
from
    project_snapshot
where
    json_extract(model, '.metadata.title') = '"Lightburn Vendor Compatibility"'
```

And note that `json_extract() will return the JSON double quotes.

The whole project snapshot is a JSON contained in the model field. Top level elements are

- `version`
- `rootPageList`
- `slotList`
- `workflowList`
- `themeList`
- `metadata`
- `appSettings`

## Download

UIB-download takes a JSON that represents a snapshot and cuts it into some
pieces so that the changes between two snapshots remain more localized. No
attempt is made to replicate the structure of an export because it would raise
unwarranted expectations that we can import the transformed snapshot which may
not necessarily be always true.
