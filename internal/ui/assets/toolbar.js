(function () {
  var currentScript = document.currentScript;
  if (!currentScript) return;
  var baseUrl = currentScript.getAttribute("data-base-url") || "/_debugbar";
  var initialId = currentScript.getAttribute("data-profile-id") || "";

  var host = document.createElement("div");
  host.id = "godebugbar-host";
  document.documentElement.appendChild(host);
  var shadow = host.attachShadow({ mode: "open" });

  var style = document.createElement("style");
  style.textContent = [
    ":host { all: initial; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; color: #e2e8f0; }",
    ".bar { position: fixed; left: 0; right: 0; bottom: 0; height: 36px; background: #0f172a; border-top: 1px solid #334155; display: flex; align-items: center; justify-content: space-between; padding: 0 12px; z-index: 2147483647; box-sizing: border-box; }",
    ".items { display: flex; align-items: center; gap: 8px; overflow-x: auto; }",
    ".brand { font-weight: 700; color: #38bdf8; margin-right: 6px; }",
    ".badge { display: inline-flex; align-items: center; gap: 6px; padding: 3px 8px; border-radius: 4px; cursor: pointer; background: #1e293b; border: 1px solid #334155; color: #e2e8f0; user-select: none; }",
    ".badge:hover, .badge.active { border-color: #38bdf8; background: #1e3a5f; }",
    ".badge.green { border-left: 3px solid #22c55e; }",
    ".badge.yellow { border-left: 3px solid #eab308; }",
    ".badge.red { border-left: 3px solid #ef4444; }",
    ".drawer { position: fixed; left: 0; right: 0; bottom: 36px; height: 340px; background: #0b1120; border-top: 1px solid #334155; z-index: 2147483646; display: none; flex-direction: column; overflow: hidden; }",
    ".drawer.open { display: flex; }",
    ".drawer-header { display: flex; align-items: center; justify-content: space-between; padding: 8px 14px; background: #1e293b; border-bottom: 1px solid #334155; font-weight: 600; }",
    ".drawer-body { padding: 12px 14px; overflow-y: auto; flex: 1; }",
    ".metrics { display: flex; gap: 12px; margin-bottom: 12px; flex-wrap: wrap; }",
    ".metric { background: #1e293b; border: 1px solid #334155; padding: 8px 12px; border-radius: 4px; min-width: 120px; }",
    ".metric-label { color: #94a3b8; font-size: 11px; }",
    ".metric-val { font-size: 14px; font-weight: 700; margin-top: 2px; }",
    "table { width: 100%; border-collapse: collapse; margin-bottom: 12px; }",
    "th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid #1e293b; vertical-align: top; word-break: break-all; }",
    "th { color: #94a3b8; font-weight: 600; width: 220px; }",
    "select { background: #1e293b; color: #e2e8f0; border: 1px solid #334155; border-radius: 4px; padding: 3px 6px; font-family: inherit; font-size: 11px; }",
    ".close-btn { cursor: pointer; color: #94a3b8; background: none; border: none; color: #e2e8f0; font-size: 14px; }"
  ].join("\n");

  var drawer = document.createElement("div");
  drawer.className = "drawer";
  var drawerHeader = document.createElement("div");
  drawerHeader.className = "drawer-header";
  var drawerTitle = document.createElement("span");
  var closeBtn = document.createElement("button");
  closeBtn.className = "close-btn";
  closeBtn.textContent = "×";
  drawerHeader.appendChild(drawerTitle);
  drawerHeader.appendChild(closeBtn);
  var drawerBody = document.createElement("div");
  drawerBody.className = "drawer-body";
  drawer.appendChild(drawerHeader);
  drawer.appendChild(drawerBody);

  var bar = document.createElement("div");
  bar.className = "bar";
  var itemsContainer = document.createElement("div");
  itemsContainer.className = "items";
  var rightContainer = document.createElement("div");
  rightContainer.className = "items";
  var historySelect = document.createElement("select");
  rightContainer.appendChild(historySelect);
  bar.appendChild(itemsContainer);
  bar.appendChild(rightContainer);

  shadow.appendChild(style);
  shadow.appendChild(drawer);
  shadow.appendChild(bar);

  var currentProfile = null;
  var activePanelName = null;
  var profilesList = [];

  closeBtn.addEventListener("click", function () {
    activePanelName = null;
    drawer.classList.remove("open");
    renderBar();
  });

  historySelect.addEventListener("change", function () {
    var id = historySelect.value;
    if (id) loadProfile(id);
  });

  function renderPanel(section) {
    drawerTitle.textContent = (section.panel && section.panel.title) || section.name;
    drawerBody.textContent = "";
    var widgets = (section.panel && section.panel.widgets) || [];
    widgets.forEach(function (w) {
      if (w.kind === "metrics" && w.metrics) {
        var mWrap = document.createElement("div");
        mWrap.className = "metrics";
        w.metrics.forEach(function (m) {
          var card = document.createElement("div");
          card.className = "metric";
          var lbl = document.createElement("div");
          lbl.className = "metric-label";
          lbl.textContent = m.label;
          var val = document.createElement("div");
          val.className = "metric-val";
          val.textContent = m.value;
          card.appendChild(lbl);
          card.appendChild(val);
          mWrap.appendChild(card);
        });
        drawerBody.appendChild(mWrap);
      } else if (w.kind === "key_value" && w.pairs) {
        var tbl = document.createElement("table");
        Object.keys(w.pairs).forEach(function (k) {
          var tr = document.createElement("tr");
          var th = document.createElement("th");
          th.textContent = k;
          var td = document.createElement("td");
          td.textContent = w.pairs[k];
          tr.appendChild(th);
          tr.appendChild(td);
          tbl.appendChild(tr);
        });
        drawerBody.appendChild(tbl);
      } else if (w.kind === "table" && w.rows) {
        var rTbl = document.createElement("table");
        if (w.columns && w.columns.length) {
          var thead = document.createElement("tr");
          w.columns.forEach(function (h) {
            var th = document.createElement("th");
            th.textContent = h;
            thead.appendChild(th);
          });
          rTbl.appendChild(thead);
        }
        w.rows.forEach(function (r) {
          var tr = document.createElement("tr");
          r.forEach(function (c) {
            var td = document.createElement("td");
            td.textContent = c;
            tr.appendChild(td);
          });
          rTbl.appendChild(tr);
        });
        drawerBody.appendChild(rTbl);
      }
    });
  }

  function renderBar() {
    itemsContainer.textContent = "";
    var brand = document.createElement("span");
    brand.className = "brand";
    brand.textContent = "GoDebugBar";
    itemsContainer.appendChild(brand);

    if (!currentProfile || !currentProfile.sections) return;
    currentProfile.sections.forEach(function (sec) {
      var item = sec.toolbar || {};
      var badge = document.createElement("div");
      badge.className = "badge " + (item.color || "default");
      if (activePanelName === sec.name) badge.className += " active";
      badge.textContent = item.label + ": " + item.value;
      badge.addEventListener("click", function () {
        if (activePanelName === sec.name) {
          activePanelName = null;
          drawer.classList.remove("open");
        } else {
          activePanelName = sec.name;
          renderPanel(sec);
          drawer.classList.add("open");
        }
        renderBar();
      });
      itemsContainer.appendChild(badge);
    });
  }

  function updateHistorySelect(selectedId) {
    historySelect.textContent = "";
    profilesList.forEach(function (p) {
      var opt = document.createElement("option");
      opt.value = p.id;
      opt.textContent = p.method + " " + p.path + " (" + p.status + ")";
      if (p.id === selectedId) opt.selected = true;
      historySelect.appendChild(opt);
    });
  }

  var origFetch = window.fetch;
  function loadProfile(id) {
    origFetch(baseUrl + "/api/profiles/" + encodeURIComponent(id))
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (data) {
        if (!data) return;
        currentProfile = data;
        var exists = profilesList.some(function (p) { return p.id === data.id; });
        if (!exists) profilesList.unshift(data);
        updateHistorySelect(data.id);
        renderBar();
        if (activePanelName) {
          var found = (data.sections || []).find(function (s) { return s.name === activePanelName; });
          if (found) renderPanel(found);
        }
      })
      .catch(function () {});
  }

  function refreshList(preferId) {
    origFetch(baseUrl + "/api/profiles")
      .then(function (r) { return r.ok ? r.json() : []; })
      .then(function (list) {
        if (Array.isArray(list)) {
          profilesList = list;
          var target = preferId || (list[0] && list[0].id);
          if (target) loadProfile(target);
        }
      })
      .catch(function () {});
  }

  window.fetch = function () {
    var args = arguments;
    var url = typeof args[0] === "string" ? args[0] : (args[0] && args[0].url) || "";
    return origFetch.apply(this, args).then(function (res) {
      if (url.indexOf(baseUrl) === -1) {
        var pid = res.headers && res.headers.get("X-Debugbar-ID");
        if (pid) loadProfile(pid);
      }
      return res;
    });
  };

  refreshList(initialId);
})();
