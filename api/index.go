package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// الهياكل الأساسية لبيانات تيليجرام
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

// دالة المعالجة الرئيسية لـ Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot is running successfully!"))
		return
	}

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := "YOUR_BOT_TOKEN" // ضع توكن البوت الخاص بك هنا

	// 1. التعامل مع أحداث ربط البوت بالحساب (منع تكرار الرسالة المتكررة)
	if update.BusinessConnection != nil {
		if update.BusinessConnection.IsEnabled {
			// يمكنك حفظ حالة الربط هنا في قاعدة بيانات لضمان عدم إرسال الترحيب مرتين، 
			// أو إرسالها فقط عند التفعيل الأول الجديد.
			sendTelegramMessage(token, update.BusinessConnection.User.ID, 
				"🎉 أهلاً بك يا عزيزي!\n\nتم تفعيل البوت بنجاح على حسابك ✅.\nأرسل /start للبدء في إدارة ملفك الشخصي وقصصك.", 
				getMainMenuKeyboard())
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. التعامل مع الضغط على الأزرار (Callback Queries)
	if update.CallbackQuery != nil {
		query := update.CallbackQuery
		chatID := query.Message.Chat.ID
		data := query.Data

		switch data {
		case "menu_profile":
			editMessageText(token, chatID, query.Message.MessageID, 
				"👤 **إدارة الملف الشخصي:**\nاختر العملية المطلوبة:", getProfileKeyboard())
		case "menu_stories":
			editMessageText(token, chatID, query.Message.MessageID, 
				"📸 **إدارة القصص:**\nاختر العملية المطلوبة:", getStoriesKeyboard())
		case "edit_bio":
			editMessageText(token, chatID, query.Message.MessageID, 
				"✍️ أرسل البايو الجديد الآن لتحديثه في حسابك:", getBackOnlyKeyboard())
		case "edit_photo":
			editMessageText(token, chatID, query.Message.MessageID, 
				"🖼️ أرسل الصورة الجديدة للملف الشخصي:", getBackOnlyKeyboard())
		case "post_story":
			editMessageText(token, chatID, query.Message.MessageID, 
				"🚀 أرسل الوسائط (صورة/فيديو) لنشرها كقصة جديدة:", getBackOnlyKeyboard())
		case "main_menu":
			editMessageText(token, chatID, query.Message.MessageID, 
				"🏠 **القائمة الرئيسية:**\nاختر من الأقسام التالية:", getMainMenuKeyboard())
		}

		answerCallbackQuery(token, query.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. التعامل مع الرسائل النصية والأوامر العادية
	if update.Message != nil {
		msg := update.Message
		chatID := msg.Chat.ID
		text := msg.Text

		if text == "/start" {
			sendTelegramMessage(token, chatID, 
				"🏠 **القائمة الرئيسية لإدارة حسابك:**\nاختر القسم المناسب:", getMainMenuKeyboard())
		} else {
			sendTelegramMessage(token, chatID, "تم استلام طلبك بنجاح ✅", getMainMenuKeyboard())
		}
	}

	w.WriteHeader(http.StatusOK)
}

// --- تصميم لوحات المفاتيح (Keyboards) مع إضافة زر الرجوع لكل قائمة ---

func getMainMenuKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "👤 الملف الشخصي", CallbackData: "menu_profile"},
				{Text: "📸 القصص", CallbackData: "menu_stories"},
			},
		},
	}
}

func getProfileKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{Text: "✏️ تعديل البايو", CallbackData: "edit_bio"},
			{Text: "🖼️ تعديل صورة الملف", CallbackData: "edit_photo"},
			{Text: "🔙 الرجوع", CallbackData: "main_menu"},
		},
	}
}

func getStoriesKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{Text: "🚀 نشر قصة جديدة", CallbackData: "post_story"},
			{Text: "🔙 الرجوع", CallbackData: "main_menu"},
		},
	}
}

func getBackOnlyKeyboard() InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{Text: "🔙 الرجوع", CallbackData: "main_menu"},
		},
	}
}

// --- دوال الاتصال بـ Telegram Bot API ---

func sendTelegramMessage(token string, chatID int64, text string, replyMarkup InlineKeyboardMarkup) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkup,
	}
	sendRequest(url, payload)
}

func editMessageText(token string, chatID int64, messageID int, text string, replyMarkup InlineKeyboardMarkup) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", token)
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkup,
	}
	sendRequest(url, payload)
}

func answerCallbackQuery(token string, callbackQueryID string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", token)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
	}
	sendRequest(url, payload)
}

// دوال تفعيل مهارات الحساب التجاري (Business Methods)
func setBusinessAccountBio(token string, businessConnectionID string, bio string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountBio", token)
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"bio":                    bio,
	}
	sendRequest(url, payload)
}

func setBusinessAccountProfilePhoto(token string, businessConnectionID string, photo string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountProfilePhoto", token)
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"photo":                  photo,
	}
	sendRequest(url, payload)
}

func postStory(token string, businessConnectionID string, content map[string]interface{}) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/postStory", token)
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"content":                content,
	}
	sendRequest(url, payload)
}

func sendRequest(url string, payload map[string]interface{}) {
	body, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}
