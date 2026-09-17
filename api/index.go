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

// ============================================================
// إعدادات عامة
// ============================================================

var httpClient = &http.Client{Timeout: 8 * time.Second}
var mediaClient = &http.Client{Timeout: 30 * time.Second}

const developerID int64 = 8705163117

const developerUsername = "Xhwe2"

const startPhotoURL = "https://od.lk/s/M18zMzMwODEzNDNfV3R3TEM/IMG_20260810_235848_327.jpg"

// ============================================================
// الذاكرة المؤقتة
// ============================================================

var (
	cooldownMu    sync.Mutex
	userCooldowns = make(map[int64]map[int64]time.Time)

	bizCacheMu sync.Mutex
	bizCache   = make(map[string]int64)
)

// ============================================================
// الاقتباسات
// ============================================================

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

// ============================================================
// الترجمة
// ============================================================

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
		"edit_first_name_btn":    "✏️ Edit Name",
		"edit_bio_btn":           "📝 Edit Bio",
		"edit_photo_btn":         "🖼️ Edit Photo",
		"edit_username_btn":      "🔗 Edit Username",
		"no_business_connection": "❌ No business account connected to the bot yet.",
		"first_name_prompt":      "✏️ Send the new first name now:",
		"bio_prompt":             "📝 Send the new bio now:",
		"username_prompt":        "🔗 Send the new username now:",
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
		"story_prompt":           "📖 Send a photo or video now:",
		"story_updated":          "✅ Story posted successfully!",
		"your_id_msg":            "Your ID is:\n`%d`",
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

// ============================================================
// الاشتراك الإجباري
// ============================================================

type ForceChannel struct {
	ChatID     int64  `json:"chat_id"`
	Username   string `json:"username"`
	Title      string `json:"title"`
	InviteLink string `json:"invite_link"`
}

type BotConfig struct {
	IsStopped      bool    `json:"is_stopped"`
	AutoReply      string  `json:"auto_reply"`
	Excluded       []int64 `json:"excluded"`
	State          string  `json:"state"`
	BusinessConnID string  `json:"business_conn_id"`
	Lang           string  `json:"lang"`

	// إعدادات الاشتراك الإجباري
	ForceChannels []ForceChannel `json:"force_channels"`

	// إشعار دخول المستخدمين
	EntryNotify      bool   `json:"entry_notify"`
	ForceEnabled     bool   `json:"force_enabled"`
	BroadcastDraftID string `json:"broadcast_draft_id"`
}

func defaultBotConfig() BotConfig {
	return BotConfig{
		IsStopped:      false,
		AutoReply:      "",
		Excluded:       []int64{},
		State:          "",
		BusinessConnID: "",
		Lang:           "ar",
		ForceChannels:  []ForceChannel{},
		EntryNotify:    true,
		ForceEnabled:   false,
	}
}

func isDeveloper(id int64) bool {
	return id == developerID
}

// ============================================================
// Telegram Models
// ============================================================

type TelegramUpdate struct {
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`

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

	MyChatMember *struct {
		Chat struct {
			ID       int64  `json:"id"`
			Type     string `json:"type"`
			Title    string `json:"title"`
			Username string `json:"username"`
		} `json:"chat"`
		From struct {
			ID int64 `json:"id"`
		} `json:"from"`
		OldChatMember struct {
			Status string `json:"status"`
		} `json:"old_chat_member"`
		NewChatMember struct {
			Status string `json:"status"`
		} `json:"new_chat_member"`
	} `json:"my_chat_member"`
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
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
	} `json:"from"`

	Text        string           `json:"text"`
	Caption     string           `json:"caption"`
	Photo       []PhotoSize      `json:"photo"`
	Video       *Video           `json:"video"`
	UsersShared *UsersSharedData `json:"users_shared"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
	From    struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
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

// ============================================================
// Handler
// ============================================================

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
		log.Println("خطأ قراءة التحديث:", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// ========================================================
	// CALLBACKS
	// ========================================================

	if update.CallbackQuery != nil {

		cb := update.CallbackQuery

		answerCallback(botToken, cb.ID)

		// ----------------------------------------------------
		// حماية لوحة المطور
		// ----------------------------------------------------

		if isProtectedAdminCallback(cb.Data) {
			if !isDeveloper(cb.From.ID) {
				sendDeveloperOnly(botToken, cb.Message.Chat.ID)
				w.WriteHeader(http.StatusOK)
				return
			}
			handleAdminCallback(botToken, cb)
			w.WriteHeader(http.StatusOK)
			return
		}

		// ----------------------------------------------------
		// زر الاقتباس
		// ----------------------------------------------------

		if cb.Data == "change_quote" {

			newQuote := quotes[rand.Intn(len(quotes))]

			updateButtonQuote(
				botToken,
				cb.Message.Chat.ID,
				cb.Message.MessageID,
				newQuote,
			)

			w.WriteHeader(http.StatusOK)
			return
		}

		deleteMessage(
			botToken,
			cb.Message.Chat.ID,
			cb.Message.MessageID,
		)

		adminID := cb.From.ID

		config, msgID := getConfig(
			botToken,
			adminID,
		)

		lang := config.Lang

		switch cb.Data {

		case "main_menu":

			config.State = ""

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "main_menu_title"),
			)

		case "stop":

			config.IsStopped = true
			config.State = ""

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "stopped_msg"),
			)

		case "start":

			config.IsStopped = false
			config.State = ""

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "started_msg"),
			)

		case "edit_text":

			config.State = "waiting_text"

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "edit_text_prompt"),
			)

		case "exclude":

			config.State = "waiting_id"

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "exclude_prompt"),
			)

		case "list_excluded":

			txt := tr(lang, "list_excluded_title")

			if len(config.Excluded) == 0 {
				txt += tr(lang, "no_excluded")
			} else {
				for _, id := range config.Excluded {
					txt += fmt.Sprintf("- `%d`\n", id)
				}
			}

			sendSubMenu(
				botToken,
				adminID,
				lang,
				txt,
			)

		case "clear_excluded":

			config.Excluded = []int64{}

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "cleared_excluded_msg"),
			)

		case "profile_menu":

			config.State = ""

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendProfileMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "profile_menu_title"),
			)

		case "edit_first_name":

			if config.BusinessConnID == "" {
				sendProfileMenu(
					botToken,
					adminID,
					lang,
					tr(lang, "no_business_connection"),
				)
				break
			}

			config.State = "waiting_first_name"

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "first_name_prompt"),
			)

		case "edit_bio":

			if config.BusinessConnID == "" {
				sendProfileMenu(
					botToken,
					adminID,
					lang,
					tr(lang, "no_business_connection"),
				)
				break
			}

			config.State = "waiting_bio"

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "bio_prompt"),
			)

		case "edit_username":

			if config.BusinessConnID == "" {
				sendProfileMenu(
					botToken,
					adminID,
					lang,
					tr(lang, "no_business_connection"),
				)
				break
			}

			config.State = "waiting_username"

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "username_prompt"),
			)

		case "edit_photo":

			if config.BusinessConnID == "" {
				sendProfileMenu(
					botToken,
					adminID,
					lang,
					tr(lang, "no_business_connection"),
				)
				break
			}

			config.State = "waiting_photo"

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				tr(lang, "photo_prompt"),
			)

		case "post_story":

			if config.BusinessConnID == "" {
				sendMenu(
					botToken,
					adminID,
					lang,
					tr(lang, "no_business_connection"),
				)
				break
			}

			sendStoryDurationMenu(
				botToken,
				adminID,
				lang,
			)

		case "story_dur_21600",
			"story_dur_43200",
			"story_dur_86400",
			"story_dur_172800":

			period := strings.TrimPrefix(
				cb.Data,
				"story_dur_",
			)

			config.State = "waiting_story_" + period

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			durationTxt := getDurationLabel(
				lang,
				period,
			)

			sendSubMenu(
				botToken,
				adminID,
				lang,
				fmt.Sprintf(
					tr(lang, "story_prompt"),
					durationTxt,
				),
			)

		case "force_check":
			if checkForceSubscription(botToken, cb.From.ID, config.ForceChannels) {
				deleteMessage(botToken, cb.Message.Chat.ID, cb.Message.MessageID)
				sendMessage(botToken, cb.Message.Chat.ID, "✅ تم التحقق من اشتراكك في جميع القنوات المطلوبة.\n\nيمكنك الآن استخدام البوت.")
				sendMenu(botToken, cb.Message.Chat.ID, lang, tr(lang, "main_menu_title"))
			} else {
				answerCallback(botToken, cb.ID)
				sendForceSubscriptionMessage(botToken, cb.Message.Chat.ID, config.ForceChannels)
			}

		case "lang_ar":

			config.Lang = "ar"
			config.State = ""

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				adminID,
				"ar",
				tr("ar", "main_menu_title"),
			)

		case "lang_en":

			config.Lang = "en"
			config.State = ""

			saveConfig(
				botToken,
				adminID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				adminID,
				"en",
				tr("en", "main_menu_title"),
			)
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// ========================================================
	// الرسائل العادية
	// ========================================================

	if update.Message != nil {

		msg := update.Message
		chatID := msg.Chat.ID

		if msg.From.ID != 0 && !isCommand(msg.Text, "/start") {
			registerTelegramUser(*msg, false)
		}

		config, msgID := getConfig(
			botToken,
			chatID,
		)

		lang := config.Lang

		// ----------------------------------------------------
		// /admin
		// ----------------------------------------------------

		if isCommand(msg.Text, "/admin") {

			if isDeveloper(msg.From.ID) {

				config.State = ""

				saveConfig(
					botToken,
					chatID,
					config,
					msgID,
				)

				sendAdminPanel(
					botToken,
					chatID,
				)

			} else {

				sendDeveloperOnly(
					botToken,
					chatID,
				)
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		// ----------------------------------------------------
		// /start
		// ----------------------------------------------------

		if isCommand(msg.Text, "/start") {
			isNewUser := registerTelegramUser(*msg, true)

			// المطور يتجاوز الاشتراك الإجباري
			if !isDeveloper(msg.From.ID) {

				if config.ForceEnabled && !checkForceSubscription(
					botToken,
					msg.From.ID,
					config.ForceChannels,
				) {

					sendForceSubscriptionMessage(
						botToken,
						chatID,
						config.ForceChannels,
					)

					// إشعار الدخول للمطور
					notifyEntry(
						botToken,
						config,
						msg,
						false,
					)

					w.WriteHeader(http.StatusOK)
					return
				}
			}

			sendStartPhoto(
				botToken,
				chatID,
				lang,
			)

			sendMenu(
				botToken,
				chatID,
				lang,
				tr(lang, "main_menu_title"),
			)

			sendUserShareKeyboard(
				botToken,
				chatID,
				lang,
			)

			if isNewUser {
				notifyEntry(botToken, config, msg, true)
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		// ----------------------------------------------------
		// التحقق من الاشتراك
		// ----------------------------------------------------

		if msg.Text == "تحقق من الاشتراك" {

			if checkForceSubscription(
				botToken,
				msg.From.ID,
				config.ForceChannels,
			) {

				sendMessage(
					botToken,
					chatID,
					"✅ تم التحقق من اشتراكك في جميع القنوات المطلوبة.\n\nيمكنك الآن استخدام البوت.",
				)

				sendMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "main_menu_title"),
				)

			} else {

				sendForceSubscriptionMessage(
					botToken,
					chatID,
					config.ForceChannels,
				)
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		// ----------------------------------------------------
		// كلمة بوت
		// ----------------------------------------------------

		if strings.TrimSpace(msg.Text) == "بوت" ||
			strings.Contains(msg.Text, "بوت") {

			sendNerdBotInfo(
				botToken,
				chatID,
			)

			w.WriteHeader(http.StatusOK)
			return
		}

		// ----------------------------------------------------
		// المستخدم المشارك
		// ----------------------------------------------------

		if msg.UsersShared != nil &&
			len(msg.UsersShared.Users) > 0 {

			su := msg.UsersShared.Users[0]

			fullName := strings.TrimSpace(
				su.FirstName + " " + su.LastName,
			)

			if fullName == "" {
				fullName = "—"
			}

			usernameLine := tr(
				lang,
				"no_username",
			)

			if su.Username != "" {
				usernameLine = "@" + su.Username
			}

			sendMessage(
				botToken,
				chatID,
				fmt.Sprintf(
					tr(lang, "user_shared_info"),
					fullName,
					usernameLine,
					su.UserID,
				),
			)

			w.WriteHeader(http.StatusOK)
			return
		}

		// ====================================================
		// حالات لوحة المطور
		// ====================================================

		if isDeveloper(msg.From.ID) &&
			strings.HasPrefix(config.State, "admin_") {

			handleAdminTextState(
				botToken,
				msg,
				&config,
				msgID,
			)

			w.WriteHeader(http.StatusOK)
			return
		}

		// ====================================================
		// حالات المستخدم العادية
		// ====================================================

		if config.State == "waiting_text" {

			config.AutoReply = msg.Text
			config.State = ""

			saveConfig(
				botToken,
				chatID,
				config,
				msgID,
			)

			sendMenu(
				botToken,
				chatID,
				lang,
				tr(lang, "saved_text_msg"),
			)

		} else if config.State == "waiting_id" {

			id, err := strconv.ParseInt(
				strings.TrimSpace(msg.Text),
				10,
				64,
			)

			if err == nil {

				alreadyExists := false

				for _, ex := range config.Excluded {
					if ex == id {
						alreadyExists = true
						break
					}
				}

				if !alreadyExists {
					config.Excluded =
						append(config.Excluded, id)
				}

				config.State = ""

				saveConfig(
					botToken,
					chatID,
					config,
					msgID,
				)

				sendMenu(
					botToken,
					chatID,
					lang,
					fmt.Sprintf(
						tr(lang, "id_added_msg"),
						id,
					),
				)

			} else {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "invalid_id_msg"),
				)
			}

		} else if config.State == "waiting_first_name" {

			parts := strings.SplitN(
				strings.TrimSpace(msg.Text),
				" ",
				2,
			)

			firstName := parts[0]
			lastName := ""

			if len(parts) > 1 {
				lastName = parts[1]
			}

			if err := setBusinessAccountName(
				botToken,
				config.BusinessConnID,
				firstName,
				lastName,
			); err != nil {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					fmt.Sprintf(
						tr(lang, "fail_name"),
						err.Error(),
					),
				)

			} else {

				config.State = ""

				saveConfig(
					botToken,
					chatID,
					config,
					msgID,
				)

				sendMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "name_updated"),
				)
			}

		} else if config.State == "waiting_bio" {

			if len([]rune(msg.Text)) > 70 {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					"❌ النبذة طويلة جداً!\nالحد الأقصى 70 حرفاً.\nأرسل نبذة أقصر:",
				)

			} else if err := setBusinessAccountBio(
				botToken,
				config.BusinessConnID,
				msg.Text,
			); err != nil {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					fmt.Sprintf(
						tr(lang, "fail_bio"),
						err.Error(),
					),
				)

			} else {

				config.State = ""

				saveConfig(
					botToken,
					chatID,
					config,
					msgID,
				)

				sendMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "bio_updated"),
				)
			}

		} else if config.State == "waiting_username" {

			username := strings.TrimPrefix(
				strings.TrimSpace(msg.Text),
				"@",
			)

			if err := setBusinessAccountUsername(
				botToken,
				config.BusinessConnID,
				username,
			); err != nil {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					fmt.Sprintf(
						tr(lang, "fail_username"),
						err.Error(),
					),
				)

			} else {

				config.State = ""

				saveConfig(
					botToken,
					chatID,
					config,
					msgID,
				)

				sendMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "username_updated"),
				)
			}

		} else if config.State == "waiting_photo" {

			if len(msg.Photo) == 0 {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "need_real_photo"),
				)

			} else {

				fileID :=
					msg.Photo[len(msg.Photo)-1].FileID

				if err := setBusinessAccountProfilePhoto(
					botToken,
					config.BusinessConnID,
					fileID,
				); err != nil {

					sendSubMenu(
						botToken,
						chatID,
						lang,
						fmt.Sprintf(
							tr(lang, "fail_photo"),
							err.Error(),
						),
					)

				} else {

					config.State = ""

					saveConfig(
						botToken,
						chatID,
						config,
						msgID,
					)

					sendMenu(
						botToken,
						chatID,
						lang,
						tr(lang, "photo_updated"),
					)
				}
			}

		} else if strings.HasPrefix(
			config.State,
			"waiting_story_",
		) {

			period := strings.TrimPrefix(
				config.State,
				"waiting_story_",
			)

			if len(msg.Photo) == 0 &&
				msg.Video == nil {

				sendSubMenu(
					botToken,
					chatID,
					lang,
					tr(lang, "need_real_media_story"),
				)

			} else {

				var err error

				if msg.Video != nil {

					err = postBusinessStory(
						botToken,
						config.BusinessConnID,
						"video",
						msg.Video.FileID,
						msg.Video.Duration,
						period,
						lang,
					)

				} else {

					fileID :=
						msg.Photo[len(msg.Photo)-1].FileID

					err = postBusinessStory(
						botToken,
						config.BusinessConnID,
						"photo",
						fileID,
						0,
						period,
						lang,
					)
				}

				if err != nil {

					sendSubMenu(
						botToken,
						chatID,
						lang,
						fmt.Sprintf(
							tr(lang, "fail_story"),
							err.Error(),
						),
					)

				} else {

					config.State = ""

					saveConfig(
						botToken,
						chatID,
						config,
						msgID,
					)

					durationTxt :=
						getDurationLabel(
							lang,
							period,
						)

					sendMenu(
						botToken,
						chatID,
						lang,
						fmt.Sprintf(
							tr(lang, "story_updated"),
							durationTxt,
						),
					)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	// ========================================================
	// Business Messages
	// ========================================================

	if update.BusinessMessage != nil {

		msg := update.BusinessMessage

		if msg.IsOutgoing ||
			msg.From.IsBot {

			w.WriteHeader(http.StatusOK)
			return
		}

		adminID := getAdminIDFromBusinessConn(
			botToken,
			msg.BusinessConnectionID,
		)

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

		config, _ := getConfig(
			botToken,
			adminID,
		)

		if config.IsStopped {
			w.WriteHeader(http.StatusOK)
			return
		}

		for _, exID := range config.Excluded {

			if exID == senderID ||
				exID == customerChatID {

				w.WriteHeader(http.StatusOK)
				return
			}
		}

		if strings.TrimSpace(msg.Text) == "بوت" ||
			strings.Contains(msg.Text, "بوت") {

			sendNerdBotInfoBusiness(
				botToken,
				customerChatID,
				msg.BusinessConnectionID,
			)

			w.WriteHeader(http.StatusOK)
			return
		}

		cooldownMu.Lock()

		if userCooldowns[adminID] == nil {
			userCooldowns[adminID] =
				make(map[int64]time.Time)
		}

		if expiry, exists :=
			userCooldowns[adminID][senderID]; exists &&
			time.Now().Before(expiry) {

			cooldownMu.Unlock()

			w.WriteHeader(http.StatusOK)
			return
		}

		userCooldowns[adminID][senderID] =
			time.Now().Add(30 * time.Minute)

		cooldownMu.Unlock()

		customerName := msg.From.FirstName

		if customerName == "" {
			customerName = "صديقي"
		}

		var detectedLang string

		if strings.TrimSpace(msg.Text) != "" {

			translatedToAr,
				dLang,
				err := translateText(
				msg.Text,
				"ar",
			)

			if err == nil &&
				dLang != "" {

				detectedLang = dLang

				if detectedLang != "ar" {

					notifyMsg := fmt.Sprintf(
						"🌐 *رسالة جديدة بلغة مترجمة (`%s`)*\n👤 *العميل:* %s (`%d`)\n\n💬 *النص الأصلي:*\n%s\n\n✨ *الترجمة للعربية:*\n%s",
						detectedLang,
						customerName,
						senderID,
						msg.Text,
						translatedToAr,
					)

					sendMessage(
						botToken,
						adminID,
						notifyMsg,
					)
				}
			}
		}

		var replyText string

		if strings.TrimSpace(msg.Text) == "" {

			replyText =
				"شكراً لتواصلك يا " +
					customerName +
					" 🌸\nاستلمت رسالتك وسأرد عليك قريباً."

		} else if config.AutoReply == "" {

			replyText =
				"أهلاً بك يا " +
					customerName +
					" 🌸\nأنا غير متوفر الآن، اترك رسالتك وسأرد عليك قريباً."

		} else {

			replyText = config.AutoReply

			replyText = strings.ReplaceAll(
				replyText,
				"{name}",
				customerName,
			)

			replyText = strings.ReplaceAll(
				replyText,
				"{الاسم}",
				customerName,
			)

			replyText = strings.ReplaceAll(
				replyText,
				"$name",
				customerName,
			)
		}

		if detectedLang != "" &&
			detectedLang != "ar" {

			if translatedReply,
				_,
				err := translateText(
				replyText,
				detectedLang,
			); err == nil &&
				translatedReply != "" {

				replyText = translatedReply
			}
		}

		sendBusinessReplyWithQuoteButton(
			botToken,
			customerChatID,
			replyText,
			msg.BusinessConnectionID,
		)

		w.WriteHeader(http.StatusOK)
		return
	}

	// ========================================================
	// Business Connection
	// ========================================================

	if update.BusinessConnection != nil {

		bc := update.BusinessConnection

		if bc.IsEnabled {

			notifyDeveloper(
				botToken,
				bc.User.ID,
				bc.User.FirstName,
				bc.User.LastName,
				bc.User.Username,
			)

			if bc.UserChatID != 0 {

				cfg, msgID :=
					getConfig(
						botToken,
						bc.UserChatID,
					)

				cfg.BusinessConnID = bc.ID

				saveConfig(
					botToken,
					bc.UserChatID,
					cfg,
					msgID,
				)
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ============================================================
// أدوات الأوامر
// ============================================================

func isCommand(text, command string) bool {

	text = strings.TrimSpace(text)

	if text == command {
		return true
	}

	if strings.HasPrefix(
		text,
		command+"@",
	) {
		return true
	}

	return false
}

// ============================================================
// لوحة المطور
// ============================================================

func sendAdminPanel(token string, chatID int64) {
	cfg, _ := getConfig(token, developerID)
	total, _ := countUsers(nil)
	bcasts, _ := countBroadcasts()
	force := "🔴"
	if cfg.ForceEnabled && len(cfg.ForceChannels) > 0 {
		force = "🟢"
	}
	entry := "🔴"
	if cfg.EntryNotify {
		entry = "🟢"
	}
	text := fmt.Sprintf("━━━━━━━━━━━━━━━━━━\n🛠 *لوحة الإدارة*\n━━━━━━━━━━━━━━━━━━\n\n📊 الإحصائيات\n👥 المستخدمون: `%d`\n📢 القنوات: `%d`\n📨 الإذاعات: `%d`\n\n🔔 إشعار الدخول: %s\n🔒 الاشتراك الإجباري: %s\n\n🇮🇶 *توقيت بغداد:*\n%s\n━━━━━━━━━━━━━━━━━━", total, len(cfg.ForceChannels), bcasts, entry, force, formatIraqDateTime(iraqNow()))
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{
		{{"text": "📊 الإحصائيات", "callback_data": "admin_stats", "style": "primary"}, {"text": "📢 الإذاعة", "callback_data": "admin_broadcast", "style": "primary"}},
		{{"text": "🔒 الاشتراك الإجباري", "callback_data": "admin_force", "style": "primary"}, {"text": "👥 المستخدمون", "callback_data": "admin_users", "style": "primary"}},
		{{"text": "🔔 إشعار الدخول", "callback_data": "admin_entry", "style": "primary"}, {"text": "⚙️ إعدادات البوت", "callback_data": "admin_settings", "style": "primary"}},
		{{"text": "🔄 تحديث", "callback_data": "admin_refresh", "style": "success"}, {"text": "❌ إغلاق", "callback_data": "admin_close", "style": "danger"}},
	}}
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": keyboard})
}

func sendDeveloperOnly(token string, chatID int64) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":  "👨‍💻 التواصل مع مطور البوت",
					"url":   "https://t.me/" + developerUsername,
					"style": "primary",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": "⛔ *هذا الأمر مخصص لمطور البوت فقط.*\n\n" +
			"لا تملك صلاحية الوصول إلى لوحة الإدارة.",
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

// ============================================================
// Callback لوحة المطور
// ============================================================

func handleAdminCallback(token string, cb *CallbackQuery) {
	if !isDeveloper(cb.From.ID) {
		sendDeveloperOnly(token, cb.Message.Chat.ID)
		return
	}
	chatID := cb.Message.Chat.ID
	config, msgID := getConfig(token, developerID)
	data := cb.Data

	switch {
	case data == "admin_close":
		config.State = ""
		config.BroadcastDraftID = ""
		saveConfig(token, developerID, config, msgID)
		editMessageText(token, chatID, cb.Message.MessageID, "تم إغلاق لوحة الإدارة.")
	case data == "admin_refresh":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendAdminPanel(token, chatID)
	case data == "admin_stats":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendAdminStats(token, chatID, config)
	case data == "admin_broadcast":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendBroadcastMenu(token, chatID)
	case data == "admin_users":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendUsersMenu(token, chatID)
	case data == "admin_settings":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendSettingsMenu(token, chatID, config)
	case data == "admin_entry":
		config.EntryNotify = !config.EntryNotify
		config.State = ""
		saveConfig(token, developerID, config, msgID)
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendAdminPanel(token, chatID)
	case data == "admin_force":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendForceAdminMenu(token, chatID, config)
	case data == "admin_force_add":
		config.State = "admin_waiting_channel"
		saveConfig(token, developerID, config, msgID)
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendMessage(token, chatID, "➕ *إضافة قناة*\n\nأرسل `@username` أو `-100...`. يجب أن يكون البوت Administrator في القناة.")
	case data == "admin_force_remove":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendForceRemoveMenu(token, chatID, config)
	case strings.HasPrefix(data, "admin_force_remove_"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_force_remove_"))
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendDeleteConfirm(token, chatID, idx, config)
	case strings.HasPrefix(data, "admin_force_confirm_"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_force_confirm_"))
		if idx >= 0 && idx < len(config.ForceChannels) {
			config.ForceChannels = append(config.ForceChannels[:idx], config.ForceChannels[idx+1:]...)
			saveConfig(token, developerID, config, msgID)
		}
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendForceAdminMenu(token, chatID, config)
	case data == "admin_force_list":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendForceChannelList(token, chatID, config)
	case data == "admin_force_toggle":
		config.ForceEnabled = !config.ForceEnabled
		saveConfig(token, developerID, config, msgID)
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendForceAdminMenu(token, chatID, config)
	case data == "admin_force_back":
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendAdminPanel(token, chatID)
	default:
		handleBroadcastAdminCallback(token, cb, &config, msgID)
	}
}

// ============================================================
// معالجة أزرار حذف القنوات
// ============================================================

func handleForceChannelCallback(
	token string,
	cb *CallbackQuery,
) bool {

	return false
}

// ============================================================
// قائمة الاشتراك الإجباري
// ============================================================

func sendForceAdminMenu(
	token string,
	chatID int64,
	config BotConfig,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{

			{
				{
					"text":          "➕ إضافة قناة",
					"callback_data": "admin_force_add",
					"style":         "success",
				},
			},

			{
				{
					"text":          "📋 عرض القنوات",
					"callback_data": "admin_force_list",
					"style":         "primary",
				},
			},

			{
				{
					"text":          "🟢/🔴 تشغيل / إيقاف",
					"callback_data": "admin_force_toggle",
					"style":         "primary",
				},
			},

			{
				{
					"text":          "➖ حذف قناة",
					"callback_data": "admin_force_remove",
					"style":         "danger",
				},
			},

			{
				{
					"text":          "🔙 رجوع",
					"callback_data": "admin_force_back",
					"style":         "danger",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": fmt.Sprintf(
			"🔐 *إدارة الاشتراك الإجباري*\n\n"+
				"عدد القنوات الحالية: `%d`\n\n"+
				"يمكنك إضافة أكثر من قناة، وسيُطلب من المستخدم الاشتراك في جميع القنوات قبل استخدام البوت.",
			len(config.ForceChannels),
		),
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

// ============================================================
// عرض القنوات
// ============================================================

func sendForceChannelList(
	token string,
	chatID int64,
	config BotConfig,
) {

	if len(config.ForceChannels) == 0 {

		sendAdminPanelWithText(
			token,
			chatID,
			"📋 *القنوات المضافة للاشتراك الإجباري*\n\n"+
				"لا توجد قنوات مضافة حالياً.",
		)

		return
	}

	text :=
		"📋 *القنوات المضافة للاشتراك الإجباري*\n\n"

	for i, ch := range config.ForceChannels {

		text += fmt.Sprintf(
			"%d. 📢 *%s*\n",
			i+1,
			escapeMarkdown(ch.Title),
		)

		if ch.Username != "" {

			text += fmt.Sprintf(
				"   🔗 @%s\n",
				strings.TrimPrefix(ch.Username, "@"),
			)
		}

		text += fmt.Sprintf(
			"   🆔 `%d`\n",
			ch.ChatID,
		)

		if ch.InviteLink != "" {
			text += "   🔗 رابط الدخول موجود\n"
		} else {
			text += "   ⚠️ لا يوجد رابط دخول محفوظ\n"
		}

		text += "\n"
	}

	text +=
		"⚠️ *ملاحظة:*\n" +
			"لن يتم حذف أي قناة من هنا إلا بعد الضغط على زر الحذف ثم تأكيد الحذف."

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          "➖ حذف قناة",
					"callback_data": "admin_force_remove",
					"style":         "danger",
				},
			},
			{
				{
					"text":          "🔙 رجوع",
					"callback_data": "admin_force_back",
					"style":         "danger",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		text,
		keyboard,
	)
}

// ============================================================
// حذف قناة
// ============================================================

func sendForceRemoveMenu(
	token string,
	chatID int64,
	config BotConfig,
) {

	if len(config.ForceChannels) == 0 {

		sendAdminPanelWithText(
			token,
			chatID,
			"➖ لا توجد قنوات يمكن حذفها.",
		)

		return
	}

	rows := [][]map[string]interface{}{}

	for i, ch := range config.ForceChannels {

		title := ch.Title

		if title == "" {
			title = fmt.Sprintf(
				"القناة %d",
				i+1,
			)
		}

		rows = append(
			rows,
			[]map[string]interface{}{
				{
					"text": fmt.Sprintf(
						"❌ %s",
						title,
					),
					"callback_data": fmt.Sprintf(
						"admin_force_remove_%d",
						i,
					),
					"style": "danger",
				},
			},
		)
	}

	rows = append(
		rows,
		[]map[string]interface{}{
			{
				"text":          "🔙 رجوع",
				"callback_data": "admin_force_back",
				"style":         "primary",
			},
		},
	)

	keyboard := map[string]interface{}{
		"inline_keyboard": rows,
	}

	sendCustomMessage(
		token,
		chatID,
		"➖ *اختر القناة التي تريد حذفها*\n\n"+
			"⚠️ لن يتم حذفها مباشرة.\n"+
			"سيظهر لك تأكيد قبل الحذف النهائي.",
		keyboard,
	)
}

// ============================================================
// تأكيد حذف القناة
// ============================================================

func sendDeleteConfirm(
	token string,
	chatID int64,
	index int,
	config BotConfig,
) {

	if index < 0 ||
		index >= len(config.ForceChannels) {

		sendMessage(
			token,
			chatID,
			"❌ القناة غير موجودة.",
		)

		return
	}

	ch := config.ForceChannels[index]

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text": "✅ نعم، احذف القناة",
					"callback_data": fmt.Sprintf(
						"admin_force_confirm_%d",
						index,
					),
					"style": "danger",
				},
			},
			{
				{
					"text":          "❌ إلغاء",
					"callback_data": "admin_force_remove",
					"style":         "primary",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		fmt.Sprintf(
			"⚠️ *تأكيد حذف القناة*\n\n"+
				"📢 القناة: *%s*\n"+
				"🆔 `%d`\n\n"+
				"هل أنت متأكد أنك تريد إزالة هذه القناة من الاشتراك الإجباري؟",
			escapeMarkdown(ch.Title),
			ch.ChatID,
		),
		keyboard,
	)
}

// ============================================================
// إضافة قناة
// ============================================================

func addForceChannel(
	token string,
	input string,
) (ForceChannel, error) {

	input = strings.TrimSpace(input)

	if input == "" {
		return ForceChannel{}, fmt.Errorf(
			"أرسل @username أو ID القناة",
		)
	}

	chatID := input

	if strings.HasPrefix(
		input,
		"https://t.me/",
	) {

		u, err := url.Parse(input)

		if err == nil {

			path := strings.Trim(
				u.Path,
				"/",
			)

			if path != "" &&
				!strings.HasPrefix(
					path,
					"+",
				) {

				chatID = "@" + path
			}
		}
	}

	if !strings.HasPrefix(chatID, "@") &&
		!strings.HasPrefix(chatID, "-100") {

		return ForceChannel{}, fmt.Errorf(
			"صيغة القناة غير صحيحة.\nاستخدم @username أو -100...",
		)
	}

	chat, err := getTelegramChat(
		token,
		chatID,
	)

	if err != nil {
		return ForceChannel{}, err
	}

	if chat.Type != "channel" {

		return ForceChannel{}, fmt.Errorf(
			"❌ هذا المعرف ليس لقناة Telegram.\nنوع المحادثة: %s",
			chat.Type,
		)
	}

	result := ForceChannel{
		ChatID:   chat.ID,
		Title:    chat.Title,
		Username: chat.Username,
	}

	// --------------------------------------------------------
	// رابط القناة
	// --------------------------------------------------------

	if chat.Username != "" {

		result.InviteLink =
			"https://t.me/" +
				strings.TrimPrefix(
					chat.Username,
					"@",
				)

	} else {

		link, err :=
			createBotInviteLink(
				token,
				chat.ID,
			)

		if err != nil {

			return ForceChannel{}, fmt.Errorf(
				"تم العثور على القناة، لكن لم أستطع إنشاء رابط دخول لها.\n\n"+
					"تأكد أن البوت مشرف في القناة ولديه صلاحية دعوة المستخدمين.\n\n"+
					"خطأ Telegram: %s",
				err.Error(),
			)
		}

		result.InviteLink = link
	}

	return result, nil
}

func getBotID(token string) (int64, error) {
	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID int64 `json:"id"`
		} `json:"result"`
		Description string `json:"description"`
	}
	resp, err := httpClient.Get("https://api.telegram.org/bot" + token + "/getMe")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, err
	}
	if !res.Ok {
		return 0, fmt.Errorf("%s", res.Description)
	}
	return res.Result.ID, nil
}

func getChatMemberStatus(token string, chatID, userID int64) (string, error) {
	payload := map[string]interface{}{"chat_id": chatID, "user_id": userID}
	b, _ := json.Marshal(payload)
	resp, err := httpClient.Post("https://api.telegram.org/bot"+token+"/getChatMember", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if !res.Ok {
		return "", fmt.Errorf("%s", res.Description)
	}
	return res.Result.Status, nil
}

// ============================================================
// getChat
// ============================================================

type TelegramChat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Username string `json:"username"`
}

func getTelegramChat(
	token string,
	chatID string,
) (TelegramChat, error) {

	var res struct {
		Ok          bool         `json:"ok"`
		Result      TelegramChat `json:"result"`
		Description string       `json:"description"`
	}

	endpoint :=
		"https://api.telegram.org/bot" +
			token +
			"/getChat?chat_id=" +
			url.QueryEscape(chatID)

	resp, err :=
		httpClient.Get(endpoint)

	if err != nil {
		return TelegramChat{}, err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return TelegramChat{}, err
	}

	if !res.Ok {

		return TelegramChat{}, fmt.Errorf(
			"%s",
			res.Description,
		)
	}

	return res.Result, nil
}

// ============================================================
// إنشاء رابط دخول بواسطة البوت
// ============================================================

func createBotInviteLink(
	token string,
	chatID int64,
) (string, error) {

	payload := map[string]interface{}{
		"chat_id": chatID,
		"name":    "Bot Force Subscription",
	}

	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			InviteLink string `json:"invite_link"`
		} `json:"result"`
		Description string `json:"description"`
	}

	b, _ := json.Marshal(payload)

	resp, err := httpClient.Post(
		"https://api.telegram.org/bot"+
			token+
			"/createChatInviteLink",
		"application/json",
		bytes.NewBuffer(b),
	)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return "", err
	}

	if !res.Ok ||
		res.Result.InviteLink == "" {

		return "",
			fmt.Errorf(
				"%s",
				res.Description,
			)
	}

	return res.Result.InviteLink, nil
}

// ============================================================
// التحقق من الاشتراك
// ============================================================

func checkForceSubscription(
	token string,
	userID int64,
	channels []ForceChannel,
) bool {

	if isDeveloper(userID) {
		return true
	}

	if len(channels) == 0 {
		return true
	}

	for _, channel := range channels {

		if !isUserMember(
			token,
			channel.ChatID,
			userID,
		) {

			return false
		}
	}

	return true
}

func isUserMember(
	token string,
	chatID int64,
	userID int64,
) bool {

	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			Status   string `json:"status"`
			IsMember bool   `json:"is_member"`
		} `json:"result"`
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"user_id": userID,
	}

	b, _ := json.Marshal(payload)

	resp, err := httpClient.Post(
		"https://api.telegram.org/bot"+
			token+
			"/getChatMember",
		"application/json",
		bytes.NewBuffer(b),
	)

	if err != nil {
		log.Println(
			"خطأ getChatMember:",
			err,
		)
		return false
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return false
	}

	if !res.Ok {
		return false
	}

	switch res.Result.Status {

	case "creator":
		return true

	case "administrator":
		return true

	case "member":
		return true

	case "restricted":
		return res.Result.IsMember

	default:
		return false
	}
}

// ============================================================
// رسالة الاشتراك الإجباري
// ============================================================

func sendForceSubscriptionMessage(
	token string,
	chatID int64,
	channels []ForceChannel,
) {

	if len(channels) == 0 {
		return
	}

	rows := [][]map[string]interface{}{}

	for i, channel := range channels {

		title := channel.Title

		if title == "" {
			title = fmt.Sprintf(
				"القناة %d",
				i+1,
			)
		}

		if channel.InviteLink != "" {

			rows = append(
				rows,
				[]map[string]interface{}{
					{
						"text":  "📢 " + title,
						"url":   channel.InviteLink,
						"style": "primary",
					},
				},
			)
		}
	}

	rows = append(
		rows,
		[]map[string]interface{}{
			{
				"text":          "🔄 تحقق من الاشتراك",
				"callback_data": "force_check",
				"style":         "success",
			},
		},
	)

	keyboard := map[string]interface{}{
		"inline_keyboard": rows,
	}

	sendCustomMessage(
		token,
		chatID,
		"🔐 *الاشتراك الإجباري*\n\n"+
			"لاستخدام البوت يجب عليك الاشتراك في جميع القنوات التالية أولاً.\n\n"+
			"بعد الاشتراك اضغط على زر *تحقق من الاشتراك*.",
		keyboard,
	)
}

// ============================================================
// إشعار الدخول
// ============================================================

func notifyEntry(token string, config BotConfig, msg *Message, passed bool) {
	if !config.EntryNotify || isDeveloper(msg.From.ID) {
		return
	}
	username := "لا يوجد"
	if msg.From.Username != "" {
		username = "@" + msg.From.Username
	}
	status := "⛔ لم يكمل الاشتراك"
	if passed {
		status = "✅ أكمل الاشتراك"
	}
	name := strings.TrimSpace(msg.From.FirstName + " " + msg.From.LastName)
	if name == "" {
		name = "غير معروف"
	}
	text := fmt.Sprintf("🔔 *مستخدم جديد دخل إلى البوت*\n\n👤 الاسم: %s\n🆔 ID: `%d`\n🔗 Username: %s\n📅 التاريخ: %s\n🕐 الوقت: %s\n🇮🇶 بغداد\n\n📌 الحالة: %s", escapeMarkdown(name), msg.From.ID, username, formatIraqDate(iraqNow()), formatIraqTime(iraqNow()), status)
	keyboard := map[string]interface{}{"inline_keyboard": [][]map[string]interface{}{{{"text": "👤 فتح المستخدم", "url": "tg://user?id=" + strconv.FormatInt(msg.From.ID, 10)}}}}
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": developerID, "text": text, "parse_mode": "Markdown", "reply_markup": keyboard})
}

// ============================================================
// معالجة حالة إضافة القناة
// ============================================================

func handleAdminTextState(token string, msg *Message, config *BotConfig, msgID int) {
	if !isDeveloper(msg.From.ID) {
		return
	}
	chatID := msg.Chat.ID
	if strings.HasPrefix(config.State, "broadcast_schedule_time:") {
		setBroadcastScheduleTime(token, config, msgID, msg.Text)
		return
	}
	if strings.HasPrefix(config.State, "broadcast_button_url:") {
		finishBroadcastButton(token, config, msgID, msg.Text)
		return
	}
	switch config.State {
	case "admin_waiting_channel":
		channel, err := addForceChannel(token, msg.Text)
		if err != nil {
			sendMessage(token, developerID, "❌ *فشل إضافة القناة*\n\n"+err.Error())
			return
		}
		for _, ex := range config.ForceChannels {
			if ex.ChatID == channel.ChatID {
				config.State = ""
				saveConfig(token, developerID, *config, msgID)
				sendMessage(token, developerID, "⚠️ هذه القناة موجودة مسبقاً.")
				return
			}
		}
		config.ForceChannels = append(config.ForceChannels, channel)
		config.State = ""
		config.ForceEnabled = true
		saveConfig(token, developerID, *config, msgID)
		sendForceAdminMenu(token, chatID, *config)
	case "broadcast_wait_text":
		finishBroadcastText(token, msg.Text, config, msgID)
	case "broadcast_wait_photo":
		if len(msg.Photo) == 0 {
			sendMessage(token, chatID, "❌ أرسل صورة فعلية.")
			return
		}
		finishBroadcastMedia(token, "photo", msg.Photo[len(msg.Photo)-1].FileID, "", config, msgID)
	case "broadcast_wait_photo_caption":
		if strings.TrimSpace(msg.Text) == "" {
			sendMessage(token, chatID, "❌ أرسل Caption أو استخدم زر ⏭ تخطي.")
			return
		}
		setBroadcastCaption(token, config, msgID, msg.Text)
	case "broadcast_wait_video":
		if msg.Video == nil {
			sendMessage(token, chatID, "❌ أرسل فيديو فعلي.")
			return
		}
		finishBroadcastMedia(token, "video", msg.Video.FileID, "", config, msgID)
	case "broadcast_wait_video_caption":
		if strings.TrimSpace(msg.Text) == "" {
			sendMessage(token, chatID, "❌ أرسل Caption أو استخدم زر ⏭ تخطي.")
			return
		}
		setBroadcastCaption(token, config, msgID, msg.Text)
	case "broadcast_schedule_date":
		setBroadcastScheduleDate(token, config, msgID, msg.Text)
	case "broadcast_schedule_time":
		setBroadcastScheduleTime(token, config, msgID, msg.Text)
	case "broadcast_button_text":
		startBroadcastButtonURL(token, config, msgID, msg.Text)
	case "broadcast_button_url":
		finishBroadcastButton(token, config, msgID, msg.Text)
	}
}

// ============================================================
// Helpers
// ============================================================

func forceStatus(config BotConfig) string {
	if !config.ForceEnabled || len(config.ForceChannels) == 0 {
		return "🔴 متوقف"
	}
	return "🟢 مفعّل"
}

func boolStatus(value bool) string {

	if value {
		return "🟢 مفعلة"
	}

	return "🔴 متوقفة"
}

func escapeMarkdown(text string) string {

	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, "*", "\\*")
	text = strings.ReplaceAll(text, "`", "\\`")
	text = strings.ReplaceAll(text, "[", "\\[")

	return text
}

// ============================================================
// Config
// ============================================================

func getConfig(
	token string,
	chatID int64,
) (BotConfig, int) {

	defaultCfg := defaultBotConfig()

	if chatID == 0 {
		return defaultCfg, 0
	}

	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getChat?chat_id=%d",
		token,
		chatID,
	)

	resp, err :=
		httpClient.Get(endpoint)

	if err != nil {
		log.Println(
			"خطأ getChat:",
			err,
		)

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

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return defaultCfg, 0
	}

	if res.Result.PinnedMessage.MessageID != 0 {

		var cfg BotConfig

		if err := json.Unmarshal(
			[]byte(
				res.Result.PinnedMessage.Text,
			),
			&cfg,
		); err == nil {

			if cfg.Lang == "" {
				cfg.Lang = "ar"
			}

			if cfg.Excluded == nil {
				cfg.Excluded = []int64{}
			}

			if cfg.ForceChannels == nil {
				cfg.ForceChannels =
					[]ForceChannel{}
			}

			// إذا كانت إعدادات قديمة ولم يكن
			// الحقل موجوداً نجعله مفعلاً.
			if cfg.Lang == "" {
				cfg.Lang = "ar"
			}

			return cfg,
				res.Result.PinnedMessage.MessageID
		}
	}

	return defaultCfg, 0
}

func saveConfig(
	token string,
	chatID int64,
	cfg BotConfig,
	pinnedMsgID int,
) {

	if chatID == 0 {
		return
	}

	b, err := json.Marshal(cfg)

	if err != nil {
		return
	}

	cfgText := string(b)

	if pinnedMsgID > 0 {

		payload := map[string]interface{}{
			"chat_id":    chatID,
			"message_id": pinnedMsgID,
			"text":       cfgText,
		}

		postJSON(
			token,
			"editMessageText",
			payload,
		)

	} else {

		payload := map[string]interface{}{
			"chat_id": chatID,
			"text":    cfgText,
		}

		var res struct {
			Ok     bool `json:"ok"`
			Result struct {
				MessageID int `json:"message_id"`
			} `json:"result"`
		}

		postJSONResult(
			token,
			"sendMessage",
			payload,
			&res,
		)

		if res.Result.MessageID != 0 {

			pinPayload := map[string]interface{}{
				"chat_id":              chatID,
				"message_id":           res.Result.MessageID,
				"disable_notification": true,
			}

			postJSON(
				token,
				"pinChatMessage",
				pinPayload,
			)
		}
	}
}

// ============================================================
// Generic Telegram POST
// ============================================================

func postJSON(
	token string,
	method string,
	payload map[string]interface{},
) {

	b, _ := json.Marshal(payload)

	_, err :=
		httpClient.Post(
			"https://api.telegram.org/bot"+
				token+
				"/"+
				method,
			"application/json",
			bytes.NewBuffer(b),
		)

	if err != nil {
		log.Println(
			"خطأ",
			method,
			err,
		)
	}
}

func postJSONResult(
	token string,
	method string,
	payload map[string]interface{},
	result interface{},
) {

	b, _ := json.Marshal(payload)

	resp, err :=
		httpClient.Post(
			"https://api.telegram.org/bot"+
				token+
				"/"+
				method,
			"application/json",
			bytes.NewBuffer(b),
		)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(
		resp.Body,
	).Decode(result); err != nil {
		log.Println(err)
	}
}

func sendCustomMessage(
	token string,
	chatID int64,
	text string,
	keyboard map[string]interface{},
) {

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

func sendAdminPanelWithText(
	token string,
	chatID int64,
	text string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          "🔙 لوحة المطور",
					"callback_data": "admin_refresh",
					"style":         "primary",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		text,
		keyboard,
	)
}

// ============================================================
// الترجمة
// ============================================================

func translateText(
	text,
	targetLang string,
) (string, string, error) {

	if strings.TrimSpace(text) == "" {
		return "", "", nil
	}

	endpoint := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=%s&dt=t&q=%s",
		targetLang,
		url.QueryEscape(text),
	)

	resp, err :=
		httpClient.Get(endpoint)

	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result []interface{}

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&result); err != nil {

		return "", "", err
	}

	if len(result) == 0 {
		return "", "", fmt.Errorf("فشل الترجمة")
	}

	translatedText := ""

	if sentences, ok :=
		result[0].([]interface{}); ok {

		for _, sentence := range sentences {

			if s, ok :=
				sentence.([]interface{}); ok &&
				len(s) > 0 {

				if tText, ok :=
					s[0].(string); ok {

					translatedText += tText
				}
			}
		}
	}

	detectedLang := ""

	if len(result) > 2 {

		if lang, ok :=
			result[2].(string); ok {

			detectedLang = lang
		}
	}

	return translatedText,
		detectedLang,
		nil
}

// ============================================================
// القوائم القديمة
// ============================================================

func sendStartPhoto(
	token string,
	chatID int64,
	lang string,
) {

	payload := map[string]interface{}{
		"chat_id": chatID,
		"photo":   startPhotoURL,
		"caption": tr(lang, "welcome"),
	}

	postJSON(
		token,
		"sendPhoto",
		payload,
	)
}

func sendUserShareKeyboard(
	token string,
	chatID int64,
	lang string,
) {

	keyboard := map[string]interface{}{
		"keyboard": [][]map[string]interface{}{
			{
				{
					"text": tr(lang, "share_user_btn"),
					"request_users": map[string]interface{}{
						"request_id":       1,
						"request_name":     true,
						"request_username": true,
					},
					"style": "success",
				},
			},
		},
		"resize_keyboard": true,
		"is_persistent":   true,
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "share_user_prompt"),
		"reply_markup": keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

func sendMenu(
	token string,
	chatID int64,
	lang,
	text string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          tr(lang, "stop_btn"),
					"callback_data": "stop",
					"style":         "danger",
				},
				{
					"text":          tr(lang, "start_btn"),
					"callback_data": "start",
					"style":         "success",
				},
			},
			{
				{
					"text":          tr(lang, "edit_text_btn"),
					"callback_data": "edit_text",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "exclude_btn"),
					"callback_data": "exclude",
					"style":         "primary",
				},
				{
					"text":          tr(lang, "list_excluded_btn"),
					"callback_data": "list_excluded",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "clear_excluded_btn"),
					"callback_data": "clear_excluded",
					"style":         "danger",
				},
			},
			{
				{
					"text":          tr(lang, "profile_menu_btn"),
					"callback_data": "profile_menu",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "post_story_btn"),
					"callback_data": "post_story",
					"style":         "primary",
				},
			},
			{
				{
					"text": fmt.Sprintf(
						"%s (%d)",
						tr(lang, "id_copy_btn"),
						chatID,
					),
					"copy_text": map[string]interface{}{
						"text": fmt.Sprintf(
							"%d",
							chatID,
						),
					},
					"style": "primary",
				},
			},
			{
				{
					"text":          tr(lang, "lang_ar_btn"),
					"callback_data": "lang_ar",
					"style":         "primary",
				},
				{
					"text":          tr(lang, "lang_en_btn"),
					"callback_data": "lang_en",
					"style":         "primary",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

func sendStoryDurationMenu(
	token string,
	chatID int64,
	lang string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          "⏱️ " + tr(lang, "dur_6h"),
					"callback_data": "story_dur_21600",
					"style":         "primary",
				},
				{
					"text":          "⏱️ " + tr(lang, "dur_12h"),
					"callback_data": "story_dur_43200",
					"style":         "primary",
				},
			},
			{
				{
					"text":          "⏱️ " + tr(lang, "dur_24h"),
					"callback_data": "story_dur_86400",
					"style":         "primary",
				},
				{
					"text":          "⏱️ " + tr(lang, "dur_48h"),
					"callback_data": "story_dur_172800",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "back_btn"),
					"callback_data": "main_menu",
					"style":         "danger",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		tr(lang, "select_story_duration"),
		keyboard,
	)
}

func sendProfileMenu(
	token string,
	chatID int64,
	lang,
	text string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          tr(lang, "edit_first_name_btn"),
					"callback_data": "edit_first_name",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "edit_bio_btn"),
					"callback_data": "edit_bio",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "edit_photo_btn"),
					"callback_data": "edit_photo",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "edit_username_btn"),
					"callback_data": "edit_username",
					"style":         "primary",
				},
			},
			{
				{
					"text":          tr(lang, "back_btn"),
					"callback_data": "main_menu",
					"style":         "danger",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		text,
		keyboard,
	)
}

func sendSubMenu(
	token string,
	chatID int64,
	lang,
	text string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          tr(lang, "back_btn"),
					"callback_data": "main_menu",
					"style":         "danger",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		text,
		keyboard,
	)
}

func sendMessage(
	token string,
	chatID int64,
	text string,
) {

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

// ============================================================
// Nerd Bot
// ============================================================

func sendNerdBotInfo(
	token string,
	chatID int64,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":  "فعلني من هنا",
					"url":   "https://t.me/Xhwe2/10",
					"style": "success",
				},
			},
		},
	}

	sendCustomMessage(
		token,
		chatID,
		"انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز",
		keyboard,
	)
}

func sendNerdBotInfoBusiness(
	token string,
	chatID int64,
	bizID string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":  "فعلني من هنا",
					"url":   "https://t.me/Xhwe2/10",
					"style": "success",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id":                chatID,
		"text":                   "انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز",
		"business_connection_id": bizID,
		"reply_markup":           keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

// ============================================================
// Business
// ============================================================

func getAdminIDFromBusinessConn(
	token string,
	connID string,
) int64 {

	if connID == "" {
		return 0
	}

	bizCacheMu.Lock()

	if id, ok := bizCache[connID]; ok {

		bizCacheMu.Unlock()

		return id
	}

	bizCacheMu.Unlock()

	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getBusinessConnection?business_connection_id=%s",
		token,
		connID,
	)

	resp, err :=
		httpClient.Get(endpoint)

	if err != nil {
		return 0
	}

	defer resp.Body.Close()

	var res BusinessConnectionResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return 0
	}

	var adminID int64

	if res.Result.UserChatID != 0 {
		adminID =
			res.Result.UserChatID
	} else {
		adminID =
			res.Result.User.ID
	}

	if adminID != 0 {

		bizCacheMu.Lock()

		bizCache[connID] =
			adminID

		bizCacheMu.Unlock()
	}

	return adminID
}

func sendBusinessReplyWithQuoteButton(
	token string,
	chatID int64,
	text,
	bizID string,
) {

	initialQuote :=
		quotes[rand.Intn(len(quotes))]

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          "✨ " + initialQuote,
					"callback_data": "change_quote",
					"style":         "primary",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id":                chatID,
		"text":                   text,
		"business_connection_id": bizID,
		"reply_markup":           keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

// ============================================================
// Business Account API
// ============================================================

type apiResult struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

func callBusinessAPI(
	token,
	method string,
	payload map[string]interface{},
) error {

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/%s",
		token,
		method,
	)

	b, _ := json.Marshal(payload)

	resp, err :=
		httpClient.Post(
			url,
			"application/json",
			bytes.NewBuffer(b),
		)

	if err != nil {
		return fmt.Errorf(
			"تعذر الاتصال بتليجرام",
		)
	}

	defer resp.Body.Close()

	var res apiResult

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return fmt.Errorf(
			"رد غير متوقع من تليجرام",
		)
	}

	if !res.Ok {
		return fmt.Errorf(
			"%s",
			res.Description,
		)
	}

	return nil
}

func setBusinessAccountName(
	token,
	businessConnID,
	firstName,
	lastName string,
) error {

	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"first_name":             firstName,
	}

	if lastName != "" {
		payload["last_name"] = lastName
	}

	return callBusinessAPI(
		token,
		"setBusinessAccountName",
		payload,
	)
}

func setBusinessAccountBio(
	token,
	businessConnID,
	bio string,
) error {

	return callBusinessAPI(
		token,
		"setBusinessAccountBio",
		map[string]interface{}{
			"business_connection_id": businessConnID,
			"bio":                    bio,
		},
	)
}

func setBusinessAccountUsername(
	token,
	businessConnID,
	username string,
) error {

	return callBusinessAPI(
		token,
		"setBusinessAccountUsername",
		map[string]interface{}{
			"business_connection_id": businessConnID,
			"username":               username,
		},
	)
}

// ============================================================
// الملفات والقصص
// ============================================================

func downloadTelegramFile(
	token,
	fileID string,
) ([]byte, error) {

	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getFile?file_id=%s",
		token,
		fileID,
	)

	resp, err :=
		mediaClient.Get(endpoint)

	if err != nil {
		return nil,
			fmt.Errorf(
				"تعذر الاتصال بتليجرام لجلب الملف",
			)
	}

	defer resp.Body.Close()

	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&res); err != nil {

		return nil,
			fmt.Errorf(
				"رد غير متوقع عند جلب الملف",
			)
	}

	if !res.Ok ||
		res.Result.FilePath == "" {

		return nil,
			fmt.Errorf(
				"%s",
				res.Description,
			)
	}

	fileURL := fmt.Sprintf(
		"https://api.telegram.org/file/bot%s/%s",
		token,
		res.Result.FilePath,
	)

	fResp, err :=
		mediaClient.Get(fileURL)

	if err != nil {
		return nil,
			fmt.Errorf(
				"تعذر تنزيل الملف",
			)
	}

	defer fResp.Body.Close()

	data, err :=
		io.ReadAll(fResp.Body)

	if err != nil {
		return nil,
			fmt.Errorf(
				"تعذر قراءة بيانات الملف",
			)
	}

	return data, nil
}

func postMultipartBusinessAPI(
	token,
	method string,
	fields map[string]string,
	fileFieldName,
	fileName string,
	fileBytes []byte,
) error {

	body := &bytes.Buffer{}

	writer :=
		multipart.NewWriter(body)

	for k, v := range fields {

		if err :=
			writer.WriteField(k, v); err != nil {

			return fmt.Errorf(
				"خطأ تجهيز الطلب",
			)
		}
	}

	part, err :=
		writer.CreateFormFile(
			fileFieldName,
			fileName,
		)

	if err != nil {
		return err
	}

	if _, err :=
		part.Write(fileBytes); err != nil {

		return err
	}

	if err :=
		writer.Close(); err != nil {

		return err
	}

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/%s",
		token,
		method,
	)

	req, err :=
		http.NewRequest(
			"POST",
			url,
			body,
		)

	if err != nil {
		return err
	}

	req.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	resp, err :=
		mediaClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var res apiResult

	if err :=
		json.NewDecoder(
			resp.Body,
		).Decode(&res); err != nil {

		return err
	}

	if !res.Ok {
		return fmt.Errorf(
			"%s",
			res.Description,
		)
	}

	return nil
}

func setBusinessAccountProfilePhoto(
	token,
	businessConnID,
	fileID string,
) error {

	data, err :=
		downloadTelegramFile(
			token,
			fileID,
		)

	if err != nil {
		return err
	}

	fields := map[string]string{
		"business_connection_id": businessConnID,
		"photo":                  `{"type":"static","photo":"attach://photo"}`,
	}

	return postMultipartBusinessAPI(
		token,
		"setBusinessAccountProfilePhoto",
		fields,
		"photo",
		"photo.jpg",
		data,
	)
}

func postBusinessStory(
	token,
	businessConnID,
	mediaType,
	fileID string,
	durationSeconds int,
	activePeriod string,
	lang string,
) error {

	if mediaType == "video" &&
		durationSeconds > 60 {

		return fmt.Errorf(
			"الفيديو أطول من 60 ثانية",
		)
	}

	data, err :=
		downloadTelegramFile(
			token,
			fileID,
		)

	if err != nil {
		return err
	}

	var contentJSON string
	var fileName string

	if mediaType == "video" {

		contentJSON =
			fmt.Sprintf(
				`{"type":"video","video":"attach://content","duration":%d}`,
				durationSeconds,
			)

		fileName = "story.mp4"

	} else {

		contentJSON =
			`{"type":"photo","photo":"attach://content"}`

		fileName = "story.jpg"
	}

	if activePeriod == "" {
		activePeriod = "86400"
	}

	fields := map[string]string{
		"business_connection_id": businessConnID,
		"content":                contentJSON,
		"active_period":          activePeriod,
	}

	return postMultipartBusinessAPI(
		token,
		"postStory",
		fields,
		"content",
		fileName,
		data,
	)
}

// ============================================================
// الرسائل
// ============================================================

func editMessageText(token string, chatID int64, messageID int, text string) {
	postJSON(token, "editMessageText", map[string]interface{}{"chat_id": chatID, "message_id": messageID, "text": text, "parse_mode": "Markdown"})
}

func deleteMessage(
	token string,
	chatID int64,
	msgID int,
) {

	postJSON(
		token,
		"deleteMessage",
		map[string]interface{}{
			"chat_id":    chatID,
			"message_id": msgID,
		},
	)
}

func updateButtonQuote(
	token string,
	chatID int64,
	msgID int,
	newQuote string,
) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text":          "✨ " + newQuote,
					"callback_data": "change_quote",
					"style":         "primary",
				},
			},
		},
	}

	postJSON(
		token,
		"editMessageReplyMarkup",
		map[string]interface{}{
			"chat_id":      chatID,
			"message_id":   msgID,
			"reply_markup": keyboard,
		},
	)
}

func answerCallback(
	token,
	callbackID string,
) {

	postJSON(
		token,
		"answerCallbackQuery",
		map[string]interface{}{
			"callback_query_id": callbackID,
		},
	)
}

// ============================================================
// إشعار Business
// ============================================================

func notifyDeveloper(
	token string,
	userID int64,
	firstName,
	lastName,
	username string,
) {

	fullName := firstName

	if lastName != "" {
		fullName += " " + lastName
	}

	if fullName == "" {
		fullName = "غير معروف"
	}

	usernameLine := "لا يوجد يوزر"

	if username != "" {
		usernameLine =
			"@" + username
	}

	text := fmt.Sprintf(
		"🔔 *تفعيل جديد للبوت*\n\n"+
			"👤 الاسم: %s\n"+
			"🆔 الايدي: `%d`\n"+
			"🔗 اليوزر: %s",
		fullName,
		userID,
		usernameLine,
	)

	sendMessage(
		token,
		developerID,
		text,
	)
}

const iraqTimezone = "Asia/Baghdad"
const broadcastBatchSize = 60
const broadcastRateDelay = 45 * time.Millisecond

var broadcastMu sync.Mutex
var activeBroadcasts = make(map[string]contextCancel)

type contextCancel struct{ stop chan struct{} }

type BroadcastButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type BroadcastDraft struct {
	Kind    string
	Text    string
	FileID  string
	Caption string
	Buttons []BroadcastButton
}

func iraqLocation() *time.Location {
	loc, err := time.LoadLocation(iraqTimezone)
	if err != nil {
		return time.FixedZone("Asia/Baghdad", 3*60*60)
	}
	return loc
}
func iraqNow() time.Time                { return time.Now().In(iraqLocation()) }
func formatIraqDate(t time.Time) string { return t.In(iraqLocation()).Format("02/01/2006") }
func formatIraqTime(t time.Time) string { return t.In(iraqLocation()).Format("03:04 PM") }
func formatIraqDateTime(t time.Time) string {
	return t.In(iraqLocation()).Format("02/01/2006 - 03:04 PM")
}

func isProtectedAdminCallback(data string) bool {
	for _, p := range []string{"admin_", "broadcast_", "users_", "settings_"} {
		if strings.HasPrefix(data, p) {
			return true
		}
	}
	return false
}

func registerTelegramUser(msg Message, markStart bool) bool {
	if msg.From.ID == 0 || !storageConfigured() {
		return false
	}
	old, err := getUser(msg.From.ID)
	isNew := err == nil && old.UserID == 0
	firstSeen := old.FirstSeen
	if firstSeen == "" {
		firstSeen = time.Now().UTC().Format(time.RFC3339)
	}
	lang := "ar"
	if old.Language != "" {
		lang = old.Language
	}
	_, _ = upsertUser(StoredUser{UserID: msg.From.ID, FirstName: msg.From.FirstName, LastName: msg.From.LastName, Username: msg.From.Username, Language: lang, FirstSeen: firstSeen}, markStart)
	return isNew
}

func adminKeyboard(rows ...[]map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"inline_keyboard": rows}
}
func adminButton(text, data, style string) map[string]interface{} {
	return map[string]interface{}{"text": text, "callback_data": data, "style": style}
}
func adminURLButton(text, u string) map[string]interface{} {
	return map[string]interface{}{"text": text, "url": u}
}

func sendBroadcastMenu(token string, chatID int64) {
	kb := adminKeyboard(
		[]map[string]interface{}{adminButton("📝 إذاعة نص", "broadcast_new_text", "primary"), adminButton("🖼 إذاعة صورة", "broadcast_new_photo", "primary")},
		[]map[string]interface{}{adminButton("🎥 إذاعة فيديو", "broadcast_new_video", "primary"), adminButton("📦 إذاعات مجدولة", "broadcast_scheduled", "primary")},
		[]map[string]interface{}{adminButton("📋 سجل الإذاعات", "broadcast_history_0", "primary"), adminButton("📋 معاينة المسودة", "broadcast_preview", "primary")},
		[]map[string]interface{}{adminButton("🔙 رجوع", "broadcast_back", "primary")},
	)
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": "📢 *لوحة الإذاعة*\n\nاختر نوع الإذاعة أو افتح السجل والجدولة.", "parse_mode": "Markdown", "reply_markup": kb})
}

func sendUsersMenu(token string, chatID int64) {
	total, _ := countUsers(nil)
	active := true
	act, _ := countUsers(&active)
	inactive := total - act
	kb := adminKeyboard([]map[string]interface{}{adminButton("🔄 تحديث", "users_refresh", "primary")}, []map[string]interface{}{adminButton("🔙 رجوع", "users_back", "primary")})
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": fmt.Sprintf("👥 *المستخدمون*\n\n👥 الإجمالي: `%d`\n🟢 النشطون: `%d`\n🔴 غير النشطين: `%d`\n\n🇮🇶 %s", total, act, inactive, formatIraqDateTime(iraqNow())), "parse_mode": "Markdown", "reply_markup": kb})
}

func sendSettingsMenu(token string, chatID int64, cfg BotConfig) {
	force := "🔴"
	if cfg.ForceEnabled {
		force = "🟢"
	}
	entry := "🔴"
	if cfg.EntryNotify {
		entry = "🟢"
	}
	kb := adminKeyboard([]map[string]interface{}{adminButton("🔔 إشعار الدخول: "+entry, "settings_entry", "primary")}, []map[string]interface{}{adminButton("🔒 الاشتراك الإجباري: "+force, "settings_force", "primary")}, []map[string]interface{}{adminButton("🌐 اللغة: العربية", "settings_language", "primary")}, []map[string]interface{}{adminButton("🕐 Asia/Baghdad", "settings_timezone", "primary")}, []map[string]interface{}{adminButton("🔙 رجوع", "settings_back", "primary")})
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": "⚙️ *إعدادات البوت*\n\n🔔 إشعار الدخول: " + entry + "\n🔒 الاشتراك الإجباري: " + force + "\n🌐 اللغة: العربية\n🕐 المنطقة الزمنية: Asia/Baghdad", "parse_mode": "Markdown", "reply_markup": kb})
}

func sendAdminStats(token string, chatID int64, cfg BotConfig) {
	total, _ := countUsers(nil)
	active := true
	act, _ := countUsers(&active)
	bc, _ := countBroadcasts()
	failed, _ := countFailedSends()
	sending, _ := getActiveSendingBroadcasts(1)
	status := "⚪ لا توجد إذاعة جارية"
	if len(sending) > 0 {
		status = "📡 إذاعة قيد التنفيذ: " + sending[0].ID
	}
	text := fmt.Sprintf("📊 *إحصائيات البوت*\n\n👥 إجمالي المستخدمين: `%d`\n🟢 النشطون: `%d`\n🔴 غير النشطين: `%d`\n📢 قنوات الاشتراك: `%d`\n📨 إجمالي الإذاعات: `%d`\n❌ إجمالي فشل الإرسال: `%d`\n%s\n🔔 إشعار الدخول: %s\n🔒 الاشتراك الإجباري: %s\n\n📅 التاريخ: %s\n🕐 الوقت: %s\n🇮🇶 التوقيت: العراق - بغداد", total, act, total-act, len(cfg.ForceChannels), bc, failed, status, boolStatus(cfg.EntryNotify), boolStatus(cfg.ForceEnabled && len(cfg.ForceChannels) > 0), formatIraqDate(iraqNow()), formatIraqTime(iraqNow()))
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": adminKeyboard([]map[string]interface{}{adminButton("🔙 رجوع", "admin_refresh", "primary")})})
}

func newBroadcastID() string {
	return "BC-" + iraqNow().Format("20060102-150405") + "-" + strconv.FormatInt(iraqNow().UnixNano()%100000, 10)
}

func saveDraft(token string, cfg *BotConfig, msgID int, d BroadcastDraft) {
	buttons, _ := json.Marshal(d.Buttons)
	id := cfg.BroadcastDraftID
	if id == "" {
		id = newBroadcastID()
		cfg.BroadcastDraftID = id
		_ = createBroadcast(BroadcastRecord{ID: id, Kind: d.Kind, Text: d.Text, FileID: d.FileID, Caption: d.Caption, ButtonsJSON: string(buttons), Status: "draft", DeveloperID: developerID, CreatedAt: time.Now().UTC().Format(time.RFC3339), ProgressChat: developerID})
	} else {
		_ = updateBroadcast(id, map[string]interface{}{"kind": d.Kind, "text": d.Text, "file_id": d.FileID, "caption": d.Caption, "buttons_json": string(buttons), "status": "draft"})
	}
	saveConfig(token, developerID, *cfg, msgID)
}

func loadDraft(cfg BotConfig) (BroadcastRecord, error) {
	if cfg.BroadcastDraftID == "" {
		return BroadcastRecord{}, fmt.Errorf("لا توجد مسودة")
	}
	return getBroadcast(cfg.BroadcastDraftID)
}
func buttonsFromJSON(s string) []BroadcastButton {
	var b []BroadcastButton
	_ = json.Unmarshal([]byte(s), &b)
	return b
}
func buttonsMarkup(buttons []BroadcastButton) map[string]interface{} {
	if len(buttons) == 0 {
		return nil
	}
	rows := make([][]map[string]interface{}, 0, len(buttons))
	for _, b := range buttons {
		rows = append(rows, []map[string]interface{}{{"text": b.Text, "url": b.URL}})
	}
	return map[string]interface{}{"inline_keyboard": rows}
}

func previewBroadcast(token string, chatID int64, b BroadcastRecord) {
	controls := [][]map[string]interface{}{
		{adminButton("🚀 نشر الآن", "broadcast_publish_confirm", "success"), adminButton("⏰ جدولة", "broadcast_schedule", "primary")},
		{adminButton("➕ إضافة زر", "broadcast_add_button", "primary"), adminButton("✏️ تعديل", "broadcast_edit", "primary")},
		{adminButton("❌ إلغاء", "broadcast_cancel", "danger")},
	}
	contentButtons := buttonsFromJSON(b.ButtonsJSON)
	for _, btn := range contentButtons {
		controls = append([][]map[string]interface{}{{{"text": btn.Text, "url": btn.URL}}}, controls...)
	}
	kb := map[string]interface{}{"inline_keyboard": controls}
	text := "📢 معاينة الإذاعة\n\n" + b.Text
	if b.Kind == "photo" {
		text = "🖼 معاينة الإذاعة\n\n" + b.Caption
	}
	if b.Kind == "video" {
		text = "🎥 معاينة الإذاعة\n\n" + b.Caption
	}
	if b.Kind == "text" {
		postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": kb})
		return
	}
	if b.Kind == "photo" {
		postJSON(token, "sendPhoto", map[string]interface{}{"chat_id": chatID, "photo": b.FileID, "caption": b.Caption, "reply_markup": kb})
		return
	}
	postJSON(token, "sendVideo", map[string]interface{}{"chat_id": chatID, "video": b.FileID, "caption": b.Caption, "reply_markup": kb})
}

func finishBroadcastText(token, text string, cfg *BotConfig, msgID int) {
	if len([]rune(text)) == 0 || len([]rune(text)) > 4096 {
		sendMessage(token, developerID, "❌ نص الإذاعة يجب أن يكون بين 1 و4096 حرفاً.")
		return
	}
	d := BroadcastDraft{Kind: "text", Text: text}
	saveDraft(token, cfg, msgID, d)
	cfg.State = ""
	saveConfig(token, developerID, *cfg, msgID)
	b, _ := loadDraft(*cfg)
	previewBroadcast(token, developerID, b)
}
func finishBroadcastMedia(token, kind, fileID, caption string, cfg *BotConfig, msgID int) {
	d := BroadcastDraft{Kind: kind, FileID: fileID, Caption: caption}
	saveDraft(token, cfg, msgID, d)
	cfg.State = "broadcast_wait_" + kind + "_caption"
	saveConfig(token, developerID, *cfg, msgID)
	sendMessageWithKeyboard(token, developerID, "📝 أرسل Caption الآن أو اضغط ⏭ تخطي.", adminKeyboard([]map[string]interface{}{adminButton("⏭ تخطي", "broadcast_skip_caption", "primary")}, []map[string]interface{}{adminButton("❌ إلغاء", "broadcast_cancel", "danger")}))
}
func setBroadcastCaption(token string, cfg *BotConfig, msgID int, caption string) {
	b, err := loadDraft(*cfg)
	if err != nil {
		sendMessage(token, developerID, "❌ لا توجد مسودة.")
		return
	}
	if (b.Kind == "photo" || b.Kind == "video") && len([]rune(caption)) > 1024 {
		sendMessage(token, developerID, "❌ Caption الحد الأقصى له 1024 حرفاً.")
		return
	}
	_ = updateBroadcast(b.ID, map[string]interface{}{"caption": caption})
	cfg.State = ""
	saveConfig(token, developerID, *cfg, msgID)
	b, _ = getBroadcast(b.ID)
	previewBroadcast(token, developerID, b)
}

func sendMessageWithKeyboard(token string, chatID int64, text string, kb map[string]interface{}) {
	postJSON(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": kb})
}

func handleBroadcastAdminCallback(token string, cb *CallbackQuery, cfg *BotConfig, msgID int) {
	data := cb.Data
	chatID := cb.Message.Chat.ID
	if strings.HasPrefix(data, "broadcast_preview_") {
		id := strings.TrimPrefix(data, "broadcast_preview_")
		if b, err := getBroadcast(id); err == nil {
			previewBroadcast(token, chatID, b)
		} else {
			sendMessage(token, chatID, "❌ الإذاعة غير موجودة.")
		}
		return
	}
	if strings.HasPrefix(data, "broadcast_history_") && data != "broadcast_history_0" && data != "broadcast_history_next" {
		n, _ := strconv.Atoi(strings.TrimPrefix(data, "broadcast_history_"))
		sendBroadcastHistory(token, chatID, n)
		return
	}
	switch data {
	case "broadcast_back":
		deleteMessage(token, chatID, cb.Message.MessageID)
		cfg.State = ""
		saveConfig(token, developerID, *cfg, msgID)
		sendAdminPanel(token, chatID)
	case "broadcast_new_text":
		cfg.State = "broadcast_wait_text"
		saveConfig(token, developerID, *cfg, msgID)
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendMessageWithKeyboard(token, chatID, "✏️ أرسل نص الإذاعة الآن.", adminKeyboard([]map[string]interface{}{adminButton("❌ إلغاء", "broadcast_cancel", "danger")}))
	case "broadcast_new_photo":
		cfg.State = "broadcast_wait_photo"
		saveConfig(token, developerID, *cfg, msgID)
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendMessageWithKeyboard(token, chatID, "📷 أرسل الصورة.", adminKeyboard([]map[string]interface{}{adminButton("❌ إلغاء", "broadcast_cancel", "danger")}))
	case "broadcast_new_video":
		cfg.State = "broadcast_wait_video"
		saveConfig(token, developerID, *cfg, msgID)
		deleteMessage(token, chatID, cb.Message.MessageID)
		sendMessageWithKeyboard(token, chatID, "🎥 أرسل الفيديو.", adminKeyboard([]map[string]interface{}{adminButton("❌ إلغاء", "broadcast_cancel", "danger")}))
	case "broadcast_skip_caption":
		cfg.State = ""
		saveConfig(token, developerID, *cfg, msgID)
		if b, err := loadDraft(*cfg); err == nil {
			_ = updateBroadcast(b.ID, map[string]interface{}{"caption": ""})
			b, _ = getBroadcast(b.ID)
			previewBroadcast(token, chatID, b)
		}
	case "broadcast_preview":
		if b, err := loadDraft(*cfg); err == nil {
			previewBroadcast(token, chatID, b)
		} else {
			sendMessage(token, chatID, "❌ لا توجد مسودة.")
		}
	case "broadcast_publish_confirm":
		if b, err := loadDraft(*cfg); err == nil {
			total, _ := countUsers(&[]bool{true}[0])
			_ = total
			showPublishConfirm(token, chatID, b)
		}
	case "broadcast_publish_yes":
		startBroadcast(token, chatID, cfg, msgID, false)
	case "broadcast_schedule":
		cfg.State = "broadcast_schedule_date"
		saveConfig(token, developerID, *cfg, msgID)
		sendMessageWithKeyboard(token, chatID, "📅 أرسل التاريخ بصيغة `DD/MM/YYYY` حسب توقيت بغداد.", adminKeyboard([]map[string]interface{}{adminButton("❌ إلغاء", "broadcast_cancel", "danger")}))
	case "broadcast_add_button":
		cfg.State = "broadcast_button_text"
		saveConfig(token, developerID, *cfg, msgID)
		sendMessage(token, chatID, "🔘 أرسل اسم الزر.")
	case "broadcast_edit":
		editBroadcastPrompt(token, chatID, cfg, msgID)
	case "broadcast_cancel":
		cancelDraft(token, chatID, cfg, msgID)
	case "broadcast_stop_confirm":
		sendMessageWithKeyboard(token, chatID, "⚠️ *هل تريد إيقاف الإذاعة الحالية؟*\n\nلن يتم إرسال مستخدمين جدد، أما الرسائل التي أُرسلت بالفعل فلن يتم سحبها.", adminKeyboard([]map[string]interface{}{adminButton("✅ نعم، أوقفها", "broadcast_stop_yes", "danger")}, []map[string]interface{}{adminButton("❌ لا", "broadcast_status", "primary")}))
	case "broadcast_stop_yes":
		stopCurrentBroadcast(token, chatID, cfg, msgID)
	case "broadcast_history_0":
		sendBroadcastHistory(token, chatID, 0)
	case "broadcast_history_next":
		sendBroadcastHistory(token, chatID, 1)
	case "broadcast_scheduled":
		sendScheduledBroadcasts(token, chatID)
	case "broadcast_scheduled_delete":
		cfg.State = ""
		sendMessage(token, chatID, "🗑 للحذف استخدم سجل الجدولة في النسخة الحالية.")
	}
}

func showPublishConfirm(token string, chatID int64, b BroadcastRecord) {
	active := true
	total, _ := countUsers(&active)
	kb := adminKeyboard([]map[string]interface{}{adminButton("✅ نعم، ابدأ الإذاعة", "broadcast_publish_yes", "success")}, []map[string]interface{}{adminButton("❌ إلغاء", "broadcast_cancel", "danger")})
	sendMessageWithKeyboard(token, chatID, fmt.Sprintf("⚠️ *تأكيد الإذاعة*\n\nسيتم إرسال المحتوى إلى جميع المستخدمين النشطين.\n\n👥 العدد: `%d`\n🆔 %s", total, b.ID), kb)
}

func startBroadcast(token string, chatID int64, cfg *BotConfig, msgID int, scheduled bool) {
	if cfg.BroadcastDraftID == "" {
		sendMessage(token, chatID, "❌ لا توجد مسودة.")
		return
	}
	b, err := getBroadcast(cfg.BroadcastDraftID)
	if err != nil {
		sendMessage(token, chatID, err.Error())
		return
	}
	if b.Status == "sending" {
		sendMessage(token, chatID, "⚠️ الإذاعة قيد التنفيذ بالفعل.")
		return
	}
	started, err := startBroadcastAtomic(b.ID, "sending")
	if err != nil {
		sendMessage(token, chatID, "⚠️ الإذاعة بدأت مسبقًا أو لم تعد متاحة.")
		return
	}
	active := true
	total, _ := countUsers(&active)
	_ = updateBroadcast(started.ID, map[string]interface{}{"total": total, "developer_id": developerID, "progress_chat_id": chatID})
	if err := supabaseRPC("enqueue_broadcast_recipients", map[string]interface{}{"p_broadcast_id": started.ID}); err != nil {
		_ = updateBroadcast(started.ID, map[string]interface{}{"status": "failed", "error_text": err.Error()})
		sendMessage(token, chatID, "❌ تعذر تجهيز قائمة المستخدمين: "+err.Error())
		return
	}
	cfg.State = "broadcast_sending"
	saveConfig(token, developerID, *cfg, msgID)
	sendMessageWithKeyboard(token, chatID, "📡 *بدأت الإذاعة*\n\nسيتم التنفيذ عبر Queue/Worker محفوظة في قاعدة البيانات.", adminKeyboard([]map[string]interface{}{adminButton("🛑 إيقاف الإذاعة", "broadcast_stop_confirm", "danger")}, []map[string]interface{}{adminButton("🔄 تحديث", "broadcast_status", "primary")}))
	_ = processBroadcastQueueOnce(token, started.ID)
	_ = scheduled
}

func cancelDraft(token string, chatID int64, cfg *BotConfig, msgID int) {
	if cfg.BroadcastDraftID != "" {
		_ = updateBroadcast(cfg.BroadcastDraftID, map[string]interface{}{"status": "cancelled", "completed_at": time.Now().UTC().Format(time.RFC3339)})
	}
	cfg.BroadcastDraftID = ""
	cfg.State = ""
	saveConfig(token, developerID, *cfg, msgID)
	deleteMessage(token, chatID, msgID)
	sendBroadcastMenu(token, chatID)
}
func editBroadcastPrompt(token string, chatID int64, cfg *BotConfig, msgID int) {
	b, err := loadDraft(*cfg)
	if err != nil {
		sendMessage(token, chatID, "❌ لا توجد مسودة.")
		return
	}
	if b.Kind == "text" {
		cfg.State = "broadcast_wait_text"
	} else if b.Kind == "photo" {
		cfg.State = "broadcast_wait_photo"
	} else {
		cfg.State = "broadcast_wait_video"
	}
	saveConfig(token, developerID, *cfg, msgID)
	sendMessage(token, chatID, "✏️ أرسل المحتوى الجديد.")
}

func setBroadcastScheduleDate(token string, cfg *BotConfig, msgID int, text string) {
	d, err := time.ParseInLocation("02/01/2006", strings.TrimSpace(text), iraqLocation())
	if err != nil {
		sendMessage(token, developerID, "❌ التاريخ غير صحيح. استخدم DD/MM/YYYY.")
		return
	}
	cfg.State = "broadcast_schedule_time:" + d.Format("2006-01-02")
	saveConfig(token, developerID, *cfg, msgID)
	sendMessage(token, developerID, "🕐 أرسل الوقت بصيغة `HH:MM` حسب توقيت بغداد.")
}
func setBroadcastScheduleTime(token string, cfg *BotConfig, msgID int, text string) {
	parts := strings.Split(strings.TrimSpace(text), ":")
	if len(parts) != 2 {
		sendMessage(token, developerID, "❌ الوقت غير صحيح.")
		return
	}
	h, e1 := strconv.Atoi(parts[0])
	m, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		sendMessage(token, developerID, "❌ الوقت غير صحيح. استخدم HH:MM.")
		return
	}
	date := strings.TrimPrefix(cfg.State, "broadcast_schedule_time:")
	loc := iraqLocation()
	t, err := time.ParseInLocation("2006-01-02 15:04", date+fmt.Sprintf(" %02d:%02d", h, m), loc)
	if err != nil {
		sendMessage(token, developerID, "❌ تعذر حفظ الموعد.")
		return
	}
	if !t.After(iraqNow()) {
		sendMessage(token, developerID, "❌ لا يمكن جدولة إذاعة في وقت مضى.")
		return
	}
	b, err := loadDraft(*cfg)
	if err != nil {
		sendMessage(token, developerID, "❌ لا توجد مسودة.")
		return
	}
	_ = updateBroadcast(b.ID, map[string]interface{}{"status": "scheduled", "scheduled_at": t.UTC().Format(time.RFC3339), "total": 0})
	cfg.State = ""
	saveConfig(token, developerID, *cfg, msgID)
	sendMessageWithKeyboard(token, developerID, fmt.Sprintf("⏰ *الإذاعة مجدولة*\n\n📅 %s\n🕐 %s\n🇮🇶 بغداد\n🆔 %s", formatIraqDate(t), formatIraqTime(t), b.ID), adminKeyboard([]map[string]interface{}{adminButton("📢 لوحة الإذاعة", "admin_broadcast", "primary")}))
}

func startBroadcastButtonURL(token string, cfg *BotConfig, msgID int, text string) {
	if strings.TrimSpace(text) == "" {
		sendMessage(token, developerID, "❌ اسم الزر لا يمكن أن يكون فارغاً.")
		return
	}
	cfg.State = "broadcast_button_url:" + url.QueryEscape(strings.TrimSpace(text))
	saveConfig(token, developerID, *cfg, msgID)
	sendMessage(token, developerID, "🔗 أرسل رابط HTTPS للزر.")
}
func finishBroadcastButton(token string, cfg *BotConfig, msgID int, text string) {
	prefix := strings.TrimPrefix(cfg.State, "broadcast_button_url:")
	name, _ := url.QueryUnescape(prefix)
	u := strings.TrimSpace(text)
	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		sendMessage(token, developerID, "❌ الرابط غير صالح. استخدم https://...")
		return
	}
	b, err := loadDraft(*cfg)
	if err != nil {
		sendMessage(token, developerID, "❌ لا توجد مسودة.")
		return
	}
	buttons := buttonsFromJSON(b.ButtonsJSON)
	if len(buttons) >= 8 {
		sendMessage(token, developerID, "❌ الحد الأقصى 8 أزرار.")
		return
	}
	buttons = append(buttons, BroadcastButton{Text: name, URL: u})
	raw, _ := json.Marshal(buttons)
	_ = updateBroadcast(b.ID, map[string]interface{}{"buttons_json": string(raw)})
	cfg.State = ""
	saveConfig(token, developerID, *cfg, msgID)
	b, _ = getBroadcast(b.ID)
	previewBroadcast(token, developerID, b)
}

func sendBroadcastHistory(token string, chatID int64, offset int) {
	list, err := listBroadcasts(8, offset*8)
	if err != nil {
		sendMessage(token, chatID, "❌ تعذر قراءة السجل.")
		return
	}
	if len(list) == 0 {
		sendMessage(token, chatID, "📋 لا توجد إذاعات.")
		return
	}
	text := "📋 *سجل الإذاعات*\n\n"
	for _, b := range list {
		text += fmt.Sprintf("📢 `%s`\n📝 %s | 🟢 %s\n👥 %d | ✅ %d | ❌ %d\n📅 %s\n\n", b.ID, b.Kind, b.Status, b.Total, b.SuccessCount, b.FailedCount, formatIraqDateTime(parseTime(b.CreatedAt)))
	}
	rows := [][]map[string]interface{}{}
	if len(list) == 8 {
		rows = append(rows, []map[string]interface{}{adminButton("➡️ التالي", "broadcast_history_next", "primary")})
	}
	rows = append(rows, []map[string]interface{}{adminButton("🔙 رجوع", "admin_broadcast", "primary")})
	sendMessageWithKeyboard(token, chatID, text, adminKeyboard(rows...))
}
func sendScheduledBroadcasts(token string, chatID int64) {
	list, err := getScheduledBroadcasts(time.Now(), 20)
	if err != nil {
		sendMessage(token, chatID, "❌ تعذر قراءة الجدولة.")
		return
	}
	if len(list) == 0 {
		sendMessageWithKeyboard(token, chatID, "⏰ لا توجد إذاعات مجدولة قادمة.", adminKeyboard([]map[string]interface{}{adminButton("🔙 رجوع", "admin_broadcast", "primary")}))
		return
	}
	text := "⏰ *الإذاعات المجدولة*\n\n"
	rows := [][]map[string]interface{}{}
	for _, b := range list {
		t := parseTime(b.ScheduledAt)
		text += fmt.Sprintf("📢 `%s`\n📅 %s\n🕐 %s\n🇮🇶 بغداد\n🟡 مجدولة\n\n", b.ID, formatIraqDate(t), formatIraqTime(t))
		rows = append(rows, []map[string]interface{}{adminButton("👁 معاينة", "broadcast_preview_"+b.ID, "primary"), adminButton("✏️ تعديل", "broadcast_scheduled_edit_"+b.ID, "primary"), adminButton("🗑 حذف", "broadcast_scheduled_delete_"+b.ID, "danger")})
	}
	rows = append(rows, []map[string]interface{}{adminButton("🔙 رجوع", "admin_broadcast", "primary")})
	sendMessageWithKeyboard(token, chatID, text, adminKeyboard(rows...))
}
func parseTime(s string) time.Time { t, _ := time.Parse(time.RFC3339, s); return t }

func stopCurrentBroadcast(token string, chatID int64, cfg *BotConfig, msgID int) {
	b, err := getActiveSendingBroadcast()
	if err != nil {
		sendMessage(token, chatID, "ℹ️ لا توجد إذاعة قيد التنفيذ.")
		return
	}
	_ = updateBroadcast(b.ID, map[string]interface{}{"status": "cancelled", "completed_at": time.Now().UTC().Format(time.RFC3339)})
	broadcastMu.Lock()
	if c, ok := activeBroadcasts[b.ID]; ok {
		close(c.stop)
		delete(activeBroadcasts, b.ID)
	}
	broadcastMu.Unlock()
	cfg.State = ""
	saveConfig(token, developerID, *cfg, msgID)
	sendMessage(token, chatID, "🛑 تم إيقاف الإذاعة الحالية. لن يتم إرسال مستخدمين جدد.")
}
func getActiveSendingBroadcast() (BroadcastRecord, error) {
	list, err := getActiveSendingBroadcasts(1)
	if err != nil || len(list) == 0 {
		return BroadcastRecord{}, fmt.Errorf("none")
	}
	return list[0], nil
}

func processBroadcastQueue(token, broadcastID string) {
	for i := 0; i < 8; i++ {
		if err := processBroadcastQueueOnce(token, broadcastID); err != nil {
			return
		}
		b, err := getBroadcast(broadcastID)
		if err != nil || b.Status != "sending" {
			return
		}
	}
}

func processBroadcastQueueOnce(token, broadcastID string) error {
	if !storageConfigured() {
		return fmt.Errorf("storage not configured")
	}
	b, err := getBroadcast(broadcastID)
	if err != nil {
		return err
	}
	if b.Status != "sending" {
		return nil
	}
	ids, err := claimBroadcastRecipients(broadcastID, broadcastBatchSize)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return finishBroadcast(token, b)
	}
	stop := make(chan struct{})
	processBroadcastBatch(token, b, ids, stop)
	pending, err := getPendingRecipients(broadcastID, 1)
	if err == nil && len(pending) == 0 {
		latest, _ := getBroadcast(broadcastID)
		return finishBroadcast(token, latest)
	}
	return nil
}

func claimBroadcastRecipients(broadcastID string, limit int) ([]int64, error) {
	var rows []struct {
		UserID int64 `json:"user_id"`
	}
	if err := supabaseRPCInto("claim_broadcast_recipients", map[string]interface{}{"p_broadcast_id": broadcastID, "p_limit": limit}, &rows); err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.UserID)
	}
	return out, nil
}
func processBroadcastBatch(token string, b BroadcastRecord, ids []int64, stop <-chan struct{}) {
	success, fail, pinsOK, pinsFail := 0, 0, 0, 0
	sentIDs := []int64{}
	failedIDs := []int64{}
	for _, uid := range ids {
		select {
		case <-stop:
			return
		default:
		}
		res, err := sendBroadcastToUser(token, uid, b)
		if err != nil {
			fail++
			failedIDs = append(failedIDs, uid)
			_ = insertBroadcastError(b.ID, uid, err.Error())
			if isInactiveTelegramError(err.Error()) {
				_ = markUserInactive(uid, true)
			}
		} else {
			success++
			sentIDs = append(sentIDs, uid)
			if res {
				pinsOK++
			} else {
				pinsFail++
			}
		}
		time.Sleep(broadcastRateDelay)
	}
	_ = markRecipientStatus(b.ID, sentIDs, "sent")
	_ = markRecipientStatus(b.ID, failedIDs, "failed")
	_ = updateBroadcast(b.ID, map[string]interface{}{"success_count": b.SuccessCount + success, "failed_count": b.FailedCount + fail, "pin_success_count": b.PinSuccessCount + pinsOK, "pin_failed_count": b.PinFailedCount + pinsFail})
	updateBroadcastProgress(token, b.ID)
}

func sendBroadcastToUser(token string, userID int64, b BroadcastRecord) (bool, error) {
	var messageID int
	var err error
	markup := buttonsMarkup(buttonsFromJSON(b.ButtonsJSON))
	switch b.Kind {
	case "text":
		messageID, err = sendBroadcastTextResult(token, userID, b.Text, markup)
	case "photo":
		messageID, err = sendBroadcastPhotoResult(token, userID, b.FileID, b.Caption, markup)
	case "video":
		messageID, err = sendBroadcastVideoResult(token, userID, b.FileID, b.Caption, markup)
	default:
		err = fmt.Errorf("unsupported broadcast kind %s", b.Kind)
	}
	if err != nil {
		return false, err
	}
	pinErr := pinBroadcastMessage(token, userID, messageID)
	return pinErr == nil, nil
}

func sendBroadcastTextResult(token string, chatID int64, text string, markup map[string]interface{}) (int, error) {
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": markup}
	return telegramMessageResult(token, "sendMessage", payload)
}
func sendBroadcastPhotoResult(token string, chatID int64, fileID, caption string, markup map[string]interface{}) (int, error) {
	payload := map[string]interface{}{"chat_id": chatID, "photo": fileID, "caption": caption, "reply_markup": markup}
	return telegramMessageResult(token, "sendPhoto", payload)
}
func sendBroadcastVideoResult(token string, chatID int64, fileID, caption string, markup map[string]interface{}) (int, error) {
	payload := map[string]interface{}{"chat_id": chatID, "video": fileID, "caption": caption, "reply_markup": markup}
	return telegramMessageResult(token, "sendVideo", payload)
}

type telegramSendResult struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
	Description string `json:"description"`
	Parameters  struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

func telegramMessageResult(token, method string, payload map[string]interface{}) (int, error) {
	for attempt := 0; attempt < 4; attempt++ {
		b, _ := json.Marshal(payload)
		resp, err := httpClient.Post("https://api.telegram.org/bot"+token+"/"+method, "application/json", strings.NewReader(string(b)))
		if err != nil {
			if attempt == 3 {
				return 0, err
			}
			time.Sleep(time.Second)
			continue
		}
		var r telegramSendResult
		err = json.NewDecoder(resp.Body).Decode(&r)
		resp.Body.Close()
		if r.Ok {
			return r.Result.MessageID, nil
		}
		if r.Parameters.RetryAfter > 0 {
			time.Sleep(time.Duration(r.Parameters.RetryAfter) * time.Second)
			continue
		}
		return 0, fmt.Errorf("%s", r.Description)
	}
	return 0, fmt.Errorf("telegram retry limit exceeded")
}
func pinBroadcastMessage(token string, chatID int64, messageID int) error {
	payload := map[string]interface{}{"chat_id": chatID, "message_id": messageID, "disable_notification": true}
	b, _ := json.Marshal(payload)
	resp, err := httpClient.Post("https://api.telegram.org/bot"+token+"/pinChatMessage", "application/json", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if !r.Ok {
		return fmt.Errorf("%s", r.Description)
	}
	return nil
}
func isInactiveTelegramError(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "forbidden") || strings.Contains(s, "bot was blocked") || strings.Contains(s, "chat not found") || strings.Contains(s, "user is deactivated")
}

func finishBroadcast(token string, b BroadcastRecord) error {
	now := time.Now().UTC()
	if b.Status != "sending" {
		return nil
	}
	if err := updateBroadcast(b.ID, map[string]interface{}{"status": "completed", "completed_at": now.Format(time.RFC3339)}); err != nil {
		return err
	}
	done := b.SuccessCount + b.FailedCount
	pct := 0.0
	if b.Total > 0 {
		pct = float64(b.SuccessCount) * 100 / float64(b.Total)
	}
	text := fmt.Sprintf("✅ *انتهت الإذاعة*\n\n━━━━━━━━━━━━━━\n👥 الإجمالي: %d\n✅ تم الإرسال: %d\n❌ فشل: %d\n📊 النجاح: %.2f%%\n📌 التثبيت: %d ناجح | %d فشل\n🕐 وقت الانتهاء: %s\n🇮🇶 توقيت بغداد", b.Total, b.SuccessCount, b.FailedCount, pct, b.PinSuccessCount, b.PinFailedCount, formatIraqTime(now))
	sendMessageWithKeyboard(token, b.ProgressChat, text, adminKeyboard([]map[string]interface{}{adminButton("🔙 لوحة الإدارة", "admin_refresh", "primary")}))
	_ = done
	return nil
}
func updateBroadcastProgress(token, broadcastID string) {
	b, err := getBroadcast(broadcastID)
	if err != nil {
		return
	}
	total := b.Total
	done := b.SuccessCount + b.FailedCount
	pct := 0.0
	if total > 0 {
		pct = float64(done) * 100 / float64(total)
	}
	text := fmt.Sprintf("📢 *جاري نشر الإذاعة...*\n\n━━━━━━━━━━━━━━\n👥 الإجمالي: %d\n✅ تم الإرسال: %d\n❌ فشل: %d\n⏳ متبقي: %d\n📊 التقدم: %.1f%%\n🕐 الوقت: %s\n🇮🇶 بغداد", total, b.SuccessCount, b.FailedCount, maxInt(total-done, 0), pct, formatIraqTime(iraqNow()))
	if b.ProgressMessage == 0 {
		id, _ := sendMessageResult(token, b.ProgressChat, text, adminKeyboard([]map[string]interface{}{adminButton("🛑 إيقاف الإذاعة", "broadcast_stop_confirm", "danger")}))
		_ = updateBroadcast(b.ID, map[string]interface{}{"progress_message_id": id})
	} else {
		editMessageText(token, b.ProgressChat, int(b.ProgressMessage), text)
	}
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func sendMessageResult(token string, chatID int64, text string, kb map[string]interface{}) (int, error) {
	return telegramMessageResult(token, "sendMessage", map[string]interface{}{"chat_id": chatID, "text": text, "parse_mode": "Markdown", "reply_markup": kb})
}

func startBroadcastAtomic(id, status string) (BroadcastRecord, error) {
	var rows []BroadcastRecord
	if err := supabaseRPCInto("start_broadcast", map[string]interface{}{"p_broadcast_id": id, "p_status": status}, &rows); err != nil {
		return BroadcastRecord{}, err
	}
	if len(rows) == 0 {
		return BroadcastRecord{}, fmt.Errorf("broadcast unavailable")
	}
	return rows[0], nil
}

func processScheduledBroadcasts(token string) error {
	if !storageConfigured() {
		return nil
	}
	list, err := getScheduledBroadcasts(time.Now(), 5)
	if err != nil {
		return err
	}
	for _, b := range list {
		started, err := startBroadcastAtomic(b.ID, "sending")
		if err != nil {
			continue
		}
		active := true
		total, _ := countUsers(&active)
		_ = updateBroadcast(started.ID, map[string]interface{}{"total": total, "developer_id": developerID, "progress_chat_id": developerID})
		if err := supabaseRPC("enqueue_broadcast_recipients", map[string]interface{}{"p_broadcast_id": started.ID}); err != nil {
			_ = updateBroadcast(started.ID, map[string]interface{}{"status": "failed", "error_text": err.Error()})
			continue
		}
		_ = processBroadcastQueueOnce(token, started.ID)
	}
	for _, b := range mustGetSendingForCron() {
		_ = processBroadcastQueueOnce(token, b.ID)
	}
	return nil
}
func mustGetSendingForCron() []BroadcastRecord { list, _ := getActiveSendingBroadcasts(5); return list }

func supabaseRPC(fn string, body interface{}) error {
	resp, err := supabaseRequest(http.MethodPost, "rpc/"+fn, "", body, "return=representation")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}
func supabaseRPCInto(fn string, body interface{}, out interface{}) error {
	resp, err := supabaseRequest(http.MethodPost, "rpc/"+fn, "", body, "return=representation")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, out)
}

// Persistent storage is backed by Supabase PostgREST. The service-role key is
// used only from the Vercel server function; it must never be exposed to users.
func storageConfigured() bool {
	return strings.TrimSpace(os.Getenv("SUPABASE_URL")) != "" && strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY")) != ""
}

func supabaseBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/") + "/rest/v1/"
}

func supabaseRequest(method, table, query string, body interface{}, prefer string) (*http.Response, error) {
	if !storageConfigured() {
		return nil, fmt.Errorf("persistent storage is not configured: SUPABASE_URL/SUPABASE_SERVICE_ROLE_KEY")
	}
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	u := supabaseBaseURL() + url.PathEscape(table)
	if query != "" {
		u += "?" + query
	}
	req, err := http.NewRequest(method, u, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	req.Header.Set("Content-Type", "application/json")
	if prefer != "" {
		req.Header.Set("Prefer", prefer)
	}
	return httpClient.Do(req)
}

func decodeSupabaseResponse(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	if out != nil && len(b) > 0 {
		if err := json.Unmarshal(b, out); err != nil {
			return err
		}
	}
	return nil
}

type StoredUser struct {
	UserID       int64  `json:"user_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	Language     string `json:"language"`
	FirstSeen    string `json:"first_seen"`
	LastActivity string `json:"last_activity"`
	LastStart    string `json:"last_start"`
	Active       bool   `json:"active"`
	Blocked      bool   `json:"blocked"`
}

func upsertUser(u StoredUser, markStart bool) (isNew bool, err error) {
	if !storageConfigured() {
		return false, nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	row := map[string]interface{}{
		"user_id": u.UserID, "first_name": u.FirstName, "last_name": u.LastName,
		"username": u.Username, "language": u.Language, "last_activity": now,
		"active": true, "blocked": false,
	}
	if markStart {
		row["last_start"] = now
	}
	if u.FirstSeen != "" {
		row["first_seen"] = u.FirstSeen
	}
	var out []StoredUser
	resp, err := supabaseRequest(http.MethodPost, "bot_users", "on_conflict=user_id", []interface{}{row}, "resolution=merge-duplicates,return=representation")
	if err != nil {
		return false, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return false, err
	}
	if len(out) == 0 {
		return false, nil
	}
	// A separate existence check is intentionally avoided: the database's
	// unique key is authoritative and the caller can use last_start to avoid
	// repeated entry notifications.
	return u.FirstSeen == "", nil
}

func getUser(userID int64) (StoredUser, error) {
	var out []StoredUser
	q := "select=*&user_id=eq." + strconv.FormatInt(userID, 10) + "&limit=1"
	resp, err := supabaseRequest(http.MethodGet, "bot_users", q, nil, "")
	if err != nil {
		return StoredUser{}, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return StoredUser{}, err
	}
	if len(out) == 0 {
		return StoredUser{}, nil
	}
	return out[0], nil
}

func markUserInactive(userID int64, blocked bool) error {
	if !storageConfigured() {
		return nil
	}
	row := map[string]interface{}{"active": false, "blocked": blocked, "last_activity": time.Now().UTC().Format(time.RFC3339)}
	resp, err := supabaseRequest(http.MethodPatch, "bot_users", "user_id=eq."+strconv.FormatInt(userID, 10), row, "return=minimal")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}

func getActiveUsers(limit, offset int) ([]StoredUser, error) {
	var out []StoredUser
	q := fmt.Sprintf("select=*&active=eq.true&blocked=eq.false&order=user_id.asc&limit=%d&offset=%d", limit, offset)
	resp, err := supabaseRequest(http.MethodGet, "bot_users", q, nil, "")
	if err != nil {
		return nil, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func countUsers(activeOnly *bool) (int, error) {
	q := "select=user_id&limit=1"
	if activeOnly != nil {
		q += "&active=eq." + strconv.FormatBool(*activeOnly) + "&blocked=eq.false"
	}
	resp, err := supabaseRequest(http.MethodGet, "bot_users", q, nil, "count=exact")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	cr := resp.Header.Get("Content-Range")
	if cr == "" {
		return 0, nil
	}
	slash := strings.LastIndex(cr, "/")
	if slash < 0 {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(cr[slash+1:]))
	return n, nil
}

type BroadcastRecord struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	Text            string `json:"text"`
	FileID          string `json:"file_id"`
	Caption         string `json:"caption"`
	ButtonsJSON     string `json:"buttons_json"`
	Status          string `json:"status"`
	DeveloperID     int64  `json:"developer_id"`
	CreatedAt       string `json:"created_at"`
	ScheduledAt     string `json:"scheduled_at"`
	StartedAt       string `json:"started_at"`
	CompletedAt     string `json:"completed_at"`
	Total           int    `json:"total"`
	SuccessCount    int    `json:"success_count"`
	FailedCount     int    `json:"failed_count"`
	PinSuccessCount int    `json:"pin_success_count"`
	PinFailedCount  int    `json:"pin_failed_count"`
	ProgressMessage int64  `json:"progress_message_id"`
	ProgressChat    int64  `json:"progress_chat_id"`
	ErrorText       string `json:"error_text"`
}

func createBroadcast(b BroadcastRecord) error {
	resp, err := supabaseRequest(http.MethodPost, "broadcasts", "", b, "return=minimal")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}

func updateBroadcast(id string, fields map[string]interface{}) error {
	q := "id=eq." + url.QueryEscape(id)
	resp, err := supabaseRequest(http.MethodPatch, "broadcasts", q, fields, "return=minimal")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}

func getBroadcast(id string) (BroadcastRecord, error) {
	var out []BroadcastRecord
	q := "select=*&id=eq." + url.QueryEscape(id) + "&limit=1"
	resp, err := supabaseRequest(http.MethodGet, "broadcasts", q, nil, "")
	if err != nil {
		return BroadcastRecord{}, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return BroadcastRecord{}, err
	}
	if len(out) == 0 {
		return BroadcastRecord{}, fmt.Errorf("broadcast not found")
	}
	return out[0], nil
}

func listBroadcasts(limit, offset int) ([]BroadcastRecord, error) {
	var out []BroadcastRecord
	q := fmt.Sprintf("select=*&order=created_at.desc&limit=%d&offset=%d", limit, offset)
	resp, err := supabaseRequest(http.MethodGet, "broadcasts", q, nil, "")
	if err != nil {
		return nil, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func countBroadcasts() (int, error) {
	resp, err := supabaseRequest(http.MethodGet, "broadcasts", "select=id&limit=1", nil, "count=exact")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	cr := resp.Header.Get("Content-Range")
	slash := strings.LastIndex(cr, "/")
	if slash < 0 {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(cr[slash+1:]))
	return n, nil
}

func createBroadcastRecipients(broadcastID string, users []StoredUser) error {
	if len(users) == 0 {
		return nil
	}
	rows := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		rows = append(rows, map[string]interface{}{"broadcast_id": broadcastID, "user_id": u.UserID, "status": "pending"})
	}
	resp, err := supabaseRequest(http.MethodPost, "broadcast_recipients", "on_conflict=broadcast_id,user_id", rows, "resolution=ignore-duplicates,return=minimal")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}

func getPendingRecipients(broadcastID string, limit int) ([]int64, error) {
	var rows []struct {
		UserID int64 `json:"user_id"`
	}
	q := fmt.Sprintf("select=user_id&broadcast_id=eq.%s&status=eq.pending&limit=%d&order=user_id.asc", url.QueryEscape(broadcastID), limit)
	resp, err := supabaseRequest(http.MethodGet, "broadcast_recipients", q, nil, "")
	if err != nil {
		return nil, err
	}
	if err := decodeSupabaseResponse(resp, &rows); err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.UserID)
	}
	return out, nil
}

func markRecipientStatus(broadcastID string, ids []int64, status string) error {
	if len(ids) == 0 {
		return nil
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}
	q := "broadcast_id=eq." + url.QueryEscape(broadcastID) + "&user_id=in.(" + strings.Join(parts, ",") + ")"
	row := map[string]interface{}{"status": status, "updated_at": time.Now().UTC().Format(time.RFC3339)}
	resp, err := supabaseRequest(http.MethodPatch, "broadcast_recipients", q, row, "return=minimal")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}

func insertBroadcastError(broadcastID string, userID int64, errText string) error {
	row := map[string]interface{}{"broadcast_id": broadcastID, "user_id": userID, "error_text": errText, "created_at": time.Now().UTC().Format(time.RFC3339)}
	resp, err := supabaseRequest(http.MethodPost, "broadcast_errors", "", row, "return=minimal")
	if err != nil {
		return err
	}
	return decodeSupabaseResponse(resp, nil)
}

func countFailedSends() (int, error) {
	resp, err := supabaseRequest(http.MethodGet, "broadcast_errors", "select=id&limit=1", nil, "count=exact")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	cr := resp.Header.Get("Content-Range")
	slash := strings.LastIndex(cr, "/")
	if slash < 0 {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(cr[slash+1:]))
	return n, nil
}

func getScheduledBroadcasts(now time.Time, limit int) ([]BroadcastRecord, error) {
	var out []BroadcastRecord
	q := fmt.Sprintf("select=*&status=eq.scheduled&scheduled_at=lte.%s&order=scheduled_at.asc&limit=%d", url.QueryEscape(now.UTC().Format(time.RFC3339)), limit)
	resp, err := supabaseRequest(http.MethodGet, "broadcasts", q, nil, "")
	if err != nil {
		return nil, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func getActiveSendingBroadcasts(limit int) ([]BroadcastRecord, error) {
	var out []BroadcastRecord
	q := fmt.Sprintf("select=*&status=eq.sending&order=started_at.asc&limit=%d", limit)
	resp, err := supabaseRequest(http.MethodGet, "broadcasts", q, nil, "")
	if err != nil {
		return nil, err
	}
	if err := decodeSupabaseResponse(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}
