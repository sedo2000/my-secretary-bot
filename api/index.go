package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- ثوابت المطور ---
const DeveloperID int64 = 8705163117
const DeveloperUsername = "Xhwe2"

// عميل HTTP عام مع timeout قصير للطلبات النصية العادية
var httpClient = &http.Client{Timeout: 8 * time.Second}

// عميل بـ timeout أطول لعمليات تنزيل/رفع الصور والفيديوهات
var mediaClient = &http.Client{Timeout: 30 * time.Second}

// متغيرات نظام التهدئة (Cooldown) والتخزين المؤقت لاتصال الأعمال
var (
	cooldownMu    sync.Mutex
	userCooldowns = make(map[int64]map[int64]time.Time) // map[adminID]map[senderID]expiryTime

	bizCacheMu sync.Mutex
	bizCache   = make(map[string]int64)
)

// صورة الترحيب التي تُرسل عند الضغط على /start
const startPhotoURL = "https://od.lk/s/M18zMzMwODEzNDNfV3R3TEM/IMG_20260810_235848_327.jpg"

// قائمة الاقتباسات
var quotes = []string{
	"قاوم ما تكره لتصل الى ما تحب",
	"الحرب بين أنت ضد أنت",
	"لا تسألني من أنا",
	"أبنِ نفسك بنفسك لنفسك",
	"ميخالف",
	"حتى لو متأخر تگدر..!",
	"من يعيش في خوف لن يكون حراً ابداً",
	"لا أبرح حتى أبلغ",
	"لا أجدني بينهم",
	"كل شيء يريدك عندما لاتريد شيئاً",
	"أنه مبرمج فحسب",
	"أنا لا افكر فيك ابداً",
	"المرء نتاج خلواته",
	"لا مزيد من الأصدقاء المزيفين",
}

// --- قاموس الترجمة: عربي (افتراضي) وإنجليزي ---
var translations = map[string]map[string]string{
	"ar": {
		"main_menu_title":        "القائمة الرئيسية 🤖:",
		"welcome":                "أهلاً بك في البوت 🤖\nاستخدم الأزرار أدناه:",
		"stop_btn":               "🛑 إيقاف الرد",
		"start_btn":              "🟢 تشغيل الرد",
		"edit_text_btn":          "📝 تعديل نص الرد",
		"exclude_btn":            "👤 استثناء حساب",
		"list_excluded_btn":      "📋 عرض المستثنين",
		"clear_excluded_btn":     "🧹 مسح المستثنين",
		"profile_menu_btn":       "🧑 إدارة الملف الشخصي",
		"post_story_btn":         "📖 نشر قصة",
		"lang_ar_btn":            "🇮🇶 العربية",
		"lang_en_btn":            "🇺🇸 English",
		"back_btn":               "🔙 رجوع",
		"stopped_msg":            "🛑 تم إيقاف الرد التلقائي بنجاح.",
		"started_msg":            "🟢 تم تشغيل الرد التلقائي بنجاح.",
		"edit_text_prompt":       "📝 أرسل الآن نص الرد التلقائي الجديد:",
		"saved_text_msg":         "✅ تم حفظ نص الرد التلقائي الجديد بنجاح!",
		"exclude_prompt":         "👤 أرسل ايدي الحساب المراد استثناؤه الآن:",
		"invalid_id_msg":         "❌ أرقام فقط! أرسل الايدي بشكل صحيح.",
		"id_added_msg":           "✅ تم إضافة الايدي `%d` إلى قائمة الاستثناء.",
		"list_excluded_title":    "📋 **قائمة الحسابات المستثناة:**\n",
		"no_excluded":            "لا يوجد حسابات مستثناة حالياً.",
		"cleared_excluded_msg":   "🧹 تم مسح جميع الاستثناءات بنجاح.",
		"profile_menu_title":     "🧑 إدارة الملف الشخصي - اختر ما تريد تعديله:",
		"edit_first_name_btn":    "✏️ تعديل الاسم",
		"edit_bio_btn":           "📝 تعديل النبذة",
		"edit_photo_btn":         "🖼️ تعديل الصورة",
		"edit_username_btn":      "🔗 تعديل اليوزر",
		"no_business_connection": "❌ لم يتم ربط حساب تجاري بعد بالبوت.",
		"first_name_prompt":      "✏️ أرسل الآن الاسم الأول الجديد (والاسم الأخير بعده بمسافة، اختياري):",
		"bio_prompt":             "📝 أرسل الآن النبذة الجديدة (حد أقصى 70 حرف):",
		"username_prompt":        "🔗 أرسل الآن اسم المستخدم الجديد (بدون @):",
		"photo_prompt":           "🖼️ أرسل الآن الصورة الجديدة لملفك الشخصي:",
		"name_updated":           "✅ تم تعديل الاسم بنجاح!",
		"bio_updated":            "✅ تم تعديل النبذة بنجاح!",
		"username_updated":       "✅ تم تعديل اسم المستخدم بنجاح!",
		"photo_updated":          "✅ تم تعديل صورة الملف الشخصي بنجاح!",
		"select_story_duration":  "⏱️ اختر مدة ظهور القصة المطلوبة:",
		"dur_6h":                 "6 ساعات",
		"dur_12h":                "12 ساعة",
		"dur_24h":                "24 ساعة",
		"dur_48h":                "48 ساعة",
		"story_prompt":           "📖 أرسل الآن صورة أو فيديو (حد أقصى 60 ثانية) لنشره كقصة (ستبقى ظاهرة لمدة %s):",
		"story_updated":          "✅ تم نشر القصة بنجاح! ستبقى ظاهرة لمدة %s.",
		"your_id_msg":            "الايدي الخاص بك هو:\n`%d`",
		"fail_name":              "❌ فشل تعديل الاسم: %s",
		"fail_bio":               "❌ فشل تعديل النبذة: %s",
		"fail_username":          "❌ فشل تعديل اليوزر: %s",
		"fail_photo":             "❌ فشل تعديل الصورة: %s",
		"fail_story":             "❌ فشل نشر القصة: %s",
		"need_real_photo":        "❌ أرسل صورة فعلية (لا يقبل ملفات أو نصوص).",
		"need_real_media_story":  "❌ أرسل صورة أو فيديو فعلي لنشره كقصة.",
		"video_too_long_error":   "الفيديو أطول من 60 ثانية، وهذا الحد الأقصى المسموح لقصص تليجرام",
		"id_copy_btn":            "🆔 نسخ الآيدي",
		"share_user_btn":         "👤 مشاركة مستخدم",
		"share_user_prompt":      "👇 استخدم هذا الزر لمشاركة أي مستخدم من قائمة محادثاتك مع البوت، وسيتم استخراج بياناته تلقائياً:",
		"user_shared_info":       "👤 *معلومات المستخدم المُشارك:*\n\nالاسم: %s\nاليوزر: %s\nالآيدي: `%d`",
		"no_username":            "لا يوجد يوزر",
	},
	"en": {
		"main_menu_title":        "Main Menu 🤖:",
		"welcome":                "Welcome to the bot 🤖\nChoose from the buttons below:",
		"stop_btn":               "🛑 Stop Reply",
		"start_btn":              "🟢 Start Reply",
		"edit_text_btn":          "📝 Edit Reply",
		"exclude_btn":            "👤 Exclude",
		"list_excluded_btn":      "📋 View Excluded",
		"clear_excluded_btn":     "🧹 Clear Excluded",
		"profile_menu_btn":       "🧑 Profile",
		"post_story_btn":         "📖 Post Story",
		"lang_ar_btn":            "🇮🇶 العربية",
		"lang_en_btn":            "🇺🇸 English",
		"back_btn":               "🔙 Back",
		"stopped_msg":            "🛑 Auto-reply has been stopped.",
		"started_msg":            "🟢 Auto-reply has been started.",
		"edit_text_prompt":       "📝 Send new auto-reply text:",
		"saved_text_msg":         "✅ Auto-reply saved!",
		"exclude_prompt":         "👤 Send account ID to exclude:",
		"invalid_id_msg":         "❌ Numbers only!",
		"id_added_msg":           "✅ ID `%d` added.",
		"list_excluded_title":    "📋 **Excluded Accounts:**\n",
		"no_excluded":            "None.",
		"cleared_excluded_msg":   "🧹 Exclusions cleared.",
		"profile_menu_title":     "🧑 Profile Management:",
		"edit_first_name_btn":    "✏️ Edit Name",
		"edit_bio_btn":           "📝 Edit Bio",
		"edit_photo_btn":         "🖼️ Edit Photo",
		"edit_username_btn":      "🔗 Edit Username",
		"no_business_connection": "❌ No business account connected.",
		"first_name_prompt":      "✏️ Send new first name:",
		"bio_prompt":             "📝 Send new bio (max 70 chars):",
		"username_prompt":        "🔗 Send new username:",
		"photo_prompt":           "🖼️ Send new photo:",
		"name_updated":           "✅ Name updated!",
		"bio_updated":            "✅ Bio updated!",
		"username_updated":       "✅ Username updated!",
		"photo_updated":          "✅ Photo updated!",
		"select_story_duration":  "⏱️ Select duration:",
		"dur_6h":                 "6 Hours",
		"dur_12h":                "12 Hours",
		"dur_24h":                "24 Hours",
		"dur_48h":                "48 Hours",
		"story_prompt":           "📖 Send photo/video (max 60s):",
		"story_updated":          "✅ Story posted!",
		"your_id_msg":            "Your ID:\n`%d`",
		"fail_name":              "❌ Failed: %s",
		"fail_bio":               "❌ Failed: %s",
		"fail_username":          "❌ Failed: %s",
		"fail_photo":             "❌ Failed: %s",
		"fail_story":             "❌ Failed: %s",
		"need_real_photo":        "❌ Send actual photo.",
		"need_real_media_story":  "❌ Send actual photo/video.",
		"video_too_long_error":   "Video > 60s",
		"id_copy_btn":            "🆔 Copy ID",
		"share_user_btn":         "👤 Share User",
		"share_user_prompt":      "👇 Use button to share user:",
		"user_shared_info":       "👤 *User Info:*\n\nName: %s\nUsername: %s\nID: `%d`",
		"no_username":            "No username",
	},
}

func tr(lang, key string) string {
	if lang != "en" {
		lang = "ar"
	}
	if val, ok := translations[lang][key]; ok {
		return val
	}
	return key
}

func getDurationLabel(lang, period string) string {
	switch period {
	case "21600":
		return tr(lang, "dur_6h")
	case "43200":
		return tr(lang, "dur_12h")
	case "86400":
		return tr(lang, "dur_24h")
	case "172800":
		return tr(lang, "dur_48h")
	default:
		return tr(lang, "dur_24h")
	}
}

// دالة الترجمة الفورية والكشف التلقائي عن لغة النص
func translateText(text, targetLang string) (string, string, error) {
	if strings.TrimSpace(text) == "" {
		return "", "", nil
	}
	endpoint := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s",
		targetLang, url.QueryEscape(text),
	)

	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if len(result) == 0 {
		return "", "", fmt.Errorf("فشل الترجمة")
	}

	translatedText := ""
	if sentences, ok := result[0].([]interface{}); ok {
		for _, sentence := range sentences {
			if s, ok := sentence.([]interface{}); ok && len(s) > 0 {
				if tText, ok := s[0].(string); ok {
					translatedText += tText
				}
			}
		}
	}

	detectedLang := ""
	if len(result) > 2 {
		if lang, ok := result[2].(string); ok {
			detectedLang = lang
		}
	}

	return translatedText, detectedLang, nil
}

// --- هياكل البيانات ---

type ForceChannel struct {
	ChatID     int64  `json:"chat_id"`
	Username   string `json:"username"`
	Title      string `json:"title"`
	InviteLink string `json:"invite_link"`
}

type UserRecord struct {
	ID   int64 `json:"id"`
	Date int64 `json:"date"`
}

type BotConfig struct {
	// إعدادات Business & Auto-Reply الأصلية
	IsStopped      bool    `json:"is_stopped"`
	AutoReply      string  `json:"auto_reply"`
	Excluded       []int64 `json:"excluded"`
	State          string  `json:"state"`
	BusinessConnID string  `json:"business_conn_id"`
	Lang           string  `json:"lang"`

	// إعدادات لوحة المطور الجديدة
	ForceSubscribe bool           `json:"force_subscribe"`
	ForceChannels  []ForceChannel `json:"force_channels"`
	EntryNotify    bool           `json:"entry_notify"`
	Users          []UserRecord   `json:"users"`
}

type TelegramUpdate struct {
	Message         *Message       `json:"message"`
	CallbackQuery   *CallbackQuery `json:"callback_query"`
	BusinessMessage *struct {
		MessageID int `json:"message_id"`
		Chat      struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			IsBot     bool   `json:"is_bot"`
		} `json:"from"`
		Text                 string `json:"text"`
		IsOutgoing           bool   `json:"is_outgoing"`
		BusinessConnectionID string `json:"business_connection_id"`
	} `json:"business_message"`
	BusinessConnection *struct {
		ID   string `json:"id"`
		User struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
		} `json:"user"`
		UserChatID int64 `json:"user_chat_id"`
		Date       int64 `json:"date"`
		IsEnabled  bool  `json:"is_enabled"`
	} `json:"business_connection"`
}

type PhotoSize struct {
	FileID       string `json:"file_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size"`
}

type Video struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
}

type SharedUserInfo struct {
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type UsersSharedData struct {
	RequestID int64            `json:"request_id"`
	Users     []SharedUserInfo `json:"users"`
}

type Message struct {
	MessageID int `json:"message_id"`
	Chat      struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"chat"`
	From struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"from"`
	Text        string           `json:"text"`
	Photo       []PhotoSize      `json:"photo"`
	Video       *Video           `json:"video"`
	UsersShared *UsersSharedData `json:"users_shared"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
	From    struct {
		ID int64 `json:"id"`
	} `json:"from"`
}

type BusinessConnectionResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
		UserChatID int64 `json:"user_chat_id"`
	} `json:"result"`
}

type ChatMemberResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Status   string `json:"status"`
		IsMember bool   `json:"is_member"`
	} `json:"result"`
}

type ChatResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		ID       int64  `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title"`
		Username string `json:"username"`
	} `json:"result"`
}

// --- الدالة الرئيسية (Handler) ---

func Handler(w http.ResponseWriter, r *http.Request) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if secret := os.Getenv("TELEGRAM_WEBHOOK_SECRET"); secret != "" {
		if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
			log.Println("رفض طلب: secret token غير مطابق")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	var update TelegramUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Println("خطأ في قراءة التحديث:", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. معالجة الضغط على الأزرار الشفافة
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		answerCallback(botToken, cb.ID, "")

		// === حماية لوحة المطور ===
		if strings.HasPrefix(cb.Data, "admin_") {
			if cb.From.ID != DeveloperID {
				answerCallback(botToken, cb.ID, "⛔ هذه اللوحة مخصصة لمطور البوت فقط.")
				w.WriteHeader(http.StatusOK)
				return
			}
			handleAdminCallback(botToken, cb)
			w.WriteHeader(http.StatusOK)
			return
		}

		if cb.Data == "force_check" {
			handleForceCheckCallback(botToken, cb)
			w.WriteHeader(http.StatusOK)
			return
		}

		// === وظائف الأزرار السابقة (باقية كما هي) ===
		if cb.Data == "change_quote" {
			newQuote := quotes[rand.Intn(len(quotes))]
			updateButtonQuote(botToken, cb.Message.Chat.ID, cb.Message.MessageID, newQuote)
			w.WriteHeader(http.StatusOK)
			return
		}

		deleteMessage(botToken, cb.Message.Chat.ID, cb.Message.MessageID)
		adminID := cb.From.ID
		config, msgID := getConfig(botToken, adminID)
		lang := config.Lang

		switch cb.Data {
		case "main_menu":
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "main_menu_title"))
		case "stop":
			config.IsStopped = true
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "stopped_msg"))
		case "start":
			config.IsStopped = false
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "started_msg"))
		case "edit_text":
			config.State = "waiting_text"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "edit_text_prompt"))
		case "exclude":
			config.State = "waiting_id"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "exclude_prompt"))
		case "list_excluded":
			txt := tr(lang, "list_excluded_title")
			if len(config.Excluded) == 0 {
				txt += tr(lang, "no_excluded")
			} else {
				for _, id := range config.Excluded {
					txt += fmt.Sprintf("- `%d`\n", id)
				}
			}
			sendSubMenu(botToken, adminID, lang, txt)
		case "clear_excluded":
			config.Excluded = []int64{}
			saveConfig(botToken, adminID, config, msgID)مرحباً بك. لقد قمت ببناء لوحة المطور الاحترافية المطلوبة مع دمجها بشكل متكامل وآمن داخل الكود الحالي الخاص بك. لم يتم حذف أو تعطيل أي ميزة من ميزات البوت السابقة (نظام Business، الترجمة، الرد التلقائي، Stories، Cooldown، وغيرها).

إليك الحل الشامل والكامل في ملف واحد كما طلبت.

### 1. الكود الكامل (`api/index.go`)
انسخ هذا الكود بالكامل وضعه في ملف `api/index.go`:

```go
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ثوابت المطور
const DeveloperID int64 = 8705163117
const BotUsername = "@Nerd8bot"

// عميل HTTP عام مع timeout قصير للطلبات النصية العادية
var httpClient = &http.Client{Timeout: 8 * time.Second}

// عميل بـ timeout أطول لعمليات تنزيل/رفع الصور والفيديوهات
var mediaClient = &http.Client{Timeout: 30 * time.Second}

// متغيرات نظام التهدئة (Cooldown) والتخزين المؤقت لاتصال الأعمال
var (
	cooldownMu    sync.Mutex
	userCooldowns = make(map[int64]map[int64]time.Time) // map[adminID]map[senderID]expiryTime

	bizCacheMu    sync.Mutex
	bizCache      = make(map[string]int64)

	botIDCache   int64
	botIDCacheMu sync.Mutex
)

// صورة الترحيب التي تُرسل عند الضغط على /start
const startPhotoURL = "[https://od.lk/s/M18zMzMwODEzNDNfV3R3TEM/IMG_20260810_235848_327.jpg](https://od.lk/s/M18zMzMwODEzNDNfV3R3TEM/IMG_20260810_235848_327.jpg)"

// قائمة الاقتباسات
var quotes = []string{
	"قاوم ما تكره لتصل الى ما تحب",
	"الحرب بين أنت ضد أنت",
	"لا تسألني من أنا",
	"أبنِ نفسك بنفسك لنفسك",
	"ميخالف",
	"حتى لو متأخر تگدر..!",
	"من يعيش في خوف لن يكون حراً ابداً",
	"لا أبرح حتى أبلغ",
	"لا أجدني بينهم",
	"كل شيء يريدك عندما لاتريد شيئاً",
	"أنه مبرمج فحسب",
	"أنا لا افكر فيك ابداً",
	"المرء نتاج خلواته",
	"لا مزيد من الأصدقاء المزيفين",
}

// --- قاموس الترجمة: عربي (افتراضي) وإنجليزي ---
var translations = map[string]map[string]string{
	"ar": {
		"main_menu_title":        "القائمة الرئيسية 🤖:",
		"welcome":                "أهلاً بك في لوحة تحكم البوت 🤖\nاختر من الأزرار أدناه للتحكم الكامل:",
		"stop_btn":               "🛑 إيقاف الرد",
		"start_btn":              "🟢 تشغيل الرد",
		"edit_text_btn":          "📝 تعديل نص الرد",
		"exclude_btn":            "👤 استثناء حساب",
		"list_excluded_btn":      "📋 عرض المستثنين",
		"clear_excluded_btn":     "🧹 مسح المستثنين",
		"profile_menu_btn":       "🧑 إدارة الملف الشخصي",
		"post_story_btn":         "📖 نشر قصة",
		"lang_ar_btn":            "🇮🇶 العربية",
		"lang_en_btn":            "🇺🇸 English",
		"back_btn":               "🔙 رجوع",
		"stopped_msg":            "🛑 تم إيقاف الرد التلقائي بنجاح.",
		"started_msg":            "🟢 تم تشغيل الرد التلقائي بنجاح.",
		"edit_text_prompt":       "📝 أرسل الآن نص الرد التلقائي الجديد:",
		"saved_text_msg":         "✅ تم حفظ نص الرد التلقائي الجديد بنجاح!",
		"exclude_prompt":         "👤 أرسل ايدي الحساب المراد استثناؤه الآن:",
		"invalid_id_msg":         "❌ أرقام فقط! أرسل الايدي بشكل صحيح.",
		"id_added_msg":           "✅ تم إضافة الايدي `%d` إلى قائمة الاستثناء.",
		"list_excluded_title":    "📋 **قائمة الحسابات المستثناة:**\n",
		"no_excluded":            "لا يوجد حسابات مستثناة حالياً.",
		"cleared_excluded_msg":   "🧹 تم مسح جميع الاستثناءات بنجاح.",
		"profile_menu_title":     "🧑 إدارة الملف الشخصي - اختر ما تريد تعديله:",
		"edit_first_name_btn":    "✏️ تعديل الاسم",
		"edit_bio_btn":           "📝 تعديل النبذة",
		"edit_photo_btn":         "🖼️ تعديل الصورة",
		"edit_username_btn":      "🔗 تعديل اليوزر",
		"no_business_connection": "❌ لم يتم ربط حساب تجاري بعد بالبوت.",
		"first_name_prompt":      "✏️ أرسل الآن الاسم الأول الجديد (والاسم الأخير بعده بمسافة، اختياري):",
		"bio_prompt":             "📝 أرسل الآن النبذة الجديدة (حد أقصى 70 حرف):",
		"username_prompt":        "🔗 أرسل الآن اسم المستخدم الجديد (بدون @):",
		"photo_prompt":           "🖼️ أرسل الآن الصورة الجديدة لملفك الشخصي:",
		"name_updated":           "✅ تم تعديل الاسم بنجاح!",
		"bio_updated":            "✅ تم تعديل النبذة بنجاح!",
		"username_updated":       "✅ تم تعديل اسم المستخدم بنجاح!",
		"photo_updated":          "✅ تم تعديل صورة الملف الشخصي بنجاح!",
		"select_story_duration":  "⏱️ اختر مدة ظهور القصة المطلوبة:",
		"dur_6h":                 "6 ساعات",
		"dur_12h":                "12 ساعة",
		"dur_24h":                "24 ساعة",
		"dur_48h":                "48 ساعة",
		"story_prompt":           "📖 أرسل الآن صورة أو فيديو (حد أقصى 60 ثانية) لنشره كقصة (ستبقى ظاهرة لمدة %s):",
		"story_updated":          "✅ تم نشر القصة بنجاح! ستبقى ظاهرة لمدة %s.",
		"your_id_msg":            "الايدي الخاص بك هو:\n`%d`",
		"fail_name":              "❌ فشل تعديل الاسم: %s",
		"fail_bio":               "❌ فشل تعديل النبذة: %s",
		"fail_username":          "❌ فشل تعديل اليوزر: %s",
		"fail_photo":             "❌ فشل تعديل الصورة: %s",
		"fail_story":             "❌ فشل نشر القصة: %s",
		"need_real_photo":        "❌ أرسل صورة فعلية (لا يقبل ملفات أو نصوص).",
		"need_real_media_story":  "❌ أرسل صورة أو فيديو فعلي لنشره كقصة.",
		"video_too_long_error":   "الفيديو أطول من 60 ثانية، وهذا الحد الأقصى المسموح لقصص تليجرام",
		"id_copy_btn":            "🆔 نسخ الآيدي",
		"share_user_btn":         "👤 User",
		"share_user_prompt":      "👇 استخدم هذا الزر لمشاركة أي مستخدم من قائمة محادثاتك مع البوت، وسيتم استخراج اسمه ويوزره وآيديه تلقائياً:",
		"user_shared_info":       "👤 *معلومات المستخدم المُشارك:*\n\nالاسم: %s\nاليوزر: %s\nالآيدي: `%d`",
		"no_username":            "لا يوجد يوزر",
	},
	"en": {
		"main_menu_title":        "Main Menu 🤖:",
		"welcome":                "Welcome to the bot control panel 🤖\nChoose from the buttons below for full control:",
		"stop_btn":               "🛑 Stop Auto-Reply",
		"start_btn":              "🟢 Start Auto-Reply",
		"edit_text_btn":          "📝 Edit Reply Text",
		"exclude_btn":            "👤 Exclude Account",
		"list_excluded_btn":      "📋 View Excluded",
		"clear_excluded_btn":     "🧹 Clear Excluded",
		"profile_menu_btn":       "🧑 Manage Profile",
		"post_story_btn":         "📖 Post Story",
		"lang_ar_btn":            "🇮🇶 العربية",
		"lang_en_btn":            "🇺🇸 English",
		"back_btn":               "🔙 Back",
		"stopped_msg":            "🛑 Auto-reply has been stopped.",
		"started_msg":            "🟢 Auto-reply has been started.",
		"edit_text_prompt":       "📝 Send the new auto-reply text now:",
		"saved_text_msg":         "✅ New auto-reply text saved successfully!",
		"exclude_prompt":         "👤 Send the account ID to exclude now:",
		"invalid_id_msg":         "❌ Numbers only! Please send a valid ID.",
		"id_added_msg":           "✅ ID `%d` added to the exclusion list.",
		"list_excluded_title":    "📋 **Excluded Accounts:**\n",
		"no_excluded":            "No excluded accounts currently.",
		"cleared_excluded_msg":   "🧹 All exclusions cleared successfully.",
		"profile_menu_title":     "🧑 Manage Profile - choose what to edit:",
		"edit_first_name_btn":    "✏️ Edit Name",
		"edit_bio_btn":           "📝 Edit Bio",
		"edit_photo_btn":         "🖼️ Edit Photo",
		"edit_username_btn":      "🔗 Edit Username",
		"no_business_connection": "❌ No business account connected to the bot yet.",
		"first_name_prompt":      "✏️ Send the new first name now (optionally followed by a last name):",
		"bio_prompt":             "📝 Send the new bio now (max 70 characters):",
		"username_prompt":        "🔗 Send the new username now (without @):",
		"photo_prompt":           "🖼️ Send the new profile photo now:",
		"name_updated":           "✅ Name updated successfully!",
		"bio_updated":            "✅ Bio updated successfully!",
		"username_updated":       "✅ Username updated successfully!",
		"photo_updated":          "✅ Profile photo updated successfully!",
		"select_story_duration":  "⏱️ Select story duration:",
		"dur_6h":                 "6 Hours",
		"dur_12h":                "12 Hours",
		"dur_24h":                "24 Hours",
		"dur_48h":                "48 Hours",
		"story_prompt":           "📖 Send a photo or video now (max 60 seconds) to post as a story (visible for %s):",
		"story_updated":          "✅ Story posted successfully! It will remain visible for %s.",
		"your_id_msg":            "Your ID is:\n`%d`",
		"fail_name":              "❌ Failed to update name: %s",
		"fail_bio":               "❌ Failed to update bio: %s",
		"fail_username":          "❌ Failed to update username: %s",
		"fail_photo":             "❌ Failed to update photo: %s",
		"fail_story":             "❌ Failed to post story: %s",
		"need_real_photo":        "❌ Please send an actual photo (files or text not accepted).",
		"need_real_media_story":  "❌ Please send an actual photo or video to post as a story.",
		"video_too_long_error":   "The video is longer than 60 seconds, which is Telegram's maximum allowed for stories",
		"id_copy_btn":            "🆔 Copy ID",
		"share_user_btn":         "👤 User",
		"share_user_prompt":      "👇 Use this button to share any user from your chat list with the bot — their name, username and ID will be extracted automatically:",
		"user_shared_info":       "👤 *Shared User Info:*\n\nName: %s\nUsername: %s\nID: `%d`",
		"no_username":            "No username",
	},
}

func tr(lang, key string) string {
	if lang != "en" {
		lang = "ar"
	}
	if val, ok := translations[lang][key]; ok {
		return val
	}
	return key
}

func getDurationLabel(lang, period string) string {
	switch period {
	case "21600":
		return tr(lang, "dur_6h")
	case "43200":
		return tr(lang, "dur_12h")
	case "86400":
		return tr(lang, "dur_24h")
	case "172800":
		return tr(lang, "dur_48h")
	default:
		return tr(lang, "dur_24h")
	}
}

// الترجمة الفورية
func translateText(text, targetLang string) (string, string, error) {
	if strings.TrimSpace(text) == "" {
		return "", "", nil
	}
	endpoint := fmt.Sprintf(
		"[https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s](https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s)",
		targetLang, url.QueryEscape(text),
	)

	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if len(result) == 0 {
		return "", "", fmt.Errorf("فشل الترجمة")
	}

	translatedText := ""
	if sentences, ok := result[0].([]interface{}); ok {
		for _, sentence := range sentences {
			if s, ok := sentence.([]interface{}); ok && len(s) > 0 {
				if tText, ok := s[0].(string); ok {
					translatedText += tText
				}
			}
		}
	}

	detectedLang := ""
	if len(result) > 2 {
		if lang, ok := result[2].(string); ok {
			detectedLang = lang
		}
	}
	return translatedText, detectedLang, nil
}

// هياكل البيانات
type ForceChannel struct {
	ChatID     int64  `json:"chat_id"`
	Username   string `json:"username"`
	Title      string `json:"title"`
	InviteLink string `json:"invite_link"`
}

type BotConfig struct {
	IsStopped      bool           `json:"is_stopped"`
	AutoReply      string         `json:"auto_reply"`
	Excluded       []int64        `json:"excluded"`
	State          string         `json:"state"`
	BusinessConnID string         `json:"business_conn_id"`
	Lang           string         `json:"lang"`
	ForceSubscribe bool           `json:"force_subscribe"`
	ForceChannels  []ForceChannel `json:"force_channels"`
	EntryNotify    bool           `json:"entry_notify"`
	KnownUsers     []int64        `json:"known_users"` // لتتبع دخول المستخدمين الجدد
}

type TelegramUpdate struct {
	Message         *Message       `json:"message"`
	CallbackQuery   *CallbackQuery `json:"callback_query"`
	BusinessMessage *struct {
		MessageID int `json:"message_id"`
		Chat      struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			IsBot     bool   `json:"is_bot"`
		} `json:"from"`
		Text                 string `json:"text"`
		IsOutgoing           bool   `json:"is_outgoing"`
		BusinessConnectionID string `json:"business_connection_id"`
	} `json:"business_message"`
	BusinessConnection *struct {
		ID   string `json:"id"`
		User struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
		} `json:"user"`
		UserChatID int64 `json:"user_chat_id"`
		Date       int64 `json:"date"`
		IsEnabled  bool  `json:"is_enabled"`
	} `json:"business_connection"`
}

type PhotoSize struct {
	FileID string `json:"file_id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Video struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
}

type SharedUserInfo struct {
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type UsersSharedData struct {
	RequestID int64            `json:"request_id"`
	Users     []SharedUserInfo `json:"users"`
}

type Message struct {
	MessageID int `json:"message_id"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
	} `json:"from"`
	Text        string           `json:"text"`
	Photo       []PhotoSize      `json:"photo"`
	Video       *Video           `json:"video"`
	UsersShared *UsersSharedData `json:"users_shared"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
	From    struct {
		ID int64 `json:"id"`
	} `json:"from"`
}

type BusinessConnectionResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		User struct {
			ID int64 `json:"user_id"`
		} `json:"user"`
		UserChatID int64 `json:"user_chat_id"`
	} `json:"result"`
}

type apiResult struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

// ----------------- الدالة الرئيسية -----------------
func Handler(w http.ResponseWriter, r *http.Request) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if secret := os.Getenv("TELEGRAM_WEBHOOK_SECRET"); secret != "" {
		if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	var update TelegramUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. معالجة الضغط على الأزرار الشفافة
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		answerCallback(botToken, cb.ID, "")

		// === حماية لوحة المطور ===
		if strings.HasPrefix(cb.Data, "admin_") {
			if cb.From.ID != DeveloperID {
				answerCallback(botToken, cb.ID, "⛔ هذه اللوحة مخصصة لمطور البوت فقط.")
				w.WriteHeader(http.StatusOK)
				return
			}
			handleAdminCallback(botToken, cb)
			w.WriteHeader(http.StatusOK)
			return
		}

		if cb.Data == "change_quote" {
			newQuote := quotes[rand.Intn(len(quotes))]
			updateButtonQuote(botToken, cb.Message.Chat.ID, cb.Message.MessageID, newQuote)
			w.WriteHeader(http.StatusOK)
			return
		}

		if cb.Data == "force_check" {
			config, _ := getConfig(botToken, cb.From.ID) // جلب الإعدادات (باستخدام ID لأن التخزين مرتبط به)
			if config.ForceSubscribe && cb.From.ID != DeveloperID {
				devConfig, _ := getConfig(botToken, DeveloperID)
				isSubbed, _ := checkForceSub(botToken, cb.From.ID, devConfig)
				if isSubbed {
					deleteMessage(botToken, cb.Message.Chat.ID, cb.Message.MessageID)
					sendMenu(botToken, cb.Message.Chat.ID, config.Lang, tr(config.Lang, "main_menu_title"))
				} else {
					answerCallback(botToken, cb.ID, "❌ لا يزال هناك قنوات لم تشترك بها.")
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		deleteMessage(botToken, cb.Message.Chat.ID, cb.Message.MessageID)
		adminID := cb.From.ID
		config, msgID := getConfig(botToken, adminID)
		lang := config.Lang

		switch cb.Data {
		case "main_menu":
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "main_menu_title"))
		case "stop":
			config.IsStopped = true
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "stopped_msg"))
		case "start":
			config.IsStopped = false
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "started_msg"))
		case "edit_text":
			config.State = "waiting_text"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "edit_text_prompt"))
		case "exclude":
			config.State = "waiting_id"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "exclude_prompt"))
		case "list_excluded":
			txt := tr(lang, "list_excluded_title")
			if len(config.Excluded) == 0 {
				txt += tr(lang, "no_excluded")
			} else {
				for _, id := range config.Excluded {
					txt += fmt.Sprintf("- `%d`\n", id)
				}
			}
			sendSubMenu(botToken, adminID, lang, txt)
		case "clear_excluded":
			config.Excluded = []int64{}
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "cleared_excluded_msg"))
		case "profile_menu":
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendProfileMenu(botToken, adminID, lang, tr(lang, "profile_menu_title"))
		case "edit_first_name":
			if config.BusinessConnID == "" {
				sendProfileMenu(botToken, adminID, lang, tr(lang, "no_business_connection"))
				break
			}
			config.State = "waiting_first_name"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "first_name_prompt"))
		case "edit_bio":
			if config.BusinessConnID == "" {
				sendProfileMenu(botToken, adminID, lang, tr(lang, "no_business_connection"))
				break
			}
			config.State = "waiting_bio"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "bio_prompt"))
		case "edit_username":
			if config.BusinessConnID == "" {
				sendProfileMenu(botToken, adminID, lang, tr(lang, "no_business_connection"))
				break
			}
			config.State = "waiting_username"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "username_prompt"))
		case "edit_photo":
			if config.BusinessConnID == "" {
				sendProfileMenu(botToken, adminID, lang, tr(lang, "no_business_connection"))
				break
			}
			config.State = "waiting_photo"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "photo_prompt"))
		case "post_story":
			if config.BusinessConnID == "" {
				sendMenu(botToken, adminID, lang, tr(lang, "no_business_connection"))
				break
			}
			sendStoryDurationMenu(botToken, adminID, lang)
		case "story_dur_21600", "story_dur_43200", "story_dur_86400", "story_dur_172800":
			period := strings.TrimPrefix(cb.Data, "story_dur_")
			config.State = "waiting_story_" + period
			saveConfig(botToken, adminID, config, msgID)
			durationTxt := getDurationLabel(lang, period)
			sendSubMenu(botToken, adminID, lang, fmt.Sprintf(tr(lang, "story_prompt"), durationTxt))
		case "lang_ar":
			config.Lang = "ar"
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, "ar", tr("ar", "main_menu_title"))
		case "lang_en":
			config.Lang = "en"
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, "en", tr("en", "main_menu_title"))
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. معالجة محادثة التحكم الخاصة بك (الرسائل المباشرة للبوت)
	if update.Message != nil {
		msg := update.Message
		chatID := msg.Chat.ID
		userID := msg.From.ID

		config, msgID := getConfig(botToken, chatID)
		devConfig, devMsgID := getConfig(botToken, DeveloperID)
		lang := config.Lang

		// --- لوحة المطور ---
		if msg.Text == "/admin" {
			if userID != DeveloperID {
				sendMessage(botToken, chatID, "⛔ هذه اللوحة مخصصة لمطور البوت فقط.\n\nمطور البوت: [Xhwe2](https://t.me/Xhwe2)")
				w.WriteHeader(http.StatusOK)
				return
			}
			devConfig.State = ""
			saveConfig(botToken, DeveloperID, devConfig, devMsgID)
			sendAdminDashboard(botToken, chatID, devConfig)
			w.WriteHeader(http.StatusOK)
			return
		}

		// إدخال قناة جديدة للاشتراك الإجباري
		if userID == DeveloperID && devConfig.State == "admin_wait_channel" && msg.Text != "" {
			handleAdminAddChannel(botToken, chatID, msg.Text, devConfig, devMsgID)
			w.WriteHeader(http.StatusOK)
			return
		}

		// --- التحقق من الاشتراك الإجباري (يستثنى المطور) ---
		if userID != DeveloperID && devConfig.ForceSubscribe && len(devConfig.ForceChannels) > 0 {
			isSubbed, missing := checkForceSub(botToken, userID, devConfig)
			if !isSubbed {
				sendForceSubMessage(botToken, chatID, missing)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		// --- إشعار دخول مستخدم جديد ---
		if msg.Text == "/start" && userID != DeveloperID {
			isKnown := false
			for _, id := range devConfig.KnownUsers {
				if id == userID {
					isKnown = true
					break
				}
			}
			if !isKnown {
				devConfig.KnownUsers = append(devConfig.KnownUsers, userID)
				saveConfig(botToken, DeveloperID, devConfig, devMsgID)
				if devConfig.EntryNotify {
					usernameLine := "@" + msg.From.Username
					if msg.From.Username == "" {
						usernameLine = "لا يوجد"
					}
					notifyTxt := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━\n🔔 مستخدم جديد\n━━━━━━━━━━━━━━━━━━━━\n👤 الاسم:\n%s\n🔗 Username:\n%s\n🆔 ID:\n`%d`\n📅 التاريخ:\n%s\n━━━━━━━━━━━━━━━━━━━━",
						msg.From.FirstName, usernameLine, userID, time.Now().Format("2006-01-02"))

					keyboard := map[string]interface{}{
						"inline_keyboard": [][]map[string]interface{}{
							{{"text": "👤 فتح حساب المستخدم", "url": fmt.Sprintf("tg://user?id=%d", userID)}},
						},
					}
					payload := map[string]interface{}{"chat_id": DeveloperID, "text": notifyTxt, "parse_mode": "Markdown", "reply_markup": keyboard}
					b, _ := json.Marshal(payload)
					httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+botToken+"/sendMessage", "application/json", bytes.NewBuffer(b))
				}
			}
		}

		// الرد على كلمة "بوت"
		if strings.TrimSpace(msg.Text) == "بوت" || strings.Contains(msg.Text, "بوت") {
			sendNerdBotInfo(botToken, chatID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if msg.UsersShared != nil && len(msg.UsersShared.Users) > 0 {
			su := msg.UsersShared.Users[0]
			fullName := strings.TrimSpace(su.FirstName + " " + su.LastName)
			if fullName == "" {
				fullName = "—"
			}
			usernameLine := tr(lang, "no_username")
			if su.Username != "" {
				usernameLine = "@" + su.Username
			}
			sendMessage(botToken, chatID, fmt.Sprintf(tr(lang, "user_shared_info"), fullName, usernameLine, su.UserID))
			w.WriteHeader(http.StatusOK)
			return
		}

		if msg.Text == "/start" {
			sendStartPhoto(botToken, chatID, lang)
			sendMenu(botToken, chatID, lang, tr(lang, "main_menu_title"))
			sendUserShareKeyboard(botToken, chatID, lang)
			w.WriteHeader(http.StatusOK)
			return
		}

		if msg.Text == "/id" {
			sendMessage(botToken, chatID, fmt.Sprintf(tr(lang, "your_id_msg"), msg.From.ID))
			w.WriteHeader(http.StatusOK)
			return
		}

		if config.State == "waiting_text" {
			config.AutoReply = msg.Text
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)
			sendMenu(botToken, chatID, lang, tr(lang, "saved_text_msg"))
		} else if config.State == "waiting_id" {
			id, err := strconv.ParseInt(strings.TrimSpace(msg.Text), 10, 64)
			if err == nil {
				alreadyExists := false
				for _, ex := range config.Excluded {
					if ex == id {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					config.Excluded = append(config.Excluded, id)
				}
				config.State = ""
				saveConfig(botToken, chatID, config, msgID)
				sendMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "id_added_msg"), id))
			} else {
				sendSubMenu(botToken, chatID, lang, tr(lang, "invalid_id_msg"))
			}
		} else if config.State == "waiting_first_name" {
			parts := strings.SplitN(strings.TrimSpace(msg.Text), " ", 2)
			firstName := parts[0]
			lastName := ""
			if len(parts) > 1 {
				lastName = parts[1]
			}
			if err := setBusinessAccountName(botToken, config.BusinessConnID, firstName, lastName); err != nil {
				sendSubMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "fail_name"), err.Error()))
			} else {
				config.State = ""
				saveConfig(botToken, chatID, config, msgID)
				sendMenu(botToken, chatID, lang, tr(lang, "name_updated"))
			}
		} else if config.State == "waiting_bio" {
			if len([]rune(msg.Text)) > 70 {
				sendSubMenu(botToken, chatID, lang, "❌ النبذة طويلة جداً! الحد الأقصى المسموح به من تيليجرام هو 70 حرفاً فقط.\nأرسل نبذة أقصر:")
			} else if err := setBusinessAccountBio(botToken, config.BusinessConnID, msg.Text); err != nil {
				sendSubMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "fail_bio"), err.Error()))
			} else {
				config.State = ""
				saveConfig(botToken, chatID, config, msgID)
				sendMenu(botToken, chatID, lang, tr(lang, "bio_updated"))
			}
		} else if config.State == "waiting_username" {
			username := strings.TrimPrefix(strings.TrimSpace(msg.Text), "@")
			if err := setBusinessAccountUsername(botToken, config.BusinessConnID, username); err != nil {
				sendSubMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "fail_username"), err.Error()))
			} else {
				config.State = ""
				saveConfig(botToken, chatID, config, msgID)
				sendMenu(botToken, chatID, lang, tr(lang, "username_updated"))
			}
		} else if config.State == "waiting_photo" {
			if len(msg.Photo) == 0 {
				sendSubMenu(botToken, chatID, lang, tr(lang, "need_real_photo"))
			} else {
				fileID := msg.Photo[len(msg.Photo)-1].FileID
				if err := setBusinessAccountProfilePhoto(botToken, config.BusinessConnID, fileID); err != nil {
					sendSubMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "fail_photo"), err.Error()))
				} else {
					config.State = ""
					saveConfig(botToken, chatID, config, msgID)
					sendMenu(botToken, chatID, lang, tr(lang, "photo_updated"))
				}
			}
		} else if strings.HasPrefix(config.State, "waiting_story_") {
			period := strings.TrimPrefix(config.State, "waiting_story_")
			if len(msg.Photo) == 0 && msg.Video == nil {
				sendSubMenu(botToken, chatID, lang, tr(lang, "need_real_media_story"))
			} else {
				var err error
				if msg.Video != nil {
					err = postBusinessStory(botToken, config.BusinessConnID, "video", msg.Video.FileID, msg.Video.Duration, period, lang)
				} else {
					fileID := msg.Photo[len(msg.Photo)-1].FileID
					err = postBusinessStory(botToken, config.BusinessConnID, "photo", fileID, 0, period, lang)
				}
				if err != nil {
					sendSubMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "fail_story"), err.Error()))
				} else {
					config.State = ""
					saveConfig(botToken, chatID, config, msgID)
					durationTxt := getDurationLabel(lang, period)
					sendMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "story_updated"), durationTxt))
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. معالجة رسائل العملاء (Business Messages)
	if update.BusinessMessage != nil {
		msg := update.BusinessMessage

		if msg.IsOutgoing || msg.From.IsBot {
			w.WriteHeader(http.StatusOK)
			return
		}

		adminID := getAdminIDFromBusinessConn(botToken, msg.BusinessConnectionID)
		if adminID == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}

		senderID := msg.From.ID
		customerChatID := msg.Chat.ID

		if senderID == adminID {
			w.WriteHeader(http.StatusOK)
			return
		}

		config, _ := getConfig(botToken, adminID)

		if config.IsStopped {
			w.WriteHeader(http.StatusOK)
			return
		}

		for _, exID := range config.Excluded {
			if exID == senderID || exID == customerChatID {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		if strings.TrimSpace(msg.Text) == "بوت" || strings.Contains(msg.Text, "بوت") {
			sendNerdBotInfoBusiness(botToken, customerChatID, msg.BusinessConnectionID)
			w.WriteHeader(http.StatusOK)
			return
		}

		// ⏱️ نظام التهدئة (Cooldown)
		cooldownMu.Lock()
		if userCooldowns[adminID] == nil {
			userCooldowns[adminID] = make(map[int64]time.Time)
		}
		if expiry, exists := userCooldowns[adminID][senderID]; exists && time.Now().Before(expiry) {
			cooldownMu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
		userCooldowns[adminID][senderID] = time.Now().Add(30 * time.Minute)
		cooldownMu.Unlock()

		customerName := msg.From.FirstName
		if customerName == "" {
			customerName = "صديقي"
		}

		var detectedLang string
		if strings.TrimSpace(msg.Text) != "" {
			translatedToAr, dLang, err := translateText(msg.Text, "ar")
			if err == nil && dLang != "" {
				detectedLang = dLang
				if detectedLang != "ar" && adminID != 0 {
					notifyMsg := fmt.Sprintf(
						"🌐 *رسالة جديدة بلغة مترجمة (`%s`)*\n👤 *العميل:* %s (`%d`)\n\n💬 *النص الأصلي:*\n%s\n\n✨ *الترجمة للعربية:*\n%s",
						detectedLang, customerName, senderID, msg.Text, translatedToAr,
					)
					sendMessage(botToken, adminID, notifyMsg)
				}
			}
		}

		var replyText string
		if strings.TrimSpace(msg.Text) == "" {
			replyText = "شكراً لتواصلك يا " + customerName + " 🌸\nاستلمت رسالتك وسأرد عليك قريباً."
		} else if config.AutoReply == "" {
			replyText = "أهلاً بك يا " + customerName + " 🌸\nأنا غير متوفر الآن، اترك رسالتك وسأرد عليك قريباً."
		} else if strings.Contains(config.AutoReply, "{name}") || strings.Contains(config.AutoReply, "{الاسم}") || strings.Contains(config.AutoReply, "$name") {
			replyText = config.AutoReply
			replyText = strings.ReplaceAll(replyText, "{name}", customerName)
			replyText = strings.ReplaceAll(replyText, "{الاسم}", customerName)
			replyText = strings.ReplaceAll(replyText, "$name", customerName)
		} else {
			replyText = "أهلاً بك يا " + customerName + " 🌸\n" + config.AutoReply
		}

		if detectedLang != "" && detectedLang != "ar" {
			if translatedReply, _, err := translateText(replyText, detectedLang); err == nil && translatedReply != "" {
				replyText = translatedReply
			}
		}

		sendBusinessReplyWithQuoteButton(botToken, customerChatID, replyText, msg.BusinessConnectionID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 4. رصد تفعيل ربط حساب تجاري
	if update.BusinessConnection != nil {
		bc := update.BusinessConnection
		if bc.IsEnabled {
			if bc.UserChatID != 0 {
				cfg, msgID := getConfig(botToken, bc.UserChatID)
				cfg.BusinessConnID = bc.ID
				saveConfig(botToken, bc.UserChatID, cfg, msgID)
			}
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ----------------- دوال لوحة المطور -----------------

func handleAdminCallback(token string, cb *CallbackQuery) {
	config, msgID := getConfig(token, DeveloperID)

	switch cb.Data {
	case "admin_main":
		config.State = ""
		saveConfig(token, DeveloperID, config, msgID)
		editAdminDashboard(token, cb.Message.Chat.ID, cb.Message.MessageID, config)

	case "admin_fsub":
		editAdminForceSub(token, cb.Message.Chat.ID, cb.Message.MessageID, config)

	case "admin_fsub_on":
		config.ForceSubscribe = true
		saveConfig(token, DeveloperID, config, msgID)
		editAdminForceSub(token, cb.Message.Chat.ID, cb.Message.MessageID, config)

	case "admin_fsub_off":
		config.ForceSubscribe = false
		saveConfig(token, DeveloperID, config, msgID)
		editAdminForceSub(token, cb.Message.Chat.ID, cb.Message.MessageID, config)

	case "admin_fsub_add":
		config.State = "admin_wait_channel"
		saveConfig(token, DeveloperID, config, msgID)
		text := "━━━━━━━━━━━━━━━━━━━━\n➕ إضافة قناة\n━━━━━━━━━━━━━━━━━━━━\n\أرسل:\n• Username القناة مثل:\n`@mychannel`\nأو:\n• ID القناة مثل:\n`-1001234567890`\n━━━━━━━━━━━━━━━━━━━━"
		editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, [][]map[string]interface{}{
			{{"text": "🔙 رجوع", "callback_data": "admin_fsub"}},
		})

	case "admin_fsub_list":
		text := "━━━━━━━━━━━━━━━━━━━━\n📢 القنوات المضافة\n━━━━━━━━━━━━━━━━━━━━\n"
		if len(config.ForceChannels) == 0 {
			text += "\nلا توجد قنوات مضافة حالياً."
		} else {
			for i, ch := range config.ForceChannels {
				text += fmt.Sprintf("\n%d️⃣ %s\n🆔 `%d`\n", i+1, ch.Title, ch.ChatID)
			}
		}
		text += "\n━━━━━━━━━━━━━━━━━━━━"
		editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, [][]map[string]interface{}{
			{{"text": "🔙 رجوع", "callback_data": "admin_fsub"}},
		})

	case "admin_fsub_links":
		text := "━━━━━━━━━━━━━━━━━━━━\n🔗 روابط القنوات\n━━━━━━━━━━━━━━━━━━━━\n"
		var kb [][]map[string]interface{}
		for _, ch := range config.ForceChannels {
			kb = append(kb, []map[string]interface{}{
				{"text": "🔗 فتح " + ch.Title, "url": ch.InviteLink},
			})
		}
		kb = append(kb, []map[string]interface{}{{"text": "🔙 رجوع", "callback_data": "admin_fsub"}})
		editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, kb)

	case "admin_fsub_del":
		text := "━━━━━━━━━━━━━━━━━━━━\n🗑️ اختر قناة لحذفها\n━━━━━━━━━━━━━━━━━━━━"
		var kb [][]map[string]interface{}
		for i, ch := range config.ForceChannels {
			kb = append(kb, []map[string]interface{}{
				{"text": "🗑️ " + ch.Title, "callback_data": fmt.Sprintf("admin_fsub_rem_%d", i)},
			})
		}
		kb = append(kb, []map[string]interface{}{{"text": "🔙 رجوع", "callback_data": "admin_fsub"}})
		editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, kb)

	case "admin_notify":
		status := "🔴 متوقف"
		if config.EntryNotify {
			status = "🟢 مفعّل"
		}
		text := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━\n🔔 إشعارات دخول المستخدمين\n━━━━━━━━━━━━━━━━━━━━\n\nالحالة الحالية:\n%s\n", status)
		kb := [][]map[string]interface{}{
			{{"text": "🟢 تشغيل", "callback_data": "admin_notif_on"}, {"text": "🔴 إيقاف", "callback_data": "admin_notif_off"}},
			{{"text": "🔙 رجوع", "callback_data": "admin_main"}},
		}
		editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, kb)

	case "admin_notif_on":
		config.EntryNotify = true
		saveConfig(token, DeveloperID, config, msgID)
		handleAdminCallback(token, &CallbackQuery{Data: "admin_notify", Message: cb.Message, From: cb.From})

	case "admin_notif_off":
		config.EntryNotify = false
		saveConfig(token, DeveloperID, config, msgID)
		handleAdminCallback(token, &CallbackQuery{Data: "admin_notify", Message: cb.Message, From: cb.From})

	case "admin_stats":
		busStatus := "غير متصل"
		if config.BusinessConnID != "" {
			busStatus = "متصل"
		}
		fsStatus := "متوقف"
		if config.ForceSubscribe {
			fsStatus = "مفعّل"
		}
		notifStatus := "متوقف"
		if config.EntryNotify {
			notifStatus = "مفعّل"
		}
		botStatus := "متوقف"
		if !config.IsStopped {
			botStatus = "يعمل"
		}
		usersCount := len(config.KnownUsers)
		text := fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━
📊 إحصائيات البوت
━━━━━━━━━━━━━━━━━━━━

👥 المستخدمون:
%d

🔐 قنوات الاشتراك:
%d

🟢 الاشتراك الإجباري:
%s

🔔 إشعار الدخول:
%s

🏢 Business:
%s

🤖 حالة البوت:
%s

━━━━━━━━━━━━━━━━━━━━`, usersCount, len(config.ForceChannels), fsStatus, notifStatus, busStatus, botStatus)
		kb := [][]map[string]interface{}{
			{{"text": "🔄 تحديث", "callback_data": "admin_stats"}},
			{{"text": "🔙 رجوع", "callback_data": "admin_main"}},
		}
		editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, kb)

	case "admin_close":
		deleteMessage(token, cb.Message.Chat.ID, cb.Message.MessageID)

	default:
		// معالجة حذف القناة
		if strings.HasPrefix(cb.Data, "admin_fsub_rem_") {
			idxStr := strings.TrimPrefix(cb.Data, "admin_fsub_rem_")
			idx, err := strconv.Atoi(idxStr)
			if err == nil && idx >= 0 && idx < len(config.ForceChannels) {
				ch := config.ForceChannels[idx]
				text := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━\n⚠️ تأكيد الحذف\n━━━━━━━━━━━━━━━━━━━━\nهل تريد إزالة:\n📢 %s\nمن نظام الاشتراك الإجباري؟\n\n⚠️ لن يتم حذف القناة من Telegram.\nسيتم فقط إزالتها من قائمة الاشتراك الإجباري.", ch.Title)
				kb := [][]map[string]interface{}{
					{{"text": "⚠️ نعم، احذف القناة", "callback_data": fmt.Sprintf("admin_fsub_conf_%d", idx)}},
					{{"text": "❌ إلغاء", "callback_data": "admin_fsub_del"}},
				}
				editAdminMarkup(token, cb.Message.Chat.ID, cb.Message.MessageID, text, kb)
			}
		} else if strings.HasPrefix(cb.Data, "admin_fsub_conf_") {
			idxStr := strings.TrimPrefix(cb.Data, "admin_fsub_conf_")
			idx, err := strconv.Atoi(idxStr)
			if err == nil && idx >= 0 && idx < len(config.ForceChannels) {
				config.ForceChannels = append(config.ForceChannels[:idx], config.ForceChannels[idx+1:]...)
				saveConfig(token, DeveloperID, config, msgID)
				handleAdminCallback(token, &CallbackQuery{Data: "admin_fsub_del", Message: cb.Message, From: cb.From})
			}
		}
	}
}

func sendAdminDashboard(token string, chatID int64, config BotConfig) {
	status := "يعمل"
	if config.IsStopped {
		status = "متوقف"
	}
	fsStatus := "متوقف"
	if config.ForceSubscribe {
		fsStatus = "مفعّل"
	}
	notifStatus := "متوقف"
	if config.EntryNotify {
		notifStatus = "مفعّل"
	}
	text := fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━
🛠️ لوحة تحكم المطور
━━━━━━━━━━━━━━━━━━━━

🤖 البوت: %s
👨‍💻 المطور: Xhwe2
🆔 Developer ID: 8705163117

━━━━━━━━━━━━━━━━━━━━
📊 حالة البوت
🟢 البوت %s

🔐 الاشتراك الإجباري: %s
📢 عدد القنوات: %d
🔔 إشعار الدخول: %s

━━━━━━━━━━━━━━━━━━━━
اختر القسم الذي تريد إدارته:`, BotUsername, status, fsStatus, len(config.ForceChannels), notifStatus)

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "⚙️ إعدادات البوت", "callback_data": "main_menu"}, {"text": "📊 الإحصائيات", "callback_data": "admin_stats"}},
			{{"text": "🔐 الاشتراك الإجباري", "callback_data": "admin_fsub"}, {"text": "📢 القنوات", "callback_data": "admin_fsub_list"}},
			{{"text": "🔔 إشعارات الدخول", "callback_data": "admin_notify"}},
			{{"text": "❌ إغلاق اللوحة", "callback_data": "admin_close"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func editAdminDashboard(token string, chatID int64, msgID int, config BotConfig) {
	status := "يعمل"
	if config.IsStopped {
		status = "متوقف"
	}
	fsStatus := "متوقف"
	if config.ForceSubscribe {
		fsStatus = "مفعّل"
	}
	notifStatus := "متوقف"
	if config.EntryNotify {
		notifStatus = "مفعّل"
	}
	text := fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━
🛠️ لوحة تحكم المطور
━━━━━━━━━━━━━━━━━━━━

🤖 البوت: %s
👨‍💻 المطور: Xhwe2
🆔 Developer ID: 8705163117

━━━━━━━━━━━━━━━━━━━━
📊 حالة البوت
🟢 البوت %s

🔐 الاشتراك الإجباري: %s
📢 عدد القنوات: %d
🔔 إشعار الدخول: %s

━━━━━━━━━━━━━━━━━━━━
اختر القسم الذي تريد إدارته:`, BotUsername, status, fsStatus, len(config.ForceChannels), notifStatus)

	kb := [][]map[string]interface{}{
		{{"text": "⚙️ إعدادات البوت", "callback_data": "main_menu"}, {"text": "📊 الإحصائيات", "callback_data": "admin_stats"}},
		{{"text": "🔐 الاشتراك الإجباري", "callback_data": "admin_fsub"}, {"text": "📢 القنوات", "callback_data": "admin_fsub_list"}},
		{{"text": "🔔 إشعارات الدخول", "callback_data": "admin_notify"}},
		{{"text": "❌ إغلاق اللوحة", "callback_data": "admin_close"}},
	}
	editAdminMarkup(token, chatID, msgID, text, kb)
}

func editAdminForceSub(token string, chatID int64, msgID int, config BotConfig) {
	status := "🔴 متوقف"
	if config.ForceSubscribe {
		status = "🟢 مفعّل"
	}
	text := fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━
🔐 الاشتراك الإجباري
━━━━━━━━━━━━━━━━━━━━

الحالة:
%s

عدد القنوات:
%d

━━━━━━━━━━━━━━━━━━━━`, status, len(config.ForceChannels))

	kb := [][]map[string]interface{}{
		{{"text": "🟢 تشغيل الاشتراك", "callback_data": "admin_fsub_on"}, {"text": "🔴 إيقاف الاشتراك", "callback_data": "admin_fsub_off"}},
		{{"text": "➕ إضافة قناة", "callback_data": "admin_fsub_add"}, {"text": "📋 قائمة القنوات", "callback_data": "admin_fsub_list"}},
		{{"text": "🔗 روابط القنوات", "callback_data": "admin_fsub_links"}, {"text": "🗑️ حذف قناة", "callback_data": "admin_fsub_del"}},
		{{"text": "🔙 رجوع", "callback_data": "admin_main"}},
	}
	editAdminMarkup(token, chatID, msgID, text, kb)
}

func editAdminMarkup(token string, chatID int64, msgID int, text string, kb [][]map[string]interface{}) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": msgID,
		"text":       text,
		"parse_mode": "Markdown",
		"reply_markup": map[string]interface{}{
			"inline_keyboard": kb,
		},
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/editMessageText", "application/json", bytes.NewBuffer(b))
}

func handleAdminAddChannel(token string, chatID int64, input string, devConfig BotConfig, devMsgID int) {
	input = strings.TrimSpace(input)

	// 1. Get Chat
	urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/getChat?chat_id=%s](https://api.telegram.org/bot%s/getChat?chat_id=%s)", token, url.QueryEscape(input))
	resp, err := httpClient.Get(urlAPI)
	if err != nil {
		sendMessage(token, chatID, "❌ حدث خطأ في الاتصال بالشبكة.")
		return
	}
	defer resp.Body.Close()

	var getChatRes struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID       int64  `json:"id"`
			Type     string `json:"type"`
			Title    string `json:"title"`
			Username string `json:"username"`
		} `json:"result"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&getChatRes)

	if !getChatRes.Ok || getChatRes.Result.Type != "channel" {
		sendMessage(token, chatID, "❌ القناة غير صحيحة أو البوت غير موجود فيها.\n\n"+getChatRes.Description)
		return
	}

	chID := getChatRes.Result.ID
	title := getChatRes.Result.Title
	username := getChatRes.Result.Username

	// 2. Get Bot ID (Cached)
	botIDCacheMu.Lock()
	if botIDCache == 0 {
		meRes, _ := httpClient.Get(fmt.Sprintf("[https://api.telegram.org/bot%s/getMe](https://api.telegram.org/bot%s/getMe)", token))
		var me struct {
			Result struct {
				ID int64 `json:"id"`
			} `json:"result"`
		}
		json.NewDecoder(meRes.Body).Decode(&me)
		meRes.Body.Close()
		botIDCache = me.Result.ID
	}
	myBotID := botIDCache
	botIDCacheMu.Unlock()

	// 3. Get Chat Member (Check Admin)
	memAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d](https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d)", token, chID, myBotID)
	mResp, _ := httpClient.Get(memAPI)
	defer mResp.Body.Close()
	var memRes struct {
		Ok     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	json.NewDecoder(mResp.Body).Decode(&memRes)

	if !memRes.Ok || (memRes.Result.Status != "administrator" && memRes.Result.Status != "creator") {
		sendMessage(token, chatID, "❌ لا يمكن إضافة القناة\nالبوت موجود في القناة ولكنه ليس Administrator.\nيرجى رفع البوت إلى Administrator ثم إعادة المحاولة.")
		return
	}

	// 4. Invite Link Setup
	inviteLink := ""
	if username != "" {
		inviteLink = "[https://t.me/](https://t.me/)" + username
	} else {
		// Private channel, create invite link
		lnkAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/createChatInviteLink](https://api.telegram.org/bot%s/createChatInviteLink)", token)
		payload := map[string]interface{}{"chat_id": chID}
		b, _ := json.Marshal(payload)
		lResp, _ := httpClient.Post(lnkAPI, "application/json", bytes.NewBuffer(b))
		defer lResp.Body.Close()
		var lnkRes struct {
			Ok     bool `json:"ok"`
			Result struct {
				InviteLink string `json:"invite_link"`
			} `json:"result"`
			Description string `json:"description"`
		}
		json.NewDecoder(lResp.Body).Decode(&lnkRes)

		if !lnkRes.Ok {
			sendMessage(token, chatID, "❌ فشل إنشاء رابط الدعوة. تأكد من إعطاء البوت صلاحية Invite Users via Link.\n\n"+lnkRes.Description)
			return
		}
		inviteLink = lnkRes.Result.InviteLink
	}

	newCh := ForceChannel{
		ChatID:     chID,
		Username:   username,
		Title:      title,
		InviteLink: inviteLink,
	}

	// Avoid duplicates
	exists := false
	for _, c := range devConfig.ForceChannels {
		if c.ChatID == chID {
			exists = true
			break
		}
	}
	if !exists {
		devConfig.ForceChannels = append(devConfig.ForceChannels, newCh)
	}

	devConfig.State = ""
	saveConfig(token, DeveloperID, devConfig, devMsgID)
	sendMessage(token, chatID, fmt.Sprintf("✅ تمت إضافة القناة: %s بنجاح.", title))
	sendAdminDashboard(token, chatID, devConfig)
}

func checkForceSub(token string, userID int64, devConfig BotConfig) (bool, []ForceChannel) {
	if !devConfig.ForceSubscribe || len(devConfig.ForceChannels) == 0 {
		return true, nil
	}

	var missing []ForceChannel
	for _, ch := range devConfig.ForceChannels {
		urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d](https://api.telegram.org/bot%s/getChatMember?chat_id=%d&user_id=%d)", token, ch.ChatID, userID)
		resp, err := httpClient.Get(urlAPI)
		if err != nil {
			continue // في حال فشل الاتصال، نتجاهل لعدم حظر المستخدم خطأً
		}

		var res struct {
			Ok     bool `json:"ok"`
			Result struct {
				Status   string `json:"status"`
				IsMember bool   `json:"is_member"`
			} `json:"result"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()

		if res.Ok {
			st := res.Result.Status
			if st == "left" || st == "kicked" {
				missing = append(missing, ch)
			} else if st == "restricted" && !res.Result.IsMember {
				missing = append(missing, ch)
			}
		} else {
			// Failed to get member, assume not subscribed just in case (safe approach)
			missing = append(missing, ch)
		}
	}
	return len(missing) == 0, missing
}

func sendForceSubMessage(token string, chatID int64, missing []ForceChannel) {
	text := "━━━━━━━━━━━━━━━━━━━━\n🔐 الاشتراك مطلوب\n━━━━━━━━━━━━━━━━━━━━\n\nلاستخدام البوت، يجب الاشتراك في جميع القنوات التالية:\n\n"
	var kb [][]map[string]interface{}

	for i, ch := range missing {
		text += fmt.Sprintf("📢 %s\n", ch.Title)
		kb = append(kb, []map[string]interface{}{
			{"text": fmt.Sprintf("📢 الاشتراك في قناة %d", i+1), "url": ch.InviteLink},
		})
	}
	text += "\n━━━━━━━━━━━━━━━━━━━━\nبعد الاشتراك في جميع القنوات اضغط:\n🔄 تحقق من الاشتراك"

	kb = append(kb, []map[string]interface{}{
		{"text": "🔄 تحقق من الاشتراك", "callback_data": "force_check"},
	})

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": map[string]interface{}{"inline_keyboard": kb},
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func deleteMessage(token string, chatID int64, msgID int) {
	urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/deleteMessage](https://api.telegram.org/bot%s/deleteMessage)", token)
	payload := map[string]interface{}{"chat_id": chatID, "message_id": msgID}
	b, _ := json.Marshal(payload)
	httpClient.Post(urlAPI, "application/json", bytes.NewBuffer(b))
}

func answerCallback(token, callbackQueryID, text string) {
	urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/answerCallbackQuery](https://api.telegram.org/bot%s/answerCallbackQuery)", token)
	payload := map[string]interface{}{"callback_query_id": callbackQueryID}
	if text != "" {
		payload["text"] = text
		payload["show_alert"] = true
	}
	b, _ := json.Marshal(payload)
	httpClient.Post(urlAPI, "application/json", bytes.NewBuffer(b))
}

// ----------------- دوال النظام القديمة -----------------

func getAdminIDFromBusinessConn(token string, connID string) int64 {
	if connID == "" {
		return 0
	}
	bizCacheMu.Lock()
	if id, ok := bizCache[connID]; ok {
		bizCacheMu.Unlock()
		return id
	}
	bizCacheMu.Unlock()

	urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/getBusinessConnection?business_connection_id=%s](https://api.telegram.org/bot%s/getBusinessConnection?business_connection_id=%s)", token, connID)
	resp, err := httpClient.Get(urlAPI)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var res BusinessConnectionResponse
	json.NewDecoder(resp.Body).Decode(&res)

	var adminID int64
	if res.Result.UserChatID != 0 {
		adminID = res.Result.UserChatID
	} else {
		adminID = res.Result.User.ID
	}

	if adminID != 0 {
		bizCacheMu.Lock()
		bizCache[connID] = adminID
		bizCacheMu.Unlock()
	}
	return adminID
}

func getConfig(token string, chatID int64) (BotConfig, int) {
	defaultCfg := BotConfig{
		IsStopped:      false,
		AutoReply:      "",
		Excluded:       []int64{},
		State:          "",
		BusinessConnID: "",
		Lang:           "ar",
		ForceSubscribe: false,
		ForceChannels:  []ForceChannel{},
		EntryNotify:    false,
		KnownUsers:     []int64{},
	}

	if chatID == 0 {
		return defaultCfg, 0
	}

	urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/getChat?chat_id=%d](https://api.telegram.org/bot%s/getChat?chat_id=%d)", token, chatID)
	resp, err := httpClient.Get(urlAPI)
	if err != nil {
		return defaultCfg, 0
	}
	defer resp.Body.Close()

	var res struct {
		Result struct {
			PinnedMessage struct {
				MessageID int    `json:"message_id"`
				Text      string `json:"text"`
			} `json:"pinned_message"`
		} `json:"result"`
	}

	json.NewDecoder(resp.Body).Decode(&res)
	if res.Result.PinnedMessage.MessageID != 0 {
		var cfg BotConfig
		if err := json.Unmarshal([]byte(res.Result.PinnedMessage.Text), &cfg); err == nil {
			if cfg.Lang == "" {
				cfg.Lang = "ar"
			}
			return cfg, res.Result.PinnedMessage.MessageID
		}
	}
	return defaultCfg, 0
}

func saveConfig(token string, chatID int64, cfg BotConfig, pinnedMsgID int) {
	if chatID == 0 {
		return
	}
	b, _ := json.Marshal(cfg)
	cfgText := string(b)

	if pinnedMsgID > 0 {
		urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/editMessageText](https://api.telegram.org/bot%s/editMessageText)", token)
		payload := map[string]interface{}{"chat_id": chatID, "message_id": pinnedMsgID, "text": cfgText}
		pBytes, _ := json.Marshal(payload)
		httpClient.Post(urlAPI, "application/json", bytes.NewBuffer(pBytes))
	} else {
		urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/sendMessage](https://api.telegram.org/bot%s/sendMessage)", token)
		payload := map[string]interface{}{"chat_id": chatID, "text": cfgText}
		pBytes, _ := json.Marshal(payload)
		resp, err := httpClient.Post(urlAPI, "application/json", bytes.NewBuffer(pBytes))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		var res struct {
			Result struct {
				MessageID int `json:"message_id"`
			} `json:"result"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Result.MessageID != 0 {
			pinUrl := fmt.Sprintf("[https://api.telegram.org/bot%s/pinChatMessage](https://api.telegram.org/bot%s/pinChatMessage)", token)
			pinPayload := map[string]interface{}{"chat_id": chatID, "message_id": res.Result.MessageID, "disable_notification": true}
			pPinBytes, _ := json.Marshal(pinPayload)
			httpClient.Post(pinUrl, "application/json", bytes.NewBuffer(pPinBytes))
		}
	}
}

func sendNerdBotInfo(token string, chatID int64) {
	text := "انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز"
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{{{"text": "فعلني من هنا", "url": "[https://t.me/Xhwe2/10](https://t.me/Xhwe2/10)", "style": "success"}}}}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendNerdBotInfoBusiness(token string, chatID int64, bizID string) {
	text := "انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز"
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{{{"text": "فعلني من هنا", "url": "[https://t.me/Xhwe2/10](https://t.me/Xhwe2/10)", "style": "success"}}}}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "business_connection_id": bizID, "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendStartPhoto(token string, chatID int64, lang string) {
	payload := map[string]interface{}{"chat_id": chatID, "photo": startPhotoURL, "caption": tr(lang, "welcome")}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendPhoto", "application/json", bytes.NewBuffer(b))
}

func sendUserShareKeyboard(token string, chatID int64, lang string) {
	keyboard := map[string]interface{}{"keyboard": [][]map[string]interface{}{{{"text": tr(lang, "share_user_btn"), "request_users": map[string]interface{}{"request_id": 1, "request_name": true, "request_username": true}, "style": "success"}}}, "resize_keyboard": true, "is_persistent": true}
	payload := map[string]interface{}{"chat_id": chatID, "text": tr(lang, "share_user_prompt"), "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendMenu(token string, chatID int64, lang, text string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": tr(lang, "stop_btn"), "callback_data": "stop", "style": "danger"}, {"text": tr(lang, "start_btn"), "callback_data": "start", "style": "success"}},
			{{"text": tr(lang, "edit_text_btn"), "callback_data": "edit_text", "style": "primary"}},
			{{"text": tr(lang, "exclude_btn"), "callback_data": "exclude", "style": "primary"}, {"text": tr(lang, "list_excluded_btn"), "callback_data": "list_excluded", "style": "primary"}},
			{{"text": tr(lang, "clear_excluded_btn"), "callback_data": "clear_excluded", "style": "danger"}},
			{{"text": tr(lang, "profile_menu_btn"), "callback_data": "profile_menu", "style": "primary"}},
			{{"text": tr(lang, "post_story_btn"), "callback_data": "post_story", "style": "primary"}},
			{{"text": fmt.Sprintf("%s (%d)", tr(lang, "id_copy_btn"), chatID), "copy_text": map[string]interface{}{"text": fmt.Sprintf("%d", chatID)}, "style": "primary"}},
			{{"text": tr(lang, "lang_ar_btn"), "callback_data": "lang_ar", "style": "primary"}, {"text": tr(lang, "lang_en_btn"), "callback_data": "lang_en", "style": "primary"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendStoryDurationMenu(token string, chatID int64, lang string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "⏱️ " + tr(lang, "dur_6h"), "callback_data": "story_dur_21600", "style": "primary"}, {"text": "⏱️ " + tr(lang, "dur_12h"), "callback_data": "story_dur_43200", "style": "primary"}},
			{{"text": "⏱️ " + tr(lang, "dur_24h"), "callback_data": "story_dur_86400", "style": "primary"}, {"text": "⏱️ " + tr(lang, "dur_48h"), "callback_data": "story_dur_172800", "style": "primary"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": tr(lang, "select_story_duration"), "parse_mode": "Markdown", "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendProfileMenu(token string, chatID int64, lang, text string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": tr(lang, "edit_first_name_btn"), "callback_data": "edit_first_name", "style": "primary"}},
			{{"text": tr(lang, "edit_bio_btn"), "callback_data": "edit_bio", "style": "primary"}},
			{{"text": tr(lang, "edit_photo_btn"), "callback_data": "edit_photo", "style": "primary"}},
			{{"text": tr(lang, "edit_username_btn"), "callback_data": "edit_username", "style": "primary"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendSubMenu(token string, chatID int64, lang, text string) {
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}}}}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendMessage(token string, chatID int64, text string) {
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown"}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendBusinessReplyWithQuoteButton(token string, chatID int64, text, bizID string) {
	initialQuote := quotes[rand.Intn(len(quotes))]
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{{{"text": "✨ " + initialQuote, "callback_data": "change_quote", "style": "primary"}}}}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "business_connection_id": bizID, "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func updateButtonQuote(token string, chatID int64, msgID int, newQuote string) {
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{{{"text": "✨ " + newQuote, "callback_data": "change_quote", "style": "primary"}}}}
	payload := map[string]interface{}{"chat_id": chatID, "message_id": msgID, "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("[https://api.telegram.org/bot](https://api.telegram.org/bot)"+token+"/editMessageReplyMarkup", "application/json", bytes.NewBuffer(b))
}

func downloadTelegramFile(token, fileID string) ([]byte, error) {
	urlAPI := fmt.Sprintf("[https://api.telegram.org/bot%s/getFile?file_id=%s](https://api.telegram.org/bot%s/getFile?file_id=%s)", token, fileID)
	resp, err := mediaClient.Get(urlAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok || res.Result.FilePath == "" {
		return nil, fmt.Errorf("file not found")
	}
	fResp, err := mediaClient.Get(fmt.Sprintf("[https://api.telegram.org/file/bot%s/%s](https://api.telegram.org/file/bot%s/%s)", token, res.Result.FilePath))
	if err != nil {
		return nil, err
	}
	defer fResp.Body.Close()
	return io.ReadAll(fResp.Body)
}

func postMultipartBusinessAPI(token, method string, fields map[string]string, fileFieldName, fileName string, fileBytes []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range fields {
		writer.WriteField(k, v)
	}
	part, _ := writer.CreateFormFile(fileFieldName, fileName)
	part.Write(fileBytes)
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("[https://api.telegram.org/bot%s/%s](https://api.telegram.org/bot%s/%s)", token, method), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := mediaClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res apiResult
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func callBusinessAPI(token, method string, payload map[string]interface{}) error {
	b, _ := json.Marshal(payload)
	resp, err := httpClient.Post(fmt.Sprintf("[https://api.telegram.org/bot%s/%s](https://api.telegram.org/bot%s/%s)", token, method), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res apiResult
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func setBusinessAccountName(token, businessConnID, firstName, lastName string) error {
	payload := map[string]interface{}{"business_connection_id": businessConnID, "first_name": firstName}
	if lastName != "" {
		payload["last_name"] = lastName
	}
	return callBusinessAPI(token, "setBusinessAccountName", payload)
}

func setBusinessAccountBio(token, businessConnID, bio string) error {
	return callBusinessAPI(token, "setBusinessAccountBio", map[string]interface{}{"business_connection_id": businessConnID, "bio": bio})
}

func setBusinessAccountUsername(token, businessConnID, username string) error {
	return callBusinessAPI(token, "setBusinessAccountUsername", map[string]interface{}{"business_connection_id": businessConnID, "username": username})
}

func setBusinessAccountProfilePhoto(token, businessConnID, fileID string) error {
	data, err := downloadTelegramFile(token, fileID)
	if err != nil {
		return err
	}
	return postMultipartBusinessAPI(token, "setBusinessAccountProfilePhoto", map[string]string{"business_connection_id": businessConnID, "photo": `{"type":"static","photo":"attach://photo"}`}, "photo", "photo.jpg", data)
}

func postBusinessStory(token, businessConnID, mediaType, fileID string, durationSeconds int, activePeriod string, lang string) error {
	data, err := downloadTelegramFile(token, fileID)
	if err != nil {
		return err
	}
	mediaJSON := fmt.Sprintf(`{"type":"%s","media":"attach://media"}`, mediaType)
	fields := map[string]string{"business_connection_id": businessConnID, "media": mediaJSON, "active_period": activePeriod}
	ext := "jpg"
	if mediaType == "video" {
		ext = "mp4"
	}
	return postMultipartBusinessAPI(token, "postBusinessStory", fields, "media", "media."+ext, data)
}
