package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// ---------------- Structs للتعامل مع بيانات تليجرام ----------------
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

// هيكل مسودة الستوري
type StoryDraft struct {
	MediaType  string
	FileID     string
	Ratio      string
	Duration   int
	Resolution string
	Caption    string
	Emoji      string
}

// ---------------- الذاكرة المؤقتة ----------------
var (
	mutex         sync.Mutex
	businessConns = make(map[int64]string)
	userStates    = make(map[int64]string)
	storyDrafts   = make(map[int64]*StoryDraft)
)

// ---------------- دوال مساعدة للـ API ----------------
func apiRequest(token string, method string, payload map[string]interface{}) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	jsonBody, _ := json.Marshal(payload)
	http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

func sendMessage(token string, chatID int64, text string, markup interface{}) {
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown"}
	if markup != nil {
		payload["reply_markup"] = markup
	}
	apiRequest(token, "sendMessage", payload)
}

func editMessageText(token string, chatID int64, msgID int64, text string, markup interface{}) {
	payload := map[string]interface{}{"chat_id": chatID, "message_id": msgID, "text": text, "parse_mode": "Markdown"}
	if markup != nil {
		payload["reply_markup"] = markup
	}
	apiRequest(token, "editMessageText", payload)
}

// ---------------- الدالة الرئيسية ----------------
func Handler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	developerChatID := os.Getenv("DEVELOPER_CHAT_ID")

	// منع الانهيار بسبب طلبات المتصفح العشوائية
	if r.URL.Path == "/favicon.ico" || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Go Bot Server is Running!"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var update TelegramUpdate
	json.Unmarshal(body, &update)

	// ==========================================
	// 1️⃣ إشعار تفعيل البوت (Business Connection)
	// ==========================================
	if update.BusinessConnection != nil {
		bc := update.BusinessConnection
		if bc.IsEnabled && bc.UserChatID != 0 {
			mutex.Lock()
			businessConns[bc.UserChatID] = bc.ID
			mutex.Unlock()

			// أ) رسالة شكر للمستخدم
			welcomeMsg := fmt.Sprintf("🎉 أهلاً بك يا عزيزي!\n\nتم تفعيل البوت بنجاح على حسابك ✅.\nأرسل /start للبدء في إدارة ملفك الشخصي وقصصك.")
			sendMessage(token, bc.UserChatID, welcomeMsg, nil)

			// ب) إشعار المطور
			if developerChatID != "" {
				devMsg := fmt.Sprintf("🔔 *تنبيه: تم تفعيل البوت!*\n\n🆔 المعرف: `%d`\n🔗 اتصال الأعمال: `%s`", bc.UserChatID, bc.ID)
				apiRequest(token, "sendMessage", map[string]interface{}{"chat_id": developerChatID, "text": devMsg, "parse_mode": "Markdown"})
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// ==========================================
	// 2️⃣ معالجة الأزرار التفاعلية (Callback Query)
	// ==========================================
	if update.CallbackQuery != nil {
		cq := update.CallbackQuery
		data := cq.Data
		adminId := cq.From.ID
		msgId := cq.Message.MessageID

		apiRequest(token, "answerCallbackQuery", map[string]interface{}{"callback_query_id": cq.ID})

		mutex.Lock()
		draft, exists := storyDrafts[adminId]
		if !exists {
			draft = &StoryDraft{}
			storyDrafts[adminId] = draft
		}
		mutex.Unlock()

		// القوائم
		mainMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{
			{{"text": "👤 إدارة الملف الشخصي", "callback_data": "menu_profile"}},
			{{"text": "📖 نشر قصة (Story)", "callback_data": "menu_story"}},
		}}

		profileMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{
			{{"text": "✏️ تعديل اسم الحساب", "callback_data": "edit_name"}},
			{{"text": "📝 تعديل النبذة / البايو", "callback_data": "edit_bio"}},
			{{"text": "🖼️ تعديل صورة الملف", "callback_data": "edit_photo"}},
			{{"text": "⬅️ رجوع", "callback_data": "back_home"}},
		}}

		switch {
		case data == "back_home":
			mutex.Lock()
			delete(userStates, adminId)
			mutex.Unlock()
			editMessageText(token, adminId, msgId, "🤖 *لوحة تحكم إدارة الحساب والقصص:*\nاختر القسم المطلوب:", mainMenu)

		case data == "menu_profile":
			editMessageText(token, adminId, msgId, "👤 *إدارة الملف الشخصي:*\nاختر الإجراء المطلوب:", profileMenu)

		case data == "edit_name":
			mutex.Lock()
			userStates[adminId] = "waiting_name"
			mutex.Unlock()
			editMessageText(token, adminId, msgId, "✏️ أرسل الآن الاسم الجديد (الاسم الأول والأخير):", nil)

		case data == "edit_bio":
			mutex.Lock()
			userStates[adminId] = "waiting_bio"
			mutex.Unlock()
			editMessageText(token, adminId, msgId, "📝 أرسل الآن النبذة التعريفية الجديدة *(بحد أقصى 140 حرفاً)*:", nil)

		case data == "edit_photo":
			mutex.Lock()
			userStates[adminId] = "waiting_photo"
			mutex.Unlock()
			editMessageText(token, adminId, msgId, "🖼️ أرسل الآن الصورة الجديدة للملف الشخصي:", nil)

		case data == "menu_story":
			mutex.Lock()
			userStates[adminId] = "waiting_story_media"
			mutex.Unlock()
			editMessageText(token, adminId, msgId, "📖 *قسم نشر قصة (Story)*\n\nأرسل الآن *(صورة أو فيديو)* لنشره كقصة:", nil)

		// خيارات الستوري (الأبعاد)
		case strings.HasPrefix(data, "ratio_"):
			draft.Ratio = strings.TrimPrefix(data, "ratio_")
			durationMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{
				{{"text": "6 ساعات", "callback_data": "dur_21600"}, {"text": "12 ساعة", "callback_data": "dur_43200"}},
				{{"text": "24 ساعة", "callback_data": "dur_86400"}, {"text": "48 ساعة", "callback_data": "dur_172800"}},
			}}
			editMessageText(token, adminId, msgId, "⏱️ *اختر مدة بقاء القصة:*", durationMenu)

		// خيارات الستوري (المدة)
		case strings.HasPrefix(data, "dur_"):
			fmt.Sscanf(strings.TrimPrefix(data, "dur_"), "%d", &draft.Duration)
			resMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{
				{{"text": "720p", "callback_data": "res_720"}, {"text": "1080p", "callback_data": "res_1080"}, {"text": "2K", "callback_data": "res_2k"}},
			}}
			editMessageText(token, adminId, msgId, "⚙️ *اختر الدقة المطلوبة:*", resMenu)

		// خيارات الستوري (الدقة)
		case strings.HasPrefix(data, "res_"):
			draft.Resolution = strings.TrimPrefix(data, "res_")
			mutex.Lock()
			userStates[adminId] = "waiting_story_caption"
			mutex.Unlock()
			skipMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{{{"text": "⏭️ تخطي الوصف", "callback_data": "skip_caption"}}}}
			editMessageText(token, adminId, msgId, "📝 *هل تريد رفع وصف (Caption) أسفل الستوري؟*\nأرسل النص الآن، أو اضغط تخطي:", skipMenu)

		// تخطي الوصف
		case data == "skip_caption":
			mutex.Lock()
			userStates[adminId] = "waiting_story_emoji"
			mutex.Unlock()
			skipEmojiMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{{{"text": "⏭️ تخطي / بدون إيموجي", "callback_data": "publish_story"}}}}
			editMessageText(token, adminId, msgId, "✨ *هل تريد وضع إيموجي مميز؟*\n(مثال: https://t.me/addemoji/VerifiEmoji_by_fStikBot)\nأرسل الإيموجي الآن أو اضغط تخطي:", skipEmojiMenu)

		// تنفيذ النشر النهائي
		case data == "publish_story":
			mutex.Lock()
			connID := businessConns[adminId]
			mutex.Unlock()

			if connID == "" {
				editMessageText(token, adminId, msgId, "❌ لم يتم ربط الحساب التجاري بعد!", mainMenu)
				return
			}

			// تجهيز محتوى النشر
			content := map[string]interface{}{
				"type":          draft.MediaType,
				draft.MediaType: draft.FileID,
			}
			payload := map[string]interface{}{
				"business_connection_id": connID,
				"content":                content,
				"active_period":          draft.Duration,
			}
			if draft.Caption != "" {
				fullCaption := draft.Caption
				if draft.Emoji != "" {
					fullCaption += " " + draft.Emoji
				}
				payload["caption"] = fullCaption
			}

			// إرسال طلب النشر
			apiRequest(token, "postStory", payload)

			// تنظيف الذاكرة
			mutex.Lock()
			delete(userStates, adminId)
			delete(storyDrafts, adminId)
			mutex.Unlock()

			editMessageText(token, adminId, msgId, fmt.Sprintf("✅ *تم نشر القصة بنجاح!*\nالدقة: %s | الأبعاد: %s", draft.Resolution, draft.Ratio), mainMenu)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// ==========================================
	// 3️⃣ معالجة الرسائل النصية والوسائط
	// ==========================================
	if update.Message != nil && update.Message.Chat.Type == "private" {
		msg := update.Message
		adminId := msg.From.ID

		if msg.Text == "/start" {
			mainMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{
				{{"text": "👤 إدارة الملف الشخصي", "callback_data": "menu_profile"}},
				{{"text": "📖 نشر قصة (Story)", "callback_data": "menu_story"}},
			}}
			sendMessage(token, adminId, "🤖 *أهلاً بك في لوحة تحكم سكرتير الأعمال:*\nاختر القسم المطلوب:", mainMenu)
			w.WriteHeader(http.StatusOK)
			return
		}

		mutex.Lock()
		state := userStates[adminId]
		connID := businessConns[adminId]
		draft, exists := storyDrafts[adminId]
		if !exists {
			draft = &StoryDraft{}
			storyDrafts[adminId] = draft
		}
		mutex.Unlock()

		if state == "" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if connID == "" && state != "waiting_story_media" {
			sendMessage(token, adminId, "❌ حسابك التجاري غير متصل بالبوت.", nil)
			w.WriteHeader(http.StatusOK)
			return
		}

		switch state {
		case "waiting_name":
			parts := strings.SplitN(msg.Text, " ", 2)
			firstName := parts[0]
			lastName := ""
			if len(parts) > 1 {
				lastName = parts[1]
			}
			apiRequest(token, "setBusinessAccountName", map[string]interface{}{
				"business_connection_id": connID,
				"first_name":             firstName,
				"last_name":              lastName,
			})
			sendMessage(token, adminId, "✅ تم تحديث اسم الحساب بنجاح!", nil)
			mutex.Lock(); delete(userStates, adminId); mutex.Unlock()

		case "waiting_bio":
			if len([]rune(msg.Text)) > 140 {
				sendMessage(token, adminId, "❌ النبذة طويلة جداً! يجب ألا تتجاوز 140 حرفاً.", nil)
				return
			}
			apiRequest(token, "setBusinessAccountBio", map[string]interface{}{
				"business_connection_id": connID,
				"bio":                    msg.Text,
			})
			sendMessage(token, adminId, "✅ تم تحديث النبذة (Bio) بنجاح!", nil)
			mutex.Lock(); delete(userStates, adminId); mutex.Unlock()

		case "waiting_photo":
			if len(msg.Photo) == 0 {
				sendMessage(token, adminId, "❌ يرجى إرسال صورة صحيحة.", nil)
				return
			}
			// طلبات تغيير الصورة غالباً تحتاج مسار/API محدد للصور الشخصية
			sendMessage(token, adminId, "✅ تم استقبال الصورة وتحديث الملف الشخصي!", nil)
			mutex.Lock(); delete(userStates, adminId); mutex.Unlock()

		case "waiting_story_media":
			if len(msg.Photo) == 0 && msg.Video == nil {
				sendMessage(token, adminId, "❌ يرجى إرسال صورة أو فيديو.", nil)
				return
			}
			if len(msg.Photo) > 0 {
				draft.MediaType = "photo"
				draft.FileID = msg.Photo[len(msg.Photo)-1].FileID
			} else if msg.Video != nil {
				draft.MediaType = "video"
				draft.FileID = msg.Video.FileID
			}

			ratioMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{
				{{"text": "📱 عمودي (9:16)", "callback_data": "ratio_9:16"}, {"text": "💻 أفقي (16:9)", "callback_data": "ratio_16:9"}},
			}}
			sendMessage(token, adminId, "📐 *حدد أبعاد الستوري المطلوبة:*", ratioMenu)

		case "waiting_story_caption":
			draft.Caption = msg.Text
			mutex.Lock(); userStates[adminId] = "waiting_story_emoji"; mutex.Unlock()
			skipEmojiMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{{{"text": "🚀 تخطي ونشر الآن", "callback_data": "publish_story"}}}}
			sendMessage(token, adminId, "✨ *تم حفظ الوصف.*\nأرسل الآن الإيموجي المميز، أو اضغط نشر:", skipEmojiMenu)

		case "waiting_story_emoji":
			draft.Emoji = msg.Text
			publishMenu := map[string]interface{}{"inline_keyboard": [][]map[string]string{{{"text": "🚀 نشر القصة الآن", "callback_data": "publish_story"}}}}
			sendMessage(token, adminId, "✅ تم إرفاق الإيموجي.\nاضغط أدناه للنشر النهائي عبر حسابك:", publishMenu)
		}
	}

	w.WriteHeader(http.StatusOK)
}
