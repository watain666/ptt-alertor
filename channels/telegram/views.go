package telegram

import (
	"fmt"
	"html"
	"sort"
	"strings"

	log "github.com/Ptt-Alertor/logrus"

	"github.com/watain666/ptt-alertor/models"
	"github.com/watain666/ptt-alertor/models/subscription"
	"github.com/watain666/ptt-alertor/models/top"
)

// Menu replies are sent with HTML parse mode. Like the notifications, each
// tag must open and close on the same line, see SendHTMLMessage.

const (
	siteURL         = "https://ptt-alertor.tiaui.co"
	exampleArticle  = "https://www.ptt.cc/bbs/EZsoft/M.1708247900.A.27C.html"
	commentListSize = 50
	rankingSize     = 3
)

// menuView renders the HTML reply for menu commands by slash command name.
func menuView(name, userID string) (string, bool) {
	switch strings.ToLower(name) {
	case "start":
		return startView(), true
	case "help":
		return helpView(), true
	case "list":
		return listView(models.User().Find(userID).Subscribes), true
	case "comments":
		return commentListView(models.User().Find(userID).Subscribes), true
	case "ranking":
		return rankingView(top.ListKeywords(rankingSize), top.ListAuthors(rankingSize), top.ListPushSum(rankingSize)), true
	}
	return "", false
}

func sendMenuView(chatID int64, userID, name string) bool {
	view, ok := menuView(name, userID)
	if !ok {
		return false
	}
	log.WithFields(log.Fields{
		"account": userID,
		"command": name,
	}).Info("Command Request")
	SendHTMLMessage(chatID, view)
	return true
}

func startView() string {
	return "👋 <b>歡迎使用 Ptt Alertor</b>\n\n" +
		"PTT 有你在意的新文章，第一時間通知你。\n\n" +
		"<b>快速開始</b>\n" +
		"1. 新增追蹤：<code>新增 Lifeismoney 咖啡</code>\n" +
		"2. 查看清單：<code>清單</code>\n" +
		"3. 所有指令：<code>指令</code>\n\n" +
		"📚 使用文件：" + siteURL + "/docs"
}

func helpView() string {
	return "📖 <b>指令說明</b>\n\n" +
		"中文指令和 / 指令可以互換使用，例如 <code>新增 Lifeismoney 咖啡</code> 等於 <code>/add Lifeismoney 咖啡</code>。\n\n" +
		"<b>📋 一般</b>\n" +
		"<code>清單</code> /list 查看追蹤清單\n" +
		"<code>推文清單</code> /comments 查看推文追蹤清單\n" +
		"<code>排行</code> /ranking 熱門追蹤排行\n" +
		"<code>指令</code> /help 顯示這份說明\n" +
		"<code>顯示鍵盤</code> /showkeyboard 顯示快捷鍵盤\n" +
		"<code>隱藏鍵盤</code> /hidekeyboard 隱藏快捷鍵盤\n\n" +
		"<b>🔍 關鍵字</b>\n" +
		"<code>新增 看板 關鍵字</code> /add 新增追蹤\n" +
		"<code>刪除 看板 關鍵字</code> /del 取消追蹤\n" +
		"例：<code>新增 gossiping,movie 金城武,結衣</code>\n\n" +
		"<b>✍️ 作者</b>\n" +
		"<code>新增作者 看板 作者</code> /addauthor 新增追蹤\n" +
		"<code>刪除作者 看板 作者</code> /delauthor 取消追蹤\n" +
		"例：<code>新增作者 gossiping ffaarr,obov</code>\n\n" +
		"<b>🔥 推噓文數</b>\n" +
		"<code>新增推文數 看板 數量</code> /addpush 推文數達標時通知\n" +
		"<code>新增噓文數 看板 數量</code> /addboo 噓文數達標時通知\n" +
		"例：<code>新增推文數 joke,beauty 10</code>，數量設為 <code>0</code> 即取消\n\n" +
		"<b>💬 推文追蹤</b>\n" +
		"<code>新增推文 網址</code> /addcomment 追蹤文章的新推文\n" +
		"<code>刪除推文 網址</code> /delcomment 取消追蹤\n" +
		"<code>清理推文</code> /cleancomments 移除已刪除的文章\n\n" +
		"📚 正規表示式、指令列等進階用法，請看使用文件：\n" + siteURL + "/docs"
}

func editMenuHelp(action string) string {
	if action == "add" {
		return "➕ <b>新增追蹤</b>\n\n" +
			"關鍵字：<code>新增 Lifeismoney 咖啡</code>\n" +
			"作者：<code>新增作者 gossiping ffaarr</code>\n" +
			"推文數：<code>新增推文數 joke 10</code>\n" +
			"噓文數：<code>新增噓文數 joke 10</code>\n" +
			"推文：<code>新增推文 " + exampleArticle + "</code>\n\n" +
			"也可以用 /add、/addauthor、/addpush、/addboo、/addcomment 接同樣的參數。\n" +
			"進階參數用法：/add -h"
	}
	return "➖ <b>取消追蹤</b>\n\n" +
		"關鍵字：<code>刪除 Lifeismoney 咖啡</code>\n" +
		"作者：<code>刪除作者 gossiping ffaarr</code>\n" +
		"推文數：<code>新增推文數 joke 0</code>\n" +
		"噓文數：<code>新增噓文數 joke 0</code>\n" +
		"推文：<code>刪除推文 " + exampleArticle + "</code>\n\n" +
		"也可以用 /del、/delauthor、/addpush、/addboo、/delcomment 接同樣的參數。\n" +
		"進階參數用法：/del -h"
}

// usageView explains a command sent without its arguments.
func usageView(c botCommand) string {
	if c.slash == "add" || c.slash == "del" {
		return editMenuHelp(c.slash)
	}
	example := html.EscapeString(c.example)
	view := "<b>" + c.description + "</b>\n\n" +
		"用法：<code>" + c.chinese + " " + c.usage + "</code>\n" +
		"例：<code>" + c.chinese + " " + example + "</code>\n" +
		"也可以用 /" + c.slash + "：<code>/" + c.slash + " " + example + "</code>"
	if c.usage == "看板 數量" {
		view += "\n\n數量設為 <code>0</code> 即取消通知。"
	}
	return view
}

func listView(subs subscription.Subscriptions) string {
	subs = append(subscription.Subscriptions(nil), subs...)
	sort.Slice(subs, func(i, j int) bool { return subs[i].Board < subs[j].Board })

	var keywords, authors, pushSums []string
	var articles int
	for _, sub := range subs {
		board := "<b>" + html.EscapeString(sub.Board) + "</b>："
		if len(sub.Keywords) > 0 {
			keywords = append(keywords, board+codeList(sub.Keywords))
		}
		if len(sub.Authors) > 0 {
			authors = append(authors, board+codeList(sub.Authors))
		}
		var sums []string
		if sub.PushSum.Up != 0 {
			sums = append(sums, fmt.Sprintf("推 <code>%d</code>", sub.PushSum.Up))
		}
		if sub.PushSum.Down != 0 {
			sums = append(sums, fmt.Sprintf("噓 <code>%d</code>", sub.PushSum.Down))
		}
		if len(sums) > 0 {
			pushSums = append(pushSums, board+strings.Join(sums, "・"))
		}
		articles += len(sub.Articles)
	}
	if len(keywords)+len(authors)+len(pushSums) == 0 && articles == 0 {
		return "📋 <b>追蹤清單</b>\n\n尚未建立清單。傳送 <code>指令</code> 查看新增方法。"
	}

	var b strings.Builder
	b.WriteString("📋 <b>追蹤清單</b>\n")
	writeSection(&b, "🔍 關鍵字", keywords)
	writeSection(&b, "✍️ 作者", authors)
	writeSection(&b, "🔥 推噓文數", pushSums)
	if articles > 0 {
		fmt.Fprintf(&b, "\n💬 另有 %d 篇推文追蹤，傳送 <code>推文清單</code> 查看", articles)
	}
	return strings.TrimRight(b.String(), "\n")
}

func commentListView(subs subscription.Subscriptions) string {
	subs = append(subscription.Subscriptions(nil), subs...)
	sort.Slice(subs, func(i, j int) bool { return subs[i].Board < subs[j].Board })

	var b strings.Builder
	fmt.Fprintf(&b, "💬 <b>推文追蹤清單</b>（上限 %d 篇）\n", commentListSize)
	var count int
	for _, sub := range subs {
		if len(sub.Articles) == 0 {
			continue
		}
		codes := append([]string(nil), sub.Articles...)
		sort.Strings(codes)
		b.WriteString("\n<b>" + html.EscapeString(sub.Board) + "</b>\n")
		for _, code := range codes {
			link := "https://www.ptt.cc/bbs/" + sub.Board + "/" + code + ".html"
			b.WriteString("• " + html.EscapeString(link) + "\n")
			count++
		}
	}
	if count == 0 {
		return "💬 <b>推文追蹤清單</b>\n\n尚未追蹤任何文章。\n例：<code>新增推文 " + exampleArticle + "</code>"
	}
	b.WriteString("\n傳送 <code>清理推文</code> 可移除已刪除的文章。")
	return b.String()
}

func rankingView(keywords, authors, pushSums []string) string {
	var b strings.Builder
	b.WriteString("🏆 <b>熱門追蹤排行</b>\n")
	writeSection(&b, "🔍 關鍵字", rankLines(keywords, codeWord))
	writeSection(&b, "✍️ 作者", rankLines(authors, codeWord))
	writeSection(&b, "🔥 推噓文數", rankLines(pushSums, func(word string) string {
		if strings.HasPrefix(word, "-") {
			return "噓 <code>" + html.EscapeString(word[1:]) + "</code>"
		}
		return "推 <code>" + html.EscapeString(word) + "</code>"
	}))
	b.WriteString("\n📊 完整 TOP 100：" + siteURL + "/top")
	return b.String()
}

// rankLines formats the top three ranking entries stored as "board:word".
func rankLines(entries []string, formatWord func(string) string) []string {
	medals := []string{"🥇", "🥈", "🥉"}
	lines := make([]string, 0, len(medals))
	for i, entry := range entries {
		if i == len(medals) {
			break
		}
		board, word, _ := strings.Cut(entry, ":")
		lines = append(lines, medals[i]+" <b>"+html.EscapeString(board)+"</b> "+formatWord(word))
	}
	return lines
}

func writeSection(b *strings.Builder, title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	b.WriteString("\n<b>" + title + "</b>\n")
	for _, line := range lines {
		b.WriteString(line + "\n")
	}
}

func codeWord(word string) string {
	return "<code>" + html.EscapeString(word) + "</code>"
}

func codeList(words []string) string {
	words = append([]string(nil), words...)
	sort.Strings(words)
	for i, w := range words {
		words[i] = codeWord(w)
	}
	return strings.Join(words, "、")
}
