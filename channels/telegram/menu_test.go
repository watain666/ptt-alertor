package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSetupMenu(t *testing.T) {
	var endpoints []string
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Errorf("request method = %q, want POST", request.Method)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}

		endpoint := strings.TrimPrefix(request.URL.Path, "/bottest/")
		endpoints = append(endpoints, endpoint)
		switch endpoint {
		case "setMyCommands":
			if got := values.Get("scope"); got != `{"type":"all_private_chats"}` {
				t.Errorf("scope = %q", got)
			}
			var commands []menuCommand
			if err := json.Unmarshal([]byte(values.Get("commands")), &commands); err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, command := range commands {
				if command.Description == "" {
					t.Errorf("missing description for %q", command.Command)
				}
				names = append(names, command.Command)
			}
			want := []string{"start", "help", "list", "comments", "ranking", "add", "del", "addauthor", "delauthor", "addpush", "addboo", "addcomment", "delcomment", "cleancomments", "showkeyboard", "hidekeyboard"}
			if !reflect.DeepEqual(names, want) {
				t.Errorf("commands = %v, want %v", names, want)
			}
		case "setChatMenuButton":
			if got := values.Get("menu_button"); got != `{"type":"commands"}` {
				t.Errorf("menu_button = %q", got)
			}
		default:
			t.Errorf("unexpected Telegram API endpoint %q", endpoint)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":true}`)), Header: make(http.Header)}, nil
	})}
	bot := &tgbotapi.BotAPI{Token: "test", Client: client}

	if err := setupMenu(bot); err != nil {
		t.Fatal(err)
	}
	if want := []string{"setMyCommands", "setChatMenuButton"}; !reflect.DeepEqual(endpoints, want) {
		t.Errorf("endpoints = %v, want %v", endpoints, want)
	}
}

func TestSetupMenuReportsTelegramError(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(`{"ok":false,"description":"Bad Request"}`)), Header: make(http.Header)}, nil
	})}
	bot := &tgbotapi.BotAPI{Token: "test", Client: client}

	if err := setupMenu(bot); err == nil || !strings.Contains(err.Error(), "set Telegram menu commands") {
		t.Fatalf("setupMenu error = %v, want Telegram API error", err)
	}
}

func TestMenuEditCommandsShowChineseExamplesAndKeepFlagHelp(t *testing.T) {
	previousBot := bot
	t.Cleanup(func() { bot = previousBot })

	for _, test := range []struct {
		message string
		want    string
	}{
		{message: "/add", want: "新增 Lifeismoney 咖啡"},
		{message: "/del", want: "刪除 Lifeismoney 咖啡"},
		{message: "/add -h", want: "Usage of Ptt Alertor"},
		{message: "/addauthor", want: "<code>新增作者 看板 作者</code>"},
		{message: "/addpush", want: "數量設為 <code>0</code> 即取消通知"},
		{message: "/delcomment", want: "<code>刪除推文 https://www.ptt.cc/bbs/EZsoft/M.1708247900.A.27C.html</code>"},
		{message: "/unknown", want: "無此指令"},
	} {
		t.Run(test.message, func(t *testing.T) {
			var sentText string
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if !strings.HasSuffix(request.URL.Path, "/sendMessage") {
					t.Errorf("unexpected Telegram API endpoint %q", request.URL.Path)
				}
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatal(err)
				}
				values, err := url.ParseQuery(string(body))
				if err != nil {
					t.Fatal(err)
				}
				sentText = values.Get("text")
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`)), Header: make(http.Header)}, nil
			})}
			bot = &tgbotapi.BotAPI{Token: "test", Client: client}

			commandEnd := strings.IndexByte(test.message, ' ')
			if commandEnd == -1 {
				commandEnd = len(test.message)
			}
			message := &tgbotapi.Message{
				Text:     test.message,
				Entities: &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: commandEnd}},
				From:     &tgbotapi.User{ID: 123},
				Chat:     &tgbotapi.Chat{ID: 456},
			}
			handleCommand(tgbotapi.Update{Message: message})
			if !strings.Contains(sentText, test.want) {
				t.Errorf("sent text = %q, want to contain %q", sentText, test.want)
			}
		})
	}
}

func TestSlashCommandMatchesMenu(t *testing.T) {
	for _, test := range []struct{ chinese, slash string }{
		{"指令", "help"}, {"清單", "list"}, {"推文清單", "comments"}, {"排行", "ranking"},
		{"新增", "add"}, {"刪除", "del"}, {"新增作者", "addauthor"}, {"刪除作者", "delauthor"},
		{"新增推文數", "addpush"}, {"新增噓文數", "addboo"}, {"新增推文", "addcomment"}, {"刪除推文", "delcomment"},
		{"清理推文", "cleancomments"}, {"顯示鍵盤", "showkeyboard"}, {"隱藏鍵盤", "hidekeyboard"},
	} {
		if got, ok := slashCommand(test.chinese); !ok || got != test.slash {
			t.Errorf("slashCommand(%q) = %q, %v, want %q", test.chinese, got, ok, test.slash)
		}
	}
	if _, ok := slashCommand(""); ok {
		t.Error("empty text matched a command")
	}
	for _, command := range menuCommands {
		if command.Command == "list" && command.Description != "查看追蹤清單（同「清單」）" {
			t.Errorf("list description = %q", command.Description)
		}
	}
}

func TestChineseTextRunsSameCommandAsSlash(t *testing.T) {
	previousBot := bot
	t.Cleanup(func() { bot = previousBot })

	for _, test := range []struct {
		text, wantText, wantParseMode, wantMarkup string
	}{
		{text: "指令", wantText: "指令說明", wantParseMode: "HTML"},
		{text: "新增作者", wantText: "<code>/addauthor gossiping ffaarr,obov</code>", wantParseMode: "HTML"},
		{text: "新增", wantText: "新增追蹤", wantParseMode: "HTML"},
		{text: "顯示鍵盤", wantText: "顯示小鍵盤", wantMarkup: "keyboard"},
		{text: "隱藏鍵盤", wantText: "隱藏小鍵盤", wantMarkup: "remove_keyboard"},
	} {
		t.Run(test.text, func(t *testing.T) {
			var values url.Values
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatal(err)
				}
				if values, err = url.ParseQuery(string(body)); err != nil {
					t.Fatal(err)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":1}}`)), Header: make(http.Header)}, nil
			})}
			bot = &tgbotapi.BotAPI{Token: "test", Client: client}

			handleTextCommand(456, "123", test.text)
			if !strings.Contains(values.Get("text"), test.wantText) {
				t.Errorf("text = %q, want to contain %q", values.Get("text"), test.wantText)
			}
			if got := values.Get("parse_mode"); got != test.wantParseMode {
				t.Errorf("parse_mode = %q, want %q", got, test.wantParseMode)
			}
			if !strings.Contains(values.Get("reply_markup"), test.wantMarkup) {
				t.Errorf("reply_markup = %q, want to contain %q", values.Get("reply_markup"), test.wantMarkup)
			}
		})
	}
}

func TestEveryChineseCommandHasSlashCommand(t *testing.T) {
	for chinese := range map[string]bool{
		"指令": true, "清單": true, "推文清單": true, "排行": true, "新增": true, "刪除": true,
		"新增作者": true, "刪除作者": true, "新增推文數": true, "新增噓文數": true,
		"新增推文": true, "刪除推文": true, "清理推文": true,
	} {
		if _, ok := slashCommand(chinese); !ok {
			t.Errorf("Chinese command %q has no slash command", chinese)
		}
	}
	for _, c := range botCommands {
		if c.usage != "" {
			assertLineBalancedHTML(t, usageView(c))
		}
	}
}

func TestKeyboardButtonsAreCommands(t *testing.T) {
	for _, row := range keyboardRows {
		for _, text := range row {
			if _, ok := slashCommand(text); !ok {
				t.Errorf("keyboard button %q is not a command", text)
			}
		}
	}
}
