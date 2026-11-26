package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultOpenRouterModel = "openrouter/free"
)

var aiHTTPClient = &http.Client{Timeout: 25 * time.Second}

type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
	MaxTokens   int                 `json:"max_tokens"`
}

type openRouterMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type openRouterResponse struct {
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Choices []struct {
		Message openRouterMessage `json:"message"`
	} `json:"choices"`
}

func generateSmartReply(customerMessage string) (string, error) {
	return generateSmartReplyWithContent(strings.TrimSpace(customerMessage))
}

const maxSmartReplyMediaBytes = 8 << 20

func hasSmartReplyPayload(msg *BusinessMessage) bool {
	return msg != nil && (strings.TrimSpace(msg.Text) != "" || strings.TrimSpace(msg.Caption) != "" ||
		len(msg.Photo) > 0 || msg.Video != nil || msg.Voice != nil || msg.Audio != nil ||
		msg.Document != nil || msg.Animation != nil || msg.Sticker != nil)
}

func telegramFileBytes(botToken, fileID string) ([]byte, string, error) {
	if strings.TrimSpace(fileID) == "" {
		return nil, "", fmt.Errorf("معرّف الملف غير متوفر")
	}
	fileReq, err := http.NewRequest(http.MethodGet, "https://api.telegram.org/bot"+botToken+"/getFile?file_id="+url.QueryEscape(fileID), nil)
	if err != nil { return nil, "", fmt.Errorf("تعذر تجهيز طلب تنزيل الملف") }
	fileResp, err := aiHTTPClient.Do(fileReq)
	if err != nil { return nil, "", fmt.Errorf("تعذر الاتصال بتليجرام لتنزيل الملف") }
	defer fileResp.Body.Close()
	var fileResult struct {
		Ok bool `json:"ok"`
		Description string `json:"description"`
		Result struct { FilePath string `json:"file_path"` } `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(fileResp.Body, 1<<20)).Decode(&fileResult); err != nil || !fileResult.Ok || fileResult.Result.FilePath == "" {
		if fileResult.Description != "" { return nil, "", fmt.Errorf("تعذر الحصول على الملف من Telegram: %s", fileResult.Description) }
		return nil, "", fmt.Errorf("تعذر قراءة بيانات الملف من Telegram")
	}
	downloadURL := "https://api.telegram.org/file/bot"+botToken+"/"+fileResult.Result.FilePath
	downloadReq, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil { return nil, "", fmt.Errorf("تعذر تجهيز تنزيل الملف") }
	downloadResp, err := aiHTTPClient.Do(downloadReq)
	if err != nil { return nil, "", fmt.Errorf("تعذر تنزيل الملف من Telegram") }
	defer downloadResp.Body.Close()
	if downloadResp.StatusCode < 200 || downloadResp.StatusCode >= 300 { return nil, "", fmt.Errorf("فشل تنزيل الملف من Telegram (HTTP %d)", downloadResp.StatusCode) }
	data, err := io.ReadAll(io.LimitReader(downloadResp.Body, maxSmartReplyMediaBytes+1))
	if err != nil { return nil, "", fmt.Errorf("تعذر قراءة الملف") }
	if len(data) > maxSmartReplyMediaBytes { return nil, "", fmt.Errorf("حجم الملف أكبر من الحد المدعوم حالياً (8 ميغابايت)") }
	return data, fileResult.Result.FilePath, nil
}

func smartReplyDataURI(data []byte, contentType string) string {
	if contentType == "" { contentType = "application/octet-stream" }
	return "data:"+contentType+";base64,"+base64.StdEncoding.EncodeToString(data)
}

func buildSmartReplyInput(botToken string, msg *BusinessMessage) (interface{}, error) {
	parts := make([]interface{}, 0, 4)
	textParts := make([]string, 0, 2)
	if strings.TrimSpace(msg.Text) != "" { textParts = append(textParts, strings.TrimSpace(msg.Text)) }
	if strings.TrimSpace(msg.Caption) != "" { textParts = append(textParts, "نص مرفق بالوسائط: "+strings.TrimSpace(msg.Caption)) }

	addImage := func(fileID, mimeType, label string) error {
		data, _, err := telegramFileBytes(botToken, fileID)
		if err != nil { return err }
		if mimeType == "" { mimeType = "image/jpeg" }
		parts = append(parts, map[string]interface{}{"type":"image_url","image_url":map[string]interface{}{"url":smartReplyDataURI(data,mimeType)}})
		textParts = append(textParts, label)
		return nil
	}

	switch {
	case len(msg.Photo) > 0:
		photo := msg.Photo[len(msg.Photo)-1]
		if err := addImage(photo.FileID, "image/jpeg", "أرسل العميل صورة. افحص الصورة بصرياً وأجب عن سؤاله."); err != nil { return nil, err }
	case msg.Sticker != nil:
		if msg.Sticker.IsAnimated { textParts = append(textParts, "أرسل العميل ملصقاً متحركاً؛ لا يمكن فك حركته في هذه الاستضافة. رمز الملصق: "+msg.Sticker.Emoji) } else if err := addImage(msg.Sticker.FileID, "image/webp", "أرسل العميل ملصقاً ثابتاً. افهم الصورة إن أمكن."); err != nil { return nil, err }
	case msg.Voice != nil || msg.Audio != nil:
		fileID, format := "", "ogg"
		if msg.Voice != nil { fileID = msg.Voice.FileID } else {
			fileID = msg.Audio.FileID
			if strings.Contains(strings.ToLower(msg.Audio.Title), ".") { format = strings.TrimPrefix(strings.ToLower(msg.Audio.Title[strings.LastIndex(msg.Audio.Title, "."):]), ".") }
		}
		data, filePath, err := telegramFileBytes(botToken, fileID)
		if err != nil { return nil, err }
		if ext := strings.ToLower(filePath[strings.LastIndex(filePath, ".")+1:]); ext != "" { format = ext }
		parts = append(parts, map[string]interface{}{"type":"input_audio","input_audio":map[string]interface{}{"data":base64.StdEncoding.EncodeToString(data),"format":format}})
		textParts = append(textParts, "أرسل العميل رسالة صوتية/ملفاً صوتياً. استمع إلى محتواه، واستخرج الكلام والمعنى ثم أجب بلغة مناسبة.")
	case msg.Video != nil:
		data, filePath, err := telegramFileBytes(botToken, msg.Video.FileID)
		if err != nil { return nil, err }
		mimeType := "video/mp4"
		if strings.HasSuffix(strings.ToLower(filePath), ".mov") { mimeType = "video/quicktime" }
		parts = append(parts, map[string]interface{}{"type":"video_url","video_url":map[string]interface{}{"url":smartReplyDataURI(data,mimeType)}})
		textParts = append(textParts, "أرسل العميل فيديو. حلل محتواه المرئي والصوتي إذا كان النموذج يدعم الفيديو.")
	case msg.Document != nil:
		data, filePath, err := telegramFileBytes(botToken, msg.Document.FileID)
		if err != nil { return nil, err }
		ext := strings.ToLower(filePath)
		if strings.HasSuffix(ext, ".txt") || strings.HasSuffix(ext, ".csv") || strings.HasSuffix(ext, ".json") || strings.HasSuffix(ext, ".md") || strings.HasSuffix(ext, ".log") {
			if !utf8.Valid(data) { return nil, fmt.Errorf("المستند النصي ليس بترميز UTF-8 مدعوم") }
			content := string(data)
			if len([]rune(content)) > 12000 { content = string([]rune(content)[:12000]) }
			textParts = append(textParts, "محتوى المستند "+filepath.Base(filePath)+":\n"+content)
		} else {
			return nil, fmt.Errorf("نوع المستند %s غير مدعوم للقراءة حالياً؛ المدعوم نصوص TXT وCSV وJSON وMD وLOG", filepath.Base(filePath))
		}
	case msg.Animation != nil:
		return nil, fmt.Errorf("ملفات GIF المتحركة غير مدعومة للتحليل حالياً؛ أرسلها كفيديو أو صورة ثابتة")
	}
	if len(textParts) == 0 { textParts = append(textParts, "أرسل العميل رسالة لا تحتوي نصاً قابلاً للاستخراج. اطلب منه توضيح طلبه بلطف.") }
	parts = append([]interface{}{map[string]interface{}{"type":"text","text":strings.Join(textParts,"\n\n")}}, parts...)
	return parts, nil
}

func generateSmartReplyWithContent(customerContent interface{}) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		return "", fmt.Errorf("مفتاح OpenRouter غير مضبوط في متغيرات Vercel: OPENROUTER_API_KEY")
	}
		model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL"))
	if model == "" {
		model = defaultOpenRouterModel
	}

	requestBody := openRouterRequest{
		Model: model,
		Temperature: 0.65,
		MaxTokens: 350,
		Messages: []openRouterMessage{
			{
				Role: "system",
				Content: "أنت مساعد يكتب رداً قصيراً وطبيعياً نيابة عن صاحب حساب Telegram Business. افهم النصوص والصور والصوت والفيديو المرفق إن كان النموذج يدعمها. اكتب بلغة العميل، وإذا كانت الرسالة بالعربية فاستخدم عربية طبيعية ولهجة عراقية خفيفة عند ملاءمتها. أخرج نص الرد فقط دون عناوين أو علامات اقتباس. لا تخترع أسعاراً أو وعوداً أو معلومات غير معروفة، ولا تدّعِ أنك نفذت إجراءً. إذا لم تتمكن من فهم الوسائط، اطلب توضيحاً بلطف. إذا كانت الرسالة شكوى أو طلباً حساساً، اقترح رداً مهذباً يوضح أن صاحب الحساب سيتابع الأمر.",
			},
			{Role: "user", Content: customerContent},
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("تعذر تجهيز طلب الذكاء الاصطناعي")
	}
	req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("تعذر تجهيز اتصال OpenRouter")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "Nerd Telegram Business Bot")

	var resp *http.Response
	for attempt := 0; attempt < 2; attempt++ {
		resp, err = aiHTTPClient.Do(req)
		if err != nil {
			if attempt == 0 && req.GetBody != nil {
				req.Body, _ = req.GetBody()
				time.Sleep(300 * time.Millisecond)
				continue
			}
			return "", fmt.Errorf("تعذر الاتصال بـ OpenRouter بعد محاولتين: %v", err)
		}
		// أعد المحاولة مرة واحدة عند حدود الاستخدام أو أخطاء الخادم المؤقتة.
		if attempt == 0 && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError) {
			resp.Body.Close()
			if req.GetBody != nil {
				req.Body, _ = req.GetBody()
			}
			time.Sleep(400 * time.Millisecond)
			continue
		}
		break
	}
	defer resp.Body.Close()

	var result openRouterResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("تعذر قراءة استجابة OpenRouter: %v", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := "استجابة غير ناجحة"
		if result.Error != nil && result.Error.Message != "" {
			message = result.Error.Message
		}
		return "", fmt.Errorf("فشل OpenRouter (HTTP %d): %s", resp.StatusCode, message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("لم يُرجع نموذج الذكاء الاصطناعي أي اقتراح")
	}
	replyText, ok := result.Choices[0].Message.Content.(string)
	if !ok { return "", fmt.Errorf("استجابة النموذج ليست نصاً") }
	reply := strings.TrimSpace(replyText)
	if reply == "" {
		return "", fmt.Errorf("عاد اقتراح الذكاء الاصطناعي فارغاً")
	}
	if len([]rune(reply)) > 2500 {
		reply = string([]rune(reply)[:2500])
	}
	return reply, nil
}

func forcedSubscriptionChannel() string {
	return strings.TrimSpace(os.Getenv("FORCE_SUBSCRIPTION_CHANNEL"))
}

func isMemberStatus(status string, isMember bool) bool {
	switch status {
	case "creator", "administrator", "member":
		return true
	case "restricted":
		return isMember
	default:
		return false
	}
}

func isUserSubscribed(botToken string, userID int64) (bool, error) {
	channel := forcedSubscriptionChannel()
	if channel == "" {
		return true, nil
	}
	values := url.Values{}
	values.Set("chat_id", channel)
	values.Set("user_id", fmt.Sprintf("%d", userID))
	endpoint := "https://api.telegram.org/bot" + botToken + "/getChatMember?" + values.Encode()
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return false, fmt.Errorf("تعذر الاتصال بتليجرام للتحقق من الاشتراك")
	}
	defer resp.Body.Close()

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			Status   string `json:"status"`
			IsMember bool   `json:"is_member"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("تعذر قراءة نتيجة التحقق من الاشتراك")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || !result.Ok {
		detail := result.Description
		if detail == "" {
			detail = "تأكد من أن @nerd8bot مشرف في " + channel
		}
		return false, fmt.Errorf("فشل التحقق من الاشتراك: %s. تأكد من إضافة @nerd8bot مشرفاً في %s", detail, channel)
	}
	return isMemberStatus(result.Result.Status, result.Result.IsMember), nil
}

func sendSubscriptionGate(token string, chatID int64, detail string) {
	channel := strings.TrimPrefix(forcedSubscriptionChannel(), "@")
	text := "🔒 لاستخدام Nerd، اشترك أولاً في قناة @" + channel + " ثم اضغط «تحققت من الاشتراك»."
	if detail != "" {
		text += "\n\n⚠️ " + detail
	}
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "📢 الاشتراك في القناة", "url": "https://t.me/" + url.PathEscape(channel), "style": "primary"}},
			{{"text": "✅ تحققت من الاشتراك", "callback_data": "check_subscription", "style": "success"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": keyboard}
	if _, err := postTelegramJSON(token, "sendMessage", payload); err != nil {
		log.Printf("فشل إرسال شاشة الاشتراك الإجباري: %v", err)
	}
}

func sendActivationGuide(token string, chatID int64, lang string) {
	text := "🛠️ طريقة تفعيل Nerd على Telegram Business\n\n" +
		"1. افتح إعدادات Telegram ثم Telegram Business / الأعمال.\n" +
		"2. افتح Chatbots / روبوتات المحادثة وأضف @nerd8bot.\n" +
		"3. فعّل أذونات إدارة الرسائل كاملة (5/5).\n" +
		"4. فعّل أذونات إدارة الملف الشخصي كاملة (4/4).\n" +
		"5. فعّل إدارة القصص أيضاً.\n" +
		"6. اضغط Save / حفظ أو علامة ✓ لتأكيد الأذونات.\n\n" +
		"قد تختلف أسماء القوائم قليلاً حسب إصدار Telegram. هذه الخطوات تُنفّذ من إعدادات حسابك، ولا يستطيع البوت منح نفسه الأذونات."
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "🤖 فتح @nerd8bot", "url": "https://t.me/nerd8bot", "style": "success"}},
			{{"text": "🔙 رجوع للقائمة", "callback_data": "main_menu", "style": "primary"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": keyboard}
	if _, err := postTelegramJSON(token, "sendMessage", payload); err != nil {
		log.Printf("فشل إرسال دليل التفعيل: %v", err)
	}
}

func sendSmartReplyMenu(token string, chatID int64, lang string, cfg BotConfig, notice string) {
	status := "🔴 متوقفة"
	if cfg.SmartReplyEnabled {
		status = "🟢 مفعّلة (معاينة وموافقة)"
	}
	keyStatus := "غير موجود"
	if strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")) != "" {
		keyStatus = "موجود في بيئة التشغيل"
	}
	text := "🧠 المحادثة الآلية الذكية\nالحالة: " + status + "\nمفتاح OpenRouter: " + keyStatus +
		"\nالنموذج: " + defaultOpenRouterModel + " (يمكن تغييره عبر OPENROUTER_MODEL)"
	if notice != "" {
		text = notice + "\n\n" + text
	}
	buttonText := "🟢 تشغيل الرد الذكي"
	style := "success"
	if cfg.SmartReplyEnabled {
		buttonText = "🔴 إيقاف الرد الذكي"
		style = "danger"
	}
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": buttonText, "callback_data": "smart_reply_toggle", "style": style}},
			{{"text": "📘 طريقة تفعيل البوت", "callback_data": "activation_guide", "style": "primary"}},
			{{"text": "🔙 رجوع", "callback_data": "main_menu", "style": "primary"}},
		},
	}
	payload := map[string]interface{}{"chat_id": chatID, "text": text, "reply_markup": keyboard}
	if _, err := postTelegramJSON(token, "sendMessage", payload); err != nil {
		log.Printf("فشل إرسال قائمة الرد الذكي: %v", err)
	}
}

func sendSmartReplyPreview(token string, adminID int64, customerChatID int64, customerName, reply string) {
	if len([]rune(reply)) > 2300 {
		reply = string([]rune(reply)[:2300]) + "…"
	}
	text := "👁️ معاينة رد Nerd الذكي\nلن يُرسل الرد للعميل إلا بعد موافقتك.\n\n👤 العميل: " + customerName +
		fmt.Sprintf("\n🆔 المحادثة: %d\n\n", customerChatID) + reply
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{
				{"text": "✅ موافقة وإرسال", "callback_data": "smart_preview_confirm", "style": "success"},
				{"text": "✏️ تعديل", "callback_data": "smart_preview_edit", "style": "primary"},
			},
			{{"text": "❌ إلغاء", "callback_data": "smart_preview_cancel", "style": "danger"}},
		},
	}
	payload := map[string]interface{}{"chat_id": adminID, "text": text, "reply_markup": keyboard}
	if _, err := postTelegramJSON(token, "sendMessage", payload); err != nil {
		log.Printf("فشل إرسال معاينة الرد الذكي: %v", err)
	}
}
