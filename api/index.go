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
	"time"
)

// عميل HTTP عام مع timeout قصير للطلبات النصية العادية
var httpClient = &http.Client{Timeout: 8 * time.Second}

// عميل بـ timeout أطول لعمليات تنزيل/رفع الصور والفيديوهات
var mediaClient = &http.Client{Timeout: 30 * time.Second}

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
		"main_menu_title":       "القائمة الرئيسية 🤖:",
		"welcome":               "أهلاً بك يا %s هذه لوحة التحكم الخاصة بحسابك 🤖\nاختر من الأزرار أدناه للتحكم الكامل:",
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
		"started_msg":           "🟢 تم تشغيل الرد التلقائي بنجاح.",
		"edit_text_prompt":      "📝 أرسل الآن نص الرد التلقائي الجديد:\n(ملاحظة: سيتم إضافة دالة `{name}` تلقائياً إذا لم تقم بكتابتها).",
		"saved_text_msg":        "✅ تم حفظ نص الرد التلقائي الجديد بنجاح!",
		"exclude_prompt":        "👤 أرسل ايدي الحساب المراد استثناؤه الآن:\n(يمكنك أيضاً الضغط على زر User أدناه لتحديد حساب أو بوت لاستخراج ايديه).",
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
		"no_business_connection": "❌ لم يتم ربط حساب تجاري بعد بالبوت.",
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
		"your_id_msg":           "🆔 **الايدي الخاص بالحساب:**\n`%d`\n\nاضغط على الزر أدناه لنسخ الـ ID تلقائياً 👇",
		"fail_name":             "❌ فشل تعديل الاسم: %s",
		"fail_bio":              "❌ فشل تعديل النبذة: %s",
		"fail_username":         "❌ فشل تعديل اليوزر: %s",
		"fail_photo":            "❌ فشل تعديل الصورة: %s",
		"fail_story":            "❌ فشل نشر القصة: %s",
		"need_real_photo":       "❌ أرسل صورة فعلية (لا يقبل ملفات أو نصوص).",
		"need_real_media_story": "❌ أرسل صورة أو فيديو فعلي لنشره كقصة.",
		"video_too_long_error":  "الفيديو أطول من 60 ثانية، وهذا الحد الأقصى المسموح لقصص تليجرام",
	},
	"en": {
		"main_menu_title":       "Main Menu 🤖:",
		"welcome":               "Welcome %s, this is your account control panel 🤖\nChoose from the buttons below for full control:",
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
		"started_msg":           "🟢 Auto-reply has been started.",
		"edit_text_prompt":      "📝 Send the new auto-reply text now:\n(Note: `{name}` tag will be added automatically if omitted).",
		"saved_text_msg":        "✅ New auto-reply text saved successfully!",
		"exclude_prompt":        "👤 Send the account ID to exclude now:\n(You can also click User button below to pick a user/bot).",
		"invalid_id_msg":        "❌ Numbers only! Please send a valid ID.",
		"id_added_msg":          "✅ ID `%d` added to the exclusion list.",
		"list_excluded_title":   "📋 **Excluded Accounts:**\n",
		"no_excluded":           "No excluded accounts currently.",
		"cleared_excluded_msg":  "🧹 All exclusions cleared successfully.",
		"profile_menu_title":    "🧑 Manage Profile - choose what to edit:",
		"edit_first_name_btn":   "✏️ Edit Name",
		"edit_bio_btn":          "📝 Edit Bio",
		"edit_photo_btn":        "🖼️ Edit Photo",
		"edit_username_btn":     "🔗 Edit Username",
		"no_business_connection": "❌ No business account connected to the bot yet.",
		"first_name_prompt":     "✏️ Send the new first name now (optionally followed by a last name):",
		"bio_prompt":            "📝 Send the new bio now (max 70 characters):",
		"username_prompt":       "🔗 Send the new username now (without @):",
		"photo_prompt":          "🖼️ Send the new profile photo now:",
		"name_updated":          "✅ Name updated successfully!",
		"bio_updated":           "✅ Bio updated successfully!",
		"username_updated":      "✅ Username updated successfully!",
		"photo_updated":         "✅ Profile photo updated successfully!",
		"select_story_duration": "⏱️ Select story duration:",
		"dur_6h":                "6 Hours",
		"dur_12h":               "12 Hours",
		"dur_24h":               "24 Hours",
		"dur_48h":               "48 Hours",
		"story_prompt":          "📖 Send a photo or video now (max 60 seconds) to post as a story (visible for %s):",
		"story_updated":         "✅ Story posted successfully! It will remain visible for %s.",
		"your_id_msg":           "🆔 **Account ID:**\n`%d`\n\nTap the button below to copy the ID automatically 👇",
		"fail_name":             "❌ Failed to update name: %s",
		"fail_bio":              "❌ Failed to update bio: %s",
		"fail_username":         "❌ Failed to update username: %s",
		"fail_photo":            "❌ Failed to update photo: %s",
		"fail_story":            "❌ Failed to post story: %s",
		"need_real_photo":       "❌ Please send an actual photo (files or text not accepted).",
		"need_real_media_story": "❌ Please send an actual photo or video to post as a story.",
		"video_too_long_error":  "The video is longer than 60 seconds, which is Telegram's maximum allowed for stories",
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

type BotConfig struct {
	IsStopped      bool    `json:"is_stopped"`
	AutoReply      string  `json:"auto_reply"`
	Excluded       []int64 `json:"excluded"`
	State          string  `json:"state"`
	BusinessConnID string  `json:"business_conn_id"`
	Lang           string  `json:"lang"`
}

type CopyTextButton struct {
	Text string `json:"text"`
}

type InlineKeyboardButton struct {
	Text         string          `json:"text"`
	CallbackData string          `json:"callback_data,omitempty"`
	CopyText     *CopyTextButton `json:"copy_text,omitempty"`
}

type SharedUser struct {
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type UsersShared struct {
	RequestID int32        `json:"request_id"`
	Users     []SharedUser `json:"users"`
}

type UserShared struct {
	RequestID int32 `json:"request_id"`
	UserID    int64 `json:"user_id"`
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

type Message struct {
	MessageID   int          `json:"message_id"`
	Chat        struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From        struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
	} `json:"from"`
	Text        string       `json:"text"`
	Photo       []PhotoSize  `json:"photo"`
	Video       *Video       `json:"video"`
	UserShared  *UserShared  `json:"user_shared"`
	UsersShared *UsersShared `json:"users_shared"`
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

	// 1. معالجة الضغط على الأزرار الشفافة Inline Buttons
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		answerCallback(botToken, cb.ID)

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
			sendSubMenuWithIDCopy(botToken, adminID, lang, tr(lang, "exclude_prompt"), adminID)
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

	// 2. معالجة الرسائل والأوامر والمشاركة من كيبورد User
	if update.Message != nil {
		msg := update.Message
		chatID := msg.Chat.ID

		config, msgID := getConfig(botToken, chatID)
		lang := config.Lang

		// أ) الاستجابة لمشاركة حساب أو بوت عبر زر User بالكيبورد السفلي
		if msg.UsersShared != nil && len(msg.UsersShared.Users) > 0 {
			targetID := msg.UsersShared.Users[0].UserID
			sendIDMessage(botToken, chatID, targetID, lang)
			w.WriteHeader(http.StatusOK)
			return
		}
		if msg.UserShared != nil {
			targetID := msg.UserShared.UserID
			sendIDMessage(botToken, chatID, targetID, lang)
			w.WriteHeader(http.StatusOK)
			return
		}

		// ب) أمر /start بترحيب بالصورة والنص والكيبورد
		if msg.Text == "/start" {
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)
			userName := msg.From.FirstName
			if userName == "" {
				userName = "صديقي"
			}
			sendStartPhotoMenu(botToken, chatID, userName, lang)
			w.WriteHeader(http.StatusOK)
			return
		}

		// ج) أمر /id لإرسال الايدي الخاص بك مع زر النسخ تلقائياً
		if msg.Text == "/id" {
			sendIDMessage(botToken, chatID, msg.From.ID, lang)
			w.WriteHeader(http.StatusOK)
			return
		}

		// د) معالجة التعديلات والإدخالات
		if config.State == "waiting_text" {
			replyInput := strings.TrimSpace(msg.Text)
			// إضافة دالة {name} تلقائياً إذا لم تكن موجودة
			if !strings.Contains(replyInput, "{name}") && !strings.Contains(replyInput, "{الاسم}") {
				replyInput = replyInput + " {name}"
			}
			config.AutoReply = replyInput
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)
			sendMenu(botToken, chatID, lang, tr(lang, "saved_text_msg")+"\n\n💬 *النص المحفوظ:*\n"+replyInput)
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
				sendSubMenuWithIDCopy(botToken, chatID, lang, tr(lang, "invalid_id_msg"), chatID)
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
				sendSubMenu(botToken, chatID, lang, "❌ النبذة طويلة جداً! الحد الأقصى المسموح به من تيليجرام هو 70 حرفاً فقط.")
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

		if msg.IsOutgoing {
			w.WriteHeader(http.StatusOK)
			return
		}

		trimmedText := strings.TrimSpace(msg.Text)

		// تجاهل كل أمر يبدأ بـ / في محادثات البزنس لمنع التداخل والتكرار
		if strings.HasPrefix(trimmedText, "/") {
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

		if senderID == adminID || customerChatID == adminID {
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

		customerName := msg.From.FirstName
		if customerName == "" {
			customerName = "صديقي"
		}

		var detectedLang string
		if len(trimmedText) > 3 {
			translatedToAr, dLang, err := translateText(trimmedText, "ar")
			if err == nil && dLang != "" {
				detectedLang = dLang
				if detectedLang != "ar" && detectedLang != "und" && adminID != 0 {
					notifyMsg := fmt.Sprintf(
						"🌐 *رسالة جديدة بلغة مترجمة (`%s`)*\n👤 *العميل:* %s (`%d`)\n\n💬 *النص الأصلي:*\n%s\n\n✨ *الترجمة للعربية:*\n%s",
						detectedLang, customerName, senderID, trimmedText, translatedToAr,
					)
					sendMessage(botToken, adminID, notifyMsg)
				}
			}
		}

		var replyText string
		if config.AutoReply == "" {
			replyText = "أهلاً بك يا " + customerName + " 🌸\nأنا غير متوفر الآن، اترك رسالتك وسأرد عليك قريباً."
		} else {
			replyText = config.AutoReply
			replyText = strings.ReplaceAll(replyText, "{name}", customerName)
			replyText = strings.ReplaceAll(replyText, "{الاسم}", customerName)
		}

		if detectedLang != "" && detectedLang != "ar" && detectedLang != "und" {
			if translatedReply, _, err := translateText(replyText, detectedLang); err == nil && translatedReply != "" {
				replyText = translatedReply
			}
		}

		sendBusinessReplyWithQuoteButton(botToken, customerChatID, replyText, msg.BusinessConnectionID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 4. رصد تفعيل الربط بالبزنس
	if update.BusinessConnection != nil {
		bc := update.BusinessConnection
		if bc.IsEnabled {
			notifyDeveloper(botToken, bc.User.ID, bc.User.FirstName, bc.User.LastName, bc.User.Username)

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

func getAdminIDFromBusinessConn(token string, connID string) int64 {
	if connID == "" {
		return 0
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getBusinessConnection?business_connection_id=%s", token, connID)
	resp, err := httpClient.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var res BusinessConnectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0
	}
	if res.Result.UserChatID != 0 {
		return res.Result.UserChatID
	}
	return res.Result.User.ID
}

func getConfig(token string, chatID int64) (BotConfig, int) {
	defaultCfg := BotConfig{
		IsStopped:      false,
		AutoReply:      "",
		Excluded:       []int64{},
		State:          "",
		BusinessConnID: "",
		Lang:           "ar",
	}

	if chatID == 0 {
		return defaultCfg, 0
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%d", token, chatID)
	resp, err := httpClient.Get(url)
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

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return defaultCfg, 0
	}

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
		url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", token)
		payload := map[string]interface{}{
			"chat_id":    chatID,
			"message_id": pinnedMsgID,
			"text":       cfgText,
		}
		pBytes, _ := json.Marshal(payload)
		httpClient.Post(url, "application/json", bytes.NewBuffer(pBytes))
	} else {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
		payload := map[string]interface{}{
			"chat_id": chatID,
			"text":    cfgText,
		}
		pBytes, _ := json.Marshal(payload)
		resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(pBytes))
		if err != nil {
			return
		}
		defer resp.Body.Close()
		var res struct {
			Result struct {
				MessageID int `json:"message_id"`
			} `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return
		}
		if res.Result.MessageID != 0 {
			pinUrl := fmt.Sprintf("https://api.telegram.org/bot%s/pinChatMessage", token)
			pinPayload := map[string]interface{}{
				"chat_id":              chatID,
				"message_id":           res.Result.MessageID,
				"disable_notification": true,
			}
			pPinBytes, _ := json.Marshal(pinPayload)
			httpClient.Post(pinUrl, "application/json", bytes.NewBuffer(pPinBytes))
		}
	}
}

// دالة الترحيب بطلب /start من خلال صورة مخصصة مع النص والكيبورد
func sendStartPhotoMenu(token string, chatID int64, userName, lang string) {
	photoURL := "https://od.lk/s/M18zMzMwODEzNDNfV3R3TEM/IMG_20260810_235848_327.jpg"
	caption := fmt.Sprintf(tr(lang, "welcome"), userName)

	keyboard := getMainMenuKeyboard(lang)

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"photo":        photoURL,
		"caption":      caption,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendPhoto", "application/json", bytes.NewBuffer(b))

	// إرسال الكيبورد السفلي الذي يحتوي على زر User
	sendUserReplyKeyboard(token, chatID)
}

// إرسال الكيبورد السفلي (Reply Keyboard) بزر User
func sendUserReplyKeyboard(token string, chatID int64) {
	keyboard := map[string]interface{}{
		"keyboard": [][]map[string]interface{}{
			{
				{
					"text": "👤 User",
					"request_users": map[string]interface{}{
						"request_id":   1,
						"max_quantity": 1,
					},
				},
			},
		},
		"resize_keyboard": true,
		"is_persistent":  true,
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         "👇 يمكنك استخدام زر User في الكيبورد أدناه للنسخ واستخراج ايدي أي شخص أو بوت بسهولة:",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

// دالة إرسال الـ ID بزر النسخ الشفاف التفاعلي
func sendIDMessage(token string, chatID int64, userID int64, lang string) {
	idStr := strconv.FormatInt(userID, 10)
	msgText := fmt.Sprintf(tr(lang, "your_id_msg"), userID)

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]InlineKeyboardButton{
			{
				{
					Text:     fmt.Sprintf("📋 Copy ID%s", idStr),
					CopyText: &CopyTextButton{Text: idStr},
				},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         msgText,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendSubMenuWithIDCopy(token string, chatID int64, lang, text string, userID int64) {
	idStr := strconv.FormatInt(userID, 10)
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]InlineKeyboardButton{
			{
				{
					Text:     fmt.Sprintf("📋 Copy ID%s", idStr),
					CopyText: &CopyTextButton{Text: idStr},
				},
			},
			{
				{Text: tr(lang, "back_btn"), CallbackData: "main_menu"},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func getMainMenuKeyboard(lang string) map[string]interface{} {
	return map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": tr(lang, "stop_btn"), "callback_data": "stop"},
				{"text": tr(lang, "start_btn"), "callback_data": "start"},
			},
			{
				{"text": tr(lang, "edit_text_btn"), "callback_data": "edit_text"},
			},
			{
				{"text": tr(lang, "exclude_btn"), "callback_data": "exclude"},
				{"text": tr(lang, "list_excluded_btn"), "callback_data": "list_excluded"},
			},
			{
				{"text": tr(lang, "clear_excluded_btn"), "callback_data": "clear_excluded"},
			},
			{
				{"text": tr(lang, "profile_menu_btn"), "callback_data": "profile_menu"},
			},
			{
				{"text": tr(lang, "post_story_btn"), "callback_data": "post_story"},
			},
			{
				{"text": tr(lang, "lang_ar_btn"), "callback_data": "lang_ar"},
				{"text": tr(lang, "lang_en_btn"), "callback_data": "lang_en"},
			},
		},
	}
}

func sendMenu(token string, chatID int64, lang, text string) {
	keyboard := getMainMenuKeyboard(lang)

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendStoryDurationMenu(token string, chatID int64, lang string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "⏱️ " + tr(lang, "dur_6h"), "callback_data": "story_dur_21600"},
				{"text": "⏱️ " + tr(lang, "dur_12h"), "callback_data": "story_dur_43200"},
			},
			{
				{"text": "⏱️ " + tr(lang, "dur_24h"), "callback_data": "story_dur_86400"},
				{"text": "⏱️ " + tr(lang, "dur_48h"), "callback_data": "story_dur_172800"},
			},
			{
				{"text": tr(lang, "back_btn"), "callback_data": "main_menu"},
			},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "select_story_duration"),
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendProfileMenu(token string, chatID int64, lang, text string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": tr(lang, "edit_first_name_btn"), "callback_data": "edit_first_name"}},
			{{"text": tr(lang, "edit_bio_btn"), "callback_data": "edit_bio"}},
			{{"text": tr(lang, "edit_photo_btn"), "callback_data": "edit_photo"}},
			{{"text": tr(lang, "edit_username_btn"), "callback_data": "edit_username"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu"}},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendSubMenu(token string, chatID int64, lang, text string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu"}},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendMessage(token string, chatID int64, text string) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendBusinessReplyWithQuoteButton(token string, chatID int64, text, bizID string) {
	initialQuote := quotes[rand.Intn(len(quotes))]

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "✨ " + initialQuote, "callback_data": "change_quote"}},
		},
	}

	payload := map[string]interface{}{
		"chat_id":                chatID,
		"text":                   text,
		"business_connection_id": bizID,
		"reply_markup":           keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func updateButtonQuote(token string, chatID int64, msgID int, newQuote string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "✨ " + newQuote, "callback_data": "change_quote"}},
		},
	}

	payload := map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   msgID,
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/editMessageReplyMarkup", "application/json", bytes.NewBuffer(b))
}

func notifyDeveloper(token string, userID int64, firstName, lastName, username string) {
	devChatID := os.Getenv("DEVELOPER_CHAT_ID")
	if devChatID == "" {
		return
	}
	devID, err := strconv.ParseInt(devChatID, 10, 64)
	if err != nil {
		return
	}

	fullName := firstName
	if lastName != "" {
		fullName += " " + lastName
	}
	if fullName == "" {
		fullName = "غير معروف"
	}

	usernameLine := "لا يوجد يوزر"
	if username != "" {
		usernameLine = "@" + username
	}

	text := fmt.Sprintf(
		"🔔 *تفعيل جديد للبوت*\n\n👤 الاسم: %s\n🆔 الايدي: `%d`\n🔗 اليوزر: %s",
		fullName, userID, usernameLine,
	)

	sendMessage(token, devID, text)
}

type apiResult struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

func downloadTelegramFile(token, fileID string) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", token, fileID)
	resp, err := mediaClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("تعذر الاتصال بتليجرام لجلب الملف")
	}
	defer resp.Body.Close()

	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("رد غير متوقع عند جلب الملف")
	}
	if !res.Ok || res.Result.FilePath == "" {
		return nil, fmt.Errorf(res.Description)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, res.Result.FilePath)
	fResp, err := mediaClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("تعذر تنزيل الملف من تليجرام")
	}
	defer fResp.Body.Close()

	data, err := io.ReadAll(fResp.Body)
	if err != nil {
		return nil, fmt.Errorf("تعذر قراءة بيانات الملف")
	}
	return data, nil
}

func postMultipartBusinessAPI(token, method string, fields map[string]string, fileFieldName, fileName string, fileBytes []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return fmt.Errorf("خطأ داخلي في تجهيز الطلب")
		}
	}

	part, err := writer.CreateFormFile(fileFieldName, fileName)
	if err != nil {
		return fmt.Errorf("خطأ داخلي في تجهيز الملف")
	}
	if _, err := part.Write(fileBytes); err != nil {
		return fmt.Errorf("خطأ داخلي في كتابة الملف")
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("خطأ داخلي في إغلاق الطلب")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("تعذر تجهيز طلب الرفع")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := mediaClient.Do(req)
	if err != nil {
		return fmt.Errorf("تعذر الاتصال بتليجرام")
	}
	defer resp.Body.Close()

	var res apiResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("رد غير متوقع من تليجرام")
	}
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func callBusinessAPI(token, method string, payload map[string]interface{}) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	b, _ := json.Marshal(payload)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("تعذر الاتصال بتليجرام")
	}
	defer resp.Body.Close()

	var res apiResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("رد غير متوقع من تليجرام")
	}
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func setBusinessAccountName(token, businessConnID, firstName, lastName string) error {
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"first_name":             firstName,
	}
	if lastName != "" {
		payload["last_name"] = lastName
	}
	return callBusinessAPI(token, "setBusinessAccountName", payload)
}

func setBusinessAccountBio(token, businessConnID, bio string) error {
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"bio":                    bio,
	}
	return callBusinessAPI(token, "setBusinessAccountBio", payload)
}

func setBusinessAccountUsername(token, businessConnID, username string) error {
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"username":               username,
	}
	return callBusinessAPI(token, "setBusinessAccountUsername", payload)
}

func setBusinessAccountProfilePhoto(token, businessConnID, fileID string) error {
	data, err := downloadTelegramFile(token, fileID)
	if err != nil {
		return err
	}

	photoJSON := `{"type":"static","photo":"attach://photo"}`
	fields := map[string]string{
		"business_connection_id": businessConnID,
		"photo":                  photoJSON,
	}
	return postMultipartBusinessAPI(token, "setBusinessAccountProfilePhoto", fields, "photo", "photo.jpg", data)
}

func postBusinessStory(token, businessConnID, mediaType, fileID string, durationSeconds int, activePeriod string, lang string) error {
	if mediaType == "video" && durationSeconds > 60 {
		return fmt.Errorf(tr(lang, "video_too_long_error"))
	}

	data, err := downloadTelegramFile(token, fileID)
	if err != nil {
		return err
	}

	var contentJSON, fileName string
	if mediaType == "video" {
		if durationSeconds > 0 {
			contentJSON = fmt.Sprintf(`{"type":"video","video":"attach://content","duration":%d}`, durationSeconds)
		} else {
			contentJSON = `{"type":"video","video":"attach://content"}`
		}
		fileName = "story.mp4"
	} else {
		contentJSON = `{"type":"photo","photo":"attach://content"}`
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
	return postMultipartBusinessAPI(token, "postStory", fields, "content", fileName, data)
}

func deleteMessage(token string, chatID int64, msgID int) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": msgID,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/deleteMessage", "application/json", bytes.NewBuffer(b))
}

func answerCallback(token, callbackID string) {
	payload := map[string]string{"callback_query_id": callbackID}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/answerCallbackQuery", "application/json", bytes.NewBuffer(b))
}
