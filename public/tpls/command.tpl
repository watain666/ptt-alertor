{{define "command"}}
<article class="markdown-body">
    <h2>指令</h2>

    <h3>關鍵字相關</h3>
    <ul class="list-disc">
        <li><code>新增 看板 關鍵字</code>：新增看板關鍵字。</li>
        <li><code>刪除 看板 關鍵字</code>：刪除看板關鍵字。</li>
    </ul>
    <p>範例： <code>新增 gossiping,movie 金城武,結衣</code></p>

    <h3>作者相關</h3>
    <ul class="list-disc">
        <li><code>新增作者 看板 作者</code>：新增看板作者。</li>
        <li><code>刪除作者 看板 作者</code>：刪除看板作者。</li>
    </ul>
    <p>範例：<code>新增作者 gossiping ffaarr,obov</code></p>

    <h3>推文數相關</h3>
    <ul class="list-disc">
        <li><code>新增推文數 看板 總數</code>：新增看板推文數</li>
        <li><code>新增噓文數 看板 總數</code>：新增看板噓文數</li>
    </ul>
    <p>範例：<code>新增推文數 beauty,joke 10</code></p>

    <h3>推文追蹤</h3>
    <ul class="list-disc">
        <li><code>新增推文 https://www.ptt.cc/bbs/EZsoft/M.1497363598.A.74E.html</code></li>
    </ul>

    <h3>一般</h3>
    <ul class="list-disc">
        <li><code>指令</code>：可使用的指令清單</li>
        <li><code>清單</code>：設定的看板、關鍵字、作者</li>
        <li><code>排行</code>：前五名追蹤的關鍵字、作者</li>
    </ul>
</article>
{{end}}