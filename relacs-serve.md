# relacs serve

New command to present release related information in web pages.
A local Go HTTP webserver is started and the home page is served at `http://localhost:8080/`.
The pages are rendered server-side, using data from the underlying functions of the `relacs` CLI.

Example:

```
relacs serve
```

## Home page

The page contains a widgets to go to detail pages to release phases.
For example, a widget to go to the RC phase page, where the user can enter the release (e.g. `5.0.0` and an RC number).
On click, this navigates to the RC phase page explained below.

## Tag

At `http://localhost:8080/tag?tag=<tag>`. Get here from another widget on the home page where user can enter the full tag.

### Data from:

- `relacs github get-commit-for-tag --tag 5.0.0-rc.1 --repo stackrox` (To get the GitHub link to the commit)
- `relacs konflux get-custom-snapshot --tag 5.0.0-rc.1` (To get the custom snapshot for the tag)
- `relacs konflux get-operator-bundle-image --tag 5.0.0-rc.1` (To get the operator bundle image for the tag)
- `relacs konflux get-operator-index-image --tag 5.0.0-rc.1` (To get the operator index image for the tag)
- `relacs konflux list-pipelineruns-for-tag --tag 5.0.0-rc.1` (To list the pipeline runs for the tag)
- `relacs konflux list-conforma-violations-for-tag --tag 5.0.0-rc.1` (To list the conformance violations for the tag - NEED TO ALLOW FILTERING BY CONFORMA SCENARIO)

### Widgets

* Top row:
  - Overall status
    - All Konflux pipelines passed? (i.e. custom snapshot present, operator index pipeline succeeded)
    - No blocking conforma violations?
  - Artifacts
    - Custom snapshot
    - Operator bundle image
    - Operator index image
    - Link to GitHub commit for the tag

* Rows:
  1. Pipeline runs
    - Grouped by original pipeline (e.g. "operator-index-on-push", "main-on-push")
  2. Conforma violations

## Phase: RC

At `http://localhost:8080/phase/rc?release=<release>&rc=<rc>`

### Data from:

- `relacs github get-commit-for-tag --tag 5.0.0-rc.1 --repo stackrox` (To get the GitHub link to the commit)
- `relacs konflux get-custom-snapshot --tag 5.0.0-rc.1` (To get the custom snapshot for the tag)
- `relacs konflux get-operator-bundle-image --tag 5.0.0-rc.1` (To get the operator bundle image for the tag)
- `relacs konflux get-operator-index-image --tag 5.0.0-rc.1` (To get the operator index image for the tag - NEW COMMAND)
- `relacs konflux list-pipelineruns-for-tag --tag 5.0.0-rc.1` (To list the pipeline runs for the tag)
- `relacs konflux list-conforma-violations-for-tag --tag 5.0.0-rc.1` (To list the conformance violations for the tag - NEED TO ALLOW FILTERING BY CONFORMA SCENARIO)

### Widgets

* Top row:
  - Overall status
    - All Konflux pipelines passed? (i.e. custom snapshot present, operator index pipeline succeeded)
    - No blocking conforma violations?
  - Artifacts
    - Custom snapshot
    - Operator bundle image
    - Operator index image
    - Link to GitHub commit for the tag

* Rows:
  1. Pipeline runs
    - Grouped by original pipeline (e.g. "operator-index-on-push", "main-on-push")
  2. Conforma violations

## Phase Release

At `http://localhost:8080/phase/release?release=<release>`.

### Data from:

* Same data as for RC phase, but querying just for `<release>` as the `--tag` parameter.
* Add sections with widgets for:
  - stage release: one widget for image release, one for CDN
  - prod release: one widget for image release, one for CDN
  Each widget showing name, status, link to pipelinerun, link to internal and external advisory

  - FBC releases (status), list of FBC releases (like the PipelineRun list, with shorthand status).

Add mock data into the new widgets first.
