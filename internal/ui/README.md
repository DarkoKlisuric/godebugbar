# ui

Embedded browser assets.

- `assets/render.js`: shared widget renderer (`window.GoDebugBarRender`) for metrics, key/value, table, code and timeline widgets, plus cURL export and EXPLAIN calls.
- `assets/toolbar.js`: the Shadow DOM bottom toolbar. It tracks the current page (SPA navigation included) and shows numbers for that page only: requests, server time, queries, errors, logs and outgoing calls, plus a page timeline and request list. They reset on navigation. Served together with the renderer as `/_debugbar/assets/toolbar.js`.
- `assets/dashboard.html`: the full-page profiler at `/_debugbar`, with server-side search (text, method, status, minimum duration, problem flags) kept in the URL.

The assets are plain ES2020 with no build step.
