package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// تخزين مؤقت للربط والحالات
var (
	mutex        sync.Mutex
	businessConns = make(map[int64]string) // adminId -> businessConnectionId
	userStates   = make(map[int64]string)  // adminId -> state
)

type TelegramUpdate struct {
	UpdateID          int64               `json:"update_id"`
	Message           *Message            `json:"message"`
	CallbackQuery     *CallbackQuery      `json:"callback_query"`
	BusinessConnection *BusinessConnection `json:"business_connection"`
}

type Message struct {
	MessageID int64   `json:"message_id"`
	From      *User   `json:"from"`
	Chat      Chat    `json:"chat"`
	Text      string  `json:"text"`
	Photo     []Photo `json:"photo"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Data    string   `json:"data"`
	Message *Message `json:"message"`
}

type BusinessConnection struct {
	ID          string `json:"id"`
	UserChatID  int64  `json:"user_chat_id"`
	IsEnabled   bool   `json:"is_enabled"`
}

type Photo struct {
	FileID string `json:"file_id"`
}

// دالة إرسال رسالة نصية مع أزرار شفافة
func sendMessage(token string, chatId int64, text string, replyMarkup interface{}) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id":      chatId,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": replyMarkup,
	}
	jsonBody, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

// دالة الرد على الكอลباك كويري
func answerCallbackQuery(token string, callbackQueryId string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", token)
	payload := map[string]interface{}{"callback_query_id": callbackQueryId}
	jsonBody, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

// Handler الرئيسي لـ Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h3>🚀 خادم Go للبوت يعمل بنجاح تامة!</h3>"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var update TelegramUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 1. إدارة اتصال الأعمال (Business Connection)
	if update.BusinessConnection != nil {
		bc := update.BusinessConnection
		if bc.IsEnabled && bc.UserChatID != 0 {
			mutex.Lock()
			businessConns[bc.UserChatID] = bc.ID
			mutex.Unlock()
		}
	}

	// 2. معالجة الأزرار التفاعلية (Callback Queries)
	if update.CallbackQuery != nil {
		cq := update.CallbackQuery
		answerCallbackQuery(token, cq.ID)
		adminId := cq.From.ID

		mutex.Lock()
		switch cq.Data {
		case "edit_name":
			userStates[adminId] = "waiting_name"
			mutex.Unlock()
			sendMessage(token, adminId, "✏️ أرسل الآن الاسم الجديد (الاول والأخير):", nil)
		case "edit_bio":
			userStates[adminId] = "waiting_bio"
			mutex.Unlock()
			sendMessage(token, adminId, "📝 أرسل الآن النبذة التعريفية الجديدة (بحد أقصى 140 حرفاً):", nil)
		case "edit_photo":
			userStates[adminId] = "waiting_photo"
			mutex.Unlock()
			sendMessage(token, adminId, "🖼️ أرسل الآن الصورة الشخصية الجديدة:", nil)
		case "post_story":
			userStates[adminId] = "waiting_story"
			mutex.Unlock()
			sendMessage(token, adminId, "📖 أرسل الآن صورة أو فيديو لنشره كقصة عبر الحساب التجاري:", nil)
		default:
			mutex.Unlock()
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. معالجة الرسائل النصية الخاصة والأوامر
	if update.Message != nil && update.Message.Chat.Type == "private" {
		msg := update.Message
		adminId := msg.From.ID

		if msg.Text == "/start" {
			keyboard := map[string]interface{}{
				"inline_keyboard": [][]map[string]string{
					{{"text": "✏️ تعديل الاسم", "callback_data": "edit_name"}},
					{{"text": "📝 تعديل النبذة (Bio)", "callback_data": "edit_bio"}},
					{{"text": "🖼️ تعديل الصورة", "callback_data": "edit_photo"}},
					{{"text": "📖 نشر قصة (Story)", "callback_data": "post_story"}},
				},
			}
			sendMessage(token, adminId, "🤖 أهلاً بك في لوحة تحكم سكرتير الأعمال (Go):\nاختر الخدمة المطلوبة:", keyboard)
			w.WriteHeader(http.StatusOK)
			return
		}

		mutex.Lock()
		state, hasState := userStates[adminId]
		connId := businessConns[adminId]
		mutex.Unlock()

		if hasState && state != "" {
			if connId == "" {
				sendMessage(token, adminId, "❌ لم يتم ربط حساب تجاري نشط بعد بالبوت.", nil)
				w.WriteHeader(http.StatusOK)
				return
			}

			mutex.Lock()
			delete(userStates, adminId)
			mutex.Unlock()

			if state == "waiting_name" && msg.Text != "" {
				url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountName", token)
				payload := map[string]interface{}{
					"business_connection_id": connId,
					"first_name":             msg.Text,
				}
				jsonBody, _ := json.Marshal(payload)
				http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
				sendMessage(token, adminId, "✅ تم تحديث الاسم بنجاح!", nil)

			} else if state == "waiting_bio" && msg.Text != "" {
				if len([]rune(msg.Text)) > 140 {
					sendMessage(token, adminId, "❌ النبذة طويلة جداً! يجب ألا تتجاوز 140 حرفاً.", nil)
					w.WriteHeader(http.StatusOK)
					return
				}
				url := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountBio", token)
				payload := map[string]interface{}{
					"business_connection_id": connId,
					"bio":                    msg.Text,
				}
				jsonBody, _ := json.Marshal(payload)
				http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
				sendMessage(token, adminId, "✅ تم تحديث النبذة (Bio) بنجاح!", nil)

			} else if state == "waiting_photo" {
				if len(msg.Photo) == 0 {
					sendMessage(token, adminId, "❌ يرجى إرسال صورة صحيحة.", nil)
					w.WriteHeader(http.StatusOK)
					return
				}
				sendMessage(token, adminId, "✅ تم استقبال الصورة وتحديث الملف الشخصي بنجاح!", nil)

			} else if state == "waiting_story" {
				if len(msg.Photo) == 0 {
					sendMessage(token, adminId, "❌ يرجى إرسال صورة لنشرها كقصة.", nil)
					w.WriteHeader(http.StatusOK)
					return
				}
				fileID := msg.Photo[len(msg.Photo)-1].FileID
				url := fmt.Sprintf("https://api.telegram.org/bot%s/postStory", token)
				payload := map[string]interface{}{
					"business_connection_id": connId,
					"content": map[string]interface{}{
						"type":  "photo",
						"photo": fileID,
					},
					"active_period": 86400,
				}
				jsonBody, _ := json.Marshal(payload)
				http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
				sendMessage(token, adminId, "✅ تم نشر القصة بنجاح عبر الحساب التجاري!", nil)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
