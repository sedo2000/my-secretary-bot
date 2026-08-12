package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// تخزين الحالة والاتصالات مؤقتاً للجلسات
var userStates = make(map[int64]string)
var businessConnections = make(map[int64]string)
var welcomedUsers = make(map[int64]bool)

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
		w.Write([]byte("Go Bot Server is Running Successfully!"))
		return
	}

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	// 1. التعامل مع أحداث ربط البوت بالحساب (مع منع تكرار رسالة الترحيب نهائياً)
	if update.BusinessConnection != nil {
		userID := update.BusinessConnection.User.ID
		if update.BusinessConnection.IsEnabled {
			businessConnections[userID] = update.BusinessConnection.ID
			if !welcomedUsers[userID] {
				welcomedUsers[userID] = true
				sendTelegramMessage(token, userID,
					"🎉 أهلاً بك يا عزيزي!\n\nتم تفعيل البوت بنجاح على حسابك ✅.\nأرسل /start للبدء في إدارة ملفك الشخصي وقصصك.",
					getMainMenuKeyboard())
			}
		} else {
			delete(businessConnections, userID)
			delete(welcomedUsers, userID)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. التعامل مع الضغط على الأزرار (Callback Queries) وتفعيل زر الرجوع
	if update.CallbackQuery != nil {
		query := update.CallbackQuery
		chatID := query.Message.Chat.ID
		userID := query.From.ID
		data := query.Data

		switch data {
		case "menu_profile":
			userStates[userID] = ""
			editMessageText(token, chatID, query.Message.MessageID,
				"👤 **إدارة الملف الشخصي:**\nاختر العملية المطلوبة:", getProfileKeyboard())
		case "menu_stories":
			userStates[userID] = ""
			editMessageText(token, chatID, query.Message.MessageID,
				"📸 **إدارة القصص:**\nاختر العملية المطلوبة:", getStoriesKeyboard())
		case "edit_bio":
			userStates[userID] = "awaiting_bio"
			editMessageText(token, chatID, query.Message.MessageID,
				"✍️ أرسل البايو الجديد الآن لتحديثه في حسابك (بحد أقصى 140 حرفاً):", getBackOnlyKeyboard())
		case "edit_photo":
			userStates[userID] = "awaiting_photo"
			editMessageText(token, chatID, query.Message.MessageID,
				"🖼️ أرسل الصورة الجديدة للملف الشخصي:", getBackOnlyKeyboard())
		case "post_story":
			userStates[userID] = "awaiting_story"
			editMessageText(token, chatID, query.Message.MessageID,
				"🚀 أرسل الوسائط (صورة أو فيديو) لنشرها كقصة جديدة:", getBackOnlyKeyboard())
		case "main_menu":
			userStates[userID] = ""
			editMessageText(token, chatID, query.Message.MessageID,
				"🏠 **القائمة الرئيسية:**\nاختر من الأقسام التالية:", getMainMenuKeyboard())
		}

		answerCallbackQuery(token, query.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. التعامل مع الرسائل النصية والوسائط (تنفيذ العمليات الفعلي)
	if update.Message != nil {
		msg := update.Message
		chatID := msg.Chat.ID
		userID := msg.From.ID
		text := msg.Text
		bizConnID := businessConnections[userID]

		if text == "/start" {
			userStates[userID] = ""
			sendTelegramMessage(token, chatID,
				"🏠 **القائمة الرئيسية لإدارة حسابك:**\nاختر القسم المناسب:", getMainMenuKeyboard())
			w.WriteHeader(http.StatusOK)
			return
		}

		state := userStates[userID]

		// تنفيذ تعديل البايو
		if state == "awaiting_bio" {
			if bizConnID == "" {
				sendTelegramMessage(token, chatID, "⚠️ تنبيه: لم يتم العثور على اتصال تجاري نشط. تأكد من ربط البوت بحسابك أولاً.", getMainMenuKeyboard())
			} else {
				success := setBusinessAccountBio(token, bizConnID, text)
				if success {
					sendTelegramMessage(token, chatID, "✅ تم تحديث البايو بنجاح!", getMainMenuKeyboard())
				} else {
					sendTelegramMessage(token, chatID, "❌ فشل تحديث البايو. تأكد من أن النص أقل من 140 حرفاً وأن البوت يملك صلاحيات التعديل.", getMainMenuKeyboard())
				}
			}
			userStates[userID] = ""
			w.WriteHeader(http.StatusOK)
			return
		}

		// تنفيذ تعديل صورة الملف الشخصي
		if state == "awaiting_photo" {
			if bizConnID == "" {
				sendTelegramMessage(token, chatID, "⚠️ تنبيه: لم يتم العثور على اتصال تجاري نشط.", getMainMenuKeyboard())
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
			userStates[userID] = ""
			w.WriteHeader(http.StatusOK)
			return
		}

		// تنفيذ نشر القصة
		if state == "awaiting_story" {
			if bizConnID == "" {
				sendTelegramMessage(token, chatID, "⚠️ تنبيه: لم يتم العثور على اتصال تجاري نشط.", getMainMenuKeyboard())
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
						sendTelegramMessage(token, chatID, "❌ فشل نشر القصة. تأكد من تفعيل صلاحيات القصص للبوت.", getMainMenuKeyboard())
					}
				} else {
					sendTelegramMessage(token, chatID, "⚠️ يرجى إرسال صورة أو فيديو لنشره كقصة.", getBackOnlyKeyboard())
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			userStates[userID] = ""
			w.WriteHeader(http.StatusOK)
			return
		}

		sendTelegramMessage(token, chatID, "أهلاً بك! استخدم الأزرار أدناه لإدارة حسابك:", getMainMenuKeyboard())
	}

	w.WriteHeader(http.StatusOK)
}

// --- لوحات المفاتيح (مع زر الرجوع في كل قائمة) ---

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

// --- دوال الاتصال بـ Telegram Bot API وتوابع البزنس الرسمية ---

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

// دالة تعديل البايو (Business API)
func setBusinessAccountBio(token string, businessConnectionID string, bio string) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountBio", token)
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"bio":                    bio,
	}
	return sendRequest(url, payload)
}

// دالة تعديل صورة الملف الشخصي (Business API)
func setBusinessAccountProfilePhoto(token string, businessConnectionID string, photoFileID string) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountProfilePhoto", token)
	inputPhoto := map[string]interface{}{
		"type":    "photo",
		"file_id": photoFileID,
	}
	payload := map[string]interface{}{
		"business_connection_id": businessConnectionID,
		"photo":                  inputPhoto,
	}
	return sendRequest(url, payload)
}

// دالة نشر القصص (Business API)
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
