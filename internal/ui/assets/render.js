window.GoDebugBarRender = window.GoDebugBarRender || (() => {
  'use strict';

  const css = `
    .gdb-widget-title { font-size: 11px; font-weight: 700; color: var(--gdb-muted); text-transform: uppercase; letter-spacing: .04em; margin: 16px 0 8px; display: flex; align-items: center; gap: 8px; }
    .gdb-metrics { display: flex; gap: 10px; flex-wrap: wrap; margin-bottom: 12px; }
    .gdb-metric { background: var(--gdb-panel); border: 1px solid var(--gdb-border); border-radius: 6px; padding: 8px 12px; min-width: 120px; }
    .gdb-metric-label { color: var(--gdb-muted); font-size: 11px; }
    .gdb-metric-value { font-size: 15px; font-weight: 700; margin-top: 2px; }
    .gdb-table { width: 100%; border-collapse: collapse; margin-bottom: 12px; background: var(--gdb-panel); border: 1px solid var(--gdb-border); }
    .gdb-table th, .gdb-table td { text-align: left; padding: 6px 8px; border-bottom: 1px solid var(--gdb-border); vertical-align: top; word-break: break-word; }
    .gdb-table th { color: var(--gdb-muted); font-weight: 600; background: var(--gdb-head); position: sticky; top: 0; }
    .gdb-table th.gdb-sortable { cursor: pointer; user-select: none; }
    .gdb-table th.gdb-sortable:hover { color: var(--gdb-text); }
    .gdb-kv th { width: 220px; }
    .gdb-table code { white-space: pre-wrap; font-family: inherit; }
    .gdb-row-red td { background: var(--gdb-red-bg); }
    .gdb-row-red td:first-child { box-shadow: inset 3px 0 var(--gdb-red); }
    .gdb-row-yellow td { background: var(--gdb-yellow-bg); }
    .gdb-row-yellow td:first-child { box-shadow: inset 3px 0 var(--gdb-yellow); }
    .gdb-row-green td:first-child { box-shadow: inset 3px 0 var(--gdb-green); }
    .gdb-code { position: relative; background: var(--gdb-panel); border: 1px solid var(--gdb-border); border-radius: 4px; padding: 10px 12px; overflow: auto; max-height: 420px; margin: 0 0 12px; white-space: pre; }
    .gdb-btn { background: var(--gdb-panel); color: var(--gdb-text); border: 1px solid var(--gdb-border); border-radius: 4px; padding: 2px 8px; font: inherit; font-size: 11px; cursor: pointer; }
    .gdb-btn:hover, .gdb-btn:focus-visible { border-color: var(--gdb-accent); outline: none; }
    .gdb-filter { background: var(--gdb-panel); color: var(--gdb-text); border: 1px solid var(--gdb-border); border-radius: 4px; padding: 2px 6px; font: inherit; font-size: 11px; text-transform: none; letter-spacing: 0; font-weight: 400; }
    .gdb-plan td { background: var(--gdb-head); }
    .gdb-timeline { display: grid; grid-template-columns: minmax(160px, 32%) 1fr; gap: 3px 10px; align-items: center; margin-bottom: 12px; }
    .gdb-tl-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--gdb-muted); }
    .gdb-tl-track { position: relative; overflow: hidden; height: 14px; background: var(--gdb-panel); border-radius: 3px; }
    .gdb-tl-bar { position: absolute; top: 2px; bottom: 2px; min-width: 2px; border-radius: 2px; background: var(--gdb-accent); }
    .gdb-tl-request { background: var(--gdb-muted); }
    .gdb-tl-query { background: #a78bfa; }
    .gdb-tl-http { background: #f472b6; }
    .gdb-tl-span { background: var(--gdb-accent); }
    .gdb-tl-red { background: var(--gdb-red); }
    .gdb-tl-yellow { background: var(--gdb-yellow); }
    .gdb-empty { color: var(--gdb-muted); padding: 24px; text-align: center; }
  `;

  const theme = `
    --gdb-bg: #0b1120; --gdb-panel: #0f172a; --gdb-head: #111a2e; --gdb-border: #1e293b; --gdb-text: #e2e8f0;
    --gdb-muted: #94a3b8; --gdb-accent: #38bdf8; --gdb-green: #22c55e; --gdb-yellow: #eab308; --gdb-red: #ef4444;
    --gdb-red-bg: rgba(239, 68, 68, .12); --gdb-yellow-bg: rgba(234, 179, 8, .10);
  `;

  const h = (tag, props = {}, ...children) => {
    const el = document.createElement(tag);
    const { className, text, attrs, on, dataset } = props;
    if (className) el.className = className;
    if (text !== undefined && text !== null) el.textContent = String(text);
    Object.entries(attrs || {}).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== false) el.setAttribute(key, value === true ? '' : String(value));
    });
    Object.entries(dataset || {}).forEach(([key, value]) => { el.dataset[key] = value; });
    Object.entries(on || {}).forEach(([event, handler]) => el.addEventListener(event, handler));
    children.flat(Infinity).forEach((child) => {
      if (child === null || child === undefined || child === false) return;
      el.appendChild(child instanceof Node ? child : document.createTextNode(String(child)));
    });
    return el;
  };

  const tooltip = (item) => Object.entries((item && item.tooltip) || {})
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n');

  const statusClass = (code) => {
    if (code >= 500) return 's5';
    if (code >= 400) return 's4';
    if (code >= 300) return 's3';
    if (code >= 200) return 's2';
    return 's1';
  };

  const formatMs = (ms) => `${(Number(ms) || 0).toFixed(2)} ms`;

  const numericValue = (text) => {
    const match = /^\s*(-?\d+(?:\.\d+)?)\s*(?:ms|MB|B|%)?\s*$/.exec(String(text));
    return match ? parseFloat(match[1]) : null;
  };

  const copyText = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      return false;
    }
  };

  const flash = (button, label) => {
    const original = button.textContent;
    button.textContent = label;
    setTimeout(() => { button.textContent = original; }, 1200);
  };

  const renderMetrics = (widget) => h('div', { className: 'gdb-metrics' },
    (widget.metrics || []).map((m) => h('div', { className: 'gdb-metric' },
      h('div', { className: 'gdb-metric-label', text: m.label }),
      h('div', { className: 'gdb-metric-value', text: m.value }))));

  const renderKeyValue = (widget) => h('table', { className: 'gdb-table gdb-kv' },
    h('tbody', {}, Object.keys(widget.pairs || {}).sort().map((key) => h('tr', {},
      h('th', { text: key }),
      h('td', {}, h('code', { text: widget.pairs[key] }))))));

  const sortRows = (rows, index, direction) => rows.slice().sort((a, b) => {
    const left = a.cells[index];
    const right = b.cells[index];
    const ln = numericValue(left);
    const rn = numericValue(right);
    const cmp = ln !== null && rn !== null ? ln - rn : String(left).localeCompare(String(right));
    return cmp * direction;
  });

  const renderTable = (widget, options) => {
    const rows = (widget.rows || []).map((cells, i) => ({ cells, color: (widget.row_colors || [])[i] || '' }));
    const explain = options.explainable && widget.title === 'Executed SQL Statements' && options.onExplain;
    const columns = widget.columns || [];
    const tbody = h('tbody');
    const table = h('table', { className: 'gdb-table' });
    let sort = { index: -1, direction: 1 };
    let filter = '';

    const draw = () => {
      tbody.textContent = '';
      let list = rows;
      if (filter) list = list.filter((row) => row.cells.some((c) => String(c).toLowerCase().includes(filter)));
      if (sort.index >= 0) list = sortRows(list, sort.index, sort.direction);
      list.forEach((row) => {
        const tr = h('tr', { className: row.color && row.color !== 'default' ? `gdb-row-${row.color}` : '' },
          row.cells.map((cell) => h('td', {}, String(cell).length > 40 ? h('code', { text: cell }) : String(cell))));
        if (explain) {
          const button = h('button', { className: 'gdb-btn', text: 'Explain', attrs: { type: 'button' } });
          button.addEventListener('click', () => {
            const existing = tr.nextElementSibling;
            if (existing && existing.classList.contains('gdb-plan')) { existing.remove(); return; }
            button.disabled = true;
            options.onExplain(row.cells[0])
              .then((plan) => tr.after(h('tr', { className: 'gdb-plan' },
                h('td', { attrs: { colspan: columns.length + 1 } }, h('pre', { className: 'gdb-code', text: plan })))))
              .finally(() => { button.disabled = false; });
          });
          tr.appendChild(h('td', {}, button));
        }
        tbody.appendChild(tr);
      });
    };

    if (columns.length) {
      const head = h('tr', {}, columns.map((column, index) => h('th', {
        className: 'gdb-sortable',
        text: column,
        attrs: { title: 'Sort', scope: 'col' },
        on: {
          click: () => {
            sort = { index, direction: sort.index === index ? -sort.direction : 1 };
            draw();
          },
        },
      })));
      if (explain) head.appendChild(h('th', { text: '' }));
      table.appendChild(h('thead', {}, head));
    }
    table.appendChild(tbody);
    draw();

    const wrap = h('div', {}, table);
    if (rows.length > 8) {
      const input = h('input', { className: 'gdb-filter', attrs: { type: 'search', placeholder: 'Filter rows…', 'aria-label': 'Filter rows' } });
      input.addEventListener('input', () => { filter = input.value.toLowerCase(); draw(); });
      wrap.dataset.filter = 'true';
      wrap.prepend(input);
    }
    return wrap;
  };

  const renderCode = (widget) => {
    const pre = h('pre', { className: 'gdb-code', text: widget.code });
    const button = h('button', { className: 'gdb-btn', text: 'Copy', attrs: { type: 'button' } });
    button.addEventListener('click', () => copyText(widget.code).then((ok) => flash(button, ok ? 'Copied' : 'Copy failed')));
    return h('div', {}, h('div', { className: 'gdb-widget-title' }, button), pre);
  };

  const renderTimeline = (widget) => {
    const events = widget.events || [];
    const total = Math.max(1, ...events.map((e) => (e.start_ms || 0) + (e.duration_ms || 0)));
    return h('div', { className: 'gdb-timeline' }, events.map((e) => {
      const left = Math.min(Math.max(((e.start_ms || 0) / total) * 100, 0), 99.7);
      const width = Math.min(Math.max(((e.duration_ms || 0) / total) * 100, 0.3), 100 - left);
      const colorClass = e.color === 'red' || e.color === 'yellow' ? ` gdb-tl-${e.color}` : '';
      const label = `${e.label} · ${formatMs(e.duration_ms)} at ${formatMs(e.start_ms)}`;
      return [
        h('div', { className: 'gdb-tl-label', text: e.label, attrs: { title: label } }),
        h('div', { className: 'gdb-tl-track', attrs: { title: label } },
          h('div', { className: `gdb-tl-bar gdb-tl-${e.category}${colorClass}`, attrs: { style: `left:${left}%;width:${width}%` } })),
      ];
    }));
  };

  const renderers = {
    metrics: renderMetrics,
    key_value: renderKeyValue,
    table: renderTable,
    code: renderCode,
    timeline: renderTimeline,
  };

  const renderWidgets = (container, widgets, options = {}) => {
    container.textContent = '';
    if (!widgets || !widgets.length) {
      container.appendChild(h('div', { className: 'gdb-empty', text: 'This section has no details.' }));
      return;
    }
    widgets.forEach((widget) => {
      const render = renderers[widget.kind];
      if (!render) return;
      if (widget.title && widget.kind !== 'metrics') {
        container.appendChild(h('div', { className: 'gdb-widget-title', text: widget.title }));
      }
      container.appendChild(render(widget, options));
    });
  };

  const findWidget = (profile, sectionName, title) => {
    const section = (profile.sections || []).find((s) => s.name === sectionName);
    const widgets = (section && section.panel && section.panel.widgets) || [];
    return widgets.find((w) => w.title === title);
  };

  const shellQuote = (value) => `'${String(value).replace(/'/g, `'\\''`)}'`;

  const curl = (profile, origin) => {
    const headers = (findWidget(profile, 'request', 'Request Headers') || {}).pairs || {};
    const query = (findWidget(profile, 'request', 'Query Parameters') || {}).pairs || {};
    const body = (findWidget(profile, 'request', 'Request Body') || {}).code;
    const params = new URLSearchParams();
    Object.keys(query).sort().forEach((key) => params.append(key, query[key]));
    const qs = params.toString();
    const parts = [`curl -X ${profile.method} ${shellQuote(`${origin}${profile.path}${qs ? `?${qs}` : ''}`)}`];
    Object.keys(headers).sort()
      .filter((key) => !['Content-Length', 'Connection', 'Accept-Encoding'].includes(key))
      .forEach((key) => parts.push(`-H ${shellQuote(`${key}: ${headers[key]}`)}`));
    if (body) parts.push(`--data-raw ${shellQuote(body)}`);
    return parts.join(' \\\n  ');
  };

  const fetchJSON = async (url) => {
    const res = await fetch(url, { credentials: 'include', headers: { Accept: 'application/json' } });
    const data = await res.json().catch(() => null);
    if (!res.ok) throw new Error((data && data.error) || `HTTP ${res.status}`);
    return data;
  };

  const explainer = (baseUrl, profileId, fetcher = fetchJSON) => (index) => fetcher(`${baseUrl}/api/profiles/${encodeURIComponent(profileId)}/queries/${encodeURIComponent(index)}/explain`)
    .then((data) => (data.plan || []).join('\n'))
    .catch((err) => `EXPLAIN failed: ${err.message}`);

  return { css, theme, h, tooltip, statusClass, formatMs, renderWidgets, curl, copyText, flash, fetchJSON, explainer };
})();
