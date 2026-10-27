package telegram

import (
	"regexp"
	"strings"
	"testing"

	"github.com/watain666/ptt-alertor/models/subscription"
)

var tagPattern = regexp.MustCompile(`</?([a-z]+)[^>]*>`)

// assertLineBalancedHTML checks every tag closes on its own line, so
// SendHTMLMessage can split long views by line break.
func assertLineBalancedHTML(t *testing.T, view string) {
	t.Helper()
	for _, line := range strings.Split(view, "\n") {
		var open []string
		for _, m := range tagPattern.FindAllStringSubmatch(line, -1) {
			if strings.HasPrefix(m[0], "</") {
				if len(open) == 0 || open[len(open)-1] != m[1] {
					t.Fatalf("unbalanced closing tag %s in line %q", m[0], line)
				}
				open = open[:len(open)-1]
				continue
			}
			open = append(open, m[1])
		}
		if len(open) != 0 {
			t.Fatalf("unclosed tags %v in line %q", open, line)
		}
	}
}

func TestStaticViews(t *testing.T) {
	for name, view := range map[string]string{
		"start": startView(),
		"help":  helpView(),
		"add":   editMenuHelp("add"),
		"del":   editMenuHelp("del"),
	} {
		t.Run(name, func(t *testing.T) {
			assertLineBalancedHTML(t, view)
			if strings.Contains(view, "giphy") {
				t.Error("view still links the removed demo GIF")
			}
		})
	}
	if !strings.Contains(editMenuHelp("del"), "<code>刪除推文 "+exampleArticle+"</code>") {
		t.Error("del help is missing the comment example")
	}
}

func TestListView(t *testing.T) {
	subs := subscription.Subscriptions{
		{Board: "stock", Authors: []string{"tccred", "test520"}, PushSum: subscription.PushSum{Up: 10, Down: 20}},
		{Board: "Lifeismoney", Keywords: []string{"優惠", "<咖啡>"}, Articles: []string{"M.1.A.1", "M.2.A.2"}},
	}
	want := "📋 <b>追蹤清單</b>\n" +
		"\n<b>🔍 關鍵字</b>\n" +
		"<b>Lifeismoney</b>：<code>&lt;咖啡&gt;</code>、<code>優惠</code>\n" +
		"\n<b>✍️ 作者</b>\n" +
		"<b>stock</b>：<code>tccred</code>、<code>test520</code>\n" +
		"\n<b>🔥 推噓文數</b>\n" +
		"<b>stock</b>：推 <code>10</code>・噓 <code>20</code>\n" +
		"\n💬 另有 2 篇推文追蹤，傳送 <code>推文清單</code> 查看"
	got := listView(subs)
	if got != want {
		t.Errorf("listView() =\n%s\nwant\n%s", got, want)
	}
	assertLineBalancedHTML(t, got)
	if subs[0].Board != "stock" {
		t.Error("listView sorted the caller's subscriptions in place")
	}

	if got := listView(nil); !strings.Contains(got, "尚未建立清單") {
		t.Errorf("empty listView() = %q", got)
	}
}

func TestCommentListView(t *testing.T) {
	subs := subscription.Subscriptions{
		{Board: "EZsoft", Articles: []string{"M.2.A.2", "M.1.A.1"}},
		{Board: "Gossiping", Keywords: []string{"問卦"}},
	}
	want := "💬 <b>推文追蹤清單</b>（上限 50 篇）\n" +
		"\n<b>EZsoft</b>\n" +
		`• <a href="https://www.ptt.cc/bbs/EZsoft/M.1.A.1.html">M.1.A.1</a>` + "\n" +
		`• <a href="https://www.ptt.cc/bbs/EZsoft/M.2.A.2.html">M.2.A.2</a>` + "\n" +
		"\n傳送 <code>清理推文</code> 可移除已刪除的文章。"
	got := commentListView(subs)
	if got != want {
		t.Errorf("commentListView() =\n%s\nwant\n%s", got, want)
	}
	assertLineBalancedHTML(t, got)

	if got := commentListView(subs[1:]); !strings.Contains(got, "尚未追蹤任何文章") {
		t.Errorf("empty commentListView() = %q", got)
	}
}

func TestRankingView(t *testing.T) {
	got := rankingView(
		[]string{"stock:台積電", "Lifeismoney:regexp:咖啡|茶", "joke:笑話", "movie:心得"},
		[]string{"stock:yinglinga"},
		[]string{"stock:10", "stock:-10"},
	)
	want := "🏆 <b>熱門追蹤排行</b>\n" +
		"\n<b>🔍 關鍵字</b>\n" +
		"🥇 <b>stock</b> <code>台積電</code>\n" +
		"🥈 <b>Lifeismoney</b> <code>regexp:咖啡|茶</code>\n" +
		"🥉 <b>joke</b> <code>笑話</code>\n" +
		"\n<b>✍️ 作者</b>\n" +
		"🥇 <b>stock</b> <code>yinglinga</code>\n" +
		"\n<b>🔥 推噓文數</b>\n" +
		"🥇 <b>stock</b> 推 <code>10</code>\n" +
		"🥈 <b>stock</b> 噓 <code>10</code>\n" +
		"\n📊 <a href=\"https://ptt-alertor.tiaui.co/top\">查看完整 TOP 100</a>"
	if got != want {
		t.Errorf("rankingView() =\n%s\nwant\n%s", got, want)
	}
	assertLineBalancedHTML(t, got)
}
