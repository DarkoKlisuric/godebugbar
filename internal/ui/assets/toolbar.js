(() => {
  'use strict';

  const script = document.currentScript;
  const R = window.GoDebugBarRender;
  if (!script || !R || document.getElementById('godebugbar-host')) return;

  const PROFILE_HEADER = 'X-Debugbar-ID';
  const STORAGE_KEY = 'godebugbar:toolbar';
  const MAX_HISTORY = 50;
  const MAX_PAGE_REQUESTS = 200;
  const PAGE_SETTLE_MS = 1200;
  const BAR_HEIGHT = 36;
  const baseUrl = (script.dataset.baseUrl || '/_debugbar').replace(/\/+$/, '');
  const initialId = script.dataset.profileId || '';
  const origFetch = window.fetch.bind(window);
  const now = () => performance.now();

  const readState = () => {
    try {
      return { minimized: false, height: 340, ...JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}') };
    } catch {
      return { minimized: false, height: 340 };
    }
  };
  const state = readState();
  const saveState = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ minimized: state.minimized, height: state.height }));
    } catch {
      return undefined;
    }
    return undefined;
  };

  const api = async (path) => {
    const res = await origFetch(`${baseUrl}${path}`, { credentials: 'include', headers: { Accept: 'application/json' } });
    const data = await res.json().catch(() => null);
    if (!res.ok) throw new Error((data && data.error) || `HTTP ${res.status}`);
    return data;
  };

  const { h } = R;
  const host = h('div', { attrs: { id: 'godebugbar-host' } });
  const spacer = h('div', { attrs: { id: 'godebugbar-spacer', 'aria-hidden': 'true' } });
  const shadow = host.attachShadow({ mode: 'open' });

  shadow.appendChild(h('style', { text: `
    :host { all: initial; ${R.theme} font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; color: var(--gdb-text); }
    * { box-sizing: border-box; }
    button { font: inherit; color: inherit; }
    a { color: var(--gdb-accent); }
    .bar { position: fixed; left: 0; right: 0; bottom: 0; height: ${BAR_HEIGHT}px; background: var(--gdb-panel); border-top: 1px solid var(--gdb-border); display: flex; align-items: center; gap: 8px; padding: 0 8px; z-index: 2147483647; }
    .items { display: flex; align-items: center; gap: 6px; overflow-x: auto; flex: 1; scrollbar-width: none; }
    .tools { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }
    .brand { font-weight: 700; color: var(--gdb-accent); background: none; border: none; cursor: pointer; padding: 0 4px; }
    .page { background: none; border: none; cursor: pointer; padding: 0; color: var(--gdb-text); font-weight: 600; white-space: nowrap; flex-shrink: 0; max-width: 30%; overflow: hidden; text-overflow: ellipsis; }
    .badge { display: inline-flex; align-items: center; gap: 6px; padding: 3px 8px; border-radius: 4px; cursor: pointer; background: #1e293b; border: 1px solid #334155; white-space: nowrap; }
    .badge:hover, .badge:focus-visible, .badge.active { border-color: var(--gdb-accent); background: #1e3a5f; outline: none; }
    .badge.green, .tab.green { border-left: 3px solid var(--gdb-green); }
    .badge.yellow, .tab.yellow { border-left: 3px solid var(--gdb-yellow); }
    .badge.red, .tab.red { border-left: 3px solid var(--gdb-red); }
    .muted { color: var(--gdb-muted); }
    .notice { color: var(--gdb-muted); font-style: italic; white-space: nowrap; }
    .error { color: var(--gdb-red); }
    select, .icon { background: #1e293b; border: 1px solid #334155; border-radius: 4px; padding: 3px 6px; font: inherit; font-size: 11px; color: var(--gdb-text); max-width: 260px; }
    .icon { cursor: pointer; text-decoration: none; }
    .drawer { position: fixed; left: 0; right: 0; bottom: ${BAR_HEIGHT}px; background: var(--gdb-bg); border-top: 1px solid var(--gdb-border); z-index: 2147483646; display: none; flex-direction: column; }
    .drawer.open { display: flex; }
    .resize { height: 5px; cursor: ns-resize; background: transparent; }
    .resize:hover { background: var(--gdb-accent); }
    .drawer-header { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 6px 12px; background: #1e293b; border-bottom: 1px solid var(--gdb-border); font-weight: 600; }
    .drawer-body { padding: 10px 12px; overflow: auto; flex: 1; }
    .tabs { display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 10px; }
    .tab { background: #1e293b; border: 1px solid #334155; border-radius: 4px; padding: 3px 8px; cursor: pointer; }
    .tab[aria-selected="true"] { border-color: var(--gdb-accent); background: #1e3a5f; }
    .rows tr { cursor: pointer; }
    .rows tr:hover td { background: #162238; }
    .mini { position: fixed; right: 12px; bottom: 12px; z-index: 2147483647; width: 40px; height: 40px; border-radius: 50%; border: 2px solid var(--gdb-accent); background: var(--gdb-panel); color: var(--gdb-accent); font-weight: 700; cursor: pointer; box-shadow: 0 2px 8px rgba(0,0,0,.4); }
    .mini.red { border-color: var(--gdb-red); color: var(--gdb-red); }
    .mini.yellow { border-color: var(--gdb-yellow); color: var(--gdb-yellow); }
    ${R.css}
  ` }));

  const pageChip = h('button', { className: 'page', attrs: { type: 'button' } });
  const items = h('div', { className: 'items', attrs: { role: 'toolbar', 'aria-label': 'Numbers for this page' } });
  const historySelect = h('select', { attrs: { 'aria-label': 'Earlier requests' } });
  const openLink = h('a', { className: 'icon', text: 'Profiler ↗', attrs: { target: '_blank', rel: 'noopener', href: baseUrl, title: 'Open the full profiler' } });
  const minimizeBtn = h('button', { className: 'icon', text: '–', attrs: { type: 'button', title: 'Minimize', 'aria-label': 'Minimize toolbar' } });
  const status = h('span', { className: 'muted' });
  const brand = h('button', { className: 'brand', text: 'GoDebugBar', attrs: { type: 'button', title: 'Minimize' } });
  const bar = h('div', { className: 'bar' },
    brand, pageChip, items,
    h('div', { className: 'tools' }, status, historySelect, openLink, minimizeBtn));

  const drawerTitle = h('span');
  const drawerTools = h('span', { className: 'tools' });
  const closeBtn = h('button', { className: 'icon', text: '×', attrs: { type: 'button', title: 'Close (Esc)', 'aria-label': 'Close panel' } });
  const drawerBody = h('div', { className: 'drawer-body' });
  const resize = h('div', { className: 'resize', attrs: { title: 'Drag to resize' } });
  const drawer = h('div', { className: 'drawer', attrs: { role: 'region', 'aria-label': 'GoDebugBar panel' } },
    resize,
    h('div', { className: 'drawer-header' }, drawerTitle, h('span', { className: 'tools' }, drawerTools, closeBtn)),
    drawerBody);
  const mini = h('button', { className: 'mini', text: 'GDB', attrs: { type: 'button', title: 'Show GoDebugBar' } });

  shadow.append(drawer, bar, mini);

  const newPage = (path, startedAt) => ({ path, startedAt, requests: [], settled: false });
  let page = newPage(window.location.pathname, 0);
  let pageTimer = null;
  let view = null;
  let history = [];

  const profileUrl = (id) => `${baseUrl}?id=${encodeURIComponent(id)}`;

  const totals = () => {
    const t = { requests: 0, ms: 0, queries: 0, queryMs: 0, slow: 0, nPlusOne: 0, duplicates: 0, failed: 0, errors: 0, logs: 0, calls: 0, worstStatus: 0 };
    page.requests.forEach((r) => {
      const s = (r.profile && r.profile.stats) || {};
      t.requests += 1;
      t.ms += r.profile ? r.profile.duration_ms : 0;
      t.queries += s.queries || 0;
      t.queryMs += s.query_ms || 0;
      t.slow += s.slow_queries || 0;
      t.nPlusOne += s.n_plus_one || 0;
      t.duplicates += s.duplicate_queries || 0;
      t.failed += s.failed_queries || 0;
      t.errors += (s.errors || 0) + (s.panic ? 1 : 0);
      t.logs += s.logs || 0;
      t.calls += s.http_calls || 0;
      t.worstStatus = Math.max(t.worstStatus, r.profile ? r.profile.status : 0);
    });
    return t;
  };

  const statusColor = (code) => {
    if (code >= 500) return 'red';
    if (code >= 400) return 'yellow';
    if (code >= 200) return 'green';
    return '';
  };

  const applyLayout = () => {
    const open = !!view && !state.minimized;
    bar.hidden = state.minimized;
    mini.hidden = !state.minimized;
    const t = totals();
    mini.className = `mini ${t.errors || t.failed || t.worstStatus >= 500 ? 'red' : (t.slow || t.nPlusOne || t.worstStatus >= 400 ? 'yellow' : '')}`;
    drawer.classList.toggle('open', open);
    drawer.style.height = `${state.height}px`;
    spacer.style.height = state.minimized ? '0' : `${BAR_HEIGHT + (open ? state.height : 0)}px`;
  };

  let openView = () => undefined;

  const badge = (key, text, color, title) => {
    const active = view && view.kind === key;
    const el = h('button', {
      className: `badge ${color || ''}${active ? ' active' : ''}`,
      text,
      attrs: { type: 'button', title, 'aria-pressed': !!active },
    });
    el.addEventListener('click', () => openView(active ? null : { kind: key }));
    return el;
  };

  const renderBar = () => {
    pageChip.textContent = page.path;
    pageChip.title = `Show all requests on ${page.path}`;
    items.textContent = '';
    const t = totals();
    if (!t.requests) {
      if (page.settled) items.appendChild(h('span', { className: 'notice', text: 'No API request on this page (served from the app\'s cache).' }));
      return;
    }
    const dbColor = t.failed ? 'red' : (t.slow || t.nPlusOne || t.duplicates ? 'yellow' : 'green');
    items.append(
      badge('requests', `Requests: ${t.requests}`, statusColor(t.worstStatus), 'API requests made on this page'),
      badge('timeline', `Time: ${R.formatMs(t.ms)}`, '', 'Server time of all requests on this page; click for the page timeline'),
      badge('queries', `DB: ${t.queries} (${R.formatMs(t.queryMs)})`, dbColor,
        `Queries: ${t.queries}\nSlow: ${t.slow}\nDuplicates: ${t.duplicates}\nN+1 suspects: ${t.nPlusOne}\nFailed: ${t.failed}`),
    );
    if (t.errors) items.appendChild(badge('requests', `Errors: ${t.errors}`, 'red', 'Errors and panics on this page'));
    if (t.logs) items.appendChild(badge('requests', `Logs: ${t.logs}`, '', 'Log messages on this page'));
    if (t.calls) items.appendChild(badge('requests', `HTTP: ${t.calls}`, '', 'Outgoing HTTP calls on this page'));
  };

  const requestRows = (requests, onPick) => {
    const table = h('table', { className: 'gdb-table rows' },
      h('thead', {}, h('tr', {}, ['Method', 'Path', 'Status', 'Time', 'DB', 'Notes'].map((c) => h('th', { text: c })))));
    const body = h('tbody');
    requests.forEach((r) => {
      const p = r.profile;
      const s = (p && p.stats) || {};
      const notes = [
        s.slow_queries ? `${s.slow_queries} slow` : '',
        s.n_plus_one ? 'N+1' : '',
        s.duplicate_queries ? `${s.duplicate_queries} duplicate` : '',
        s.failed_queries ? `${s.failed_queries} failed` : '',
        s.errors ? `${s.errors} errors` : '',
        s.panic ? 'panic' : '',
      ].filter(Boolean).join(', ');
      const color = s.panic || s.failed_queries || (p && p.status >= 500) ? 'red' : (notes || (p && p.status >= 400) ? 'yellow' : '');
      const tr = h('tr', { className: color ? `gdb-row-${color}` : '' },
        h('td', { text: p ? p.method : r.method || '' }),
        h('td', { text: p ? p.path : '…' }),
        h('td', { text: p ? p.status : '' }),
        h('td', { text: p ? R.formatMs(p.duration_ms) : '' }),
        h('td', { text: p ? `${s.queries || 0} (${R.formatMs(s.query_ms)})` : '' }),
        h('td', { text: notes }));
      tr.addEventListener('click', () => onPick(r));
      body.appendChild(tr);
    });
    table.appendChild(body);
    return table;
  };

  const pageTimeline = () => ({
    kind: 'timeline',
    title: '',
    events: page.requests.map((r) => ({
      label: r.profile ? `${r.profile.method} ${r.profile.path} (${r.profile.status})` : r.method || 'request',
      category: 'http',
      start_ms: Math.max(r.clientStart - page.startedAt, 0),
      duration_ms: Math.max(r.clientEnd - r.clientStart, 0),
      color: r.profile && r.profile.status >= 500 ? 'red' : (r.profile && r.profile.status >= 400 ? 'yellow' : ''),
    })),
  });

  const renderRequest = (profile, sectionName) => {
    const sections = profile.sections || [];
    const active = sections.find((s) => s.name === sectionName) || sections[0];
    drawerTitle.textContent = `${profile.method} ${profile.path} · ${profile.status} · ${R.formatMs(profile.duration_ms)}`;
    drawerTools.textContent = '';
    drawerTools.append(
      h('button', { className: 'icon', text: '← Page', attrs: { type: 'button' }, on: { click: () => openView({ kind: 'requests' }) } }),
      h('a', { text: 'Open in profiler ↗', attrs: { target: '_blank', rel: 'noopener', href: profileUrl(profile.id) } }),
    );
    drawerBody.textContent = '';
    drawerBody.appendChild(h('div', { className: 'tabs', attrs: { role: 'tablist' } }, sections.map((s) => {
      const item = s.toolbar || {};
      return h('button', {
        className: `tab ${item.color || ''}`,
        text: `${item.label}: ${item.value}`,
        attrs: { type: 'button', role: 'tab', 'aria-selected': active && s.name === active.name ? 'true' : 'false', title: R.tooltip(item) },
        on: { click: () => openView({ kind: 'request', profile, section: s.name }) },
      });
    })));
    const body = h('div');
    drawerBody.appendChild(body);
    R.renderWidgets(body, (active && active.panel && active.panel.widgets) || [], {
      explainable: profile.explainable,
      onExplain: R.explainer(baseUrl, profile.id, api),
    });
  };

  const renderDrawer = () => {
    if (!view) return;
    drawerTools.textContent = '';
    if (view.kind === 'request') {
      renderRequest(view.profile, view.section);
      return;
    }
    const pick = (r) => r.profile && openView({ kind: 'request', profile: r.profile });
    drawerBody.textContent = '';
    if (view.kind === 'timeline') {
      drawerTitle.textContent = `Timeline of ${page.path}`;
      R.renderWidgets(drawerBody, [pageTimeline()]);
      drawerBody.appendChild(requestRows(page.requests, pick));
      return;
    }
    if (view.kind === 'queries') {
      drawerTitle.textContent = `Requests on ${page.path} with their queries`;
      const withQueries = page.requests.filter((r) => r.profile && r.profile.stats && r.profile.stats.queries);
      drawerBody.appendChild(withQueries.length
        ? requestRows(withQueries, (r) => openView({ kind: 'request', profile: r.profile, section: 'database' }))
        : h('div', { className: 'gdb-empty', text: 'No queries on this page.' }));
      return;
    }
    drawerTitle.textContent = `Requests on ${page.path}`;
    drawerBody.appendChild(page.requests.length
      ? requestRows(page.requests, pick)
      : h('div', { className: 'gdb-empty', text: 'No API request on this page yet.' }));
  };

  openView = (next) => {
    view = next;
    renderBar();
    renderDrawer();
    applyLayout();
  };

  const refresh = () => {
    renderBar();
    if (view && view.kind !== 'request') renderDrawer();
    applyLayout();
  };

  const renderHistory = () => {
    historySelect.textContent = '';
    historySelect.appendChild(h('option', { text: 'Earlier requests…', attrs: { value: '' } }));
    history.forEach((p) => historySelect.appendChild(h('option', { text: `${p.method} ${p.path} (${p.status})`, attrs: { value: p.id } })));
  };

  const remember = (profile) => {
    history = [{ id: profile.id, method: profile.method, path: profile.path, status: profile.status },
      ...history.filter((p) => p.id !== profile.id)].slice(0, MAX_HISTORY);
    renderHistory();
  };

  const showError = (err) => {
    status.textContent = 'profiler unreachable';
    status.className = 'error';
    status.title = err.message;
  };

  const track = async (id, target, clientStart, clientEnd, method) => {
    if (!id || target.requests.some((r) => r.id === id)) return;
    const entry = { id, method, clientStart, clientEnd, profile: null };
    target.requests = [...target.requests, entry].slice(-MAX_PAGE_REQUESTS);
    if (target === page) refresh();
    try {
      entry.profile = await api(`/api/profiles/${encodeURIComponent(id)}`);
      status.textContent = '';
      remember(entry.profile);
      if (target === page) refresh();
    } catch (err) {
      showError(err);
    }
  };

  const settleLater = (target) => {
    clearTimeout(pageTimer);
    pageTimer = setTimeout(() => {
      target.settled = true;
      if (target === page) refresh();
    }, PAGE_SETTLE_MS);
  };

  const onRouteChange = () => {
    const path = window.location.pathname;
    if (path === page.path) return;
    page = newPage(path, now());
    if (view && view.kind !== 'request') view = null;
    settleLater(page);
    refresh();
  };

  ['pushState', 'replaceState'].forEach((method) => {
    const original = window.history[method];
    window.history[method] = function patchedHistory(...args) {
      const result = original.apply(this, args);
      onRouteChange();
      return result;
    };
  });
  window.addEventListener('popstate', onRouteChange);

  const isOwnCall = (url) => String(url || '').includes(baseUrl);

  window.fetch = function patchedFetch(input, init) {
    const url = typeof input === 'string' ? input : (input && (input.url || input.href)) || '';
    const method = ((init && init.method) || (input && input.method) || 'GET').toUpperCase();
    const target = page;
    const startedAt = now();
    return origFetch(input, init).then((res) => {
      if (!isOwnCall(url)) track(res.headers.get(PROFILE_HEADER), target, startedAt, now(), method);
      return res;
    });
  };

  const origOpen = XMLHttpRequest.prototype.open;
  XMLHttpRequest.prototype.open = function patchedOpen(method, url, ...rest) {
    if (!isOwnCall(url)) {
      const target = page;
      const startedAt = now();
      this.addEventListener('loadend', () => {
        try {
          track(this.getResponseHeader(PROFILE_HEADER), target, startedAt, now(), String(method).toUpperCase());
        } catch {
          return undefined;
        }
        return undefined;
      });
    }
    return origOpen.call(this, method, url, ...rest);
  };

  const setMinimized = (minimized) => {
    state.minimized = minimized;
    saveState();
    applyLayout();
  };

  brand.addEventListener('click', () => setMinimized(true));
  minimizeBtn.addEventListener('click', () => setMinimized(true));
  mini.addEventListener('click', () => setMinimized(false));
  closeBtn.addEventListener('click', () => openView(null));
  pageChip.addEventListener('click', () => openView(view && view.kind === 'requests' ? null : { kind: 'requests' }));
  historySelect.addEventListener('change', async () => {
    const id = historySelect.value;
    historySelect.value = '';
    if (!id) return;
    try {
      openView({ kind: 'request', profile: await api(`/api/profiles/${encodeURIComponent(id)}`) });
    } catch (err) {
      showError(err);
    }
  });
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && view) openView(null);
  });
  resize.addEventListener('pointerdown', (event) => {
    event.preventDefault();
    const startY = event.clientY;
    const startHeight = state.height;
    const move = (e) => {
      state.height = Math.min(Math.max(startHeight + (startY - e.clientY), 120), window.innerHeight - 80);
      applyLayout();
    };
    const up = () => {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', up);
      saveState();
    };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', up);
  });

  document.documentElement.appendChild(host);
  (document.body || document.documentElement).appendChild(spacer);
  renderHistory();
  refresh();
  settleLater(page);

  const start = async () => {
    const firstPage = page;
    try {
      const list = await api(`/api/profiles?limit=${MAX_HISTORY}`);
      history = (list || []).map((p) => ({ id: p.id, method: p.method, path: p.path, status: p.status }));
      renderHistory();
    } catch (err) {
      showError(err);
    }
    if (initialId) {
      const nav = performance.getEntriesByType && performance.getEntriesByType('navigation')[0];
      track(initialId, firstPage, 0, nav ? nav.responseEnd : 0, 'GET');
    }
  };
  start();
})();
