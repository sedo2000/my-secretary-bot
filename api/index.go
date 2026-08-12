package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const (
	connFile     = "/tmp/business_connections.json"
	welcomedFile = "/tmp/welcomed_users.json"
)

var mu sync.Mutex

func loadMap(filename string, v interface{}) {
	mu.Lock()
	defer mu.Unlock()
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	json.Unmarshal(data, v)
}

func saveMap(filename string, v interface{}) {
	mu.Lock()
	defer mu.Unlock()
	data, _ := json.Marshal(v)
	os.WriteFile(filename, data, 0644)
}

func getBusinessConn(userID int64) string {
	conns := make(map[string]string)
	loadMap(connFile, &conns)
	return conns[fmt.Sprintf("%d", userID)]
}

func setBusinessConn(userID int64, connID string) {
	conns := make(map[string]string)
	loadMap(connFile, &conns)
	conns[fmt.Sprintf("%d", userID)] = connID
	saveMap(connFile, &conns)
}

func removeBusinessConn(userID int64) {
	conns := make(map[string]string)
	loadMap(connFile, &conns)
	delete(conns, fmt.Sprintf("%d", userID))
	saveMap(connFile, &conns)
}

func isWelcomed(userID int64) bool {
	welcomed := make(map[string]bool)
	loadMap(welcomedFile, &welcomed)
	return welcomed[fmt.Sprintf("%d", userID)]
}

func setWelcomed(userID int64, val bool) {
	welcomed := make(map[string]bool)
	loadMap(welcomedFile, &welcomed)
	welcomed[fmt.Sprintf("%d", userID)] = val
	saveMap(welcomedFile, &welcomed)
}

type Update struct {
	UpdateID           int                 `json:"update_id"`
	Message            *Message            `json:"message,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	BusinessConnection *BusinessConnection `json:"business_connection,omitempty"`
	BusinessMessage    *Message            `json:"business_message,omitempty"`
}

type BusinessConnection struct {
	ID        string `json:"id"`
	User      User   `json:"user"`
	IsEnabled bool   `json:"is_enabled"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`
	From      User   `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Telegram Business Bot Server is Running!"))
		return
	}

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	// 1. تفعيل البوت وربطه بالحساب (بدون تكرار رسالة الترحيب)
	if update.BusinessConnection != nil {
		userID := update.BusinessConnection.User.ID
		if update.BusinessConnection.IsEnabled {
			setBusinessConn(userID, update.BusinessConnection.ID)
			if !isWelcomed(userID) {
				setWelcomed(userID, true)
				sendTelegramMessage(token, userID, "",
					"🎉 أهلاً بك يا عزيزي!\n\nتم تفعيل البوت بنجاح على حسابك التجاري ✅.",
					getMainMenuKeyboard())
			}
		} else {
			removeBusinessConn(userID)
			setWelcomed(userID, false)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. معالجة الضغط على الأزرار
	if update.CallbackQuery != nil {
		query := update.CallbackQuery
		chatID := query.Message.Chat.ID
		data := query.Data

		switch data {
		case "menu_info":
			editMessageText(token, chatID, query.Message.MessageID,
				"ℹ️ **معلومات البوت:**\nهذا البوت مخصص لإدارة الردود على عملائك عبر ميزة تيليجرام للأعمال.", getBackOnlyKeyboard())
		case "main_menu":
			editMessageText(token, chatID, query.Message.MessageID,
				"🏠 **القائمة الرئيسية:**\nاختر من الأقسام التالية:", getMainMenuKeyboard())
		}

		answerCallbackQuery(token, query.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. معالجة الرسائل الواردة في المحادثات التجارية
	msg := update.Message
	if msg == nil {
		msg = update.BusinessMessage
	}

	if msg != nil {
		chatID := msg.Chat.ID
		text := msg.Text

		if text == "/start" {
			sendTelegramMessage(token, chatID, "",
				"🏠 **أهلاً بك في بوت إدارة الأعمال:**\nاستخدم الأيارات أدناه:", getMainMenuKeyboard())
		} else if text != "" {
			// مثال على الرد الآلي على العملاء عبر الحساب التجاري
			bizConnID := getBusinessConn(msg.From.ID)
			sendTelegramMessage(token, chatID, bizConnID, "🤖 تم استلام رسالتك وسيتم الرد عليك قريباً بواسطة صاحب الحساب.", InlineKeyboardMarkup{})
		}
	}

	w.WriteHeader(http.StatusOK)
}

func getMainMenuKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{InlineKeyboardButton{Text: "ℹ️ معلومات البوت", CallbackData: "menu_info"}},
		},
	}
}

func getBackOnlyKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{InlineKeyboardButton{Text: "🔙 الرجوع", CallbackData: "main_menu"}},
		},
	}
}

func sendTelegramMessage(token string, chatID int64, businessConnectionID string, text string, replyMarkup InlineKeyboardMarkup) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	if businessConnectionID != "" {
		payload["business_connection_id"] = businessConnectionID
	}
	if len(replyMarkup.InlineKeyboard) > 0 {
		payload["reply_markup"] = replyMarkup
	}
	return sendRequest(url, payload)
}

func editMessageText(token string, chatID int64, messageID int, text string, replyMarkup InlineKeyboardMarkup) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", token)
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkup,
	}
	return sendRequest(url, payload)
}

func answerCallbackQuery(token string, callbackQueryID string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", token)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
	}
	sendRequest(url, payload)
}

func sendRequest(url string, payload map[string]interface{}) bool {
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
