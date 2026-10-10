/* Engineering Notes enhancements.
 *
 * On a note:   "Copy" buttons on code blocks, and a link icon on each section
 *              heading so a specific answer can be shared.
 * On the index: live search over every note (titles, summaries, tags, full
 *              text including code) and in-place filtering by tag.
 *
 * All of it is progressive enhancement. Without JavaScript the notes still
 * read normally, and the tag chips are ordinary links to static tag pages.
 */
(function () {
  'use strict';

  // ---- code block copy buttons ----

  function addCopyButtons() {
    if (!navigator.clipboard || !window.isSecureContext) { return; }
    var blocks = document.querySelectorAll('.note-body pre');
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
    var heads = document.querySelectorAll('.note-body h2[id], .note-body h3[id]');
    for (var i = 0; i < heads.length; i++) {
      var a = document.createElement('a');
      a.className = 'heading-anchor';
      a.href = '#' + heads[i].id;
      a.setAttribute('aria-label', 'Link to this section: ' + heads[i].textContent);
      a.textContent = '#';
      heads[i].appendChild(a);
    }
  }

  // ---- index search and tag filter ----

  function setupIndex() {
    var data = document.getElementById('notes-index');
    var input = document.getElementById('notes-search');
    var grid = document.getElementById('notes-grid');
    if (!data || !input || !grid) { return; }

    var notes;
    try { notes = JSON.parse(data.textContent); } catch (e) { return; }

    var bySlug = {};
    for (var i = 0; i < notes.length; i++) {
      var n = notes[i];
      n.haystack = [n.title, n.summary, n.tags.join(' '), n.text].join(' ').toLowerCase();
      bySlug[n.slug] = n;
    }

    var cards = grid.querySelectorAll('.note-card');
    var chips = document.querySelectorAll('.tag-list .tag-chip[data-tag]');
    var status = document.getElementById('notes-status');
    var activeTag = '';

    function apply() {
      var words = input.value.toLowerCase().split(/\s+/).filter(Boolean);
      var shown = 0;
      for (var i = 0; i < cards.length; i++) {
        var n = bySlug[cards[i].getAttribute('data-slug')];
        var ok = !!n;
        if (ok && activeTag) { ok = n.tags.indexOf(activeTag) !== -1; }
        for (var w = 0; ok && w < words.length; w++) {
          ok = n.haystack.indexOf(words[w]) !== -1;
        }
        cards[i].hidden = !ok;
        if (ok) { shown++; }
      }
      for (var c = 0; c < chips.length; c++) {
        var on = chips[c].getAttribute('data-tag') === activeTag;
        if (on) { chips[c].setAttribute('aria-current', 'true'); } else { chips[c].removeAttribute('aria-current'); }
      }
      if (status) {
        status.textContent = (words.length || activeTag)
          ? shown + ' of ' + cards.length + ' notes' + (activeTag ? ' tagged ' + activeTag : '')
          : '';
      }
    }

    input.addEventListener('input', apply);
    for (var c = 0; c < chips.length; c++) {
      chips[c].addEventListener('click', function (ev) {
        ev.preventDefault();
        activeTag = this.getAttribute('data-tag');
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
