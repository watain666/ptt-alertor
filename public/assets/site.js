(() => {
  'use strict';

  const root = document.documentElement;
  const themeButtons = Array.from(document.querySelectorAll('[data-theme-choice]'));
  function applyTheme(theme) {
    root.dataset.theme = theme;
    themeButtons.forEach(button => button.setAttribute('aria-pressed', String(button.dataset.themeChoice === theme)));
    document.querySelector('meta[name="theme-color"]').content = theme === 'dark' ? '#0a0a0a' : '#ffffff';
  }
  applyTheme(root.dataset.theme || 'light');
  themeButtons.forEach(button => button.addEventListener('click', () => {
    const theme = button.dataset.themeChoice;
    applyTheme(theme);
    try { localStorage.setItem('ptt-alertor-theme', theme); } catch (_) {}
  }));
  window.addEventListener('storage', event => {
    if (event.key !== 'ptt-alertor-theme' && event.key !== null) return;
    applyTheme(event.newValue === 'dark' ? 'dark' : 'light');
  });

  function updateNavigation() {
    const page = window.location.pathname;
    let active = page === '/docs' ? (window.location.hash === '#self-hosting' ? 'self-hosting' : 'docs') : page === '/top' ? 'top' : 'overview';
    if (page === '/' || page === '/telegram') {
      ['features', 'how-it-works'].forEach(id => {
        const section = document.getElementById(id);
        if (section && section.getBoundingClientRect().top <= 150) active = id;
      });
    }
    document.querySelectorAll('[data-nav]').forEach(link => {
      if (link.dataset.nav === active) link.setAttribute('aria-current', 'location');
      else link.removeAttribute('aria-current');
    });
  }
  let navigationPending = false;
  window.addEventListener('scroll', () => {
    if (navigationPending) return;
    navigationPending = true;
    window.requestAnimationFrame(() => { updateNavigation(); navigationPending = false; });
  }, { passive: true });
  window.addEventListener('hashchange', updateNavigation);
  updateNavigation();

  const toast = document.getElementById('feedback-toast');
  let toastTimer;
  function notify(message) {
    if (!toast) return;
    clearTimeout(toastTimer);
    toast.textContent = message;
    toast.hidden = false;
    toastTimer = setTimeout(() => { toast.hidden = true; }, 4200);
  }

  async function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(text);
        return;
      } catch (_) {
        // Older browsers and denied clipboard permissions use the selection fallback.
      }
    }
    const field = document.createElement('textarea');
    field.value = text;
    field.setAttribute('readonly', '');
    field.style.cssText = 'position:fixed;top:0;left:-9999px';
    document.body.appendChild(field);
    const previousFocus = document.activeElement;
    field.select();
    try {
      if (!document.execCommand('copy')) throw new Error('Clipboard unavailable');
    } finally {
      field.remove();
      if (previousFocus) previousFocus.focus({ preventScroll: true });
    }
  }

  const menuToggle = document.querySelector('.menu-toggle');
  const mobileNav = document.getElementById('mobile-nav');
  function closeMenu() {
    if (!menuToggle || !mobileNav) return;
    mobileNav.hidden = true;
    menuToggle.setAttribute('aria-expanded', 'false');
    menuToggle.setAttribute('aria-label', '開啟導覽選單');
  }
  if (menuToggle && mobileNav) {
    menuToggle.addEventListener('click', () => {
      const expanded = menuToggle.getAttribute('aria-expanded') !== 'true';
      menuToggle.setAttribute('aria-expanded', String(expanded));
      menuToggle.setAttribute('aria-label', expanded ? '關閉導覽選單' : '開啟導覽選單');
      mobileNav.hidden = !expanded;
    });
    mobileNav.addEventListener('click', event => {
      if (event.target.closest('a')) closeMenu();
    });
    document.addEventListener('keydown', event => {
      if (event.key === 'Escape' && !mobileNav.hidden) {
        closeMenu();
        menuToggle.focus();
      }
    });
    window.matchMedia('(min-width: 961px)').addEventListener('change', closeMenu);
  }

  const form = document.getElementById('subscription-form');
  if (form) {
    const board = document.getElementById('board-input');
    const condition = document.getElementById('condition-input');
    const output = document.getElementById('generated-command');
    const tabs = Array.from(form.querySelectorAll('[role="tab"]'));
    const panel = document.getElementById('subscription-panel');
    const modes = {
      keyword: { command: '新增', label: '關鍵字', english: 'KEYWORD', type: 'text', value: '咖啡', placeholder: '例如：咖啡', hint: '想追蹤多個看板或關鍵字？用半形逗號「,」分隔即可。' },
      author: { command: '新增作者', label: '作者帳號', english: 'AUTHOR', type: 'text', value: 'ffaarr', placeholder: '例如：ffaarr', hint: '填入 PTT 作者帳號，多位作者可用半形逗號「,」分隔。' },
      push: { command: '新增推文數', label: '推文門檻', english: 'PUSH', type: 'number', value: '50', placeholder: '1–100', hint: '設定 1–100 的推文門檻；設為 0 則取消此看板的推文數通知。' },
      comment: { command: '新增推文', label: '文章網址', english: 'ARTICLE URL', type: 'url', value: '', placeholder: 'https://www.ptt.cc/bbs/看板/M.….html', hint: '貼上 PTT 文章的完整網址，追蹤這篇文章的後續推文。' }
    };
    const values = Object.fromEntries(Object.entries(modes).map(([key, mode]) => [key, mode.value]));
    let activeMode = 'keyword';

    function updateCommand() {
      const value = condition.value.trim();
      let error = '';
      if (activeMode === 'keyword' && value.length === 0) error = '請輸入想追蹤的關鍵字。';
      if (activeMode === 'author' && !/^[A-Za-z0-9_]+(?:,[A-Za-z0-9_]+)*$/.test(value)) error = '請使用英文作者帳號，多位作者用半形逗號分隔。';
      if (activeMode === 'comment' && !/^https?:\/\/www\.ptt\.cc\/bbs\/[\w-]+\/M\.\d+\.A\.\w+\.html$/.test(value)) error = '請貼上完整的 PTT 文章網址，例如 https://www.ptt.cc/bbs/EZsoft/M.1708247900.A.27C.html';
      condition.setCustomValidity(error);
      const parts = [modes[activeMode].command];
      if (activeMode !== 'comment') parts.push(board.value.trim() || '看板');
      parts.push(value || (activeMode === 'comment' ? '文章網址' : modes[activeMode].label));
      output.value = parts.join(' ');
    }

    function selectMode(mode) {
      values[activeMode] = condition.value;
      activeMode = mode;
      const settings = modes[mode];
      tabs.forEach(tab => {
        const selected = tab.dataset.mode === mode;
        tab.setAttribute('aria-selected', String(selected));
        tab.tabIndex = selected ? 0 : -1;
      });
      panel.setAttribute('aria-labelledby', `tab-${mode}`);
      document.getElementById('condition-label').textContent = settings.label;
      document.getElementById('condition-label-en').textContent = settings.english;
      document.getElementById('condition-hint').textContent = settings.hint;
      document.getElementById('board-field').hidden = mode === 'comment';
      board.disabled = mode === 'comment';
      condition.type = settings.type;
      condition.value = values[mode];
      condition.placeholder = settings.placeholder;
      if (mode === 'push') {
        condition.min = '0';
        condition.max = '100';
        condition.step = '1';
      } else {
        ['min', 'max', 'step'].forEach(name => condition.removeAttribute(name));
      }
      updateCommand();
    }

    tabs.forEach((tab, index) => {
      tab.addEventListener('click', () => selectMode(tab.dataset.mode));
      tab.addEventListener('keydown', event => {
        let next;
        if (event.key === 'ArrowRight') next = (index + 1) % tabs.length;
        if (event.key === 'ArrowLeft') next = (index + tabs.length - 1) % tabs.length;
        if (event.key === 'Home') next = 0;
        if (event.key === 'End') next = tabs.length - 1;
        if (next === undefined) return;
        event.preventDefault();
        selectMode(tabs[next].dataset.mode);
        tabs[next].focus();
      });
    });
    form.addEventListener('input', updateCommand);
    form.addEventListener('submit', async event => {
      event.preventDefault();
      updateCommand();
      if (!form.reportValidity()) return;
      try {
        await copyText(output.value);
        notify('已複製指令，貼到 Telegram 傳給機器人就可以囉！');
      } catch (_) {
        const selection = window.getSelection();
        const range = document.createRange();
        range.selectNodeContents(output);
        selection.removeAllRanges();
        selection.addRange(range);
        notify('無法自動複製，已選取指令，請手動複製。');
      }
    });
    document.querySelectorAll('[data-preset]').forEach(button => {
      button.addEventListener('click', () => {
        const [boardName, keyword] = button.dataset.preset.split('|');
        selectMode('keyword');
        board.value = boardName;
        condition.value = keyword;
        updateCommand();
        document.getElementById('command-builder').scrollIntoView({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth', block: 'center' });
        condition.focus({ preventScroll: true });
      });
    });
    updateCommand();
  }

  document.querySelectorAll('[data-copy-code]').forEach(button => {
    button.addEventListener('click', async () => {
      const code = document.getElementById(button.dataset.copyCode);
      if (!code) return;
      try { await copyText(code.textContent); notify('已複製指令。'); }
      catch (_) { notify('無法自動複製，請選取程式碼手動複製。'); }
    });
  });

  // Keep the existing documentation and ranking tabs working without CDN scripts.
  const legacyTabs = Array.from(document.querySelectorAll('.nav-tabs [role="tab"]'));
  function selectLegacyTab(tab) {
    legacyTabs.forEach(item => {
      const selected = item === tab;
      item.parentElement.classList.toggle('active', selected);
      item.setAttribute('aria-selected', String(selected));
      item.tabIndex = selected ? 0 : -1;
      const target = document.getElementById(item.getAttribute('aria-controls'));
      if (target) target.classList.toggle('active', selected);
    });
  }
  if (legacyTabs.length) {
    const initial = legacyTabs.find(tab => tab.getAttribute('href') === window.location.hash) || legacyTabs[0];
    selectLegacyTab(initial);
    window.addEventListener('hashchange', () => {
      const target = legacyTabs.find(tab => tab.getAttribute('href') === window.location.hash);
      if (target) selectLegacyTab(target);
    });
    legacyTabs.forEach((tab, index) => {
      tab.addEventListener('click', event => {
        event.preventDefault();
        selectLegacyTab(tab);
        window.history.replaceState(null, '', tab.getAttribute('href'));
        updateNavigation();
      });
      tab.addEventListener('keydown', event => {
        let next;
        if (event.key === 'ArrowRight') next = (index + 1) % legacyTabs.length;
        if (event.key === 'ArrowLeft') next = (index + legacyTabs.length - 1) % legacyTabs.length;
        if (event.key === 'Home') next = 0;
        if (event.key === 'End') next = legacyTabs.length - 1;
        if (next === undefined) return;
        event.preventDefault();
        selectLegacyTab(legacyTabs[next]);
        legacyTabs[next].focus();
      });
    });
  }
  document.querySelectorAll('.table .copy').forEach(button => {
    button.addEventListener('click', async () => {
      const cells = button.closest('tr').querySelectorAll('td');
      const kind = button.closest('.tab-pane').id;
      const board = cells[0].textContent.trim();
      let word = cells[1].textContent.trim();
      let command = { keywords: '新增', authors: '新增作者', pushsum: '新增推文數' }[kind];
      if (kind === 'pushsum' && Number(word) < 0) { command = '新增噓文數'; word = String(Math.abs(Number(word))); }
      try { await copyText(`${command} ${board} ${word}`); notify('已複製指令，請貼到 Telegram 對話中。'); }
      catch (_) { notify('無法自動複製，請依照表格內容手動輸入指令。'); }
    });
  });

  const counter = document.getElementById('counter');
  const counterBoard = document.getElementById('counter-board');
  if (counter && counterBoard && 'WebSocket' in window) {
    try {
      const configured = counterBoard.dataset.wsHost;
      const url = new URL(configured || '/ws', window.location.href);
      url.protocol = url.protocol === 'https:' || url.protocol === 'wss:' ? 'wss:' : 'ws:';
      if (window.location.protocol === 'https:') url.protocol = 'wss:';
      if (url.pathname === '/' || url.pathname === '') url.pathname = '/ws';
      const ws = new WebSocket(url.href);
      ws.addEventListener('message', event => {
        const raw = String(event.data).replaceAll(',', '').trim();
        if (!/^\d+$/.test(raw)) return;
        const value = Number(raw);
        if (Number.isSafeInteger(value)) counter.textContent = value.toLocaleString('en-US');
      });
      window.addEventListener('pagehide', () => ws.close(), { once: true });
    } catch (_) {
      // The server-rendered count remains available when live updates are unavailable.
    }
  }
})();
