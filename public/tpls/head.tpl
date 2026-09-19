{{define "head"}}
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Ptt Alertor — 你在意的，第一時間。</title>
    <meta name="description" content="免費追蹤 PTT 看板、關鍵字、作者與推文，透過 Telegram 接收文章通知。支援 Docker Compose 自架，讓每一則通知，由你決定。">
    <meta name="theme-color" content="#ffffff">
    <meta property="og:type" content="website">
    <meta property="og:title" content="Ptt Alertor — 你在意的，第一時間。">
    <meta property="og:description" content="從 PTT 的萬千討論中，找到你關心的事。免費、開源，直接送到你的 Telegram。">
    <script>
    (() => {
        let theme;
        try { theme = localStorage.getItem('ptt-alertor-theme'); } catch (_) {}
        if (theme !== 'light' && theme !== 'dark') theme = 'light';
        document.documentElement.dataset.theme = theme;
        document.querySelector('meta[name="theme-color"]').content = theme === 'dark' ? '#0a0a0a' : '#ffffff';
    })();
    </script>
    <link rel="icon" href="/assets/favicon.svg" type="image/svg+xml">
    <link rel="stylesheet" href="/assets/site.css?v=20260920-3">
</head>
{{end}}
