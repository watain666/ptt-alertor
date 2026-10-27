package jobs

import (
	"testing"

	"github.com/watain666/ptt-alertor/models/article"
)

func TestCheckerHTML(t *testing.T) {
	articles := article.Articles{
		{Title: "[情報] 咖啡 <買一送一> & 更多", Link: "https://www.ptt.cc/bbs/Lifeismoney/M.1.A.1.html"},
		{Title: "[情報] 第二篇", Link: "https://www.ptt.cc/bbs/Lifeismoney/M.2.A.2.html"},
	}
	for _, tt := range []struct {
		name string
		c    Checker
		want string
	}{
		{"keyword", Checker{board: "Lifeismoney", word: "咖啡", articles: articles},
			"🔔 <b>新文章通知</b> #Lifeismoney\n\n" +
				"<b>看板：</b><code>Lifeismoney</code>\n" +
				"<b>關鍵字：</b><code>咖啡</code>\n\n" +
				`• <a href="https://www.ptt.cc/bbs/Lifeismoney/M.1.A.1.html">[情報] 咖啡 &lt;買一送一&gt; &amp; 更多</a>` + "\n" +
				`• <a href="https://www.ptt.cc/bbs/Lifeismoney/M.2.A.2.html">[情報] 第二篇</a>` + "\n"},
		{"author", Checker{board: "drama-ticket", author: "abc", word: "abc", articles: articles[1:]},
			"✍️ <b>作者新文章</b> #drama_ticket\n\n" +
				"<b>看板：</b><code>drama-ticket</code>\n" +
				"<b>作者：</b><code>abc</code>\n\n" +
				`• <a href="https://www.ptt.cc/bbs/Lifeismoney/M.2.A.2.html">[情報] 第二篇</a>` + "\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.HTML(); got != tt.want {
				t.Errorf("HTML() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestPushSumCheckerHTML(t *testing.T) {
	psc := pushSumChecker{Checker: Checker{board: "Lifeismoney", word: "10", subType: "pushup", articles: article.Articles{
		{Title: "Re: [情報] 蝦皮優惠", Link: "https://www.ptt.cc/bbs/Lifeismoney/M.1.A.1.html", PushSum: 10},
		{Title: "[情報] 爆文", Link: "https://www.ptt.cc/bbs/Lifeismoney/M.2.A.2.html", PushSum: 100},
	}}}
	want := "🔥 <b>推文數通知</b> #Lifeismoney\n\n" +
		"<b>看板：</b><code>Lifeismoney</code>\n" +
		"<b>推文數門檻：</b><code>10</code>\n\n" +
		`• <code>10</code> <a href="https://www.ptt.cc/bbs/Lifeismoney/M.1.A.1.html">Re: [情報] 蝦皮優惠</a>` + "\n" +
		`• <code>爆</code> <a href="https://www.ptt.cc/bbs/Lifeismoney/M.2.A.2.html">[情報] 爆文</a>` + "\n"
	if got := psc.HTML(); got != want {
		t.Errorf("HTML() =\n%s\nwant\n%s", got, want)
	}

	psc.subType, psc.word = "pushdown", "20"
	psc.articles = article.Articles{{Title: "[問卦] 噓爆", Link: "https://www.ptt.cc/bbs/Gossiping/M.3.A.3.html", PushSum: -100}}
	want = "💢 <b>噓文數通知</b> #Lifeismoney\n\n" +
		"<b>看板：</b><code>Lifeismoney</code>\n" +
		"<b>噓文數門檻：</b><code>20</code>\n\n" +
		`• <code>XX</code> <a href="https://www.ptt.cc/bbs/Gossiping/M.3.A.3.html">[問卦] 噓爆</a>` + "\n"
	if got := psc.HTML(); got != want {
		t.Errorf("HTML() =\n%s\nwant\n%s", got, want)
	}
}

func TestCommentCheckerHTML(t *testing.T) {
	cc := commentChecker{Article: article.Article{
		Board: "Gossiping",
		Title: "[問卦] 有沒有 <b> 的八卦",
		Link:  "https://www.ptt.cc/bbs/Gossiping/M.1.A.1.html",
		Comments: article.Comments{
			{Tag: "推", UserID: "ChoDino", Content: ": 推文 <3"},
			{Tag: "→", UserID: "abc", Content: ": 補充"},
		},
	}}
	want := "💬 <b>推文更新</b> #Gossiping\n\n" +
		`<a href="https://www.ptt.cc/bbs/Gossiping/M.1.A.1.html">[問卦] 有沒有 &lt;b&gt; 的八卦</a>` + "\n\n" +
		"<b>推</b> <code>ChoDino</code>: 推文 &lt;3\n" +
		"<b>→</b> <code>abc</code>: 補充\n"
	if got := cc.HTML(); got != want {
		t.Errorf("HTML() =\n%s\nwant\n%s", got, want)
	}
}

func TestBroadcasterHTML(t *testing.T) {
	bc := Broadcaster{Msg: "維護公告 <今晚> & 明早"}
	if got, want := bc.HTML(), "維護公告 &lt;今晚&gt; &amp; 明早"; got != want {
		t.Errorf("HTML() = %q, want %q", got, want)
	}
}
