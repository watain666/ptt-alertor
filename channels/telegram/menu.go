package telegram

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type menuCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// botCommand pairs a slash command with the Chinese text command that does
// the same thing, so both can be used interchangeably. Commands with usage
// take arguments, and example shows sample arguments.
type botCommand struct {
	slash       string
	chinese     string
	description string
	usage       string
	example     string
}

var botCommands = []botCommand{
	{slash: "start", description: "開始使用機器人"},
	{slash: "help", chinese: "指令", description: "查看所有指令與範例"},
	{slash: "list", chinese: "清單", description: "查看追蹤清單"},
	{slash: "comments", chinese: "推文清單", description: "查看推文追蹤清單"},
	{slash: "ranking", chinese: "排行", description: "查看熱門追蹤排行"},
	{slash: "add", chinese: "新增", description: "新增追蹤關鍵字", usage: "看板 關鍵字", example: "gossiping,movie 金城武,結衣"},
	{slash: "del", chinese: "刪除", description: "取消追蹤關鍵字", usage: "看板 關鍵字", example: "gossiping 金城武"},
	{slash: "addauthor", chinese: "新增作者", description: "新增追蹤作者", usage: "看板 作者", example: "gossiping ffaarr,obov"},
	{slash: "delauthor", chinese: "刪除作者", description: "取消追蹤作者", usage: "看板 作者", example: "gossiping ffaarr"},
	{slash: "addpush", chinese: "新增推文數", description: "推文數達標時通知", usage: "看板 數量", example: "joke,beauty 10"},
	{slash: "addboo", chinese: "新增噓文數", description: "噓文數達標時通知", usage: "看板 數量", example: "joke 10"},
	{slash: "addcomment", chinese: "新增推文", description: "追蹤文章的新推文", usage: "網址", example: exampleArticle},
	{slash: "delcomment", chinese: "刪除推文", description: "取消追蹤文章推文", usage: "網址", example: exampleArticle},
	{slash: "cleancomments", chinese: "清理推文", description: "移除已刪除的推文追蹤"},
	{slash: "showkeyboard", chinese: "顯示鍵盤", description: "顯示快捷鍵盤"},
	{slash: "hidekeyboard", chinese: "隱藏鍵盤", description: "隱藏快捷鍵盤"},
}

// keyboardRows lays out the reply keyboard. Each button sends its Chinese
// text command; the ones that need arguments reply with their usage.
var keyboardRows = [][]string{
	{"清單", "推文清單", "排行", "指令"},
	{"新增", "新增作者", "新增推數", "新增噓數"},
	{"刪除", "刪除作者", "新增推文", "刪除推文"},
	{"清理推文", "隱藏鍵盤"},
}

// chineseAliases are shorter names for Chinese commands, used where the full
// name is too long, e.g. a reply keyboard button.
var chineseAliases = map[string]string{
	"新增推數": "新增推文數",
	"新增噓數": "新增噓文數",
}

// expandAlias replaces a command alias at the start of text with its full name.
func expandAlias(text string) string {
	text = strings.TrimSpace(text)
	name, rest, _ := strings.Cut(text, " ")
	if full, ok := chineseAliases[name]; ok {
		return strings.TrimSpace(full + " " + rest)
	}
	return text
}

var menuCommands = func() []menuCommand {
	commands := make([]menuCommand, 0, len(botCommands))
	for _, c := range botCommands {
		description := c.description
		if c.chinese != "" {
			description += "（同「" + c.chinese + "」）"
		}
		commands = append(commands, menuCommand{Command: c.slash, Description: description})
	}
	return commands
}()

// slashCommand returns the slash command for a Chinese text command.
func slashCommand(chinese string) (string, bool) {
	for _, c := range botCommands {
		if c.chinese != "" && c.chinese == chinese {
			return c.slash, true
		}
	}
	return "", false
}

func findBotCommand(slash string) (botCommand, bool) {
	for _, c := range botCommands {
		if c.slash == slash {
			return c, true
		}
	}
	return botCommand{}, false
}

func setupMenu(bot *tgbotapi.BotAPI) error {
	commands, err := json.Marshal(menuCommands)
	if err != nil {
		return fmt.Errorf("encode Telegram menu commands: %w", err)
	}

	_, err = bot.MakeRequest("setMyCommands", url.Values{
		"commands": {string(commands)},
		"scope":    {`{"type":"all_private_chats"}`},
	})
	if err != nil {
		return fmt.Errorf("set Telegram menu commands: %w", err)
	}

	_, err = bot.MakeRequest("setChatMenuButton", url.Values{
		"menu_button": {`{"type":"commands"}`},
	})
	if err != nil {
		return fmt.Errorf("set Telegram menu button: %w", err)
	}
	return nil
}
