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
		"main_menu_title":       "القائمة الرئيسية 🤖:",
		"welcome":               "أهلاً بك في لوحة تحكم البوت 🤖\nاختر من الأزرار أدناه للتحكم الكامل:",
		"stop_btn":              "🛑 إيقاف الرد",
		"start_btn":             "🟢 تشغيل الرد",
		"edit_text_btn":         "📝 تعديل نص الرد",
		"exclude_btn":           "👤 استثناء حساب",
		"list_excluded_btn":     "📋 عرض المستثنين",
		"clear_excluded_btn":    "🧹 مسح المستثنين",
		"profile_menu_btn":      "🧑 إدارة الملف الشخصي",
		"post_story_btn":        "📖 نشر قصة",
		"lang_ar_btn":           "🇮🇶 العربية",
		"lang_en_btn":           "🇺🇸 English",
		"back_btn":              "🔙 رجوع",
		"stopped_msg":           "🛑 تم إيقاف الرد التلقائي بنجاح.",
		"started_msg":            "🟢 تم تشغيل الرد التلقائي بنجاح.",
		"edit_text_prompt":      "📝 أرسل الآن نص الرد التلقائي الجديد:",
		"saved_text_msg":        "✅ تم حفظ نص الرد التلقائي الجديد بنجاح!",
		"exclude_prompt":        "👤 أرسل ايدي الحساب المراد استثناؤه الآن:",
		"invalid_id_msg":        "❌ أرقام فقط! أرسل الايدي بشكل صحيح.",
		"id_added_msg":          "✅ تم إضافة الايدي `%d` إلى قائمة الاستثناء.",
		"list_excluded_title":   "📋 **قائمة الحسابات المستثناة:**\n",
		"no_excluded":           "لا يوجد حسابات مستثناة حالياً.",
		"cleared_excluded_msg":  "🧹 تم مسح جميع الاستثناءات بنجاح.",
		"profile_menu_title":    "🧑 إدارة الملف الشخصي - اختر ما تريد تعديله:",
		"edit_first_name_btn":   "✏️ تعديل الاسم",
		"edit_bio_btn":          "📝 تعديل النبذة",
		"edit_photo_btn":        "🖼️ تعديل الصورة",
		"edit_username_btn":     "🔗 تعديل اليوزر",
		"no_business_connection":"❌ لم يتم ربط حساب تجاري بعد بالبوت.",
		"first_name_prompt":     "✏️ أرسل الآن الاسم الأول الجديد (والاسم الأخير بعده بمسافة، اختياري):",
		"bio_prompt":            "📝 أرسل الآن النبذة الجديدة (حد أقصى 70 حرف):",
		"username_prompt":       "🔗 أرسل الآن اسم المستخدم الجديد (بدون @):",
		"photo_prompt":          "🖼️ أرسل الآن الصورة الجديدة لملفك الشخصي:",
		"name_updated":          "✅ تم تعديل الاسم بنجاح!",
		"bio_updated":           "✅ تم تعديل النبذة بنجاح!",
		"username_updated":      "✅ تم تعديل اسم المستخدم بنجاح!",
		"photo_updated":         "✅ تم تعديل صورة الملف الشخصي بنجاح!",
		"select_story_duration": "⏱️ اختر مدة ظهور القصة المطلوبة:",
		"dur_6h":                "6 ساعات",
		"dur_12h":               "12 ساعة",
		"dur_24h":               "24 ساعة",
		"dur_48h":               "48 ساعة",
		"story_prompt":          "📖 أرسل الآن صورة أو فيديو (حد أقصى 60 ثانية) لنشره كقصة (ستبقى ظاهرة لمدة %s):",
		"story_updated":         "✅ تم نشر القصة بنجاح! ستبقى ظاهرة لمدة %s.",
		"your_id_msg":           "الايدي الخاص بك هو:\n`%d`",
		"fail_name":             "❌ فشل تعديل الاسم: %s",
		"fail_bio":              "❌ فشل تعديل النبذة: %s",
		"fail_username":         "❌ فشل تعديل اليوزر: %s",
		"fail_photo":            "❌ فشل تعديل الصورة: %s",
		"fail_story":            "❌ فشل نشر القصة: %s",
		"need_real_photo":       "❌ أرسل صورة فعلية (لا يقبل ملفات أو نصوص).",
		"need_real_media_story": "❌ أرسل صورة أو فيديو فعلي لنشره كقصة.",
		"video_too_long_error":  "الفيديو أطول من 60 ثانية، وهذا الحد الأقصى المسموح لقصص تليجرام",
		"id_copy_btn":           "🆔 نسخ الآيدي",
		"share_user_btn":        "👤 User",
		"share_user_prompt":     "👇 استخدم هذا الزر لمشاركة أي مستخدم من قائمة محادثاتك مع البوت، وسيتم استخراج اسمه ويوزره وآيديه تلقائياً:",
		"user_shared_info":      "👤 *معلومات المستخدم المُشارك:*\n\nالاسم: %s\nاليوزر: %s\nالآيدي: `%d`",
		"no_username":           "لا يوجد يوزر",
	},
	"en": {
		"main_menu_title":       "Main Menu 🤖:",
		"welcome":               "Welcome to the bot control panel 🤖\nChoose from the buttons below for full control:",
		"stop_btn":              "🛑 Stop Auto-Reply",
		"start_btn":             "🟢 Start Auto-Reply",
		"edit_text_btn":         "📝 Edit Reply Text",
		"exclude_btn":           "👤 Exclude Account",
		"list_excluded_btn":     "📋 View Excluded",
		"clear_excluded_btn":    "🧹 Clear Excluded",
		"profile_menu_btn":      "🧑 Manage Profile",
		"post_story_btn":        "📖 Post Story",
		"lang_ar_btn":           "🇮🇶 العربية",
		"lang_en_btn":           "🇺🇸 English",
		"back_btn":              "🔙 Back",
		"stopped_msg":           "🛑 Auto-reply has been stopped.",
		"started_msg":            "🟢 Auto-reply has been started.",
		"edit_text_prompt":      "📝 Send the new auto-reply text now:",
		"saved_text_msg":        "✅ New auto-reply text saved successfully!",
		"edit_first_name_btn":   "✏️ Edit Name",
		"edit_bio_btn":          "📝 Edit Bio",
		"edit_photo_btn":        "🖼️ Edit Photo",
		"edit_username_btn":     "🔗 Edit Username",
		"no_business_connection":"❌ No business account connected to the bot yet.",
		"first_name_prompt":     "✏️ Send the new first name now:",
		"bio_prompt":            "📝 Send the new bio now:",
		"username_prompt":       "🔗 Send the new username now:",
		"photo_prompt":          "🖼️ Send the new profile photo now:",
		"name_updated":          "✅ Name updated successfully!",
		"bio_updated":            "✅ Bio updated successfully!",
		"username_updated":       "✅ Username updated successfully!",
		"photo_updated":          "✅ Profile photo updated successfully!",
		"select_story_duration": "⏱️ Select story duration:",
		"dur_6h":                "6 Hours",
		"dur_12h":               "12 Hours",
		"dur_24h":               "24 Hours",
		"dur_48h":               "48 Hours",
		"story_prompt":          "📖 Send a photo or video now:",
		"story_updated":         "✅ Story posted successfully!",
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
	ChatID      int64  `json:"chat_id"`
	Username    string `json:"username"`
	Title       string `json:"title"`
	InviteLink  string `json:"invite_link"`
}

type BotConfig struct {
	IsStopped       bool           `json:"is_stopped"`
	AutoReply       string         `json:"auto_reply"`
	Excluded        []int64        `json:"excluded"`
	State           string         `json:"state"`
	BusinessConnID  string         `json:"business_conn_id"`
	Lang            string         `json:"lang"`

	// إعدادات الاشتراك الإجباري
	ForceChannels   []ForceChannel `json:"force_channels"`

	// إشعار دخول المستخدمين
	EntryNotify     bool           `json:"entry_notify"`
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
	}
}

func isDeveloper(id int64) bool {
	return id == developerID
}

// ============================================================
// Telegram Models
// ============================================================

type TelegramUpdate struct {
	Message         *Message `json:"message"`
	CallbackQuery   *CallbackQuery `json:"callback_query"`

	BusinessMessage *struct {
		MessageID int `json:"message_id"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID int64 `json:"id"`
			FirstName string `json:"first_name"`
			IsBot bool `json:"is_bot"`
		} `json:"from"`
		Text string `json:"text"`
		IsOutgoing bool `json:"is_outgoing"`
		BusinessConnectionID string `json:"business_connection_id"`
	} `json:"business_message"`

	BusinessConnection *struct {
		ID string `json:"id"`
		User struct {
			ID int64 `json:"id"`
			FirstName string `json:"first_name"`
			LastName string `json:"last_name"`
			Username string `json:"username"`
		} `json:"user"`
		UserChatID int64 `json:"user_chat_id"`
		Date int64 `json:"date"`
		IsEnabled bool `json:"is_enabled"`
	} `json:"business_connection"`

	MyChatMember *struct {
		Chat struct {
			ID int64 `json:"id"`
			Type string `json:"type"`
			Title string `json:"title"`
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
	Width int `json:"width"`
	Height int `json:"height"`
}

type Video struct {
	FileID string `json:"file_id"`
	Width int `json:"width"`
	Height int `json:"height"`
	Duration int `json:"duration"`
}

type SharedUserInfo struct {
	UserID int64 `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Username string `json:"username"`
}

type UsersSharedData struct {
	RequestID int64 `json:"request_id"`
	Users []SharedUserInfo `json:"users"`
}

type Message struct {
	MessageID int `json:"message_id"`
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`

	From struct {
		ID int64 `json:"id"`
	} `json:"from"`

	Text string `json:"text"`
	Photo []PhotoSize `json:"photo"`
	Video *Video `json:"video"`
	UsersShared *UsersSharedData `json:"users_shared"`
}

type CallbackQuery struct {
	ID string `json:"id"`
	Message Message `json:"message"`
	Data string `json:"data"`
	From struct {
		ID int64 `json:"id"`
		FirstName string `json:"first_name"`
		Username string `json:"username"`
	} `json:"from"`
}

type BusinessConnectionResponse struct {
	Ok bool `json:"ok"`
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

		if strings.HasPrefix(cb.Data, "admin_") {

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

			// المطور يتجاوز الاشتراك الإجباري
			if !isDeveloper(msg.From.ID) {

				if !checkForceSubscription(
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

			notifyEntry(
				botToken,
				config,
				msg,
				true,
			)

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

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{

			{
				{
					"text": "🔐 الاشتراك الإجباري",
					"callback_data": "admin_force",
					"style": "primary",
				},
			},

			{
				{
					"text": "➕ إضافة قناة",
					"callback_data": "admin_force_add",
					"style": "success",
				},
				{
					"text": "➖ حذف قناة",
					"callback_data": "admin_force_remove",
					"style": "danger",
				},
			},

			{
				{
					"text": "📋 القنوات المضافة",
					"callback_data": "admin_force_list",
					"style": "primary",
				},
			},

			{
				{
					"text": "🔔 إشعارات الدخول",
					"callback_data": "admin_entry",
					"style": "primary",
				},
			},

			{
				{
					"text": "📊 الإحصائيات الحالية",
					"callback_data": "admin_stats",
					"style": "primary",
				},
			},

			{
				{
					"text": "🤖 حالة البوت",
					"callback_data": "admin_status",
					"style": "primary",
				},
			},

			{
				{
					"text": "🔄 تحديث اللوحة",
					"callback_data": "admin_refresh",
					"style": "success",
				},
				{
					"text": "❌ إغلاق",
					"callback_data": "admin_close",
					"style": "danger",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":
			"🛠 *لوحة مطور البوت*\n\n" +
				"مرحباً بك في لوحة التحكم الخاصة بالمطور.\n\n" +
				"اختر القسم الذي تريد التحكم به:",
		"parse_mode": "Markdown",
		"reply_markup": keyboard,
	}

	postJSON(
		token,
		"sendMessage",
		payload,
	)
}

func sendDeveloperOnly(token string, chatID int64) {

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text": "👨‍💻 التواصل مع مطور البوت",
					"url": "https://t.me/" + developerUsername,
					"style": "primary",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":
			"⛔ *هذا الأمر مخصص لمطور البوت فقط.*\n\n" +
				"لا تملك صلاحية الوصول إلى لوحة الإدارة.",
		"parse_mode": "Markdown",
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

func handleAdminCallback(
	token string,
	cb *CallbackQuery,
) {

	chatID := cb.Message.Chat.ID

	config, msgID :=
		getConfig(
			token,
			developerID,
		)

	switch cb.Data {

	case "admin_close":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

	case "admin_refresh":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendAdminPanel(
			token,
			chatID,
		)

	case "admin_force":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendForceAdminMenu(
			token,
			chatID,
			config,
		)

	case "admin_force_add":

		config.State = "admin_waiting_channel"

		saveConfig(
			token,
			developerID,
			config,
			msgID,
		)

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendMessage(
			token,
			chatID,
			"➕ *إضافة قناة للاشتراك الإجباري*\n\n"+
				"أرسل الآن:\n"+
				"• `@channelusername`\n"+
				"أو\n"+
				"• رقم القناة مثل `-1001234567890`\n\n"+
				"⚠️ يجب أن يكون البوت مشرفاً في القناة حتى يستطيع التحقق من اشتراك المستخدمين.\n\n"+
				"بعد إضافتها سيحاول البوت تجهيز رابط دخول للقناة.",
		)

	case "admin_force_remove":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendForceRemoveMenu(
			token,
			chatID,
			config,
		)

	case "admin_force_list":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendForceChannelList(
			token,
			chatID,
			config,
		)

	case "admin_entry":

		config.EntryNotify = !config.EntryNotify

		saveConfig(
			token,
			developerID,
			config,
			msgID,
		)

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		status := "🔴 متوقفة"

		if config.EntryNotify {
			status = "🟢 مفعلة"
		}

		sendAdminPanelWithText(
			token,
			chatID,
			"🔔 *إشعارات الدخول*\n\nالحالة الحالية: "+status,
		)

	case "admin_stats":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		cooldownMu.Lock()
		adminUsers := len(userCooldowns[developerID])
		cooldownMu.Unlock()

		bizCacheMu.Lock()
		businessConnections := len(bizCache)
		bizCacheMu.Unlock()

		text := fmt.Sprintf(
			"📊 *الإحصائيات الحالية*\n\n"+
				"👥 مستخدمون في ذاكرة Cooldown: `%d`\n"+
				"🔗 اتصالات Business في الذاكرة: `%d`\n"+
				"📢 قنوات الاشتراك الإجباري: `%d`\n"+
				"🔔 إشعارات الدخول: `%s`\n\n"+
				"ℹ️ هذه الإحصائيات مؤقتة لأن البوت يعمل على Vercel Serverless.",
			adminUsers,
			businessConnections,
			len(config.ForceChannels),
			boolStatus(config.EntryNotify),
		)

		sendAdminPanelWithText(
			token,
			chatID,
			text,
		)

	case "admin_status":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendAdminPanelWithText(
			token,
			chatID,
			fmt.Sprintf(
				"🤖 *حالة البوت*\n\n"+
					"🟢 النظام يعمل\n"+
					"🔐 الاشتراك الإجباري: %s\n"+
					"📢 عدد القنوات: `%d`\n"+
					"🔔 إشعارات الدخول: %s\n"+
					"👨‍💻 المطور: `%d`",
				forceStatus(config),
				len(config.ForceChannels),
				boolStatus(config.EntryNotify),
				developerID,
			),
		)

	case "admin_force_back":

		deleteMessage(
			token,
			chatID,
			cb.Message.MessageID,
		)

		sendAdminPanel(
			token,
			chatID,
		)

	case "admin_force_delete_confirm":

		indexStr := strings.TrimPrefix(
			cb.Data,
			"admin_force_delete_confirm_",
		)

		_ = indexStr

	case "admin_noop":
		// لا شيء
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
					"text": "➕ إضافة قناة",
					"callback_data": "admin_force_add",
					"style": "success",
				},
			},

			{
				{
					"text": "📋 عرض القنوات",
					"callback_data": "admin_force_list",
					"style": "primary",
				},
			},

			{
				{
					"text": "➖ حذف قناة",
					"callback_data": "admin_force_remove",
					"style": "danger",
				},
			},

			{
				{
					"text": "🔙 رجوع",
					"callback_data": "admin_force_back",
					"style": "danger",
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
		"parse_mode": "Markdown",
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
		"⚠️ *ملاحظة:*\n"+
			"لن يتم حذف أي قناة من هنا إلا بعد الضغط على زر الحذف ثم تأكيد الحذف."

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{
					"text": "➖ حذف قناة",
					"callback_data": "admin_force_remove",
					"style": "danger",
				},
			},
			{
				{
					"text": "🔙 رجوع",
					"callback_data": "admin_force_back",
					"style": "danger",
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
				"text": "🔙 رجوع",
				"callback_data": "admin_force_back",
				"style": "primary",
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
					"text": "❌ إلغاء",
					"callback_data": "admin_force_remove",
					"style": "primary",
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
		ChatID: chat.ID,
		Title: chat.Title,
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

// ============================================================
// getChat
// ============================================================

type TelegramChat struct {
	ID int64 `json:"id"`
	Type string `json:"type"`
	Title string `json:"title"`
	Username string `json:"username"`
}

func getTelegramChat(
	token string,
	chatID string,
) (TelegramChat, error) {

	var res struct {
		Ok bool `json:"ok"`
		Result TelegramChat `json:"result"`
		Description string `json:"description"`
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
		"name": "Bot Force Subscription",
	}

	var res struct {
		Ok bool `json:"ok"`
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
		Ok bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
			IsMember bool `json:"is_member"`
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
						"text": "📢 " + title,
						"url": channel.InviteLink,
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
				"text": "🔄 تحقق من الاشتراك",
				"callback_data": "force_check",
				"style": "success",
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

func notifyEntry(
	token string,
	config BotConfig,
	msg *Message,
	passed bool,
) {

	if !config.EntryNotify {
		return
	}

	if isDeveloper(msg.From.ID) {
		return
	}

	status := "⛔ لم يكمل الاشتراك"

	if passed {
		status = "✅ أكمل الاشتراك"
	}

	username := "لا يوجد"

	// Message model الحالي لا يحتوي username.
	// لذلك نستخدم ID فقط بأمان.

	text := fmt.Sprintf(
		"🔔 *دخول مستخدم جديد*\n\n"+
			"👤 الاسم: مستخدم Telegram\n"+
			"🆔 الايدي: `%d`\n"+
			"🔗 اليوزر: %s\n\n"+
			"📌 الحالة: %s\n"+
			"🕐 الوقت: %s",
		msg.From.ID,
		username,
		status,
		time.Now().Format(
			"2006-01-02 15:04:05",
		),
	)

	sendMessage(
		token,
		developerID,
		text,
	)
}

// ============================================================
// معالجة حالة إضافة القناة
// ============================================================

func handleAdminTextState(
	token string,
	msg *Message,
	config *BotConfig,
	msgID int,
) {

	switch config.State {

	case "admin_waiting_channel":

		channel, err :=
			addForceChannel(
				token,
				msg.Text,
			)

		if err != nil {

			sendMessage(
				token,
				developerID,
				"❌ *فشل إضافة القناة*\n\n"+
					err.Error(),
			)

			return
		}

		// منع التكرار
		for _, existing :=
			range config.ForceChannels {

			if existing.ChatID ==
				channel.ChatID {

				config.State = ""

				saveConfig(
					token,
					developerID,
					*config,
					msgID,
				)

				sendMessage(
					token,
					developerID,
					"⚠️ هذه القناة موجودة مسبقاً ضمن الاشتراك الإجباري.",
				)

				return
			}
		}

		config.ForceChannels =
			append(
				config.ForceChannels,
				channel,
			)

		config.State = ""

		saveConfig(
			token,
			developerID,
			*config,
			msgID,
		)

		text :=
			"✅ *تمت إضافة القناة بنجاح*\n\n"+
				"📢 الاسم: *" +
				escapeMarkdown(channel.Title) +
				"*\n"+
				fmt.Sprintf(
					"🆔 `%d`\n",
					channel.ChatID,
				)

		if channel.Username != "" {

			text +=
				"🔗 @" +
					strings.TrimPrefix(
						channel.Username,
						"@",
					) +
					"\n"
		}

		if channel.InviteLink != "" {

			text +=
				"🔗 رابط الدخول: تم تجهيزه بنجاح\n\n"+
				"🟢 أصبحت القناة الآن ضمن الاشتراك الإجباري."
		}

		sendMessage(
			token,
			developerID,
			text,
		)
	}
}

// ============================================================
// Helpers
// ============================================================

func forceStatus(config BotConfig) string {

	if len(config.ForceChannels) == 0 {
		return "⚪ غير مفعّل"
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
				MessageID int `json:"message_id"`
				Text string `json:"text"`
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
			"chat_id": chatID,
			"message_id": pinnedMsgID,
			"text": cfgText,
		}

		postJSON(
			token,
			"editMessageText",
			payload,
		)

	} else {

		payload := map[string]interface{}{
			"chat_id": chatID,
			"text": cfgText,
		}

		var res struct {
			Ok bool `json:"ok"`
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
				"chat_id": chatID,
				"message_id": res.Result.MessageID,
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
		"chat_id": chatID,
		"text": text,
		"parse_mode": "Markdown",
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
					"text": "🔙 لوحة المطور",
					"callback_data": "admin_refresh",
					"style": "primary",
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
		"photo": startPhotoURL,
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
						"request_id": 1,
						"request_name": true,
						"request_username": true,
					},
					"style": "success",
				},
			},
		},
		"resize_keyboard": true,
		"is_persistent": true,
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": tr(lang, "share_user_prompt"),
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
					"text": tr(lang, "stop_btn"),
					"callback_data": "stop",
					"style": "danger",
				},
				{
					"text": tr(lang, "start_btn"),
					"callback_data": "start",
					"style": "success",
				},
			},
			{
				{
					"text": tr(lang, "edit_text_btn"),
					"callback_data": "edit_text",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "exclude_btn"),
					"callback_data": "exclude",
					"style": "primary",
				},
				{
					"text": tr(lang, "list_excluded_btn"),
					"callback_data": "list_excluded",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "clear_excluded_btn"),
					"callback_data": "clear_excluded",
					"style": "danger",
				},
			},
			{
				{
					"text": tr(lang, "profile_menu_btn"),
					"callback_data": "profile_menu",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "post_story_btn"),
					"callback_data": "post_story",
					"style": "primary",
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
					"text": tr(lang, "lang_ar_btn"),
					"callback_data": "lang_ar",
					"style": "primary",
				},
				{
					"text": tr(lang, "lang_en_btn"),
					"callback_data": "lang_en",
					"style": "primary",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": text,
		"parse_mode": "Markdown",
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
					"text": "⏱️ " + tr(lang, "dur_6h"),
					"callback_data": "story_dur_21600",
					"style": "primary",
				},
				{
					"text": "⏱️ " + tr(lang, "dur_12h"),
					"callback_data": "story_dur_43200",
					"style": "primary",
				},
			},
			{
				{
					"text": "⏱️ " + tr(lang, "dur_24h"),
					"callback_data": "story_dur_86400",
					"style": "primary",
				},
				{
					"text": "⏱️ " + tr(lang, "dur_48h"),
					"callback_data": "story_dur_172800",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "back_btn"),
					"callback_data": "main_menu",
					"style": "danger",
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
					"text": tr(lang, "edit_first_name_btn"),
					"callback_data": "edit_first_name",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "edit_bio_btn"),
					"callback_data": "edit_bio",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "edit_photo_btn"),
					"callback_data": "edit_photo",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "edit_username_btn"),
					"callback_data": "edit_username",
					"style": "primary",
				},
			},
			{
				{
					"text": tr(lang, "back_btn"),
					"callback_data": "main_menu",
					"style": "danger",
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
					"text": tr(lang, "back_btn"),
					"callback_data": "main_menu",
					"style": "danger",
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
		"chat_id": chatID,
		"text": text,
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
					"text": "فعلني من هنا",
					"url": "https://t.me/Xhwe2/10",
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
					"text": "فعلني من هنا",
					"url": "https://t.me/Xhwe2/10",
					"style": "success",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":
			"انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز",
		"business_connection_id": bizID,
		"reply_markup": keyboard,
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
					"text": "✨ " + initialQuote,
					"callback_data": "change_quote",
					"style": "primary",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": text,
		"business_connection_id": bizID,
		"reply_markup": keyboard,
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
	Ok bool `json:"ok"`
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
		"first_name": firstName,
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
			"bio": bio,
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
			"username": username,
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
		Ok bool `json:"ok"`
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
		"business_connection_id":
			businessConnID,
		"photo":
			`{"type":"static","photo":"attach://photo"}`,
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
		"business_connection_id":
			businessConnID,
		"content":
			contentJSON,
		"active_period":
			activePeriod,
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

func deleteMessage(
	token string,
	chatID int64,
	msgID int,
) {

	postJSON(
		token,
		"deleteMessage",
		map[string]interface{}{
			"chat_id": chatID,
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
					"text": "✨ " + newQuote,
					"callback_data": "change_quote",
					"style": "primary",
				},
			},
		},
	}

	postJSON(
		token,
		"editMessageReplyMarkup",
		map[string]interface{}{
			"chat_id": chatID,
			"message_id": msgID,
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
