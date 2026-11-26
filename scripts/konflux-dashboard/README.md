# Konflux dashboard (mock)

Static HTML that loads `mock.json` next to it. Browsers block `fetch()` for `file://`, so run a small HTTP server from this directory.

```bash
cd scripts/konflux-dashboard
python3 -m http.server 8765
```

Open [http://localhost:8765/](http://localhost:8765/). Edit `mock.json` and refresh to try different data.

Without a server, `index.html` falls back to embedded JSON in the page when opened as a file.
