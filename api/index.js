/**
 * ============================================================================
 * مشروع: بوت سكرتير تليجرام للأعمال وإشراف المجموعات (Enterprise Edition)
 * الإطار المستخدم: grammY (Node.js)
 * البيئة المستهدفة: Vercel Serverless Functions
 * ============================================================================
 */

const { Bot, webhookCallback, InlineKeyboard, InputFile } = require("grammy");

// 1️⃣ إعدادات البوت والتوكن
const token = process.env.TELEGRAM_BOT_TOKEN;
const bot = new Bot(token || "DUMMY_TOKEN");

// الذاكرة المؤقتة لإدارة الجلسات والإعدادات
const memoryCache = new Map();

// معالج الأخطاء العالمي لمنع الانهيار
bot.catch((err) => {
  console.error("❌ حدث خطأ داخلي في معالجة البوت:", err.error || err);
});

// القواميس والبيانات الأساسية
const badWords = ["كحبة", "مطي", "قندرة", "ساقط", "فرخ", "عير", "كس", "طيز", "زنيّم"];

const smartAnswers = {
  السعر: "ℹ️ لمعرفة الأسعار والتفاصيل الكاملة، يمكنك زيارة القناة الرسمية أو مراسلة الدعم الفني المباشر.",
  الدعم: "🛠️ للتواصل مع فريق الدعم الفني، يرجى مراسلة الحساب التجاري المباشر.",
  التسجيل: "📝 يمكنك التسجيل والاشتراك عبر فتح المحادثة الخاصة واتباع التعليمات.",
};

const quotes = [
  "قاوم ما تكره لتصل الى ما تحب",
  "الحرب بين أنت ضد أنت",
  "أبنِ نفسك بنفسك لنفسك",
  "ميخالف، عابر سبيل ستمر كل الصعاب",
  "حتى لو متأخر تگدر تبدأ من جديد..!",
  "من يعيش في خوف لن يكون حراً ابداً",
  "لا أبرح حتى أبلغ المبتغى",
  "أنه مبرمج فحسب، يصنع واقعه بيديه",
  "المرء نتاج خلواته وتأملاته العميقة",
  "لا مزيد من الأصدقاء المزيفين"
];

function getNerdChannelKeyboard() {
  return new InlineKeyboard().url("تحديثات نيرد 📢", "https://t.me/Xhwe2");
}

async function translateText(text, targetLang) {
  if (!text || !text.trim()) return { text: "", detectedLang: "" };
  try {
    const url = `https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=${targetLang}&dt=t&q=${encodeURIComponent(text)}`;
    const res = await fetch(url);
    const data = await res.json();
    let translatedText = "";
    if (data && data[0]) {
      data[0].forEach((item) => {
        if (item && item[0]) translatedText += item[0];
      });
    }
    const detectedLang = data && data[2] ? data[2] : "";
    return { text: translatedText, detectedLang };
  } catch (err) {
    return { text: text, detectedLang: "" };
  }
}

// 2️⃣ حماية المجموعات
bot.on(["message:group", "message:supergroup"], async (ctx) => {
  const messageText = ctx.message.text || ctx.message.caption || "";
  const userId = ctx.from?.id;

  if (badWords.some((w) => messageText.toLowerCase().includes(w))) {
    try { await ctx.deleteMessage(); return; } catch (e) {}
  }

  if (/(https?:\/\/[^\s]+)|(t\.me\/[^\s]+)/i.test(messageText) && userId) {
    try { await ctx.deleteMessage(); await ctx.banChatMember(userId); return; } catch (e) {}
  }

  for (const [key, answer] of Object.entries(smartAnswers)) {
    if (messageText.includes(key)) {
      await ctx.reply(answer, {
        reply_to_message_id: ctx.message.message_id,
        reply_markup: getNerdChannelKeyboard(),
      });
      return;
    }
  }
});

// 3️⃣ أوامر الخاص واللوحة
bot.command("start", async (ctx) => {
  if (ctx.chat.type !== "private") return;
  await ctx.reply("أهلاً بك في لوحة تحكم سكرتير الحساب التجاري 🤖", {
    reply_markup: new InlineKeyboard().url("تحديثات نيرد 📢", "https://t.me/Xhwe2"),
  });
});

bot.on("callback_query:data", async (ctx) => {
  await ctx.answerCallbackQuery();
  if (ctx.callbackQuery.data === "change_quote") {
    const randomQuote = quotes[Math.floor(Math.random() * quotes.length)];
    await ctx.editMessageReplyMarkup({
      reply_markup: new InlineKeyboard()
        .text(`✨ ${randomQuote}`, "change_quote").row()
        .url("تحديثات نيرد 📢", "https://t.me/Xhwe2"),
    });
  }
});

// 4️⃣ رسائل الأعمال والرد الآلي
bot.on("business_message", async (ctx) => {
  const msg = ctx.update.business_message;
  if (!msg || msg.from.is_bot || msg.is_outgoing) return;

  const customerName = msg.from.first_name || "عميلنا العزيز";
  let replyText = `أهلاً بك يا ${customerName} 🌸\nاستلمت رسالتك وسأرد عليك في أقرب وقت.`;

  if (msg.text) {
    const trRes = await translateText(msg.text, "ar");
    if (trRes.detectedLang && trRes.detectedLang !== "ar") {
      const trReply = await translateText(replyText, trRes.detectedLang);
      if (trReply.text) replyText = trReply.text;
    }
  }

  const initialQuote = quotes[Math.floor(Math.random() * quotes.length)];

  await bot.api.sendMessage(msg.chat.id, replyText, {
    business_connection_id: msg.business_connection_id,
    reply_markup: new InlineKeyboard()
      .text(`✨ ${initialQuote}`, "change_quote").row()
      .url("تحديثات نيرد 📢", "https://t.me/Xhwe2"),
  });
});

// 5️⃣ معالج الويب هوك مع الحماية لطلبات المتصفح (GET vs POST)
const handleUpdate = webhookCallback(bot, "http");

module.exports = async (req, res) => {
  if (req.method !== "POST") {
    res.statusCode = 200;
    res.setHeader("Content-Type", "text/html; charset=utf-8");
    return res.end(`
      <html dir="rtl">
        <head><title>بوت تليجرام للأعمال</title></head>
        <body style="font-family: Tahoma; text-align: center; padding-top: 50px; background: #0f172a; color: #fff;">
          <h1>🤖 بوت سكرتير تليجرام للأعمال يعمل بنجاح!</h1>
          <p>الخادم متصل وجاهز لاستقبال التحديثات من تليجرام.</p>
        </body>
      </html>
    `);
  }

  try {
    return await handleUpdate(req, res);
  } catch (error) {
    console.error("❌ خطأ في معالجة الويب هوك:", error);
    res.statusCode = 500;
    res.setHeader("Content-Type", "application/json");
    return res.end(JSON.stringify({ error: error.message }));
  }
};
