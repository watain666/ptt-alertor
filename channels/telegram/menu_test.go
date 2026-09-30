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
			want := []string{"start", "help", "list", "comments", "ranking", "add", "del", "showkeyboard", "hidekeyboard"}
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
