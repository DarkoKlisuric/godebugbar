# ui

Embedded browser assets.

- `assets/render.js`: shared widget renderer (`window.GoDebugBarRender`) for metrics, key/value, table, code and timeline widgets, plus cURL export and EXPLAIN calls.
- `assets/toolbar.js`: the Shadow DOM bottom toolbar. Served together with the renderer as `/_debugbar/assets/toolbar.js`.
- `assets/dashboard.html`: the full-page profiler at `/_debugbar`.

The assets are plain ES2020 with no build step.
