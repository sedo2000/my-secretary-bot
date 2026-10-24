package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
	Role    string `json:"role"`
	Content string `json:"content"`
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
	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		return "", fmt.Errorf("مفتاح OpenRouter غير مضبوط في متغيرات Vercel: OPENROUTER_API_KEY")
	}
	customerMessage = strings.TrimSpace(customerMessage)
	if customerMessage == "" {
		return "", fmt.Errorf("رسالة العميل فارغة")
	}
	if len([]rune(customerMessage)) > 3000 {
		customerMessage = string([]rune(customerMessage)[:3000])
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
				Content: "أنت مساعد يكتب مسودة رد قصيرة وطبيعية نيابة عن صاحب حساب Telegram Business. اكتب بلغة العميل، وإذا كانت الرسالة بالعربية فاستخدم عربية طبيعية ولهجة عراقية خفيفة عند ملاءمتها. أخرج نص الرد فقط دون عناوين أو علامات اقتباس. لا تخترع أسعاراً أو وعوداً أو معلومات غير معروفة، ولا تدّعِ أنك نفذت إجراءً. إذا كانت الرسالة شكوى أو طلباً حساساً، اقترح رداً مهذباً يوضح أن صاحب الحساب سيتابع الأمر.",
			},
			{Role: "user", Content: customerMessage},
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
	reply := strings.TrimSpace(result.Choices[0].Message.Content)
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
