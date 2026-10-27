package jobs

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/models/pushsum"
)

// Telegram notifications are sent with HTML parse mode.
// Each tag must open and close on the same line, see telegram.SendHTMLMessage.

func (c Checker) HTML() string {
	title, label := "🔔 <b>新文章通知</b>", "關鍵字"
	if c.author != "" {
		title, label = "✍️ <b>作者新文章</b>", "作者"
	}
	var b strings.Builder
	b.WriteString(title + " " + boardHashtag(c.board) + "\n\n")
	b.WriteString(htmlField("看板", c.board))
	b.WriteString(htmlField(label, c.word))
	b.WriteString("\n")
	for _, a := range c.articles {
		b.WriteString("• " + articleLink(a) + "\n")
	}
	return b.String()
}

func (psc pushSumChecker) HTML() string {
	title, label := "🔥 <b>推文數通知</b>", "推文數門檻"
	if psc.subType == "pushdown" {
		title, label = "💢 <b>噓文數通知</b>", "噓文數門檻"
	}
	var b strings.Builder
	b.WriteString(title + " " + boardHashtag(psc.board) + "\n\n")
	b.WriteString(htmlField("看板", psc.board))
	b.WriteString(htmlField(label, psc.word))
	b.WriteString("\n")
	for _, a := range psc.articles {
		b.WriteString("• <code>" + html.EscapeString(pushSumText(a.PushSum)) + "</code> " + articleLink(a) + "\n")
	}
	return b.String()
}

func (cc commentChecker) HTML() string {
	var b strings.Builder
	b.WriteString("💬 <b>推文更新</b> " + boardHashtag(cc.Article.Board) + "\n\n")
	b.WriteString(articleLink(cc.Article) + "\n\n")
	for _, cmt := range cc.Article.Comments {
		b.WriteString(fmt.Sprintf("<b>%s</b> <code>%s</code>%s\n",
			html.EscapeString(cmt.Tag), html.EscapeString(cmt.UserID), html.EscapeString(cmt.Content)))
	}
	return b.String()
}

func (bc Broadcaster) HTML() string {
	return html.EscapeString(bc.Msg)
}

func htmlField(label, value string) string {
	return "<b>" + label + "：</b><code>" + html.EscapeString(value) + "</code>\n"
}

func articleLink(a article.Article) string {
	return `<a href="` + html.EscapeString(a.Link) + `">` + html.EscapeString(a.Title) + "</a>"
}

// boardHashtag turns a board name into a Telegram hashtag, which only allows
// letters, digits and underscores.
func boardHashtag(board string) string {
	return "#" + strings.Map(func(r rune) rune {
		if r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			return r
		}
		return '_'
	}, board)
}

func pushSumText(sum int) string {
	if text, ok := pushsum.NumTextMap[sum]; ok {
		return text
	}
	return strconv.Itoa(sum)
}
