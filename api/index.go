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
	stateFile    = "/tmp/user_states.json"
	connFile     = "/tmp/business_connections.json"
	welcomedFile = "/tmp/welcomed_users.json"
)

var mu sync.Mutex

// دوال حفظ واسترجاع الحالة لضمان عدم ضياعها في Vercel Serverless
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

func getUserState(userID int64) string {
	states := make(map[string]string)
	loadMap(stateFile, &states)
	return states[fmt.Sprintf("%d", userID)]
}

func setUserState(userID int64, state string) {
	states := make(map[string]string)
	loadMap(stateFile, &states)
	states[fmt.Sprintf("%d", userID)] = state
	saveMap(stateFile, &states)
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

// هياكل بيانات تيليجرام
type Update struct {
	UpdateID           int                 `json:"update_id"`
	Message            *Message            `json:"message,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	BusinessConnection *BusinessConnection `json:"business_connection,omitempty"`
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
	MessageID int         `json:"message_id"`
	Chat      Chat        `json:"chat"`
	Text      string      `json:"text,omitempty"`
	From      User        `json:"from"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Video     *Video      `json:"video,omitempty"`
}

type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type Video struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
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

// دالة المعالجة الرئيسية لـ Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Go Bot Server is Running & Fixed Successfully!"))
		return
	}

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	// 1. معالجة ربط البوت بالحساب (مع منع التكرار تماماً)
	if update.BusinessConnection != nil {
		userID := update.BusinessConnection.User.ID
		if update.BusinessConnection.IsEnabled {
			setBusinessConn(userID, update.BusinessConnection.ID)
			if !isWelcomed(userID) {
				setWelcomed(userID, true)
				sendTelegramMessage(token, userID,
					"🎉 أهلاً بك يا عزيزي!\n\nتم تفعيل البوت بنجاح على حسابك ✅.\nأرسل /start للبدء في إدارة ملفك الشخصي وقصصك.",
					getMainMenuKeyboard())
			}
		} else {
			removeBusinessConn(userID)
			setWelcomed(userID, false)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. التعامل مع الأزرار (Callback Queries) مع تفعيل زر الرجوع لكل قائمة
	if update.CallbackQuery != nil {
		query := update.CallbackQuery
		chatID := query.Message.Chat.ID
		userID := query.From.ID
		data := query.Data

		switch data {
		case "menu_profile":
			setUserState(userID, "")
			editMessageText(token, chatID, query.Message.MessageID,
				"👤 **إدارة الملف الشخصي:**\nاختر العملية المطلوبة:", getProfileKeyboard())
		case "menu_stories":
			setUserState(userID, "")
			editMessageText(token, chatID, query.Message.MessageID,
				"📸 **إدارة القصص:**\nاختر العملية المطلوبة:", getStoriesKeyboard())
		case "edit_bio":
			setUserState(userID, "awaiting_bio")
			editMessageText(token, chatID, query.Message.MessageID,
				"✍️ أرسل البايو الجديد الآن لتحديثه في حسابك (بحد أقصى 140 حرفاً):", getBackOnlyKeyboard())
		case "edit_photo":
			setUserState(userID, "awaiting_photo")
			editMessageText(token, chatID, query.Message.MessageID,
				"🖼️ أرسل الصورة الجديدة للملف الشخصي:", getBackOnlyKeyboard())
		case "post_story":
			setUserState(userID, "awaiting_story")
			editMessageText(token, chatID, query.Message.MessageID,
				"🚀 أرسل الوسائط (صورة أو فيديو) لنشرها كقصة جديدة:", getBackOnlyKeyboard())
		case "main_menu":
			setUserState(userID, "")
			editMessageText(token, chatID, query.Message.MessageID,
				"🏠 **القائمة الرئيسية:**\nاختر من الأقسام التالية:", getMainMenuKeyboard())
		}

		answerCallbackQuery(token, query.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. التعامل مع الرسائل النصية والوسائط وتنفيذ العمليات الفعلية
	if update.Message != nil {
		msg := update.Message
		chatID := msg.Chat.ID
		userID := msg.From.ID
		text := msg.Text
		bizConnID := getBusinessConn(userID)

		if text == "/start" {
			setUserState(userID, "")
			sendTelegramMessage(token, chatID,
				"🏠 **القائمة الرئيسية لإدارة حسابك:**\nاختر القسم المناسب:", getMainMenuKeyboard())
			w.WriteHeader(http.StatusOK)
			return
		}

		state := getUserState(userID)

		// تنفيذ تعديل البايو
		if state == "awaiting_bio" {
			if bizConnID == "" {
				sendTelegramMessage(token, chatID, "⚠️ تنبيه: لم يتم ربط البوت بحسابك التجاري بعد. يرجى تفعيله من إعدادات تيليجرام.", getMainMenuKeyboard())
			} else {
				success := setBusinessAccountBio(token, bizConnID, text)
				if success {
					sendTelegramMessage(token, chatID, "✅ تم تحديث البايو بنجاح!", getMainMenuKeyboard())
				} else {
					sendTelegramMessage(token, chatID, "❌ فشل تحديث البايو. تأكد من أن النص أقل من 140 حرفاً وأن البوت يملك صلاحية التعديل.", getMainMenuKeyboard())
				}
			}
			setUserState(userID, "")
			w.WriteHeader(http.StatusOK)
			return
		}

		// تنفيذ تعديل صورة الملف الشخصي
		if state == "awaiting_photo" {
			if bizConnID == "" {
				sendTelegramMessage(token, chatID, "⚠️ تنبيه: لم يتم ربط البوت بحسابك التجاري بعد.", getMainMenuKeyboard())
			} else if len(msg.Photo) > 0 {
				photoFileID := msg.Photo[len(msg.Photo)-1].FileID
				success := setBusinessAccountProfilePhoto(token, bizConnID, photoFileID)
				if success {
					sendTelegramMessage(token, chatID, "✅ تم تحديث صورة الملف الشخصي بنجاح!", getMainMenuKeyboard())
				} else {
					sendTelegramMessage(token, chatID, "❌ فشل تحديث صورة الملف الشخصي.", getMainMenuKeyboard())
				}
			} else {
				sendTelegramMessage(token, chatID, "⚠️ يرجى إرسال صورة صحيحة لتحديث صورة الملف.", getBackOnlyKeyboard())
				w.WriteHeader(http.StatusOK)
				return
			}
			setUserState(userID, "")
			w.WriteHeader(http.StatusOK)
			return
		}

		// تنفيذ نشر القصة
		if state == "awaiting_story" {
			if bizConnID == "" {
				sendTelegramMessage(token, chatID, "⚠️ تنبيه: لم يتم ربط البوت بحسابك التجاري بعد.", getMainMenuKeyboard())
			} else {
				var content map[string]interface{}
				if len(msg.Photo) > 0 {
					photoFileID := msg.Photo[len(msg.Photo)-1].FileID
					content = map[string]interface{}{
						"type":  "photo",
						"photo": photoFileID,
					}
				} else if msg.Video != nil {
					content = map[string]interface{}{
						"type":  "video",
						"video": msg.Video.FileID,
					}
				}

				if content != nil {
					success := postStory(token, bizConnID, content)
					if success {
						sendTelegramMessage(token, chatID, "🚀 تم نشر القصة بنجاح على حسابك التجاري!", getMainMenuKeyboard())
					} else {
						sendTelegramMessage(token, chatID, "❌ فشل نشر القصة. تأكد من منح البوت صلاحية نشر القصص من إعدادات حسابك.", getMainMenuKeyboard())
					}
				} else {
					sendTelegramMessage(token, chatID, "⚠️ يرجى إرسال صورة أو فيديو صالح لنشره كقصة.", getBackOnlyKeyboard())
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			setUserState(userID, "")
			w.WriteHeader(http.StatusOK)
			return
		}

		sendTelegramMessage(token, chatID, "أهلاً بك! استخدم الأزرار أدناه لإدارة حسابك:", getMainMenuKeyboard())
	}

	w.WriteHeader(http.StatusOK)
}

// --- لوحات المفاتيح مع زر الرجوع في كل قائمة ---

func getMainMenuKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				InlineKeyboardButton{Text: "👤 الملف الشخصي", CallbackData: "menu_profile"},
				InlineKeyboardButton{Text: "📸 القصص", CallbackData: "menu_stories"},
			},
		},
	}
}

func getProfileKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{InlineKeyboardButton{Text: "✏️ تعديل البايو", CallbackData: "edit_bio"}},
			{InlineKeyboardButton{Text: "🖼️ تعديل صورة الملف", CallbackData: "edit_photo"}},
			{InlineKeyboardButton{Text: "🔙 الرجوع", CallbackData: "main_menu"}},
		},
	}
}

func getStoriesKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{InlineKeyboardButton{Text: "🚀 نشر قصة جديدة", CallbackData: "post_story"}},
			{InlineKeyboardButton{Text: "🔙 الرجوع", CallbackData: "main_menu"}},
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

// --- دوال الاتصال بـ Telegram Bot API المحدثة ---

func sendTelegramMessage(token string, chatID int64, text string, replyMarkup InlineKeyboardMarkup) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkup,
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

// تعديل البايو التجاري
func setBusinessAccountBio(token string, businessConnectionID string, bio string) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountBio", token)
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"bio":                    bio,
	}
	return sendRequest(url, payload)
}

// تعديل صورة الملف الشخصي التجاري
func setBusinessAccountProfilePhoto(token string, businessConnectionID string, photoFileID string) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountProfilePhoto", token)
	inputPhoto := map[string]interface{}{
		"type":  "static",
		"photo": photoFileID,
	}
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"photo":                  inputPhoto,
	}
	return sendRequest(url, payload)
}

// نشر القصة التجارية
func postStory(token string, businessConnectionID string, content map[string]interface{}) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/postStory", token)
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"content":                content,
	}
	return sendRequest(url, payload)
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
