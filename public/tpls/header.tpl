{{define "header"}}
<a class="skip-link" href="#main-content">跳至主要內容</a>
<header class="site-header">
    <a href="/" class="brand" aria-label="Ptt Alertor 首頁"><span class="brand-mark" aria-hidden="true"><svg viewBox="0 0 32 32"><path d="M8 5H3v22h5M24 5h5v22h-5M11 22l5-12 5 12M13 18h6"/></svg></span><span>Ptt Alertor<small>Stay in the loop.</small></span></a>
    <nav aria-label="主要導覽" class="desktop-nav">
        <span class="nav-label">EXPLORE</span>
        <a href="/" data-nav="overview">總覽 <span>↗</span></a>
        <a href="/#features" data-nav="features">功能介紹 <span>↗</span></a>
        <a href="/#how-it-works" data-nav="how-it-works">開始使用 <span>↗</span></a>
        <span class="nav-label nav-label-resources">RESOURCES</span>
        <a href="/docs" data-nav="docs">使用文件 <span>↗</span></a>
        <a href="/docs#self-hosting" data-nav="self-hosting">Docker 自架 <span>↗</span></a>
        <a href="/top" data-nav="top">熱門排行 <span>↗</span></a>
    </nav>
    <div class="sidebar-bottom">
        <a class="sidebar-github" href="https://github.com/watain666/ptt-alertor" target="_blank" rel="noopener noreferrer">Open source on GitHub <span aria-hidden="true">↗</span></a>
        <div class="theme-control" role="group" aria-label="網站主題">
            <button type="button" data-theme-choice="light" aria-label="切換白色主題" aria-pressed="true"><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M5 5l1.5 1.5M17.5 17.5 19 19M5 19l1.5-1.5M17.5 6.5 19 5"/></svg></button>
            <button type="button" data-theme-choice="dark" aria-label="切換黑色主題" aria-pressed="false"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20.5 14a9 9 0 0 1-10.5-10.5A9 9 0 1 0 20.5 14Z"/></svg></button>
        </div>
    </div>
    <button type="button" class="menu-toggle" aria-label="開啟導覽選單" aria-expanded="false" aria-controls="mobile-nav"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 8h16M4 16h16"/></svg></button>
    <nav id="mobile-nav" class="mobile-nav" aria-label="行動版導覽" hidden><a href="/">總覽</a><a href="/#features">功能介紹</a><a href="/#how-it-works">開始使用</a><a href="/docs">使用文件</a><a href="/docs#self-hosting">Docker 自架</a><a href="/top">熱門排行</a></nav>
</header>
{{end}}
