/* Notices enhancements.
 *
 * On a notice: "Copy" buttons on code blocks, and a link icon on each section
 *              heading so a specific answer can be shared.
 * On the index: live search over every notice (titles, summaries, categories, full
 *              text including code) and in-place filtering by category.
 *
 * All of it is progressive enhancement. Without JavaScript the notices still
 * read normally, and the category chips are ordinary links to static category pages.
 */
(function () {
  'use strict';

  // ---- code block copy buttons ----

  function addCopyButtons() {
    if (!navigator.clipboard || !window.isSecureContext) { return; }
    var blocks = document.querySelectorAll('.notice-body pre');
    for (var i = 0; i < blocks.length; i++) {
      (function (pre) {
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'copy-btn';
        btn.textContent = 'Copy';
        btn.setAttribute('aria-label', 'Copy code to clipboard');
        btn.addEventListener('click', function () {
          var code = pre.querySelector('code') || pre;
          navigator.clipboard.writeText(code.innerText.replace(/\n$/, '')).then(function () {
            flash(btn, 'Copied');
          }, function () {
            flash(btn, 'Press Ctrl+C');
          });
        });
        pre.classList.add('has-copy');
        pre.appendChild(btn);
      })(blocks[i]);
    }
  }

  function flash(btn, text) {
    btn.textContent = text;
    clearTimeout(btn._t);
    btn._t = setTimeout(function () { btn.textContent = 'Copy'; }, 1800);
  }

  // ---- heading anchor links ----

  function addHeadingAnchors() {
    var heads = document.querySelectorAll('.notice-body h2[id], .notice-body h3[id]');
    for (var i = 0; i < heads.length; i++) {
      var a = document.createElement('a');
      a.className = 'heading-anchor';
      a.href = '#' + heads[i].id;
      a.setAttribute('aria-label', 'Link to this section: ' + heads[i].textContent);
      a.textContent = '#';
      heads[i].appendChild(a);
    }
  }

  // ---- index search and category filter ----

  function setupIndex() {
    var data = document.getElementById('notices-index');
    var input = document.getElementById('notices-search');
    var grid = document.getElementById('notices-grid');
    if (!data || !input || !grid) { return; }

    var notices;
    try { notices = JSON.parse(data.textContent); } catch (e) { return; }

    var bySlug = {};
    for (var i = 0; i < notices.length; i++) {
      var n = notices[i];
      n.haystack = [n.title, n.summary, n.categories.join(' '), n.text].join(' ').toLowerCase();
      bySlug[n.slug] = n;
    }

    var cards = grid.querySelectorAll('.notice-card');
    var chips = document.querySelectorAll('.category-list .category-chip[data-category]');
    var status = document.getElementById('notices-status');
    var activeCategory = '';

    function apply() {
      var words = input.value.toLowerCase().split(/\s+/).filter(Boolean);
      var shown = 0;
      for (var i = 0; i < cards.length; i++) {
        var n = bySlug[cards[i].getAttribute('data-slug')];
        var ok = !!n;
        if (ok && activeCategory) { ok = n.categories.indexOf(activeCategory) !== -1; }
        for (var w = 0; ok && w < words.length; w++) {
          ok = n.haystack.indexOf(words[w]) !== -1;
        }
        cards[i].hidden = !ok;
        if (ok) { shown++; }
      }
      for (var c = 0; c < chips.length; c++) {
        var on = chips[c].getAttribute('data-category') === activeCategory;
        if (on) { chips[c].setAttribute('aria-current', 'true'); } else { chips[c].removeAttribute('aria-current'); }
      }
      if (status) {
        status.textContent = (words.length || activeCategory)
          ? shown + ' of ' + cards.length + ' notices' + (activeCategory ? ' in ' + activeCategory : '')
          : '';
      }
    }

    input.addEventListener('input', apply);
    for (var c = 0; c < chips.length; c++) {
      chips[c].addEventListener('click', function (ev) {
        ev.preventDefault();
        activeCategory = this.getAttribute('data-category');
        apply();
      });
    }
  }

  function start() {
    addCopyButtons();
    addHeadingAnchors();
    setupIndex();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', start);
  } else {
    start();
  }
})();
