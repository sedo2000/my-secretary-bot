package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type starsPlan struct {
	Days   int
	Amount int
	Label  string
}

// Baghdad uses بتوقيت بغداد+3 year-round for the bot's displayed timestamps.
var baghdadLocation = time.FixedZone("Asia/Baghdad", 3*60*60)

func init() { time.Local = baghdadLocation }

func formatBaghdadTimestamp(unixSeconds int64) string {
	return time.Unix(unixSeconds, 0).In(baghdadLocation).Format("2006-01-02 15:04")
}

const trialDuration = 7 * 24 * time.Hour

func ensureFreeTrial(config *BotConfig) bool {
	if config.TrialStartedAt != 0 { return false }
	now := time.Now()
	config.TrialStartedAt = now.Unix()
	config.TrialUntil = now.Add(trialDuration).Unix()
	return true
}

func trialActive(config BotConfig) bool { return config.TrialUntil > time.Now().Unix() }

func canUseStoryPublishing(token string, userID int64, config BotConfig) bool {
	if isOwnerTelegramID(userID) || subscriptionActive(config) || trialActive(config) { return true }
	ownerID := ownerTelegramID()
	if ownerID <= 0 || userID <= 0 { return false }
	ownerConfig, msgID := getConfig(token, ownerID)
	if msgID < 0 { return false }
	for _, exemptID := range ownerConfig.ExemptedSubscriptionIDs { if exemptID == userID { return true } }
	return false
}

func trialWelcomeText(config BotConfig) string {
	return fmt.Sprintf("🎁 بدأت تجربتك المجانية لمدة 7 أيام وتشمل جميع ميزات البوت!\n📅 تنتهي في: %s (توقيت بغداد)\n\nبعد انتهاء التجربة، تبقى جميع الميزات متاحة باستثناء نشر ستوري ونشر ستوريات متعددة، ويمكن تفعيلهما بباقة Stars.", formatBaghdadTimestamp(config.TrialUntil))
}

func sendTrialNotice(token string, chatID int64, config *BotConfig, msgID int, newlyStarted bool) {
	if newlyStarted { sendMessage(token, chatID, trialWelcomeText(*config)); return }
	now := time.Now().Unix()
	if config.TrialUntil == 0 { return }
	if now >= config.TrialUntil && !config.TrialExpiredNoticeSent {
		config.TrialExpiredNoticeSent = true
		if saveConfig(token, chatID, *config, msgID) { sendMessage(token, chatID, "⏰ انتهت تجربتك المجانية لمدة 7 أيام. يمكنك مواصلة استخدام جميع ميزات البوت، أما نشر الستوري والستوريات المتعددة فيحتاج إلى باقة Stars.") }
		return
	}
	if now < config.TrialUntil && config.TrialUntil-now <= 24*60*60 && !config.TrialReminderSent {
		config.TrialReminderSent = true
		if saveConfig(token, chatID, *config, msgID) { sendMessage(token, chatID, fmt.Sprintf("⏳ تذكير: تنتهي تجربتك المجانية خلال أقل من 24 ساعة، بتاريخ %s (توقيت بغداد). بعد انتهائها، ستحتاج إلى Stars فقط لنشر الستوري والستوريات المتعددة.", formatBaghdadTimestamp(config.TrialUntil))) }
	}
}

func starsPlanForDays(days int) (starsPlan, bool) {
	switch days {
	case 30:
		return starsPlan{Days: 30, Amount: 100, Label: "شهر"}, true
	case 90:
		return starsPlan{Days: 90, Amount: 250, Label: "3 أشهر"}, true
	case 180:
		return starsPlan{Days: 180, Amount: 500, Label: "6 أشهر"}, true
	default:
		return starsPlan{}, false
	}
}

func planDaysForCallback(data string) (int, bool) {
	switch data {
	case "buy_30":
		return 30, true
	case "buy_90":
		return 90, true
	case "buy_180":
		return 180, true
	default:
		return 0, false
	}
}

func isSubscriptionCallback(data string) bool {
	switch data {
	case "plans", "subscription_status", "buy_30", "buy_90", "buy_180":
		return true
	default:
		return false
	}
}

func subscriptionActive(config BotConfig) bool {
	return config.SubscriptionUntil > time.Now().Unix()
}

func ownerTelegramID() int64 {
	id, err := strconv.ParseInt(strings.TrimSpace(os.Getenv("OWNER_TELEGRAM_ID")), 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func isOwnerTelegramID(userID int64) bool {
	ownerID := ownerTelegramID()
	return ownerID > 0 && userID == ownerID
}

// hasPremiumAccess grants access to the configured owner, explicitly exempted
// users, or users with an active paid subscription.
func hasPremiumAccess(token string, userID int64, config BotConfig) bool {
	if isOwnerTelegramID(userID) || subscriptionActive(config) {
		return true
	}
	ownerID := ownerTelegramID()
	if ownerID == 0 || userID <= 0 {
		return false
	}
	ownerConfig, msgID := getConfig(token, ownerID)
	if msgID < 0 {
		return false
	}
	for _, exemptID := range ownerConfig.ExemptedSubscriptionIDs {
		if exemptID == userID {
			return true
		}
	}
	return false
}

func subscriptionPayload(userID int64, days int) string {
	return fmt.Sprintf("nerd_sub:%d:%d", userID, days)
}

func parseSubscriptionPayload(payload string) (int64, starsPlan, bool) {
	parts := strings.Split(payload, ":")
	if len(parts) != 3 || parts[0] != "nerd_sub" {
		return 0, starsPlan{}, false
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || userID <= 0 {
		return 0, starsPlan{}, false
	}
	days, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, starsPlan{}, false
	}
	plan, ok := starsPlanForDays(days)
	return userID, plan, ok
}

func validStarsCheckout(userID int64, currency string, amount int, payload string) bool {
	expectedUserID, plan, ok := parseSubscriptionPayload(payload)
	return ok && expectedUserID == userID && currency == "XTR" && amount == plan.Amount
}

func sendSubscriptionMenu(token string, chatID int64, notice string) {
	text := "⭐ تفعيل نشر الستوريات\n\nتجربة مجانية لمدة 7 أيام تشمل جميع الميزات. بعد انتهائها، يبقى البوت متاحًا وتحتاج إلى باقة فقط لنشر ستوري أو نشر ستوريات متعددة.\n\nاختر باقة نشر الستوريات باستخدام Telegram Stars:\n• شهر — 100 ⭐\n• 3 أشهر — 250 ⭐\n• 6 أشهر — 500 ⭐\n\nبعد تأكيد الدفع من تيليجرام، تُفتح ميزات النشر المدفوعة تلقائيًا."
	if notice != "" {
		text = notice + "\n\n" + text
	}
	rows := [][]map[string]interface{}{
		{{"text": "📅 شهر · 100 ⭐", "callback_data": "buy_30"}},
		{{"text": "🗓️ 3 أشهر · 250 ⭐", "callback_data": "buy_90"}},
		{{"text": "👑 6 أشهر · 500 ⭐", "callback_data": "buy_180"}},
		{{"text": "ℹ️ حالة اشتراكي", "callback_data": "subscription_status"}},
	}
	// Owner-only testing: grants access for QA without creating a payment or revenue record.
	if isOwnerTelegramID(chatID) {
		rows = append(rows, []map[string]interface{}{{"text": "🧪 اختبار الاشتراك مجانًا (للمالك)", "callback_data": "owner_test_subscription"}})
	}
	keyboard := map[string]interface{}{"inline_keyboard": rows}
	_, err := postTelegramJSON(token, "sendMessage", map[string]interface{}{
		"chat_id": chatID, "text": text, "reply_markup": keyboard,
	})
	if err != nil {
		logSubscriptionError("sendSubscriptionMenu", chatID, err)
	}
}

func sendSubscriptionStatus(token string, chatID int64, config BotConfig) {
	if !subscriptionActive(config) {
		if trialActive(config) {
			remaining := time.Until(time.Unix(config.TrialUntil, 0))
			days := int((remaining + 24*time.Hour - 1) / (24 * time.Hour))
			notice := fmt.Sprintf("🎁 تجربتك المجانية فعّالة وتشمل جميع الميزات.\n⏳ المتبقي: نحو %d يومًا.\n📅 تنتهي في: %s (توقيت بغداد).\n\nبعد انتهاء التجربة، الدفع مطلوب فقط لنشر الستوري والستوريات المتعددة.", days, formatBaghdadTimestamp(config.TrialUntil))
			sendSubscriptionMenu(token, chatID, notice)
		} else {
			sendSubscriptionMenu(token, chatID, "📭 انتهت التجربة المجانية. ما زالت ميزات البوت متاحة، وتحتاج إلى باقة فقط لنشر الستوري والستوريات المتعددة.")
		}
		return
	}
	remaining := time.Until(time.Unix(config.SubscriptionUntil, 0))
	days := int((remaining + 24*time.Hour - 1) / (24 * time.Hour))
	text := fmt.Sprintf("✅ اشتراك نشر الستوريات فعّال.\n⏳ متبقٍ نحو %d يومًا.\n📅 ينتهي في: %s (توقيت بغداد)", days, formatBaghdadTimestamp(config.SubscriptionUntil))
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "➕ تمديد الاشتراك", "callback_data": "plans"}},
		},
	}
	if _, err := postTelegramJSON(token, "sendMessage", map[string]interface{}{
		"chat_id": chatID, "text": text, "reply_markup": keyboard,
	}); err != nil {
		logSubscriptionError("sendSubscriptionStatus", chatID, err)
	}
}

func sendStarsInvoice(token string, chatID int64, days int) error {
	plan, ok := starsPlanForDays(days)
	if !ok {
		return fmt.Errorf("unknown subscription plan")
	}
	payload := subscriptionPayload(chatID, days)
	prices := []map[string]interface{}{
		{"label": "نشر الستوريات · " + plan.Label, "amount": plan.Amount},
	}
	_, err := postTelegramJSON(token, "sendInvoice", map[string]interface{}{
		"chat_id": chatID,
		"title": "تفعيل نشر الستوريات",
		"description": fmt.Sprintf("فتح ميزات نشر الستوريات لمدة %d يومًا.", plan.Days),
		"payload": payload,
		"provider_token": "",
		"currency": "XTR",
		"prices": prices,
	})
	return err
}

func applySuccessfulStarsPayment(token string, chatID int64, payment *SuccessfulPayment, config *BotConfig, configMsgID int) error {
	if payment == nil || payment.Currency != "XTR" {
		return fmt.Errorf("invalid Stars payment currency")
	}
	userID, plan, ok := parseSubscriptionPayload(payment.InvoicePayload)
	if !ok || userID != chatID || payment.TotalAmount != plan.Amount || payment.TelegramPaymentChargeID == "" {
		return fmt.Errorf("payment details do not match a known plan")
	}

	for _, chargeID := range config.ProcessedStarChargeIDs {
		if chargeID == payment.TelegramPaymentChargeID {
			sendSubscriptionStatus(token, chatID, *config)
			return nil
		}
	}

	now := time.Now().Unix()
	start := now
	if config.SubscriptionUntil > start {
		start = config.SubscriptionUntil
	}
	oldUntil := config.SubscriptionUntil
	oldCharges := append([]string(nil), config.ProcessedStarChargeIDs...)
	config.SubscriptionUntil = time.Unix(start, 0).Add(time.Duration(plan.Days) * 24 * time.Hour).Unix()
	config.ProcessedStarChargeIDs = append(config.ProcessedStarChargeIDs, payment.TelegramPaymentChargeID)
	if len(config.ProcessedStarChargeIDs) > 30 {
		config.ProcessedStarChargeIDs = append([]string(nil), config.ProcessedStarChargeIDs[len(config.ProcessedStarChargeIDs)-30:]...)
	}
	if !saveConfig(token, chatID, *config, configMsgID) {
		config.SubscriptionUntil = oldUntil
		config.ProcessedStarChargeIDs = oldCharges
		return fmt.Errorf("could not persist subscription config")
	}
	sendMessage(token, chatID, fmt.Sprintf("🎉 تم تأكيد الدفع بنجاح!\n💎 ميزة نشر الستوريات: %s\n⏳ المدة: %d يومًا\n📅 تاريخ الانتهاء: %s (UTC)\n\nتم فتح نشر الستوري والستوريات المتعددة.", plan.Label, plan.Days, formatBaghdadTimestamp(config.SubscriptionUntil)))
	sendMenu(token, chatID, config.Lang, tr(config.Lang, "main_menu_title"))
	return nil
}

func logSubscriptionError(action string, chatID int64, err error) {
	// Avoid logging invoice payloads or payment identifiers.
	log.Printf("%s failed for chat %d: %v", action, chatID, err)
}

func sendOwnerPaymentsMenu(token string, chatID int64) {
	rows := [][]map[string]interface{}{
		{{"text": "🧾 آخر المدفوعات", "callback_data": "owner_payments_recent"}},
		{{"text": "📊 إحصائيات النجوم", "callback_data": "owner_payments_stats"}},
		{{"text": "🔎 البحث عن مستخدم", "callback_data": "owner_payments_lookup"}},
		{{"text": "🔙 رجوع", "callback_data": "main_menu"}},
	}
	_, err := postTelegramJSON(token, "sendMessage", map[string]interface{}{
		"chat_id": chatID, "text": "💳 إدارة المدفوعات والاشتراكات\n\nتُسجَّل المدفوعات بعد تأكيد Telegram فقط. السجل يعرض أحدث العمليات المحفوظة، والإحصائيات إجمالية للعمليات التي سجّلها البوت.",
		"reply_markup": map[string]interface{}{"inline_keyboard": rows},
	})
	if err != nil { logSubscriptionError("sendOwnerPaymentsMenu", chatID, err) }
}

func sendRecentStarPayments(token string, chatID int64, config BotConfig) {
	if len(config.StarPayments) == 0 { sendMessage(token, chatID, "🧾 لا توجد مدفوعات مسجلة بعد."); return }
	baghdad := time.FixedZone("Asia/Baghdad", 3*60*60)
	var b strings.Builder
	fmt.Fprintf(&b, "🧾 أحدث المدفوعات المحفوظة (%d):\n\n", len(config.StarPayments))
	for i := len(config.StarPayments)-1; i >= 0; i-- {
		p := config.StarPayments[i]
		fmt.Fprintf(&b, "• المستخدم: %d\n  المبلغ: %d ⭐ | المدة: %d يومًا\n  الدفع: %s (بغداد)\n  ينتهي: %s (بغداد)\n\n",
			p.UserID, p.Amount, p.Days,
			formatBaghdadTimestamp(p.PaidAt),
			formatBaghdadTimestamp(p.SubscriptionUntil))
	}
	text := b.String()
	if len([]rune(text)) > 3900 { text = string([]rune(text)[:3900]) }
	sendMessage(token, chatID, text)
}

func sendStarPaymentStats(token string, chatID int64, config BotConfig) {
	balanceText := "غير متاح عبر Bot API؛ راجع رصيد Stars من إعدادات البوت في Telegram."
	if balance, err := getBotStarBalance(token); err == nil { balanceText = fmt.Sprintf("%d ⭐", balance) }
	sendMessage(token, chatID, fmt.Sprintf("📊 إحصائيات مدفوعات Stars\n\n💰 إجمالي النجوم المسجلة: %d ⭐\n🧾 عدد عمليات الدفع المسجلة: %d\n💼 رصيد Stars الحالي: %s\n\nملاحظة: إجمالي النجوم هو مجموع المدفوعات التي سجّلها هذا البوت، وليس بالضرورة الرصيد القابل للسحب بعد أي معاملات أو استردادات.", config.StarRevenueTotal, config.StarPaymentCount, balanceText))
}

func sendUserSubscriptionLookup(token string, ownerChatID, userID int64) {
	userConfig, msgID := getConfig(token, userID)
	if msgID < 0 { sendMessage(token, ownerChatID, fmt.Sprintf("⚠️ تعذر قراءة إعدادات المستخدم %d من Telegram، لذلك لم أستطع التحقق من اشتراكه.", userID)); return }
	status := "❌ لا يوجد اشتراك مدفوع فعّال"
	if subscriptionActive(userConfig) {
		status = fmt.Sprintf("✅ اشتراك مدفوع فعّال حتى %s بتوقيت بغداد", formatBaghdadTimestamp(userConfig.SubscriptionUntil))
	} else if trialActive(userConfig) {
		status = fmt.Sprintf("🎁 تجربة مجانية فعّالة حتى %s UTC (ليست اشتراكًا مدفوعًا)", formatBaghdadTimestamp(userConfig.TrialUntil))
	}
	ownerConfig, ownerMsgID := getConfig(token, ownerChatID)
	paymentHistory := "لا توجد عملية دفع لهذا المستخدم ضمن آخر العمليات المحفوظة."
	if ownerMsgID >= 0 {
		for i := len(ownerConfig.StarPayments)-1; i >= 0; i-- {
			p := ownerConfig.StarPayments[i]
			if p.UserID == userID { paymentHistory = fmt.Sprintf("🧾 آخر دفعة مسجلة: %d ⭐، باقة %d يومًا، بتاريخ %s UTC.", p.Amount, p.Days, formatBaghdadTimestamp(p.PaidAt)); break }
		}
	}
	sendMessage(token, ownerChatID, fmt.Sprintf("🔎 حالة المستخدم %d\n\n%s\n%s", userID, status, paymentHistory))
}

func recordOwnerStarPayment(token string, userID int64, payment *SuccessfulPayment, userConfig BotConfig) error {
	if payment == nil { return fmt.Errorf("missing successful payment") }
	ownerID := ownerTelegramID()
	if ownerID <= 0 { return fmt.Errorf("OWNER_TELEGRAM_ID is not configured") }
	expectedUserID, plan, ok := parseSubscriptionPayload(payment.InvoicePayload)
	if !ok || expectedUserID != userID || payment.Currency != "XTR" || payment.TotalAmount != plan.Amount || payment.TelegramPaymentChargeID == "" { return fmt.Errorf("payment does not match a configured Stars plan") }
	ownerConfig, ownerMsgID := getConfig(token, ownerID)
	if ownerMsgID < 0 { return fmt.Errorf("could not load owner config") }
	for _, id := range ownerConfig.RecordedStarChargeIDs { if id == payment.TelegramPaymentChargeID { return nil } }
	ownerConfig.RecordedStarChargeIDs = append(ownerConfig.RecordedStarChargeIDs, payment.TelegramPaymentChargeID)
	if len(ownerConfig.RecordedStarChargeIDs) > 20 { ownerConfig.RecordedStarChargeIDs = append([]string(nil), ownerConfig.RecordedStarChargeIDs[len(ownerConfig.RecordedStarChargeIDs)-20:]...) }
	ownerConfig.StarRevenueTotal += int64(payment.TotalAmount)
	ownerConfig.StarPaymentCount++
	ownerConfig.StarPayments = append(ownerConfig.StarPayments, StarPaymentRecord{UserID: userID, Amount: payment.TotalAmount, Days: plan.Days, PaidAt: time.Now().Unix(), SubscriptionUntil: userConfig.SubscriptionUntil, ChargeID: payment.TelegramPaymentChargeID})
	if len(ownerConfig.StarPayments) > 10 { ownerConfig.StarPayments = append([]StarPaymentRecord(nil), ownerConfig.StarPayments[len(ownerConfig.StarPayments)-10:]...) }
	if !saveConfig(token, ownerID, ownerConfig, ownerMsgID) { return fmt.Errorf("could not persist owner payment ledger") }
	_, err := postTelegramJSON(token, "sendMessage", map[string]interface{}{
		"chat_id": ownerID,
		"text": fmt.Sprintf("💳 تم استلام دفعة Stars جديدة!\n\n🆔 المستخدم: %d\n⭐ المبلغ: %d Stars\n📦 الباقة: %s (%d يومًا)\n📅 انتهاء الاشتراك: %s UTC\n\n📊 إجمالي النجوم المسجلة: %d ⭐", userID, payment.TotalAmount, plan.Label, plan.Days, formatBaghdadTimestamp(userConfig.SubscriptionUntil), ownerConfig.StarRevenueTotal),
	})
	if err != nil { return fmt.Errorf("payment saved but owner notification failed: %w", err) }
	return nil
}

func getBotStarBalance(token string) (int64, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMyStarBalance", token)
	resp, err := httpClient.Get(apiURL)
	if err != nil { return 0, err }
	defer resp.Body.Close()
	var result struct {
		Ok bool `json:"ok"`
		Description string `json:"description"`
		Result struct { Amount int64 `json:"amount"` } `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return 0, err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !result.Ok { return 0, fmt.Errorf("Telegram balance endpoint unavailable: %s", result.Description) }
	return result.Result.Amount, nil
}
