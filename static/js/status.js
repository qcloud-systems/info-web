/* Upstream service status, with a 90-day bar per provider.
 *
 * Two adapters:
 *   statuspage  /api/v2/status.json     current indicator
 *               /api/v2/incidents.json  recent incidents -> daily bars
 *   instatus    /summary.json           current state only, no history
 *
 * Both are served with permissive CORS headers, so this runs entirely in the
 * browser with no proxy and no credentials.
 *
 * On the bars: Statuspage publishes no uptime-history API -- the bars on
 * their own pages come from internal data. These are reconstructed from
 * reported incidents, so a day is red because the provider *posted* an
 * incident covering it. An outage nobody posted shows green. That limit is
 * stated on the page itself; do not present these as measured uptime.
 *
 * Everything is best-effort. The markup renders complete and readable before
 * this runs, and any failure -- offline, blocked, CORS change, slow provider,
 * unexpected payload -- leaves a neutral row rather than an error.
 */
(function () {
  'use strict';

  var DAYS = 90;
  var TIMEOUT_MS = 8000;
  var DAY_MS = 86400000;

  // Statuspage top-level indicator -> presentation.
  var INDICATOR = {
    none: { state: 'ok', text: 'Operational' },
    minor: { state: 'warn', text: 'Minor issue' },
    major: { state: 'bad', text: 'Major outage' },
    critical: { state: 'bad', text: 'Critical outage' },
    maintenance: { state: 'info', text: 'Maintenance' }
  };

  // Instatus page.status -> presentation.
  var INSTATUS = {
    UP: { state: 'ok', text: 'Operational' },
    HASISSUES: { state: 'bad', text: 'Service issues' },
    UNDERMAINTENANCE: { state: 'info', text: 'Maintenance' }
  };

  // Per-incident impact -> bar state. Ranked so a worse incident on the same
  // day wins.
  var IMPACT = {
    none: null,
    maintenance: null,
    minor: 'warn',
    major: 'bad',
    critical: 'bad'
  };
  var RANK = { unknown: 0, ok: 1, warn: 2, bad: 3 };

  function fetchJSON(url) {
    var controller = null;
    var timer = null;
    try {
      controller = new AbortController();
      timer = setTimeout(function () {
        controller.abort();
      }, TIMEOUT_MS);
    } catch (e) {
      controller = null;
    }

    var opts = { cache: 'no-store' };
    if (controller) {
      opts.signal = controller.signal;
    }

    return fetch(url, opts)
      .then(function (res) {
        if (!res.ok) {
          throw new Error('HTTP ' + res.status);
        }
        return res.json();
      })
      .then(
        function (body) {
          if (timer) { clearTimeout(timer); }
          return body;
        },
        function (err) {
          if (timer) { clearTimeout(timer); }
          throw err;
        }
      );
  }

  // UTC midnight for "n days before today", so every visitor sees the same
  // buckets regardless of local timezone.
  function dayStart(offsetFromToday) {
    var now = new Date();
    var utc = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
    return utc - offsetFromToday * DAY_MS;
  }

  function buildBars(container) {
    var frag = document.createDocumentFragment();
    var bars = [];
    for (var i = DAYS - 1; i >= 0; i--) {
      var bar = document.createElement('span');
      bar.className = 'status-bar';
      bar.setAttribute('data-state', 'unknown');
      var d = new Date(dayStart(i));
      bar.title = d.toISOString().slice(0, 10) + ': no data';
      frag.appendChild(bar);
      bars.push({ el: bar, start: dayStart(i), label: d.toISOString().slice(0, 10) });
    }
    container.appendChild(frag);
    return bars;
  }

  function setBar(entry, state, note) {
    if (RANK[state] <= RANK[entry.el.getAttribute('data-state')]) {
      return;
    }
    entry.el.setAttribute('data-state', state);
    entry.el.title = entry.label + ': ' + note;
  }

  function paintHead(item, state, text) {
    var dot = item.querySelector('.status-dot');
    var label = item.querySelector('.status-text');
    if (dot) { dot.setAttribute('data-state', state); }
    if (label) { label.textContent = text; }
  }

  // Mark every day an incident overlapped, worst impact winning.
  function applyIncidents(bars, incidents) {
    if (!Array.isArray(incidents)) {
      return;
    }
    var now = Date.now();
    for (var i = 0; i < incidents.length; i++) {
      var inc = incidents[i];
      if (!inc || !inc.created_at) { continue; }

      var state = IMPACT[inc.impact];
      if (!state) { continue; }

      var from = Date.parse(inc.created_at);
      var to = inc.resolved_at ? Date.parse(inc.resolved_at) : now;
      if (isNaN(from)) { continue; }
      if (isNaN(to) || to < from) { to = from; }

      for (var b = 0; b < bars.length; b++) {
        var dayFrom = bars[b].start;
        var dayTo = dayFrom + DAY_MS;
        if (from < dayTo && to >= dayFrom) {
          setBar(bars[b], state, (inc.name || 'Incident') + ' (' + inc.impact + ')');
        }
      }
    }
  }

  function markClear(bars) {
    for (var i = 0; i < bars.length; i++) {
      if (bars[i].el.getAttribute('data-state') === 'unknown') {
        setBar(bars[i], 'ok', 'no reported incident');
      }
    }
  }

  function handleStatuspage(item, bars) {
    var api = item.getAttribute('data-api');
    var incidentsURL = item.getAttribute('data-incidents');

    fetchJSON(api)
      .then(function (body) {
        var indicator = body && body.status && body.status.indicator;
        var mapped = INDICATOR[indicator];
        if (!mapped) {
          paintHead(item, 'unknown', 'Unknown');
          return;
        }
        var description = body.status.description;
        var text = mapped.text;
        if (indicator !== 'none' && typeof description === 'string' && description.length <= 40) {
          text = description;
        }
        paintHead(item, mapped.state, text);
      })
      .catch(function () {
        paintHead(item, 'unknown', 'Unavailable');
      });

    if (!incidentsURL) {
      return;
    }
    fetchJSON(incidentsURL)
      .then(function (body) {
        applyIncidents(bars, body && body.incidents);
        markClear(bars);
      })
      .catch(function () {
        /* bars stay in the neutral "no data" state */
      });
  }

  function handleInstatus(item) {
    fetchJSON(item.getAttribute('data-api'))
      .then(function (body) {
        var raw = body && body.page && body.page.status;
        var mapped = INSTATUS[raw];
        if (!mapped) {
          paintHead(item, 'unknown', 'Unknown');
          return;
        }
        paintHead(item, mapped.state, mapped.text);
      })
      .catch(function () {
        paintHead(item, 'unknown', 'Unavailable');
      });
  }

  function start() {
    var items = document.querySelectorAll('.status-item[data-api]');
    for (var i = 0; i < items.length; i++) {
      var item = items[i];
      var barHost = item.querySelector('.status-bars');
      var bars = barHost ? buildBars(barHost) : [];

      if (item.getAttribute('data-kind') === 'instatus') {
        // No public history; the bars stay neutral by design.
        if (barHost) { barHost.setAttribute('data-history', 'none'); }
        handleInstatus(item);
      } else {
        handleStatuspage(item, bars);
      }
    }
  }

  if (typeof fetch !== 'function' || typeof Promise !== 'function') {
    return;
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', start);
  } else {
    start();
  }
})();
