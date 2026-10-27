package telegram

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	log "github.com/Ptt-Alertor/logrus"

	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/julienschmidt/httprouter"
	"github.com/watain666/ptt-alertor/command"
	"github.com/watain666/ptt-alertor/myutil"
)

var (
	bot   *tgbotapi.BotAPI
	err   error
	token = os.Getenv("TELEGRAM_TOKEN")
	host  = os.Getenv("APP_HOST")
)

func init() {
	if token == "" {
		log.Warn("Telegram token is empty, telegram channel disabled")
		return
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.WithError(err).Warn("Telegram Bot Initialize Failed, telegram channel disabled")
		return
	}
	// bot.Debug = true
	log.Info("Telegram Authorized on " + bot.Self.UserName)
	if err := setupMenu(bot); err != nil {
		log.WithError(err).Warn("Telegram Menu Setup Failed")
	}

	if !strings.HasPrefix(strings.ToLower(host), "https://") {
		log.WithField("APP_HOST", host).Warn("Skip Telegram webhook setup: APP_HOST must be https")
		return
	}

	webhookConfig := tgbotapi.NewWebhook(host + "/telegram/" + token)
	webhookConfig.MaxConnections = 100
	_, err = bot.SetWebhook(webhookConfig)
	if err != nil {
		log.WithError(err).Warn("Telegram Bot Set Webhook Failed, continue without webhook")
		return
	}
	log.Info("Telegram Bot Sets Webhook Success")
}

// HandleRequest handles request from webhook
func HandleRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if bot == nil {
		http.Error(w, "telegram channel disabled", http.StatusServiceUnavailable)
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("Telegram Read Request Body Failed")
	}

	var update tgbotapi.Update
	json.Unmarshal(bytes, &update)

	if update.CallbackQuery != nil {
		handleCallbackQuery(update)
		return
	}

	if update.Message != nil {
		if update.Message.IsCommand() {
			handleCommand(update)
			return
		}
		if update.Message.Text != "" {
			handleText(update)
			return
		}
	}
}

func handleCallbackQuery(update tgbotapi.Update) {
	var responseText string
	userID := strconv.Itoa(update.CallbackQuery.From.ID)
	switch update.CallbackQuery.Data {
	case "CANCEL":
		responseText = "取消"
	default:
		responseText = command.HandleCommand(update.CallbackQuery.Data, userID, true)
	}
	SendTextMessage(update.CallbackQuery.Message.Chat.ID, responseText)
}

// help - 所有指令清單
// list - 設定清單
// comments - 推文追蹤清單
// ranking - 熱門關鍵字、作者、推文數
// add - 新增看板關鍵字、作者、推文數
// del - 刪除看板關鍵字、作者、推文數
// showkeyboard - 顯示快捷小鍵盤
// hidekeyboard - 隱藏快捷小鍵盤
func handleCommand(update tgbotapi.Update) {
	userID := strconv.Itoa(update.Message.From.ID)
	chatID := update.Message.Chat.ID

	name := update.Message.Command()
	arguments := strings.TrimSpace(update.Message.CommandArguments())
	switch name {
	case "start":
		command.HandleTelegramFollow(userID, chatID)
	case "add", "del":
		if arguments == "" {
			SendHTMLMessage(chatID, editMenuHelp(name))
			return
		}
		if strings.HasPrefix(arguments, "-") {
			SendTextMessage(chatID, command.HandleCommand(name+" "+arguments, userID, true))
			return
		}
	}
	if runMenuCommand(chatID, userID, name) {
		return
	}
	// The other slash commands work the same as their Chinese text command,
	// e.g. "/addauthor 看板 作者" is "新增作者 看板 作者".
	c, ok := findBotCommand(name)
	if !ok || c.chinese == "" {
		SendTextMessage(chatID, "無此指令，請傳送「指令」查看指令清單")
		return
	}
	if c.usage != "" && arguments == "" {
		SendHTMLMessage(chatID, usageView(c))
		return
	}
	handleTextCommand(chatID, userID, strings.TrimSpace(c.chinese+" "+arguments))
}

func handleText(update tgbotapi.Update) {
	handleTextCommand(update.Message.Chat.ID, strconv.Itoa(update.Message.From.ID), update.Message.Text)
}

func handleTextCommand(chatID int64, userID, text string) {
	if match, _ := regexp.MatchString("^(刪除|刪除作者)+\\s.*\\*+", text); match {
		sendConfirmation(chatID, text)
		return
	}
	if fields := strings.Fields(text); len(fields) == 1 {
		name := fields[0]
		if slash, ok := slashCommand(name); ok {
			name = slash
		}
		if runMenuCommand(chatID, userID, name) {
			return
		}
	}
	SendTextMessage(chatID, command.HandleCommand(text, userID, true))
}

// runMenuCommand runs commands that take no arguments, by slash name.
func runMenuCommand(chatID int64, userID, name string) bool {
	switch name {
	case "showkeyboard":
		showReplyKeyboard(chatID)
		return true
	case "hidekeyboard":
		hideReplyKeyboard(chatID)
		return true
	}
	return sendMenuView(chatID, userID, name)
}

func sendConfirmation(chatID int64, cmd string) {
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("是", cmd),
			tgbotapi.NewInlineKeyboardButtonData("否", "CANCEL"),
		))
	msg := tgbotapi.NewMessage(chatID, "確定"+cmd+"？")
	msg.ReplyMarkup = markup
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Send Confirmation Failed")
	}
}

const maxCharacters = 4096

// SendTextMessage sends text message to chatID
func SendTextMessage(chatID int64, text string) {
	for _, msg := range myutil.SplitTextByLineBreak(text, maxCharacters) {
		sendTextMessage(chatID, msg)
	}
}

// SendHTMLMessage sends HTML formatted text message to chatID.
// HTML tags must not span lines, since long text is split by line break.
func SendHTMLMessage(chatID int64, text string) {
	for _, msg := range myutil.SplitTextByLineBreak(text, maxCharacters) {
		sendMessage(chatID, msg, tgbotapi.ModeHTML)
	}
}

func sendTextMessage(chatID int64, text string) {
	sendMessage(chatID, text, "")
}

func sendMessage(chatID int64, text string, parseMode string) {
	if bot == nil {
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode
	msg.DisableWebPagePreview = true
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Send Message Failed")
	}
}

func showReplyKeyboard(chatID int64) {
	if bot == nil {
		return
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("清單"),
			tgbotapi.NewKeyboardButton("推文清單"),
			tgbotapi.NewKeyboardButton("排行"),
			tgbotapi.NewKeyboardButton("指令"),
		))
	msg := tgbotapi.NewMessage(chatID, "顯示小鍵盤")
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Show Reply Keyboard Failed")
	}
}

func hideReplyKeyboard(chatID int64) {
	if bot == nil {
		return
	}

	msg := tgbotapi.NewMessage(chatID, "隱藏小鍵盤")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, err := bot.Send(msg)
	if err != nil {
		log.WithError(err).Error("Telegram Hide Reply Keyboard Failed")
	}
}
