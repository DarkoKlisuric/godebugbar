(() => {
  'use strict';

  const script = document.currentScript;
  const R = window.GoDebugBarRender;
  if (!script || !R || document.getElementById('godebugbar-host')) return;

  const PROFILE_HEADER = 'X-Debugbar-ID';
  const STORAGE_KEY = 'godebugbar:toolbar';
  const MAX_HISTORY = 50;
  const BAR_HEIGHT = 36;
  const baseUrl = (script.dataset.baseUrl || '/_debugbar').replace(/\/+$/, '');
  const initialId = script.dataset.profileId || '';
  const origFetch = window.fetch.bind(window);

  const readState = () => {
    try {
      return { minimized: false, follow: true, height: 340, panel: '', ...JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}') };
    } catch {
      return { minimized: false, follow: true, height: 340, panel: '' };
    }
  };
  const state = readState();
  const saveState = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
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
    .badge { display: inline-flex; align-items: center; gap: 6px; padding: 3px 8px; border-radius: 4px; cursor: pointer; background: #1e293b; border: 1px solid #334155; white-space: nowrap; }
    .badge:hover, .badge:focus-visible, .badge.active { border-color: var(--gdb-accent); background: #1e3a5f; outline: none; }
    .badge.green { border-left: 3px solid var(--gdb-green); }
    .badge.yellow { border-left: 3px solid var(--gdb-yellow); }
    .badge.red { border-left: 3px solid var(--gdb-red); }
    .pill { background: var(--gdb-accent); color: #0b1120; border: none; border-radius: 10px; padding: 1px 8px; font-weight: 700; cursor: pointer; }
    .muted { color: var(--gdb-muted); }
    .page { color: var(--gdb-muted); white-space: nowrap; flex-shrink: 0; max-width: 40%; overflow: hidden; text-overflow: ellipsis; }
    .page b { color: var(--gdb-text); font-weight: 600; }
    .notice { color: var(--gdb-muted); font-style: italic; white-space: nowrap; }
    .error { color: var(--gdb-red); }
    select, .icon { background: #1e293b; border: 1px solid #334155; border-radius: 4px; padding: 3px 6px; font: inherit; font-size: 11px; color: var(--gdb-text); max-width: 260px; }
    .icon { cursor: pointer; }
    label { display: inline-flex; gap: 4px; align-items: center; color: var(--gdb-muted); cursor: pointer; }
    .drawer { position: fixed; left: 0; right: 0; bottom: ${BAR_HEIGHT}px; background: var(--gdb-bg); border-top: 1px solid var(--gdb-border); z-index: 2147483646; display: none; flex-direction: column; }
    .drawer.open { display: flex; }
    .resize { height: 5px; cursor: ns-resize; background: transparent; }
    .resize:hover { background: var(--gdb-accent); }
    .drawer-header { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 6px 12px; background: #1e293b; border-bottom: 1px solid var(--gdb-border); font-weight: 600; }
    .drawer-body { padding: 10px 12px; overflow: auto; flex: 1; }
    .mini { position: fixed; right: 12px; bottom: 12px; z-index: 2147483647; width: 40px; height: 40px; border-radius: 50%; border: 2px solid var(--gdb-accent); background: var(--gdb-panel); color: var(--gdb-accent); font-weight: 700; cursor: pointer; box-shadow: 0 2px 8px rgba(0,0,0,.4); }
    .mini.red { border-color: var(--gdb-red); color: var(--gdb-red); }
    .mini.yellow { border-color: var(--gdb-yellow); color: var(--gdb-yellow); }
    ${R.css}
  ` }));

  const items = h('div', { className: 'items', attrs: { role: 'toolbar', 'aria-label': 'GoDebugBar sections' } });
  const newPill = h('button', { className: 'pill', attrs: { type: 'button', hidden: true, title: 'Show the newest request' } });
  const followBox = h('input', { attrs: { type: 'checkbox' } });
  followBox.checked = state.follow;
  const historySelect = h('select', { attrs: { 'aria-label': 'Recorded requests' } });
  const openLink = h('a', { className: 'icon', text: 'Open ↗', attrs: { target: '_blank', rel: 'noopener', title: 'Open this request in the profiler' } });
  const minimizeBtn = h('button', { className: 'icon', text: '–', attrs: { type: 'button', title: 'Minimize', 'aria-label': 'Minimize toolbar' } });
  const status = h('span', { className: 'muted' });
  const pageChip = h('span', { className: 'page', attrs: { title: 'API requests made since this page opened' } });
  const brand = h('button', { className: 'brand', text: 'GoDebugBar', attrs: { type: 'button', title: 'Minimize' } });
  const bar = h('div', { className: 'bar' },
    brand,
    pageChip,
    items,
    h('div', { className: 'tools' }, status, newPill, h('label', { attrs: { title: 'Follow new requests' } }, followBox, 'Follow'), historySelect, openLink, minimizeBtn));

  const drawerTitle = h('span');
  const drawerOpen = h('a', { text: 'Open in profiler ↗', attrs: { target: '_blank', rel: 'noopener' } });
  const closeBtn = h('button', { className: 'icon', text: '×', attrs: { type: 'button', title: 'Close (Esc)', 'aria-label': 'Close panel' } });
  const drawerBody = h('div', { className: 'drawer-body' });
  const resize = h('div', { className: 'resize', attrs: { title: 'Drag to resize' } });
  const drawer = h('div', { className: 'drawer', attrs: { role: 'region', 'aria-label': 'GoDebugBar panel' } },
    resize,
    h('div', { className: 'drawer-header' }, drawerTitle, h('span', { className: 'tools' }, drawerOpen, closeBtn)),
    drawerBody);
  const mini = h('button', { className: 'mini', text: 'GDB', attrs: { type: 'button', title: 'Show GoDebugBar' } });

  shadow.append(drawer, bar, mini);

  let current = null;
  let history = [];
  let newCount = 0;
  let pinned = false;
  let loadSeq = 0;
  let page = { path: window.location.pathname, count: 0, settled: false };
  let pageTimer = null;
  const PAGE_SETTLE_MS = 1200;

  const profileUrl = (id) => `${baseUrl}?id=${encodeURIComponent(id)}`;
  const worstColor = (profile) => {
    const colors = (profile && profile.sections || []).map((s) => s.toolbar && s.toolbar.color);
    if (colors.includes('red')) return 'red';
    if (colors.includes('yellow')) return 'yellow';
    return '';
  };

  const applyLayout = () => {
    const open = !!state.panel && !!current && !state.minimized;
    bar.hidden = state.minimized;
    mini.hidden = !state.minimized;
    mini.className = `mini ${worstColor(current)}`;
    drawer.classList.toggle('open', open);
    drawer.style.height = `${state.height}px`;
    spacer.style.height = state.minimized ? '0' : `${BAR_HEIGHT + (open ? state.height : 0)}px`;
  };

  const renderPanel = () => {
    if (!current) {
      applyLayout();
      return;
    }
    const section = (current.sections || []).find((s) => s.name === state.panel);
    if (!section) {
      state.panel = '';
      applyLayout();
      return;
    }
    drawerTitle.textContent = (section.panel && section.panel.title) || section.name;
    drawerOpen.href = profileUrl(current.id);
    R.renderWidgets(drawerBody, (section.panel && section.panel.widgets) || [], {
      explainable: current.explainable,
      onExplain: R.explainer(baseUrl, current.id, api),
    });
  };

  const renderPage = () => {
    pageChip.textContent = '';
    let detail = '';
    if (page.count > 0) detail = ` · ${page.count} API ${page.count === 1 ? 'request' : 'requests'}`;
    else if (page.settled) detail = ' · no API request';
    pageChip.append(h('b', { text: page.path }), detail);
  };

  const renderBar = () => {
    items.textContent = '';
    if (!current && page.settled && page.count === 0) {
      items.appendChild(h('span', { className: 'notice', text: 'This page was served from the app\'s cache: no request reached the server.' }));
    }
    (current && current.sections || []).forEach((section) => {
      const item = section.toolbar || {};
      const badge = h('button', {
        className: `badge ${item.color || ''}${state.panel === section.name ? ' active' : ''}`,
        text: `${item.label}: ${item.value}`,
        attrs: { type: 'button', title: R.tooltip(item), 'aria-pressed': state.panel === section.name },
      });
      badge.addEventListener('click', () => {
        state.panel = state.panel === section.name ? '' : section.name;
        saveState();
        renderBar();
        renderPanel();
        applyLayout();
      });
      items.appendChild(badge);
    });
    if (current) openLink.href = profileUrl(current.id);
    openLink.hidden = !current;
    newPill.hidden = newCount === 0;
    newPill.textContent = `${newCount} new`;
  };

  const renderHistory = () => {
    historySelect.textContent = '';
    history.forEach((p) => historySelect.appendChild(h('option', {
      text: `${p.method} ${p.path} (${p.status})`,
      attrs: { value: p.id, selected: current && current.id === p.id },
    })));
  };

  const remember = (summary) => {
    history = [summary, ...history.filter((p) => p.id !== summary.id)].slice(0, MAX_HISTORY);
  };

  const show = async (id) => {
    const seq = ++loadSeq;
    try {
      const profile = await api(`/api/profiles/${encodeURIComponent(id)}`);
      if (seq !== loadSeq) return;
      current = profile;
      remember({ id: profile.id, method: profile.method, path: profile.path, status: profile.status });
      status.textContent = '';
      renderHistory();
      renderBar();
      renderPanel();
      applyLayout();
    } catch (err) {
      status.textContent = 'profiler unreachable';
      status.className = 'error';
      status.title = err.message;
    }
  };

  const onProfileId = async (id) => {
    if (!id || history.some((p) => p.id === id)) return;
    page.count += 1;
    renderPage();
    if (state.follow && !pinned) {
      newCount = 0;
      await show(id);
      return;
    }
    try {
      const profile = await api(`/api/profiles/${encodeURIComponent(id)}`);
      remember({ id: profile.id, method: profile.method, path: profile.path, status: profile.status });
      newCount += 1;
      renderHistory();
      renderBar();
    } catch {
      return;
    }
  };

  const onRouteChange = () => {
    const path = window.location.pathname;
    if (path === page.path) return;
    page = { path, count: 0, settled: false };
    renderPage();
    clearTimeout(pageTimer);
    pageTimer = setTimeout(() => {
      if (page.path !== path) return;
      page.settled = true;
      renderPage();
      if (page.count === 0 && state.follow && !pinned) {
        loadSeq += 1;
        current = null;
        renderHistory();
        renderBar();
        applyLayout();
      }
    }, PAGE_SETTLE_MS);
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
    return origFetch(input, init).then((res) => {
      if (!isOwnCall(url)) onProfileId(res.headers.get(PROFILE_HEADER));
      return res;
    });
  };

  const origOpen = XMLHttpRequest.prototype.open;
  XMLHttpRequest.prototype.open = function patchedOpen(method, url, ...rest) {
    if (!isOwnCall(url)) {
      this.addEventListener('loadend', () => {
        try {
          onProfileId(this.getResponseHeader(PROFILE_HEADER));
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
  closeBtn.addEventListener('click', () => {
    state.panel = '';
    saveState();
    renderBar();
    applyLayout();
  });
  const showLatest = () => {
    pinned = false;
    newCount = 0;
    if (history[0]) show(history[0].id);
  };

  followBox.addEventListener('change', () => {
    state.follow = followBox.checked;
    saveState();
    if (state.follow) showLatest();
  });
  historySelect.addEventListener('change', () => {
    pinned = !!history[0] && historySelect.value !== history[0].id;
    if (!pinned) newCount = 0;
    show(historySelect.value);
  });
  newPill.addEventListener('click', showLatest);
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && state.panel) {
      state.panel = '';
      saveState();
      renderBar();
      applyLayout();
    }
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

  renderPage();
  document.documentElement.appendChild(host);
  (document.body || document.documentElement).appendChild(spacer);
  applyLayout();

  const start = async () => {
    try {
      const list = await api('/api/profiles');
      history = (list || []).slice(0, MAX_HISTORY).map((p) => ({ id: p.id, method: p.method, path: p.path, status: p.status }));
      renderHistory();
      const target = initialId || (history[0] && history[0].id);
      if (target) await show(target);
    } catch (err) {
      status.textContent = 'profiler unreachable';
      status.className = 'error';
      status.title = err.message;
    }
  };
  start();
})();
