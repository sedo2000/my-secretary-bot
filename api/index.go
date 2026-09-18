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

// عميل HTTP عام مع timeout قصير
var httpClient = &http.Client{Timeout: 8 * time.Second}

// عميل بـ timeout أطول للوسائط
var mediaClient = &http.Client{Timeout: 60 * time.Second}

// متغيرات التهدئة والتخزين المؤقت
var (
	cooldownMu    sync.Mutex
	userCooldowns = make(map[int64]map[int64]time.Time)

	bizCacheMu sync.Mutex
	bizCache   = make(map[string]int64)

	// تخزين قائمة الستوريات في الذاكرة
	storyBatchMu sync.Mutex
	storyBatch   = make(map[int64][]StoryItem)
)

// صورة الترحيب
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

// --- قاموس الترجمة ---
var translations = map[string]map[string]string{
	"ar": {
		"main_menu_title":          "القائمة الرئيسية 🤖:",
		"welcome":                  "أهلاً بك في لوحة تحكم البوت 🤖\nاختر من الأزرار أدناه للتحكم الكامل:",
		"stop_btn":                 "🛑 إيقاف الرد",
		"start_btn":                "🟢 تشغيل الرد",
		"edit_text_btn":            "📝 تعديل نص الرد",
		"media_reply_btn":          "🎬 الرد التلقائي بالوسائط",
		"interaction_menu_btn":     "💬 الرد على التفاعلات",
		"batch_story_btn":          "📚 نشر ستوريات متعددة",
		"exclude_btn":              "👤 استثناء حساب",
		"list_excluded_btn":        "📋 عرض المستثنين",
		"clear_excluded_btn":       "🧹 مسح المستثنين",
		"profile_menu_btn":         "🧑 إدارة الملف الشخصي",
		"post_story_btn":           "📖 نشر قصة",
		"lang_ar_btn":              "🇮🇶 العربية",
		"lang_en_btn":              "🇺🇸 English",
		"back_btn":                 "🔙 رجوع",
		"stopped_msg":              "🛑 تم إيقاف الرد التلقائي بنجاح.",
		"started_msg":              "🟢 تم تشغيل الرد التلقائي بنجاح.",
		"edit_text_prompt":         "📝 أرسل الآن نص الرد التلقائي الجديد:",
		"saved_text_msg":           "✅ تم حفظ نص الرد التلقائي الجديد بنجاح!",
		"exclude_prompt":           "👤 أرسل ايدي الحساب المراد استثناؤه الآن:",
		"invalid_id_msg":           "❌ أرقام فقط! أرسل الايدي بشكل صحيح.",
		"id_added_msg":             "✅ تم إضافة الايدي `%d` إلى قائمة الاستثناء.",
		"list_excluded_title":      "📋 **قائمة الحسابات المستثناة:**\n",
		"no_excluded":              "لا يوجد حسابات مستثناة حالياً.",
		"cleared_excluded_msg":     "🧹 تم مسح جميع الاستثناءات بنجاح.",
		"profile_menu_title":       "🧑 إدارة الملف الشخصي - اختر ما تريد تعديله:",
		"edit_first_name_btn":      "✏️ تعديل الاسم",
		"edit_bio_btn":             "📝 تعديل النبذة",
		"edit_photo_btn":           "🖼️ تعديل الصورة",
		"edit_username_btn":        "🔗 تعديل اليوزر",
		"no_business_connection":   "❌ لم يتم ربط حساب تجاري بعد بالبوت.",
		"first_name_prompt":        "✏️ أرسل الآن الاسم الأول الجديد (والاسم الأخير بعده بمسافة، اختياري):",
		"bio_prompt":               "📝 أرسل الآن النبذة الجديدة (حد أقصى 70 حرف):",
		"username_prompt":          "🔗 أرسل الآن اسم المستخدم الجديد (بدون @):",
		"photo_prompt":             "🖼️ أرسل الآن الصورة الجديدة لملفك الشخصي:",
		"name_updated":             "✅ تم تعديل الاسم بنجاح!",
		"bio_updated":              "✅ تم تعديل النبذة بنجاح!",
		"username_updated":         "✅ تم تعديل اسم المستخدم بنجاح!",
		"photo_updated":            "✅ تم تعديل صورة الملف الشخصي بنجاح!",
		"select_story_duration":    "⏱️ اختر مدة ظهور القصة المطلوبة:",
		"dur_6h":                   "6 ساعات",
		"dur_12h":                  "12 ساعة",
		"dur_24h":                  "24 ساعة",
		"dur_48h":                  "48 ساعة",
		"story_prompt":             "📖 أرسل الآن صورة أو فيديو (حد أقصى 60 ثانية) لنشره كقصة (ستبقى ظاهرة لمدة %s):",
		"story_updated":            "✅ تم نشر القصة بنجاح في الدائرة العلوية! ستبقى ظاهرة لمدة %s.",
		"your_id_msg":              "الايدي الخاص بك هو:\n`%d`",
		"fail_name":                "❌ فشل تعديل الاسم: %s",
		"fail_bio":                 "❌ فشل تعديل النبذة: %s",
		"fail_username":            "❌ فشل تعديل اليوزر: %s",
		"fail_photo":               "❌ فشل تعديل الصورة: %s",
		"fail_story":               "❌ فشل نشر القصة: %s",
		"need_real_photo":          "❌ أرسل صورة فعلية (لا يقبل ملفات أو نصوص).",
		"need_real_media_story":    "❌ أرسل صورة أو فيديو فعلي لنشره كقصة.",
		"video_too_long_error":     "الفيديو أطول من 60 ثانية، وهذا الحد الأقصى المسموح لقصص تليجرام",
		"id_copy_btn":              "🆔 نسخ الآيدي",
		"share_user_btn":           "👤 User",
		"share_user_prompt":        "👇 استخدم هذا الزر لمشاركة أي مستخدم من قائمة محادثاتك مع البوت:",
		"user_shared_info":         "👤 *معلومات المستخدم المُشارك:*\n\nالاسم: %s\nاليوزر: %s\nالآيدي: `%d`",
		"no_username":              "لا يوجد يوزر",

		"media_menu_title":         "🎬 اختر نوع الوسائط للرد التلقائي:",
		"media_text_btn":           "📝 نص فقط",
		"media_voice_btn":          "🎤 رسالة صوتية",
		"media_audio_btn":          "🎵 ملف صوتي",
		"media_gif_btn":            "🎞️ GIF متحرك",
		"media_sticker_btn":        "😀 ملصق",
		"media_video_btn":          "🎥 فيديو",
		"media_photo_btn":          "🖼️ صورة",
		"media_preview_btn":        "👁️ معاينة الرد الحالي",
		"media_clear_btn":          "🗑️ حذف الوسائط والعودة للنص",
		"media_upload_prompt":      "📤 أرسل الآن %s التي تريد استخدامها كرد تلقائي:",
		"media_saved_msg":          "✅ تم حفظ الرد التلقائي بنجاح!\nالنوع: *%s*",
		"media_saved_with_caption": "✅ تم حفظ الرد التلقائي بنجاح!\nالنوع: *%s*\nالنص المصاحب: %s",
		"media_cleared_msg":        "🗑️ تم حذف الوسائط، الرد التلقائي الآن نصي فقط.",
		"current_media_info":       "📌 *الرد التلقائي الحالي:*\n\nالنوع: *%s*\nالنص المصاحب: %s",
		"no_media_set":             "⚠️ لا يوجد وسائط محددة حالياً، الرد نصي فقط.",
		"need_media_error":         "❌ يجب إرسال %s فعلي!",
		"media_type_voice":         "رسالة صوتية",
		"media_type_audio":         "ملف صوتي",
		"media_type_gif":           "GIF متحرك",
		"media_type_sticker":       "ملصق",
		"media_type_video":         "فيديو",
		"media_type_photo":         "صورة",
		"media_type_text":          "نص فقط",
		"no_caption":               "بدون نص مصاحب",

		"interaction_menu_title":   "💬 إعدادات الردود على التفاعلات:",
		"reply_story_mention_btn":  "📸 الرد على ذكر البوت في ستوري",
		"reply_reaction_btn":       "👍 الرد على الريأكشن",
		"reply_voice_btn":          "🎤 الرد على الرسائل الصوتية",
		"interaction_on":           "✅ مفعّل",
		"interaction_off":          "❌ معطّل",
		"story_mention_prompt":     "📸 أرسل نص الرد على ذكر البوت في الستوري:",
		"reaction_prompt":          "👍 أرسل نص الرد على الريأكشن:",
		"voice_reply_prompt":       "🎤 أرسل نص الرد على الرسائل الصوتية:",

		"batch_story_title":        "📚 إعداد قائمة الستوريات للنشر المتسلسل\n\nالستوريات ستُنشر في الدائرة العلوية واحدة تلو الأخرى:",
		"batch_story_add":          "➕ إضافة عنصر",
		"batch_story_list":         "📋 عرض القائمة",
		"batch_story_clear":        "🗑️ مسح القائمة",
		"batch_story_publish":      "🚀 بدء النشر",
		"batch_story_prompt":       "📤 أرسل الآن صورة أو فيديو لإضافته إلى القائمة:",
		"batch_story_added":        "✅ تمت الإضافة! المجموع: %d عنصر",
		"batch_story_empty":        "⚠️ القائمة فارغة، أضف عناصر أولاً.",
		"batch_story_list_title":   "📋 *قائمة الستوريات (%d عنصر):*\n\n⏱️ المدة: *%s*\n\n",
		"batch_story_item_line":    "%d. %s\n",
		"batch_story_cleared":      "🗑️ تم مسح القائمة بالكامل.",
		"batch_story_need_media":   "❌ أرسل صورة أو فيديو فعلي.",
		"batch_story_duration_set": "✅ تم تحديد المدة: *%s*\nالآن أضف العناصر (صور/فيديوهات).",
		"batch_story_choose_duration": "⏱️ اختر مدة ظهور الستوريات المتعددة:",

		// نصوص النشر التدريجي
		"batch_publish_start":      "🚀 *بدء النشر التدريجي*\n\n⏱️ المدة: *%s*\n📊 العدد: *%d*\n\nسيتم نشر ستوري واحد في كل ضغطة.\nاضغط الزر أدناه لنشر الأول:",
		"batch_publish_next_btn":   "▶️ نشر الستوري التالي",
		"batch_publish_first_btn":  "▶️ بدء النشر (ستوري 1)",
		"batch_publish_success":    "✅ *تم نشر الستوري بنجاح!*\n\n📊 المتبقي: *%d* من *%d*\n⏱️ المدة: *%s*\n\nهل تريد متابعة النشر؟",
		"batch_publish_item_fail":  "❌ فشل نشر الستوري: %v\n\n📊 المتبقي: %d",
		"batch_publish_done":       "🎉 *اكتمل النشر!*\n\n✅ تم نشر *%d* ستوري بنجاح في الدائرة العلوية\n⏱️ المدة: *%s*",
		"batch_publish_stop":       "⏹️ تم إيقاف النشر. القائمة محفوظة.",
	},
	"en": {
		"main_menu_title":          "Main Menu 🤖:",
		"welcome":                  "Welcome to the bot control panel 🤖\nChoose from the buttons below for full control:",
		"stop_btn":                 "🛑 Stop Auto-Reply",
		"start_btn":                "🟢 Start Auto-Reply",
		"edit_text_btn":            "📝 Edit Reply Text",
		"media_reply_btn":          "🎬 Media Auto-Reply",
		"interaction_menu_btn":     "💬 Reply to Interactions",
		"batch_story_btn":          "📚 Publish Multiple Stories",
		"exclude_btn":              "👤 Exclude Account",
		"list_excluded_btn":        "📋 View Excluded",
		"clear_excluded_btn":       "🧹 Clear Excluded",
		"profile_menu_btn":         "🧑 Manage Profile",
		"post_story_btn":           "📖 Post Story",
		"lang_ar_btn":              "🇮🇶 العربية",
		"lang_en_btn":              "🇺🇸 English",
		"back_btn":                 "🔙 Back",
		"stopped_msg":              "🛑 Auto-reply has been stopped.",
		"started_msg":              "🟢 Auto-reply has been started.",
		"edit_text_prompt":         "📝 Send the new auto-reply text now:",
		"saved_text_msg":           "✅ New auto-reply text saved successfully!",
		"exclude_prompt":           "👤 Send the account ID to exclude now:",
		"invalid_id_msg":           "❌ Numbers only! Please send a valid ID.",
		"id_added_msg":             "✅ ID `%d` added to the exclusion list.",
		"list_excluded_title":      "📋 **Excluded Accounts:**\n",
		"no_excluded":              "No excluded accounts currently.",
		"cleared_excluded_msg":     "🧹 All exclusions cleared successfully.",
		"profile_menu_title":       "🧑 Manage Profile - choose what to edit:",
		"edit_first_name_btn":      "✏️ Edit Name",
		"edit_bio_btn":             "📝 Edit Bio",
		"edit_photo_btn":           "🖼️ Edit Photo",
		"edit_username_btn":        "🔗 Edit Username",
		"no_business_connection":   "❌ No business account connected to the bot yet.",
		"first_name_prompt":        "✏️ Send the new first name now:",
		"bio_prompt":               "📝 Send the new bio now (max 70 chars):",
		"username_prompt":          "🔗 Send the new username now (without @):",
		"photo_prompt":             "🖼️ Send the new profile photo now:",
		"name_updated":             "✅ Name updated successfully!",
		"bio_updated":              "✅ Bio updated successfully!",
		"username_updated":         "✅ Username updated successfully!",
		"photo_updated":            "✅ Profile photo updated successfully!",
		"select_story_duration":    "⏱️ Select story duration:",
		"dur_6h":                   "6 Hours",
		"dur_12h":                  "12 Hours",
		"dur_24h":                  "24 Hours",
		"dur_48h":                  "48 Hours",
		"story_prompt":             "📖 Send a photo or video now (max 60 seconds) to post as story (visible for %s):",
		"story_updated":            "✅ Story posted successfully to the top circle! Visible for %s.",
		"your_id_msg":              "Your ID is:\n`%d`",
		"fail_name":                "❌ Failed to update name: %s",
		"fail_bio":                 "❌ Failed to update bio: %s",
		"fail_username":            "❌ Failed to update username: %s",
		"fail_photo":               "❌ Failed to update photo: %s",
		"fail_story":               "❌ Failed to post story: %s",
		"need_real_photo":          "❌ Please send an actual photo.",
		"need_real_media_story":    "❌ Please send an actual photo or video.",
		"video_too_long_error":     "Video is longer than 60 seconds - Telegram's max for stories",
		"id_copy_btn":              "🆔 Copy ID",
		"share_user_btn":           "👤 User",
		"share_user_prompt":        "👇 Use this button to share any user:",
		"user_shared_info":         "👤 *Shared User Info:*\n\nName: %s\nUsername: %s\nID: `%d`",
		"no_username":              "No username",

		"media_menu_title":         "🎬 Choose media type for auto-reply:",
		"media_text_btn":           "📝 Text Only",
		"media_voice_btn":          "🎤 Voice Message",
		"media_audio_btn":          "🎵 Audio File",
		"media_gif_btn":            "🎞️ GIF Animation",
		"media_sticker_btn":        "😀 Sticker",
		"media_video_btn":          "🎥 Video",
		"media_photo_btn":          "🖼️ Photo",
		"media_preview_btn":        "👁️ Preview Current Reply",
		"media_clear_btn":          "🗑️ Clear Media & Back to Text",
		"media_upload_prompt":      "📤 Send the %s you want to use as auto-reply now:",
		"media_saved_msg":          "✅ Auto-reply saved!\nType: *%s*",
		"media_saved_with_caption": "✅ Auto-reply saved!\nType: *%s*\nCaption: %s",
		"media_cleared_msg":        "🗑️ Media cleared, auto-reply is text only.",
		"current_media_info":       "📌 *Current Auto-Reply:*\n\nType: *%s*\nCaption: %s",
		"no_media_set":             "⚠️ No media set, reply is text only.",
		"need_media_error":         "❌ You must send an actual %s!",
		"media_type_voice":         "Voice Message",
		"media_type_audio":         "Audio File",
		"media_type_gif":           "GIF Animation",
		"media_type_sticker":       "Sticker",
		"media_type_video":         "Video",
		"media_type_photo":         "Photo",
		"media_type_text":          "Text Only",
		"no_caption":               "No caption",

		"interaction_menu_title":   "💬 Interaction Reply Settings:",
		"reply_story_mention_btn":  "📸 Reply to Story Mention",
		"reply_reaction_btn":       "👍 Reply to Reaction",
		"reply_voice_btn":          "🎤 Reply to Voice Messages",
		"interaction_on":           "✅ Enabled",
		"interaction_off":          "❌ Disabled",
		"story_mention_prompt":     "📸 Send reply text for story mentions:",
		"reaction_prompt":          "👍 Send reply text for reactions:",
		"voice_reply_prompt":       "🎤 Send reply text for voice messages:",

		"batch_story_title":        "📚 Setup Batch Stories Queue\n\nStories will be published to the top circle one by one:",
		"batch_story_add":          "➕ Add Item",
		"batch_story_list":         "📋 View List",
		"batch_story_clear":        "🗑️ Clear List",
		"batch_story_publish":      "🚀 Start Publishing",
		"batch_story_prompt":       "📤 Send a photo or video to add:",
		"batch_story_added":        "✅ Added! Total: %d items",
		"batch_story_empty":        "⚠️ Queue is empty.",
		"batch_story_list_title":   "📋 *Stories Queue (%d items):*\n\n⏱️ Duration: *%s*\n\n",
		"batch_story_item_line":    "%d. %s\n",
		"batch_story_cleared":      "🗑️ Queue cleared.",
		"batch_story_need_media":   "❌ Send actual photo or video.",
		"batch_story_duration_set": "✅ Duration set: *%s*\nNow add items.",
		"batch_story_choose_duration": "⏱️ Choose duration for multiple stories:",

		"batch_publish_start":      "🚀 *Starting progressive publishing*\n\n⏱️ Duration: *%s*\n📊 Count: *%d*\n\nOne story will be published per click.\nClick the button below to publish the first:",
		"batch_publish_next_btn":   "▶️ Publish Next Story",
		"batch_publish_first_btn":  "▶️ Start (Story 1)",
		"batch_publish_success":    "✅ *Story published successfully!*\n\n📊 Remaining: *%d* of *%d*\n⏱️ Duration: *%s*\n\nContinue publishing?",
		"batch_publish_item_fail":  "❌ Failed to publish: %v\n\n📊 Remaining: %d",
		"batch_publish_done":       "🎉 *Publishing complete!*\n\n✅ Published *%d* stories to the top circle\n⏱️ Duration: *%s*",
		"batch_publish_stop":       "⏹️ Publishing stopped. Queue is saved.",
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

func mediaTypeName(lang, mediaType string) string {
	switch mediaType {
	case "voice":
		return tr(lang, "media_type_voice")
	case "audio":
		return tr(lang, "media_type_audio")
	case "gif":
		return tr(lang, "media_type_gif")
	case "sticker":
		return tr(lang, "media_type_sticker")
	case "video":
		return tr(lang, "media_type_video")
	case "photo":
		return tr(lang, "media_type_photo")
	default:
		return tr(lang, "media_type_text")
	}
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

func toggleLabel(lang string, enabled bool) string {
	if enabled {
		return tr(lang, "interaction_on")
	}
	return tr(lang, "interaction_off")
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

// ============== الهياكل ==============

type BotConfig struct {
	IsStopped      bool    `json:"is_stopped"`
	AutoReply      string  `json:"auto_reply"`
	ReplyType      string  `json:"reply_type"`
	ReplyFileID    string  `json:"reply_file_id"`
	ReplyCaption   string  `json:"reply_caption"`
	Excluded       []int64 `json:"excluded"`
	State          string  `json:"state"`
	BusinessConnID string  `json:"business_conn_id"`
	Lang           string  `json:"lang"`

	ReplyToStoryMention bool   `json:"reply_to_story_mention"`
	ReplyToReaction     bool   `json:"reply_to_reaction"`
	ReplyToVoice        bool   `json:"reply_to_voice"`
	StoryMentionReply   string `json:"story_mention_reply"`
	ReactionReply       string `json:"reaction_reply"`
	VoiceReply          string `json:"voice_reply"`

	BatchStoryDuration string `json:"batch_story_duration"`
}

type StoryItem struct {
	FileID    string `json:"file_id"`
	MediaType string `json:"media_type"`
	Duration  int    `json:"duration"`
	Caption   string `json:"caption"`
}

type TelegramUpdate struct {
	Message            *Message         `json:"message"`
	CallbackQuery      *CallbackQuery   `json:"callback_query"`
	BusinessMessage    *BusinessMessage `json:"business_message"`
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

type BusinessMessage struct {
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

	Story    *Story          `json:"story"`
	Reaction *ReactionUpdate `json:"reaction"`
	Voice    *Voice          `json:"voice"`
	Photo    []PhotoSize     `json:"photo"`
	Video    *Video          `json:"video"`
}

type Story struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	ID int `json:"id"`
}

type ReactionUpdate struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	MessageID int `json:"message_id"`
	User      struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"user"`
	OldReaction []ReactionType `json:"old_reaction"`
	NewReaction []ReactionType `json:"new_reaction"`
}

type ReactionType struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
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

type Voice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}

type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	Title    string `json:"title"`
}

type Animation struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration int    `json:"duration"`
}

type Sticker struct {
	FileID     string `json:"file_id"`
	Emoji      string `json:"emoji"`
	IsAnimated bool   `json:"is_animated"`
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
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
	Caption     string           `json:"caption"`
	Photo       []PhotoSize      `json:"photo"`
	Video       *Video           `json:"video"`
	Voice       *Voice           `json:"voice"`
	Audio       *Audio           `json:"audio"`
	Animation   *Animation       `json:"animation"`
	Sticker     *Sticker         `json:"sticker"`
	Document    *Document        `json:"document"`
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

// ============== Handler الرئيسي ==============

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

	// ============ 1. معالجة الأزرار ============
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
			config.ReplyType = "text"
			config.ReplyFileID = ""
			config.ReplyCaption = ""
			config.State = "waiting_text"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "edit_text_prompt"))

		case "media_menu":
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMediaMenu(botToken, adminID, lang)

		case "media_text":
			config.ReplyType = "text"
			config.ReplyFileID = ""
			config.ReplyCaption = ""
			config.State = "waiting_text"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "edit_text_prompt"))

		case "media_voice", "media_audio", "media_gif", "media_sticker", "media_video", "media_photo":
			mediaType := strings.TrimPrefix(cb.Data, "media_")
			config.State = "waiting_media_" + mediaType
			saveConfig(botToken, adminID, config, msgID)
			typeName := mediaTypeName(lang, mediaType)
			sendSubMenu(botToken, adminID, lang, fmt.Sprintf(tr(lang, "media_upload_prompt"), typeName))

		case "media_preview":
			previewText := tr(lang, "no_media_set")
			if config.ReplyType != "" && config.ReplyType != "text" && config.ReplyFileID != "" {
				caption := config.ReplyCaption
				if caption == "" {
					caption = tr(lang, "no_caption")
				}
				previewText = fmt.Sprintf(tr(lang, "current_media_info"), mediaTypeName(lang, config.ReplyType), caption)
				sendSubMenu(botToken, adminID, lang, previewText)
				sendPreviewMedia(botToken, adminID, config.ReplyType, config.ReplyFileID, config.ReplyCaption)
				w.WriteHeader(http.StatusOK)
				return
			}
			sendSubMenu(botToken, adminID, lang, previewText)

		case "media_clear":
			config.ReplyType = "text"
			config.ReplyFileID = ""
			config.ReplyCaption = ""
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendMenu(botToken, adminID, lang, tr(lang, "media_cleared_msg"))

		case "interaction_menu":
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendInteractionMenu(botToken, adminID, lang, config)

		case "toggle_story_mention":
			config.ReplyToStoryMention = !config.ReplyToStoryMention
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendInteractionMenu(botToken, adminID, lang, config)

		case "toggle_reaction":
			config.ReplyToReaction = !config.ReplyToReaction
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendInteractionMenu(botToken, adminID, lang, config)

		case "toggle_voice":
			config.ReplyToVoice = !config.ReplyToVoice
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendInteractionMenu(botToken, adminID, lang, config)

		case "edit_story_reply":
			config.State = "waiting_story_reply"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "story_mention_prompt"))

		case "edit_reaction_reply":
			config.State = "waiting_reaction_reply"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "reaction_prompt"))

		case "edit_voice_reply":
			config.State = "waiting_voice_reply"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "voice_reply_prompt"))

		case "batch_story_menu":
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			if config.BatchStoryDuration == "" {
				config.BatchStoryDuration = "86400"
				saveConfig(botToken, adminID, config, msgID)
			}
			storyBatchMu.Lock()
			count := len(storyBatch[adminID])
			storyBatchMu.Unlock()
			sendBatchStoryMenu(botToken, adminID, lang, count, config.BatchStoryDuration)

		case "batch_dur_21600", "batch_dur_43200", "batch_dur_86400", "batch_dur_172800":
			period := strings.TrimPrefix(cb.Data, "batch_dur_")
			config.BatchStoryDuration = period
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)

			storyBatchMu.Lock()
			count := len(storyBatch[adminID])
			storyBatchMu.Unlock()

			durationTxt := getDurationLabel(lang, period)
			sendMessage(botToken, adminID, fmt.Sprintf(tr(lang, "batch_story_duration_set"), durationTxt))
			sendBatchStoryMenu(botToken, adminID, lang, count, period)

		case "batch_add":
			if config.BusinessConnID == "" {
				sendMenu(botToken, adminID, lang, tr(lang, "no_business_connection"))
				break
			}
			config.State = "waiting_batch_item"
			saveConfig(botToken, adminID, config, msgID)
			sendSubMenu(botToken, adminID, lang, tr(lang, "batch_story_prompt"))

		case "batch_list":
			storyBatchMu.Lock()
			items := storyBatch[adminID]
			storyBatchMu.Unlock()

			if len(items) == 0 {
				sendBatchStoryMenu(botToken, adminID, lang, 0, config.BatchStoryDuration)
				sendMessage(botToken, adminID, tr(lang, "batch_story_empty"))
				break
			}
			durationTxt := getDurationLabel(lang, config.BatchStoryDuration)
			txt := fmt.Sprintf(tr(lang, "batch_story_list_title"), len(items), durationTxt)
			for i, item := range items {
				mediaIcon := "🖼️"
				if item.MediaType == "video" {
					mediaIcon = "🎥"
				}
				txt += fmt.Sprintf(tr(lang, "batch_story_item_line"), i+1, mediaIcon)
			}
			sendSubMenu(botToken, adminID, lang, txt)

		case "batch_clear":
			storyBatchMu.Lock()
			storyBatch[adminID] = []StoryItem{}
			storyBatchMu.Unlock()
			config.State = ""
			saveConfig(botToken, adminID, config, msgID)
			sendBatchStoryMenu(botToken, adminID, lang, 0, config.BatchStoryDuration)

		// 🚀 بدء النشر التدريجي - يعرض زر "نشر الأول"
		case "batch_publish":
			storyBatchMu.Lock()
			items := storyBatch[adminID]
			storyBatchMu.Unlock()

			if len(items) == 0 {
				sendMessage(botToken, adminID, tr(lang, "batch_story_empty"))
				break
			}

			if config.BusinessConnID == "" {
				sendMessage(botToken, adminID, tr(lang, "no_business_connection"))
				break
			}

			activePeriod := config.BatchStoryDuration
			if activePeriod == "" {
				activePeriod = "86400"
			}

			durationTxt := getDurationLabel(lang, activePeriod)

			// إرسال رسالة البداية مع زر "نشر الأول"
			sendPublishStartMessage(botToken, adminID, lang, len(items), durationTxt)

		// 🚀 نشر ستوري واحد فقط (الزر يُضغط لكل ستوري)
		case "batch_publish_one":
			storyBatchMu.Lock()
			items := storyBatch[adminID]
			storyBatchMu.Unlock()

			if len(items) == 0 {
				sendMessage(botToken, adminID, "🎉 اكتمل النشر!")
				break
			}

			if config.BusinessConnID == "" {
				sendMessage(botToken, adminID, tr(lang, "no_business_connection"))
				break
			}

			activePeriod := config.BatchStoryDuration
			if activePeriod == "" {
				activePeriod = "86400"
			}

			durationTxt := getDurationLabel(lang, activePeriod)

			// نشر الستوري الأول من القائمة (متزامن - طلب واحد فقط)
			firstItem := items[0]

			err := postBusinessStory(
				botToken, config.BusinessConnID,
				firstItem.MediaType, firstItem.FileID,
				firstItem.Duration, activePeriod,
				lang,
			)

			// إزالة العنصر الأول من القائمة
			storyBatchMu.Lock()
			remaining := len(storyBatch[adminID]) - 1
			if remaining > 0 {
				storyBatch[adminID] = storyBatch[adminID][1:]
			} else {
				storyBatch[adminID] = []StoryItem{}
			}
			storyBatchMu.Unlock()

			if err != nil {
				// فشل النشر
				sendPublishFailMessage(botToken, adminID, lang, err.Error(), remaining)
			} else {
				// نجح النشر
				if remaining > 0 {
					sendPublishSuccessMessage(botToken, adminID, lang, remaining, len(items), durationTxt)
				} else {
					// اكتمل النشر
					sendMessage(botToken, adminID, fmt.Sprintf(
						tr(lang, "batch_publish_done"),
						len(items), durationTxt,
					))
					// مسح القائمة
					storyBatchMu.Lock()
					storyBatch[adminID] = []StoryItem{}
					storyBatchMu.Unlock()
					// العودة للقائمة
					sendBatchStoryMenu(botToken, adminID, lang, 0, activePeriod)
				}
			}

		case "batch_publish_stop":
			sendMessage(botToken, adminID, tr(lang, "batch_publish_stop"))
			sendBatchStoryMenu(botToken, adminID, lang, 0, config.BatchStoryDuration)

		// باقي الأزرار
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

	// ============ 2. معالجة الرسائل الخاصة ============
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

		// معالجة حالات الوسائط
		if strings.HasPrefix(config.State, "waiting_media_") {
			mediaType := strings.TrimPrefix(config.State, "waiting_media_")
			var fileID string
			var ok bool

			switch mediaType {
			case "voice":
				if msg.Voice != nil {
					fileID, ok = msg.Voice.FileID, true
				}
			case "audio":
				if msg.Audio != nil {
					fileID, ok = msg.Audio.FileID, true
				}
			case "gif":
				if msg.Animation != nil {
					fileID, ok = msg.Animation.FileID, true
				} else if msg.Document != nil && msg.Document.MimeType == "video/mp4" {
					fileID, ok = msg.Document.FileID, true
				}
			case "sticker":
				if msg.Sticker != nil {
					fileID, ok = msg.Sticker.FileID, true
				}
			case "video":
				if msg.Video != nil {
					fileID, ok = msg.Video.FileID, true
				}
			case "photo":
				if len(msg.Photo) > 0 {
					fileID, ok = msg.Photo[len(msg.Photo)-1].FileID, true
				}
			}

			if !ok {
				typeName := mediaTypeName(lang, mediaType)
				sendSubMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "need_media_error"), typeName))
				w.WriteHeader(http.StatusOK)
				return
			}

			config.ReplyType = mediaType
			config.ReplyFileID = fileID
			config.ReplyCaption = msg.Caption
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)

			typeName := mediaTypeName(lang, mediaType)
			if msg.Caption != "" {
				sendMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "media_saved_with_caption"), typeName, msg.Caption))
			} else {
				sendMenu(botToken, chatID, lang, fmt.Sprintf(tr(lang, "media_saved_msg"), typeName))
			}

			sendPreviewMedia(botToken, chatID, mediaType, fileID, msg.Caption)
			w.WriteHeader(http.StatusOK)
			return
		}

		// حالات الردود المخصصة للتفاعلات
		if config.State == "waiting_story_reply" {
			config.StoryMentionReply = msg.Text
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)
			sendInteractionMenu(botToken, chatID, lang, config)
		} else if config.State == "waiting_reaction_reply" {
			config.ReactionReply = msg.Text
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)
			sendInteractionMenu(botToken, chatID, lang, config)
		} else if config.State == "waiting_voice_reply" {
			config.VoiceReply = msg.Text
			config.State = ""
			saveConfig(botToken, chatID, config, msgID)
			sendInteractionMenu(botToken, chatID, lang, config)

		// إضافة عنصر للستوريات
		} else if config.State == "waiting_batch_item" {
			var fileID, mediaType string
			var duration int

			if msg.Video != nil {
				fileID = msg.Video.FileID
				mediaType = "video"
				duration = msg.Video.Duration
			} else if len(msg.Photo) > 0 {
				fileID = msg.Photo[len(msg.Photo)-1].FileID
				mediaType = "photo"
			} else {
				sendSubMenu(botToken, chatID, lang, tr(lang, "batch_story_need_media"))
				w.WriteHeader(http.StatusOK)
				return
			}

			newItem := StoryItem{
				FileID:    fileID,
				MediaType: mediaType,
				Duration:  duration,
				Caption:   msg.Caption,
			}

			storyBatchMu.Lock()
			if storyBatch[chatID] == nil {
				storyBatch[chatID] = []StoryItem{}
			}
			storyBatch[chatID] = append(storyBatch[chatID], newItem)
			count := len(storyBatch[chatID])
			storyBatchMu.Unlock()

			config.State = ""
			saveConfig(botToken, chatID, config, msgID)

			sendMessage(botToken, chatID, fmt.Sprintf(tr(lang, "batch_story_added"), count))
			sendBatchStoryMenu(botToken, chatID, lang, count, config.BatchStoryDuration)

		// الحالات القديمة
		} else if config.State == "waiting_text" {
			config.AutoReply = msg.Text
			config.ReplyType = "text"
			config.ReplyFileID = ""
			config.ReplyCaption = ""
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
				sendSubMenu(botToken, chatID, lang, "❌ النبذة طويلة جداً! الحد الأقصى 70 حرفاً.")
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

	// ============ 3. رسائل العملاء ============
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

		customerName := msg.From.FirstName
		if customerName == "" {
			customerName = "صديقي"
		}

		if msg.Story != nil && config.ReplyToStoryMention {
			handleStoryMention(botToken, adminID, config, customerChatID, senderID, customerName, msg)
			w.WriteHeader(http.StatusOK)
			return
		}

		if msg.Reaction != nil && config.ReplyToReaction {
			handleReaction(botToken, adminID, config, customerChatID, senderID, customerName, msg)
			w.WriteHeader(http.StatusOK)
			return
		}

		if msg.Voice != nil && config.ReplyToVoice {
			handleVoiceMessage(botToken, adminID, config, customerChatID, senderID, customerName, msg)
			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.TrimSpace(msg.Text) == "بوت" || strings.Contains(msg.Text, "بوت") {
			sendNerdBotInfoBusiness(botToken, customerChatID, msg.BusinessConnectionID)
			w.WriteHeader(http.StatusOK)
			return
		}

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

		if config.ReplyType != "" && config.ReplyType != "text" && config.ReplyFileID != "" {
			caption := config.ReplyCaption
			if caption != "" {
				caption = strings.ReplaceAll(caption, "{name}", customerName)
				caption = strings.ReplaceAll(caption, "{الاسم}", customerName)
				caption = strings.ReplaceAll(caption, "$name", customerName)

				if detectedLang != "" && detectedLang != "ar" {
					if translatedCaption, _, err := translateText(caption, detectedLang); err == nil && translatedCaption != "" {
						caption = translatedCaption
					}
				}
			}

			if err := sendBusinessMediaReply(botToken, customerChatID, config.ReplyType, config.ReplyFileID, caption, msg.BusinessConnectionID); err != nil {
				log.Println("خطأ إرسال وسائط الرد التلقائي:", err)
			}
			w.WriteHeader(http.StatusOK)
			return
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

	// ============ 4. ربط الحساب التجاري ============
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

// ============== معالجات التفاعلات ==============

func handleStoryMention(token string, adminID int64, config BotConfig, customerChatID, senderID int64, customerName string, msg *BusinessMessage) {
	notifyText := fmt.Sprintf(
		"📸 *ذكر جديد في ستوري!*\n\n👤 العميل: %s\n🆔 `%d`",
		customerName, senderID,
	)
	sendMessage(token, adminID, notifyText)

	replyText := config.StoryMentionReply
	if replyText == "" {
		replyText = "شكراً لذكرنا في ستوريك يا " + customerName + " 🌸"
	}
	replyText = strings.ReplaceAll(replyText, "{name}", customerName)
	replyText = strings.ReplaceAll(replyText, "{الاسم}", customerName)
	replyText = strings.ReplaceAll(replyText, "$name", customerName)

	sendBusinessMessage(token, customerChatID, replyText, msg.BusinessConnectionID)
}

func handleReaction(token string, adminID int64, config BotConfig, customerChatID, senderID int64, customerName string, msg *BusinessMessage) {
	emoji := ""
	if len(msg.Reaction.NewReaction) > 0 {
		emoji = msg.Reaction.NewReaction[0].Emoji
	}
	if emoji == "" {
		return
	}

	notifyText := fmt.Sprintf(
		"👍 *تفاعل جديد!*\n\n👤 العميل: %s\n🆔 `%d`\n😀 التفاعل: %s",
		customerName, senderID, emoji,
	)
	sendMessage(token, adminID, notifyText)

	replyText := config.ReactionReply
	if replyText == "" {
		replyText = "شكراً لتفاعلك يا " + customerName + " 🌸"
	}
	replyText = strings.ReplaceAll(replyText, "{name}", customerName)
	replyText = strings.ReplaceAll(replyText, "{emoji}", emoji)
	replyText = strings.ReplaceAll(replyText, "{الاسم}", customerName)
	replyText = strings.ReplaceAll(replyText, "$name", customerName)

	sendBusinessMessage(token, customerChatID, replyText, msg.BusinessConnectionID)
}

func handleVoiceMessage(token string, adminID int64, config BotConfig, customerChatID, senderID int64, customerName string, msg *BusinessMessage) {
	notifyText := fmt.Sprintf(
		"🎤 *رسالة صوتية جديدة!*\n\n👤 العميل: %s\n🆔 `%d`\n⏱️ المدة: %d ثانية",
		customerName, senderID, msg.Voice.Duration,
	)
	sendMessage(token, adminID, notifyText)

	replyText := config.VoiceReply
	if replyText == "" {
		replyText = "شكراً لرسالتك الصوتية يا " + customerName + " 🌸\nسأستمع لها وأرد عليك قريباً."
	}
	replyText = strings.ReplaceAll(replyText, "{name}", customerName)
	replyText = strings.ReplaceAll(replyText, "{الاسم}", customerName)
	replyText = strings.ReplaceAll(replyText, "$name", customerName)

	sendBusinessMessage(token, customerChatID, replyText, msg.BusinessConnectionID)
}

// ============== دوال مساعدة ==============

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

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getBusinessConnection?business_connection_id=%s", token, connID)
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Println("خطأ getBusinessConnection:", err)
		return 0
	}
	defer resp.Body.Close()

	var res BusinessConnectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Println("خطأ فك تشفير getBusinessConnection:", err)
		return 0
	}

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
		IsStopped:          false,
		AutoReply:          "",
		ReplyType:          "text",
		ReplyFileID:        "",
		ReplyCaption:       "",
		Excluded:           []int64{},
		State:              "",
		BusinessConnID:     "",
		Lang:               "ar",
		BatchStoryDuration: "86400",
	}

	if chatID == 0 {
		return defaultCfg, 0
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%d", token, chatID)
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Println("خطأ getChat:", err)
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
		log.Println("خطأ فك تشفير getChat:", err)
		return defaultCfg, 0
	}

	if res.Result.PinnedMessage.MessageID != 0 {
		var cfg BotConfig
		if err := json.Unmarshal([]byte(res.Result.PinnedMessage.Text), &cfg); err == nil {
			if cfg.Lang == "" {
				cfg.Lang = "ar"
			}
			if cfg.ReplyType == "" {
				cfg.ReplyType = "text"
			}
			if cfg.BatchStoryDuration == "" {
				cfg.BatchStoryDuration = "86400"
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
		if _, err := httpClient.Post(url, "application/json", bytes.NewBuffer(pBytes)); err != nil {
			log.Println("خطأ editMessageText:", err)
		}
	} else {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
		payload := map[string]interface{}{
			"chat_id": chatID,
			"text":    cfgText,
		}
		pBytes, _ := json.Marshal(payload)
		resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(pBytes))
		if err != nil {
			log.Println("خطأ sendMessage (saveConfig):", err)
			return
		}
		defer resp.Body.Close()
		var res struct {
			Result struct {
				MessageID int `json:"message_id"`
			} `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			log.Println("خطأ فك تشفير sendMessage:", err)
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
			if _, err := httpClient.Post(pinUrl, "application/json", bytes.NewBuffer(pPinBytes)); err != nil {
				log.Println("خطأ pinChatMessage:", err)
			}
		}
	}
}

// ============== إرسال الرسائل والقوائم ==============

func sendNerdBotInfo(token string, chatID int64) {
	text := "انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز"
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "فعلني من هنا", "url": "https://t.me/Xhwe2/10", "style": "success"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": keyboard}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendNerdBotInfoBusiness(token string, chatID int64, bizID string) {
	text := "انا اسمي نيرد | Nerd من خلالي رح تقدر تنشر ستوريات غير محدودة بدون اشتراك مميز"
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "فعلني من هنا", "url": "https://t.me/Xhwe2/10", "style": "success"}},
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

func sendStartPhoto(token string, chatID int64, lang string) {
	payload := map[string]interface{}{"chat_id": chatID, "photo": startPhotoURL, "caption": tr(lang, "welcome")}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendPhoto", "application/json", bytes.NewBuffer(b))
}

func sendUserShareKeyboard(token string, chatID int64, lang string) {
	keyboard := map[string]interface{}{
		"keyboard": [][]map[string]interface{}{
			{{
				"text": tr(lang, "share_user_btn"),
				"request_users": map[string]interface{}{
					"request_id":       1,
					"request_name":     true,
					"request_username": true,
				},
				"style": "success",
			}},
		},
		"resize_keyboard": true,
		"is_persistent":   true,
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "share_user_prompt"),
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendMenu(token string, chatID int64, lang, text string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": tr(lang, "stop_btn"), "callback_data": "stop", "style": "danger"},
				{"text": tr(lang, "start_btn"), "callback_data": "start", "style": "success"},
			},
			{{"text": tr(lang, "edit_text_btn"), "callback_data": "edit_text", "style": "primary"}},
			{{"text": tr(lang, "media_reply_btn"), "callback_data": "media_menu", "style": "success"}},
			{{"text": tr(lang, "interaction_menu_btn"), "callback_data": "interaction_menu", "style": "success"}},
			{{"text": tr(lang, "batch_story_btn"), "callback_data": "batch_story_menu", "style": "success"}},
			{
				{"text": tr(lang, "exclude_btn"), "callback_data": "exclude", "style": "primary"},
				{"text": tr(lang, "list_excluded_btn"), "callback_data": "list_excluded", "style": "primary"},
			},
			{{"text": tr(lang, "clear_excluded_btn"), "callback_data": "clear_excluded", "style": "danger"}},
			{{"text": tr(lang, "profile_menu_btn"), "callback_data": "profile_menu", "style": "primary"}},
			{{"text": tr(lang, "post_story_btn"), "callback_data": "post_story", "style": "primary"}},
			{{
				"text": fmt.Sprintf("%s (%d)", tr(lang, "id_copy_btn"), chatID),
				"copy_text": map[string]interface{}{"text": fmt.Sprintf("%d", chatID)},
				"style":     "primary",
			}},
			{
				{"text": tr(lang, "lang_ar_btn"), "callback_data": "lang_ar", "style": "primary"},
				{"text": tr(lang, "lang_en_btn"), "callback_data": "lang_en", "style": "primary"},
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

func sendMediaMenu(token string, chatID int64, lang string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": tr(lang, "media_text_btn"), "callback_data": "media_text", "style": "primary"}},
			{
				{"text": tr(lang, "media_voice_btn"), "callback_data": "media_voice", "style": "primary"},
				{"text": tr(lang, "media_audio_btn"), "callback_data": "media_audio", "style": "primary"},
			},
			{
				{"text": tr(lang, "media_gif_btn"), "callback_data": "media_gif", "style": "primary"},
				{"text": tr(lang, "media_sticker_btn"), "callback_data": "media_sticker", "style": "primary"},
			},
			{
				{"text": tr(lang, "media_video_btn"), "callback_data": "media_video", "style": "primary"},
				{"text": tr(lang, "media_photo_btn"), "callback_data": "media_photo", "style": "primary"},
			},
			{{"text": tr(lang, "media_preview_btn"), "callback_data": "media_preview", "style": "success"}},
			{{"text": tr(lang, "media_clear_btn"), "callback_data": "media_clear", "style": "danger"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "media_menu_title"),
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendInteractionMenu(token string, chatID int64, lang string, config BotConfig) {
	storyBtn := tr(lang, "reply_story_mention_btn") + " " + toggleLabel(lang, config.ReplyToStoryMention)
	reactionBtn := tr(lang, "reply_reaction_btn") + " " + toggleLabel(lang, config.ReplyToReaction)
	voiceBtn := tr(lang, "reply_voice_btn") + " " + toggleLabel(lang, config.ReplyToVoice)

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": storyBtn, "callback_data": "toggle_story_mention", "style": "primary"}},
			{{"text": reactionBtn, "callback_data": "toggle_reaction", "style": "primary"}},
			{{"text": voiceBtn, "callback_data": "toggle_voice", "style": "primary"}},
			{{"text": "✏️ " + tr(lang, "story_mention_prompt"), "callback_data": "edit_story_reply", "style": "success"}},
			{{"text": "✏️ " + tr(lang, "reaction_prompt"), "callback_data": "edit_reaction_reply", "style": "success"}},
			{{"text": "✏️ " + tr(lang, "voice_reply_prompt"), "callback_data": "edit_voice_reply", "style": "success"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         tr(lang, "interaction_menu_title"),
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendBatchStoryMenu(token string, chatID int64, lang string, count int, duration string) {
	durationTxt := getDurationLabel(lang, duration)
	title := tr(lang, "batch_story_title") + "\n\n⏱️ *" + durationTxt + "*"

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "⏱️ 6h", "callback_data": "batch_dur_21600", "style": "primary"},
				{"text": "⏱️ 12h", "callback_data": "batch_dur_43200", "style": "primary"},
			},
			{
				{"text": "⏱️ 24h", "callback_data": "batch_dur_86400", "style": "primary"},
				{"text": "⏱️ 48h", "callback_data": "batch_dur_172800", "style": "primary"},
			},
			{{"text": tr(lang, "batch_story_add"), "callback_data": "batch_add", "style": "primary"}},
			{{"text": fmt.Sprintf("%s (%d)", tr(lang, "batch_story_list"), count), "callback_data": "batch_list", "style": "primary"}},
			{{"text": tr(lang, "batch_story_publish"), "callback_data": "batch_publish", "style": "success"}},
			{{"text": tr(lang, "batch_story_clear"), "callback_data": "batch_clear", "style": "danger"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
		},
	}
	payload := map[string]interface{}{
		"chat_id":      chatID,
		"text":         title,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

// 🆕 رسالة بداية النشر التدريجي
func sendPublishStartMessage(token string, chatID int64, lang string, total int, durationTxt string) {
	text := fmt.Sprintf(tr(lang, "batch_publish_start"), durationTxt, total)

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": tr(lang, "batch_publish_first_btn"), "callback_data": "batch_publish_one", "style": "success"}},
			{{"text": tr(lang, "batch_publish_stop"), "callback_data": "batch_publish_stop", "style": "danger"}},
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

// 🆕 رسالة نجاح نشر ستوري
func sendPublishSuccessMessage(token string, chatID int64, lang string, remaining, total int, durationTxt string) {
	text := fmt.Sprintf(tr(lang, "batch_publish_success"), remaining, total, durationTxt)

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": fmt.Sprintf("%s (%d متبقي)", tr(lang, "batch_publish_next_btn"), remaining), "callback_data": "batch_publish_one", "style": "success"}},
			{{"text": tr(lang, "batch_publish_stop"), "callback_data": "batch_publish_stop", "style": "danger"}},
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

// 🆕 رسالة فشل نشر ستوري
func sendPublishFailMessage(token string, chatID int64, lang string, errMsg string, remaining int) {
	text := fmt.Sprintf(tr(lang, "batch_publish_item_fail"), errMsg, remaining)

	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "🔄 إعادة المحاولة", "callback_data": "batch_publish_one", "style": "success"}},
			{{"text": tr(lang, "batch_publish_stop"), "callback_data": "batch_publish_stop", "style": "danger"}},
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

func sendPreviewMedia(token string, chatID int64, mediaType, fileID, caption string) {
	var method, field string
	switch mediaType {
	case "voice":
		method, field = "sendVoice", "voice"
	case "audio":
		method, field = "sendAudio", "audio"
	case "gif":
		method, field = "sendAnimation", "animation"
	case "sticker":
		method, field = "sendSticker", "sticker"
	case "video":
		method, field = "sendVideo", "video"
	case "photo":
		method, field = "sendPhoto", "photo"
	default:
		return
	}
	payload := map[string]interface{}{"chat_id": chatID, field: fileID}
	if mediaType != "sticker" && caption != "" {
		payload["caption"] = caption
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/"+method, "application/json", bytes.NewBuffer(b))
}

func sendBusinessMediaReply(token string, chatID int64, mediaType, fileID, caption, bizID string) error {
	var method, field string
	switch mediaType {
	case "voice":
		method, field = "sendVoice", "voice"
	case "audio":
		method, field = "sendAudio", "audio"
	case "gif":
		method, field = "sendAnimation", "animation"
	case "sticker":
		method, field = "sendSticker", "sticker"
	case "video":
		method, field = "sendVideo", "video"
	case "photo":
		method, field = "sendPhoto", "photo"
	default:
		return fmt.Errorf("نوع وسائط غير مدعوم: %s", mediaType)
	}
	payload := map[string]interface{}{
		"chat_id":                chatID,
		"business_connection_id": bizID,
		field:                    fileID,
	}
	if mediaType != "sticker" && caption != "" {
		payload["caption"] = caption
		payload["parse_mode"] = "Markdown"
	}
	b, _ := json.Marshal(payload)
	resp, err := mediaClient.Post("https://api.telegram.org/bot"+token+"/"+method, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res apiResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	if !res.Ok {
		return fmt.Errorf(res.Description)
	}
	return nil
}

func sendBusinessMessage(token string, chatID int64, text, bizID string) {
	payload := map[string]interface{}{
		"chat_id":                chatID,
		"text":                   text,
		"business_connection_id": bizID,
	}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/sendMessage", "application/json", bytes.NewBuffer(b))
}

func sendStoryDurationMenu(token string, chatID int64, lang string) {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "⏱️ " + tr(lang, "dur_6h"), "callback_data": "story_dur_21600", "style": "primary"},
				{"text": "⏱️ " + tr(lang, "dur_12h"), "callback_data": "story_dur_43200", "style": "primary"},
			},
			{
				{"text": "⏱️ " + tr(lang, "dur_24h"), "callback_data": "story_dur_86400", "style": "primary"},
				{"text": "⏱️ " + tr(lang, "dur_48h"), "callback_data": "story_dur_172800", "style": "primary"},
			},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
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
			{{"text": tr(lang, "edit_first_name_btn"), "callback_data": "edit_first_name", "style": "primary"}},
			{{"text": tr(lang, "edit_bio_btn"), "callback_data": "edit_bio", "style": "primary"}},
			{{"text": tr(lang, "edit_photo_btn"), "callback_data": "edit_photo", "style": "primary"}},
			{{"text": tr(lang, "edit_username_btn"), "callback_data": "edit_username", "style": "primary"}},
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
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
			{{"text": tr(lang, "back_btn"), "callback_data": "main_menu", "style": "danger"}},
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
			{{"text": "✨ " + initialQuote, "callback_data": "change_quote", "style": "primary"}},
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
			{{"text": "✨ " + newQuote, "callback_data": "change_quote", "style": "primary"}},
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

// ============== أدوات الملفات و الرفع ==============

func downloadTelegramFile(token, fileID string) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", token, fileID)
	resp, err := mediaClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("تعذر الاتصال بتليجرام")
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
		return nil, fmt.Errorf("رد غير متوقع")
	}
	if !res.Ok || res.Result.FilePath == "" {
		return nil, fmt.Errorf(res.Description)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, res.Result.FilePath)
	fResp, err := mediaClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("تعذر تنزيل الملف")
	}
	defer fResp.Body.Close()

	data, err := io.ReadAll(fResp.Body)
	if err != nil {
		return nil, fmt.Errorf("تعذر قراءة البيانات")
	}
	return data, nil
}

func postMultipartBusinessAPI(token, method string, fields map[string]string, fileFieldName, fileName string, fileBytes []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return fmt.Errorf("خطأ داخلي")
		}
	}

	part, err := writer.CreateFormFile(fileFieldName, fileName)
	if err != nil {
		return fmt.Errorf("خطأ داخلي في الملف")
	}
	if _, err := part.Write(fileBytes); err != nil {
		return fmt.Errorf("خطأ كتابة الملف")
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("خطأ إغلاق الطلب")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("تعذر تجهيز الطلب")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := mediaClient.Do(req)
	if err != nil {
		return fmt.Errorf("تعذر الاتصال بتليجرام")
	}
	defer resp.Body.Close()

	var res apiResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("رد غير متوقع")
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
		return fmt.Errorf("رد غير متوقع")
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

// 🆕 نشر ستوري واحد في الدائرة العلوية (بدون post_to_chat_page)
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
	payload := map[string]interface{}{"chat_id": chatID, "message_id": msgID}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/deleteMessage", "application/json", bytes.NewBuffer(b))
}

func answerCallback(token, callbackID string) {
	payload := map[string]string{"callback_query_id": callbackID}
	b, _ := json.Marshal(payload)
	httpClient.Post("https://api.telegram.org/bot"+token+"/answerCallbackQuery", "application/json", bytes.NewBuffer(b))
}
