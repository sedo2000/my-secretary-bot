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

func starsPlanForDays(days int) (starsPlan, bool) {
	switch days {
	case 14:
		return starsPlan{Days: 14, Amount: 100, Label: "أسبوعان"}, true
	case 30:
		return starsPlan{Days: 30, Amount: 250, Label: "شهر"}, true
	case 180:
		return starsPlan{Days: 180, Amount: 500, Label: "6 أشهر"}, true
	default:
		return starsPlan{}, false
	}
}

func planDaysForCallback(data string) (int, bool) {
	switch data {
	case "buy_14":
		return 14, true
	case "buy_30":
		return 30, true
	case "buy_180":
		return 180, true
	default:
		return 0, false
	}
}

func isSubscriptionCallback(data string) bool {
	switch data {
	case "plans", "subscription_status", "buy_14", "buy_30", "buy_180":
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
	text := "💎 Nerd Premium\n\n🔒 ميزات الإدارة مقفلة حتى تفعيل الاشتراك.\n\n" +
		"اختر مدة الاشتراك وادفع باستخدام Telegram Stars:\n" +
		"• أسبوعان — 100 ⭐\n" +
		"• شهر — 250 ⭐\n" +
		"• 6 أشهر — 500 ⭐\n\n" +
		"بعد تأكيد الدفع من تيليجرام، تُفتح الميزات تلقائيًا."
	if notice != "" {
		text = notice + "\n\n" + text
	}
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "🗓️ أسبوعان · 100 ⭐", "callback_data": "buy_14"}},
			{{"text": "📅 شهر · 250 ⭐", "callback_data": "buy_30"}},
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
		sendSubscriptionMenu(token, chatID, "📭 لا يوجد اشتراك فعّال حاليًا.")
		return
	}
	remaining := time.Until(time.Unix(config.SubscriptionUntil, 0))
	days := int((remaining + 24*time.Hour - 1) / (24 * time.Hour))
	text := fmt.Sprintf("✅ اشتراك Nerd Premium فعّال.\n⏳ متبقٍ نحو %d يومًا.\n📅 ينتهي في: %s (UTC)", days, time.Unix(config.SubscriptionUntil, 0).UTC().Format("2006-01-02 15:04"))
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
		{"label": "Nerd Premium · " + plan.Label, "amount": plan.Amount},
	}
	_, err := postTelegramJSON(token, "sendInvoice", map[string]interface{}{
		"chat_id": chatID,
		"title": "اشتراك Nerd Premium",
		"description": fmt.Sprintf("تفعيل ميزات Nerd لمدة %d يومًا.", plan.Days),
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
	sendMessage(token, chatID, fmt.Sprintf("🎉 تم تأكيد الدفع بنجاح!\n💎 باقة Nerd Premium: %s\n⏳ المدة: %d يومًا\n📅 تاريخ الانتهاء: %s (UTC)\n\nيمكنك الآن استخدام ميزات الإدارة.", plan.Label, plan.Days, time.Unix(config.SubscriptionUntil, 0).UTC().Format("2006-01-02 15:04")))
	sendMenu(token, chatID, config.Lang, tr(config.Lang, "main_menu_title"))
	return nil
}

func logSubscriptionError(action string, chatID int64, err error) {
	// Avoid logging invoice payloads or payment identifiers.
	log.Printf("%s failed for chat %d: %v", action, chatID, err)
}
