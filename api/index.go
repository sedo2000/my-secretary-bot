package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{Timeout: 8 * time.Second}
var mediaClient = &http.Client{Timeout: 30 * time.Second}

var (
	cooldownMu    sync.Mutex
	userCooldowns = make(map[int64]map[int64]time.Time)

	bizCacheMu sync.Mutex
	bizCache   = make(map[string]int64)
)

const startPhotoURL = "https://od.lk/s/M18zMzMwODEzNDNfV3R3TEM/IMG_20260810_235848_327.jpg"

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
		"started_msg":           "🟢 تم تشغيل الرد التلقائي بنجاح.",
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
		"story_prompt":          "📖 أرسل صورة، فيديو (حد أقصى 60 ثانية)، أو **رابط منشور قناة تيليجرام** لنشره كقصة (ستبقى ظاهرة لمدة %s):",
		"story_updated":         "✅ تم نشر القصة بنجاح! ستبقى ظاهرة لمدة %s.",
		"your_id_msg":           "الايدي الخاص بك هو:\n`%d`",
		"fail_name":             "❌ فشل تعديل الاسم: %s",
		"fail_bio":              "❌ فشل تعديل النبذة: %s",
		"fail_username":         "❌ فشل تعديل اليوزر: %s",
		"fail_photo":            "❌ فشل تعديل الصورة: %s",
		"fail_story":            "❌ فشل نشر القصة: %s",
		"need_real_photo":       "❌ أرسل صورة فعلية (لا يقبل ملفات أو نصوص).",
		"need_real_media_story": "❌ أرسل صورة، فيديو، أو رابط منشور قناة صحيح لنشره كقصة.",
		"invalid_channel_link":  "❌ رابط المنشور غير صالح أو غير مدعوم. أرسل رابطاً صحيحاً أو صورة/فيديو.",
		"forward_failed":        "❌ فشل جلب المحتوى من الرابط. تأكد من أن البوت مشرف في القناة وأن الرابط صحيح.",
		"no_media_in_link":      "❌ المنشور المذكور لا يحتوي على صورة أو فيديو صالح.",
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
		"started_msg":           "🟢 Auto-reply has been started.",
		"edit_text_prompt":      "📝 Send the new auto-reply text now:",
		"saved_text_msg":        "✅ New auto-reply text saved successfully!",
		"exclude_prompt":        "👤 Send the account ID to exclude now:",
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
		"story_prompt":          "📖 Send a photo, video (max 60s), or **Telegram channel post link** to post as a story (visible for %s):",
		"story_updated":         "✅ Story posted successfully! It will remain visible for %s.",
		"your_id_msg":           "Your ID is:\n`%d`",
		"fail_name":             "❌ Failed to update name: %s",
		"fail_bio":              "❌ Failed to update bio: %s",
		"fail_username":         "❌ Failed to update username: %s",
		"fail_photo":            "❌ Failed to update photo: %s",
		"fail_story":            "❌ Failed to post story: %s",
		"need_real_photo":       "❌ Please send an actual photo (files or text not accepted).",
		"need_real_media_story": "❌ Please send an actual photo, video, or a valid channel post link.",
		"invalid_channel_link":  "❌ Invalid channel post link format. Please send a correct link or photo/video.",
		"forward_failed":        "❌ Failed to fetch content from the link. Make sure the bot is an admin in the channel.",
		"no_media_in_link":      "❌ The referenced post does not contain a valid photo or video.",
		"video_too_long_error":  "The video is longer than 60 seconds, which is Telegram's maximum allowed for stories",
		"id_copy_btn":           "🆔 Copy ID",
		"share_user_btn":        "👤 User",
		"share_user_prompt":     "👇 Use this button to share any user from your chat list with the bot — their name, username and ID will be extracted automatically:",
		"user_shared_info":      "👤 *Shared User Info:*\n\nName: %s\nUsername: %s\nID: `%d`",
		"no_username":           "No username",
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

func parseTelegramPostLink(link string) (string, int, error) {
	link = strings.TrimSpace(link)
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "www.")

	if !strings.HasPrefix(link, "t.me/") && !strings.HasPrefix(link, "telegram.me/") {
		return "", 0, fmt.Errorf("not a telegram link")
	}

	parts := strings.Split(link, "/")
	if len(parts) < 3 {
		return "", 0, fmt.Errorf("incomplete link")
	}

	if parts[1] == "c" {
		if len(parts) < 4 {
			return "", 0, fmt.Errorf("invalid private link")
		}
		chatIDStr := "-100" + parts[2]
		msgID, err := strconv.Atoi(parts[3])
		if err != nil {
			return "", 0, err
		}
		return chatIDStr, msgID, nil
	} else {
		username := "@" + parts[1]
		msgID, err := strconv.Atoi(parts[2])
		if err != nil {
			return "", 0, err
		}
		return username, msgID, nil
	}
}

type ForwardMessageResponse struct {
	Ok     bool    `json:"ok"`
	Result Message `json:"result"`
}

func forwardMessageToGetContent(botToken string, fromChatID string, messageID int, targetChatID int64) (*Message, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/forwardMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":      targetChatID,
		"from_chat_id": fromChatID,
		"message_id":   messageID,
	}
	body, _ := json.Marshal(payload)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res ForwardMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, fmt.Errorf("telegram api error")
	}
	return &res.Result, nil
}

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
		ID int64 `json:"id"`
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

func setBusinessAccountName(botToken, businessConnID, firstName, lastName string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountName", botToken)
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"first_name":             firstName,
		"last_name":              lastName,
	}
	body, _ := json.Marshal(payload)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func setBusinessAccountBio(botToken, businessConnID, bio string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountBio", botToken)
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"bio":                    bio,
	}
	body, _ := json.Marshal(payload)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func setBusinessAccountUsername(botToken, businessConnID, username string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountUsername", botToken)
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"username":               username,
	}
	body, _ := json.Marshal(payload)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func setBusinessAccountProfilePhoto(botToken, businessConnID, fileID string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setBusinessAccountProfilePhoto", botToken)
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"photo":                  fileID,
	}
	body, _ := json.Marshal(payload)
	resp, err := mediaClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func postBusinessStory(botToken, businessConnID, mediaType, fileID string, duration int, period string, lang string) error {
	if mediaType == "video" && duration > 60 {
		return fmt.Errorf(tr(lang, "video_too_long_error"))
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/postBusinessStory", botToken)
	durInt, _ := strconv.Atoi(period)
	if durInt == 0 {
		durInt = 86400
	}
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"period":                 durInt,
	}
	if mediaType == "photo" {
		payload["photo"] = fileID
	} else if mediaType == "video" {
		payload["video"] = fileID
	}
	body, _ := json.Marshal(payload)
	resp, err := mediaClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

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

	// معالجة ارتباط الحساب التجاري
	if update.BusinessConnection != nil {
		bc := update.BusinessConnection
		bizCacheMu.Lock()
		if bc.IsEnabled {
			bizCache[bc.ID] = bc.UserChatID
		} else {
			delete(bizCache, bc.ID)
		}
		bizCacheMu.Unlock()

		if bc.IsEnabled && bc.UserChatID != 0 {
			config, msgID := getConfig(botToken, bc.UserChatID)
			config.BusinessConnID = bc.ID
			saveConfig(botToken, bc.UserChatID, config, msgID)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// معالجة الرسائل عبر الحساب التجاري (Business Messages) للرد التلقائي
	if update.BusinessMessage != nil {
		bm := update.BusinessMessage
		if bm.IsOutgoing || bm.From.IsBot {
			w.WriteHeader(http.StatusOK)
			return
		}

		bizCacheMu.Lock()
		ownerChatID, found := bizCache[bm.BusinessConnectionID]
		bizCacheMu.Unlock()

		if !found {
			ownerChatID = bm.Chat.ID
		}

		config, _ := getConfig(botToken, ownerChatID)
		if config.IsStopped {
			w.WriteHeader(http.StatusOK)
			return
		}

		// التحقق من الاستثناءات
		senderID := bm.From.ID
		for _, exID := range config.Excluded {
			if exID == senderID {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		// نظام التهدئة (Cooldown) لمنع التكرار
		cooldownMu.Lock()
		if _, ok := userCooldowns[ownerChatID]; !ok {
			userCooldowns[ownerChatID] = make(map[int64]time.Time)
		}
		if expiry, ok := userCooldowns[ownerChatID][senderID]; ok && time.Now().Before(expiry) {
			cooldownMu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
		userCooldowns[ownerChatID][senderID] = time.Now().Add(10 * time.Second)
		cooldownMu.Unlock()

		replyText := config.AutoReply
		if replyText == "" {
			replyText = "أهلاً بك، تم استلام رسالتك وسيتم الرد عليك قريباً."
		}

		sendBusinessReply(botToken, bm.BusinessConnectionID, bm.Chat.ID, bm.MessageID, replyText)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. معالجة الضغط على الأزرار الشفافة
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

	// 2. معالجة محادثة التحكم الخاصة بك
	if update.Message != nil {
		msg := update.Message
		chatID := msg.Chat.ID

		config, msgID := getConfig(botToken, chatID)
		lang := config.Lang

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
			var fileID string
			var mediaType string
			var duration int

			if len(msg.Photo) > 0 {
				fileID = msg.Photo[len(msg.Photo)-1].FileID
				mediaType = "photo"
			} else if msg.Video != nil {
				fileID = msg.Video.FileID
				mediaType = "video"
				duration = msg.Video.Duration
			} else if strings.Contains(msg.Text, "t.me/") {
				fromChatID, msgIDNum, err := parseTelegramPostLink(msg.Text)
				if err != nil {
					sendSubMenu(botToken, chatID, lang, tr(lang, "invalid_channel_link"))
					w.WriteHeader(http.StatusOK)
					return
				}
				forwardedMsg, err := forwardMessageToGetContent(botToken, fromChatID, msgIDNum, chatID)
				if err != nil || forwardedMsg == nil {
					sendSubMenu(botToken, chatID, lang, tr(lang, "forward_failed"))
					w.WriteHeader(http.StatusOK)
					return
				}
				deleteMessage(botToken, chatID, forwardedMsg.MessageID)

				if len(forwardedMsg.Photo) > 0 {
					fileID = forwardedMsg.Photo[len(forwardedMsg.Photo)-1].FileID
					mediaType = "photo"
				} else if forwardedMsg.Video != nil {
					fileID = forwardedMsg.Video.FileID
					mediaType = "video"
					duration = forwardedMsg.Video.Duration
				} else {
					sendSubMenu(botToken, chatID, lang, tr(lang, "no_media_in_link"))
					w.WriteHeader(http.StatusOK)
					return
				}
			} else {
				sendSubMenu(botToken, chatID, lang, tr(lang, "need_real_media_story"))
				w.WriteHeader(http.StatusOK)
				return
			}

			err := postBusinessStory(botToken, config.BusinessConnID, mediaType, fileID, duration, period, lang)
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
}

func sendBusinessReply(botToken, businessConnID string, chatID int64, messageID int, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"business_connection_id": businessConnID,
		"chat_id":                chatID,
		"text":                   text,
		"reply_parameters": map[string]interface{}{
			"message_id": messageID,
		},
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func getConfig(botToken string, adminID int64) (BotConfig, int) {
	// يمكن ربطها بقاعدة بيانات أو تخزين مؤقت محلي حسب رغبتك
	return BotConfig{Lang: "ar"}, 0
}

func saveConfig(botToken string, adminID int64, config BotConfig, msgID int) {
	// حفظ الإعدادات
}

func sendMessage(botToken string, chatID int64, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendSubMenu(botToken string, chatID int64, lang, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu"}},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendMenu(botToken string, chatID int64, lang, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
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
				{"text": tr(lang, "post_story_btn"), "callback_data": "post_story"},
			},
			{
				{"text": tr(lang, "lang_ar_btn"), "callback_data": "lang_ar"},
				{"text": tr(lang, "lang_en_btn"), "callback_data": "lang_en"},
			},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendProfileMenu(botToken string, chatID int64, lang, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{"text": tr(lang, "edit_first_name_btn"), "callback_data": "edit_first_name"},
				{"text": tr(lang, "edit_bio_btn"), "callback_data": "edit_bio"},
			},
			{
				{"text": tr(lang, "edit_username_btn"), "callback_data": "edit_username"},
				{"text": tr(lang, "edit_photo_btn"), "callback_data": "edit_photo"},
			},
			{
				{"text": tr(lang, "back_btn"), "callback_data": "main_menu"},
			},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendStoryDurationMenu(botToken string, chatID int64, lang string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{"text": tr(lang, "dur_6h"), "callback_data": "story_dur_21600"},
				{"text": tr(lang, "dur_12h"), "callback_data": "story_dur_43200"},
			},
			{
				{"text": tr(lang, "dur_24h"), "callback_data": "story_dur_86400"},
				{"text": tr(lang, "dur_48h"), "callback_data": "story_dur_172800"},
			},
			{
				{"text": tr(lang, "back_btn"), "callback_data": "main_menu"},
			},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "select_story_duration"),
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendStartPhoto(botToken string, chatID int64, lang string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", botToken)
	quote := quotes[rand.Intn(len(quotes))]
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "✨ اقتباس جديد", "callback_data": "change_quote"}},
			{{"text": tr(lang, "id_copy_btn"), "callback_data": "get_id_action"}},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"photo":        startPhotoURL,
		"caption":      fmt.Sprintf("💡 *اقتباس اليوم:*\n\n_%s_", quote),
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendUserShareKeyboard(botToken string, chatID int64, lang string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	keyboard := map[string]interface{}{
		"keyboard": [][]map[string]interface{}{
			{
				{
					"text": tr(lang, "share_user_btn"),
					"request_users": map[string]interface{}{
						"request_id": 1,
						"max_quantity": 1,
					},
				},
			},
		},
		"resize_keyboard":   true,
		"is_persistent":     true,
		"selective":         true,
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "share_user_prompt"),
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func answerCallback(botToken, callbackID string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", botToken)
	payload := map[string]string{"callback_query_id": callbackID}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func deleteMessage(botToken string, chatID int64, messageID int) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func updateButtonQuote(botToken string, chatID int64, messageID int, newQuote string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageCaption", botToken)
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "✨ اقتباس جديد", "callback_data": "change_quote"}},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   messageID,
		"caption":      fmt.Sprintf("💡 *اقتباس اليوم:*\n\n_%s_", newQuote),
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(apiURL, "application/json", bytes.NewBuffer(body))
}

func sendNerdBotInfo(botToken string, chatID int64) {
	info := "🤖 *معلومات البوت البرمجية:*\n\n- البوت مكتوب بلغة Go.\n- يعمل بنظام Serverless Functions.\n- يدعم إدارة الحسابات التجارية، نشر القصص، والردود التلقائية."
	sendMessage(botToken, chatID, info)
}
