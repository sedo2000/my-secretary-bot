package handler

import (
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
	baghdad := time.FixedZone("Asia/Baghdad", 3*60*60)
	return fmt.Sprintf("🎁 بدأت تجربتك المجانية لمدة 7 أيام وتشمل جميع ميزات البوت!\n📅 تنتهي في: %s (توقيت بغداد)\n\nبعد انتهاء التجربة، تبقى جميع الميزات متاحة باستثناء نشر ستوري ونشر ستوريات متعددة، ويمكن تفعيلهما بباقة Stars.", time.Unix(config.TrialUntil, 0).In(baghdad).Format("2006-01-02 15:04"))
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
		baghdad := time.FixedZone("Asia/Baghdad", 3*60*60)
		if saveConfig(token, chatID, *config, msgID) { sendMessage(token, chatID, fmt.Sprintf("⏳ تذكير: تنتهي تجربتك المجانية خلال أقل من 24 ساعة، بتاريخ %s (توقيت بغداد). بعد انتهائها، ستحتاج إلى Stars فقط لنشر الستوري والستوريات المتعددة.", time.Unix(config.TrialUntil, 0).In(baghdad).Format("2006-01-02 15:04"))) }
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
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "📅 شهر · 100 ⭐", "callback_data": "buy_30"}},
			{{"text": "🗓️ 3 أشهر · 250 ⭐", "callback_data": "buy_90"}},
			{{"text": "👑 6 أشهر · 500 ⭐", "callback_data": "buy_180"}},
			{{"text": "ℹ️ حالة اشتراكي", "callback_data": "subscription_status"}},
		},
	}
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
			baghdad := time.FixedZone("Asia/Baghdad", 3*60*60)
			notice := fmt.Sprintf("🎁 تجربتك المجانية فعّالة وتشمل جميع الميزات.\\n⏳ المتبقي: نحو %d يومًا.\\n📅 تنتهي في: %s (توقيت بغداد).\\n\\nبعد انتهاء التجربة، الدفع مطلوب فقط لنشر الستوري والستوريات المتعددة.", days, time.Unix(config.TrialUntil, 0).In(baghdad).Format("2006-01-02 15:04"))
			sendSubscriptionMenu(token, chatID, notice)
		} else {
			sendSubscriptionMenu(token, chatID, "📭 انتهت التجربة المجانية. ما زالت ميزات البوت متاحة، وتحتاج إلى باقة فقط لنشر الستوري والستوريات المتعددة.")
		}
		return
	}
	remaining := time.Until(time.Unix(config.SubscriptionUntil, 0))
	days := int((remaining + 24*time.Hour - 1) / (24 * time.Hour))
	text := fmt.Sprintf("✅ اشتراك نشر الستوريات فعّال.\n⏳ متبقٍ نحو %d يومًا.\n📅 ينتهي في: %s (UTC)", days, time.Unix(config.SubscriptionUntil, 0).UTC().Format("2006-01-02 15:04"))
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
	sendMessage(token, chatID, fmt.Sprintf("🎉 تم تأكيد الدفع بنجاح!\n💎 ميزة نشر الستوريات: %s\n⏳ المدة: %d يومًا\n📅 تاريخ الانتهاء: %s (UTC)\n\nتم فتح نشر الستوري والستوريات المتعددة.", plan.Label, plan.Days, time.Unix(config.SubscriptionUntil, 0).UTC().Format("2006-01-02 15:04")))
	sendMenu(token, chatID, config.Lang, tr(config.Lang, "main_menu_title"))
	return nil
}

func logSubscriptionError(action string, chatID int64, err error) {
	// Avoid logging invoice payloads or payment identifiers.
	log.Printf("%s failed for chat %d: %v", action, chatID, err)
}
