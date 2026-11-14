(function () {
  if (window.__GODEBUGBAR_LOADED__) return;
  window.__GODEBUGBAR_LOADED__ = true;

  var scriptEl = document.currentScript || document.querySelector('script[data-profile-id]');
  var prefix = (scriptEl && scriptEl.getAttribute('data-prefix')) || '/_debugbar';
  var initialId = (scriptEl && scriptEl.getAttribute('data-profile-id')) || '';
  var origin = '';
  if (scriptEl && scriptEl.src) {
    try {
      origin = new URL(scriptEl.src, window.location.href).origin;
    } catch (_) {}
  }
  var baseUrl = origin + prefix;

  var profiles = [];
  var activeProfile = null;
  var activeTab = '';
  var collapsed = false;

  var host = document.createElement('div');
  host.id = 'godebugbar-host';
  var shadow = host.attachShadow({ mode: 'open' });

  var style = document.createElement('style');
  style.textContent = [
    ':host { all: initial; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; color: #e2e8f0; }',
    '.bar { position: fixed; left: 0; right: 0; bottom: 0; height: 36px; background: #0f172a; border-top: 1px solid #334155; display: flex; align-items: center; justify-content: space-between; padding: 0 10px; z-index: 2147483647; box-sizing: border-box; }',
    '.bar.collapsed { left: auto; width: auto; border-left: 1px solid #334155; border-top-left-radius: 6px; padding: 0 8px; }',
    '.items { display: flex; align-items: center; gap: 6px; overflow-x: auto; }',
    '.brand { font-weight: 700; color: #38bdf8; padding-right: 6px; cursor: pointer; user-select: none; }',
    '.badge { display: inline-flex; align-items: center; gap: 6px; padding: 4px 8px; border-radius: 4px; background: #1e293b; border: 1px solid #334155; cursor: pointer; white-space: nowrap; }',
    '.badge:hover, .badge.active { background: #334155; border-color: #64748b; }',
    '.dot { width: 8px; height: 8px; border-radius: 50%; background: #94a3b8; }',
    '.dot.green { background: #22c55e; }',
    '.dot.yellow { background: #eab308; }',
    '.dot.red { background: #ef4444; }',
    '.label { color: #94a3b8; font-weight: 600; }',
    '.val { color: #f8fafc; }',
    '.controls { display: flex; align-items: center; gap: 8px; }',
    'select { background: #1e293b; color: #f8fafc; border: 1px solid #334155; border-radius: 4px; padding: 3px 6px; font-family: inherit; font-size: 11px; }',
    '.btn { background: transparent; border: none; color: #94a3b8; cursor: pointer; font-size: 14px; padding: 2px 6px; }',
    '.btn:hover { color: #f8fafc; }',
    '.drawer { position: fixed; left: 0; right: 0; bottom: 36px; height: 360px; background: #0f172a; border-top: 1px solid #334155; z-index: 2147483646; display: none; flex-direction: column; }',
    '.drawer.open { display: flex; }',
    '.drawer-head { display: flex; justify-content: space-between; align-items: center; padding: 8px 14px; background: #1e293b; border-bottom: 1px solid #334155; font-weight: 700; }',
    '.drawer-body { padding: 12px 14px; overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 14px; }',
    '.metrics { display: flex; flex-wrap: wrap; gap: 10px; }',
    '.metric { background: #1e293b; border: 1px solid #334155; border-radius: 4px; padding: 8px 12px; min-width: 120px; }',
    '.metric-lbl { color: #94a3b8; font-size: 11px; }',
    '.metric-val { color: #f8fafc; font-size: 14px; font-weight: 700; margin-top: 2px; }',
    '.widget-title { font-weight: 700; color: #cbd5e1; margin-bottom: 6px; }',
    'table { width: 100%; border-collapse: collapse; font-size: 12px; }',
    'th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid #1e293b; vertical-align: top; }',
    'th { background: #1e293b; color: #94a3b8; font-weight: 600; }',
    'pre { background: #1e293b; padding: 8px; border-radius: 4px; overflow-x: auto; margin: 0; }'
  ].join('\n');

  var drawerEl = document.createElement('div');
  drawerEl.className = 'drawer';
  var barEl = document.createElement('div');
  barEl.className = 'bar';

  shadow.appendChild(style);
  shadow.appendChild(drawerEl);
  shadow.appendChild(barEl);

  function mount() {
    if (document.body && !document.getElementById('godebugbar-host')) {
      document.body.appendChild(host);
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', mount);
  } else {
    mount();
  }

  function loadProfile(id) {
    if (!id) return;
    fetch(baseUrl + '/api/profiles/' + encodeURIComponent(id), { headers: { 'X-Debugbar-Internal': '1' } })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (profile) {
        if (!profile) return;
        var exists = profiles.some(function (p) { return p.id === profile.id; });
        if (!exists) profiles.unshift(profile);
        activeProfile = profile;
        render();
      })
      .catch(function () {});
  }

  function render() {
    barEl.innerHTML = '';
    drawerEl.innerHTML = '';

    if (collapsed) {
      barEl.className = 'bar collapsed';
      var expandBtn = document.createElement('span');
      expandBtn.className = 'brand';
      expandBtn.textContent = 'GoDebugbar';
      expandBtn.onclick = function () { collapsed = false; render(); };
      barEl.appendChild(expandBtn);
      drawerEl.className = 'drawer';
      return;
    }

    barEl.className = 'bar';
    var itemsEl = document.createElement('div');
    itemsEl.className = 'items';

    var brand = document.createElement('span');
    brand.className = 'brand';
    brand.textContent = 'GoDebugbar';
    brand.onclick = function () { activeTab = ''; render(); };
    itemsEl.appendChild(brand);

    var sections = (activeProfile && activeProfile.sections) || [];
    sections.forEach(function (sec) {
      var badge = document.createElement('div');
      badge.className = 'badge' + (activeTab === sec.name ? ' active' : '');
      var dot = document.createElement('span');
      dot.className = 'dot ' + (sec.toolbar.color || 'default');
      var lbl = document.createElement('span');
      lbl.className = 'label';
      lbl.textContent = sec.toolbar.label;
      var val = document.createElement('span');
      val.className = 'val';
      val.textContent = sec.toolbar.value;

      badge.appendChild(dot);
      badge.appendChild(lbl);
      badge.appendChild(val);
      badge.onclick = function () {
        activeTab = activeTab === sec.name ? '' : sec.name;
        render();
      };
      itemsEl.appendChild(badge);
    });

    var controlsEl = document.createElement('div');
    controlsEl.className = 'controls';

    if (profiles.length > 0) {
      var sel = document.createElement('select');
      profiles.forEach(function (p) {
        var opt = document.createElement('option');
        opt.value = p.id;
        opt.textContent = p.status + ' ' + p.method + ' ' + p.path + ' (' + p.duration_ms.toFixed(1) + 'ms)';
        if (activeProfile && activeProfile.id === p.id) opt.selected = true;
        sel.appendChild(opt);
      });
      sel.onchange = function () {
        var found = profiles.find(function (p) { return p.id === sel.value; });
        if (found) {
          activeProfile = found;
          render();
        }
      };
      controlsEl.appendChild(sel);
    }

    var minBtn = document.createElement('button');
    minBtn.className = 'btn';
    minBtn.textContent = '×';
    minBtn.onclick = function () { collapsed = true; activeTab = ''; render(); };
    controlsEl.appendChild(minBtn);

    barEl.appendChild(itemsEl);
    barEl.appendChild(controlsEl);

    var currentSec = sections.find(function (s) { return s.name === activeTab; });
    if (!currentSec || !currentSec.panel) {
      drawerEl.className = 'drawer';
      return;
    }

    drawerEl.className = 'drawer open';
    var head = document.createElement('div');
    head.className = 'drawer-head';
    var titleSpan = document.createElement('span');
    titleSpan.textContent = currentSec.panel.title;
    var closeBtn = document.createElement('button');
    closeBtn.className = 'btn';
    closeBtn.textContent = 'Close';
    closeBtn.onclick = function () { activeTab = ''; render(); };
    head.appendChild(titleSpan);
    head.appendChild(closeBtn);
    drawerEl.appendChild(head);

    var body = document.createElement('div');
    body.className = 'drawer-body';

    (currentSec.panel.widgets || []).forEach(function (w) {
      var block = document.createElement('div');
      if (w.title) {
        var wt = document.createElement('div');
        wt.className = 'widget-title';
        wt.textContent = w.title;
        block.appendChild(wt);
      }

      if (w.kind === 'metrics' && w.metrics) {
        var mWrap = document.createElement('div');
        mWrap.className = 'metrics';
        w.metrics.forEach(function (m) {
          var card = document.createElement('div');
          card.className = 'metric';
          var ml = document.createElement('div');
          ml.className = 'metric-lbl';
          ml.textContent = m.label;
          var mv = document.createElement('div');
          mv.className = 'metric-val';
          mv.textContent = m.value;
          card.appendChild(ml);
          card.appendChild(mv);
          mWrap.appendChild(card);
        });
        block.appendChild(mWrap);
      } else if (w.kind === 'key_value' && w.pairs) {
        var tbl = document.createElement('table');
        Object.keys(w.pairs).sort().forEach(function (k) {
          var tr = document.createElement('tr');
          var tdK = document.createElement('th');
          tdK.style.width = '240px';
          tdK.textContent = k;
          var tdV = document.createElement('td');
          tdV.textContent = w.pairs[k];
          tr.appendChild(tdK);
          tr.appendChild(tdV);
          tbl.appendChild(tr);
        });
        block.appendChild(tbl);
      } else if (w.kind === 'table') {
        var tEl = document.createElement('table');
        if (w.columns && w.columns.length) {
          var thead = document.createElement('tr');
          w.columns.forEach(function (c) {
            var th = document.createElement('th');
            th.textContent = c;
            thead.appendChild(th);
          });
          tEl.appendChild(thead);
        }
        (w.rows || []).forEach(function (r) {
          var tr = document.createElement('tr');
          r.forEach(function (cell) {
            var td = document.createElement('td');
            td.textContent = cell;
            tr.appendChild(td);
          });
          tEl.appendChild(tr);
        });
        block.appendChild(tEl);
      } else if (w.kind === 'code') {
        var pre = document.createElement('pre');
        pre.textContent = w.code || '';
        block.appendChild(pre);
      }

      body.appendChild(block);
    });

    drawerEl.appendChild(body);
  }

  var origFetch = window.fetch;
  if (origFetch) {
    window.fetch = function () {
      var args = arguments;
      var reqInput = args[0];
      var reqInit = args[1];
      var isInternal = reqInit && reqInit.headers && reqInit.headers['X-Debugbar-Internal'];
      return origFetch.apply(this, args).then(function (res) {
        if (!isInternal && res && res.headers) {
          var debugId = res.headers.get('X-Debugbar-ID');
          if (debugId && String(reqInput).indexOf(prefix) === -1) {
            loadProfile(debugId);
          }
        }
        return res;
      });
    };
  }

  if (initialId) {
    loadProfile(initialId);
  } else {
    fetch(baseUrl + '/api/profiles', { headers: { 'X-Debugbar-Internal': '1' } })
      .then(function (r) { return r.ok ? r.json() : []; })
      .then(function (list) {
        if (Array.isArray(list) && list.length > 0) {
          profiles = list;
          activeProfile = list[0];
          render();
        } else {
          render();
        }
      })
      .catch(function () { render(); });
  }
})();
