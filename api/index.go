package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// هياكل بيانات تليجرام
type TelegramUpdate struct {
	UpdateID           int64               `json:"update_id"`
	Message            *Message            `json:"message"`
	CallbackQuery      *CallbackQuery      `json:"callback_query"`
	BusinessConnection *BusinessConnection `json:"business_connection"`
}

type Message struct {
	MessageID int64   `json:"message_id"`
	From      *User   `json:"from"`
	Chat      Chat    `json:"chat"`
	Text      string  `json:"text"`
	Photo     []Photo `json:"photo"`
	Video     *Video  `json:"video"`
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
	ID         string `json:"id"`
	UserChatID int64  `json:"user_chat_id"`
	IsEnabled  bool   `json:"is_enabled"`
}

type Photo struct {
	FileID string `json:"file_id"`
}

type Video struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}

// دالة مساعدة لإرسال الرسائل مع الأزرار
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

// دالة الرد على الأزرار الشفافة
func answerCallbackQuery(token string, callbackQueryId string, text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", token)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryId,
		"text":              text,
	}
	jsonBody, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

// المالدلر الرئيسي لـ Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	developerChatID := os.Getenv("DEVELOPER_CHAT_ID") // معرف المطور لإرسال إشعار التفعيل

	// منع انهيار طلبات المتصفح أو الأيقونات
	if r.URL.Path == "/favicon.ico" || (r.Method != http.MethodPost && !r.URL.Query().Has("hook")) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h2 dir='rtl'>🚀 خادم Go للبوت يعمل بنجاح تام!</h2>"))
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

	// 1️⃣ مراقبة تفعيل البوت (Business Connection) وإرسال التنبيهات
	if update.BusinessConnection != nil {
		bc := update.BusinessConnection
		if bc.IsEnabled && bc.UserChatID != 0 {
			userChatId := bc.UserChatID
			
			// أ) إرسال رسالة شكر للمستخدم مع ذكر اسمه
			urlUser := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%d", token, userChatId)
			resp, err := http.Get(urlUser)
			userName := "عزيزي المستخدم"
			if err == nil && resp.StatusCode == http.StatusOK {
				var chatRes struct {
					Result struct {
						FirstName string `json:"first_name"`
					} `json:"result"`
				}
				json.NewDecoder(resp.Body).Decode(&chatRes)
				if chatRes.Result.FirstName != "" {
					userName = chatRes.Result.FirstName
				}
			}

			welcomeMsg := fmt.Sprintf("🎉 أهلاً بك يا *%s*!\n\nتم تفعيل البوت على حسابك بنجاح ✅.\nيمكنك الآن إرسال /start لإدارة ملفك الشخصي ونشر القصص.", userName)
			sendMessage(token, userChatId, welcomeMsg, nil)

			// ب) إرسال إشعار لمطور البوت بأن شخصاً ما قام بتفعيل البوت
			if developerChatID != "" {
				var devChatIdInt int64
				fmt.Sscanf(developerChatID, "%d", &devChatIdInt)
				devMsg := fmt.Sprintf("🔔 *تنبيه: قام شخص بتفعيل البوت!*\n\n👤 الاسم: %s\n🆔 المعرف: `%d`\n🔗 اتصال الأعمال: `%s`", userName, userChatId, bc.ID)
				sendMessage(token, devChatIdInt, devMsg, nil)
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2️⃣ معالجة الأزرار التفاعلية (Callback Queries)
	if update.CallbackQuery != nil {
		cq := update.CallbackQuery
		answerCallbackQuery(token, cq.ID, "تم التحديد بنجاح")
		adminId := cq.From.ID

		keyboardMain := map[string]interface{}{
			"inline_keyboard": [][]map[string]string{
				{{"text": "👤 إدارة الملف الشخصي", "callback_data": "menu_profile"}},
				{{"text": "📖 نشر قصة (Story)", "callback_data": "menu_story"}},
			},
		}

		keyboardProfile := map[string]interface{}{
			"inline_keyboard": [][]map[string]string{
				{{"text": "✏️ تعديل اسم الحساب", "callback_data": "edit_name"}},
				{{"text": "📝 تعديل النبذة / البايو (140 حرف)", "callback_data": "edit_bio"}},
				{{"text": "🖼️ تعديل صورة الملف الشخصي", "callback_data": "edit_photo"}},
				{{"text": "⬅️ رجوع للقائمة الرئيسية", "callback_data": "back_home"}},
			},
		}

		switch cq.Data {
		case "menu_profile":
			sendMessage(token, adminId, "👤 *قسم إدارة الملف الشخصي:*\nاختر الإجراء المطلوب:", keyboardProfile)
		case "menu_story":
			sendMessage(token, adminId, "📖 *قسم نشر قصة (Story)*\n\nأرسل الآن *(صورة أو فيديو)* لنشره كقصة عبر حسابك التجاري:", nil)
		case "back_home":
			sendMessage(token, adminId, "🤖 *لوحة تحكم إدارة الحساب والقصص:*", keyboardMain)
		case "edit_name":
			sendMessage(token, adminId, "✏️ أرسل الآن الاسم الجديد (الاسم الأول والأخير):", nil)
		case "edit_bio":
			sendMessage(token, adminId, "📝 أرسل الآن النبذة التعريفية الجديدة *(بحد أقصى 140 حرفاً)*:", nil)
		case "edit_photo":
			sendMessage(token, adminId, "🖼️ أرسل الآن الصورة الجديدة للملف الشخصي:", nil)
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// 3️⃣ معالجة الرسائل والأوامر النصية الخاصة
	if update.Message != nil && update.Message.Chat.Type == "private" {
		msg := update.Message
		adminId := msg.From.ID

		if msg.Text == "/start" {
			keyboardMain := map[string]interface{}{
				"inline_keyboard": [][]map[string]string{
					{{"text": "👤 إدارة الملف الشخصي", "callback_data": "menu_profile"}},
					{{"text": "📖 نشر قصة (Story)", "callback_data": "menu_story"}},
				},
			}
			sendMessage(token, adminId, "🤖 *أهلاً بك في لوحة تحكم إدارة الحساب والقصص (Go):*\nاختر القسم المطلوب:", keyboardMain)
		}
	}

	w.WriteHeader(http.StatusOK)
}
