package telegram

import (
	"encoding/json"
	"fmt"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type menuCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

var menuCommands = []menuCommand{
	{Command: "start", Description: "開始使用機器人"},
	{Command: "help", Description: "查看所有指令與範例"},
	{Command: "list", Description: "查看追蹤清單"},
	{Command: "comments", Description: "查看推文追蹤清單"},
	{Command: "ranking", Description: "查看熱門追蹤排行"},
	{Command: "add", Description: "新增追蹤：查看中文指令範例"},
	{Command: "del", Description: "取消追蹤：查看中文指令範例"},
	{Command: "showkeyboard", Description: "顯示快捷鍵盤"},
	{Command: "hidekeyboard", Description: "隱藏快捷鍵盤"},
}

func editMenuHelp(action string) string {
	if action == "add" {
		return "可直接傳送中文指令新增追蹤：\n" +
			"新增 Lifeismoney 咖啡\n" +
			"新增作者 gossiping ffaarr\n" +
			"新增推文數 joke 10\n" +
			"新增噓文數 joke 10\n" +
			"新增推文 https://www.ptt.cc/bbs/EZsoft/M.1708247900.A.27C.html\n\n" +
			"進階參數用法：/add -h"
	}
	return "可直接傳送中文指令取消追蹤：\n" +
		"刪除 Lifeismoney 咖啡\n" +
		"刪除作者 gossiping ffaarr\n" +
		"刪除推文 https://www.ptt.cc/bbs/EZsoft/M.1708247900.A.27C.html\n" +
		"新增推文數 joke 0\n" +
		"新增噓文數 joke 0\n\n" +
		"進階參數用法：/del -h"
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
