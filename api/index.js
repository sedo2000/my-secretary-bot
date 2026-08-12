// ============================================================
// 📦 استيراد المكتبات والإعدادات الأساسية
// ============================================================
const { Bot, webhookCallback, InlineKeyboard, session, InputFile } = require('grammy');
const express = require('express');
const cors = require('cors');
const axios = require('axios');
const translate = require('google-translate-api-x');

require('dotenv').config();

// ============================================================
// 🔐 المتغيرات الثابتة والإعدادات
// ============================================================
const BOT_TOKEN = process.env.BOT_TOKEN;
const ADMIN_ID = parseInt(process.env.ADMIN_ID);
const CHANNEL_URL = process.env.CHANNEL_URL || 'https://t.me/Xhwe2';

// قوائم الكلمات الممنوعة
const BAD_WORDS = ['شتيمة1', 'شتيمة2', 'كلمة بذيئة']; // أضف القائمة الكاملة
const SPAM_PATTERNS = [
  /t\.me\//gi,
  /telegram\.me\//gi,
  /https?:\/\/t\.me\//gi,
  /@\w+/gi
];

// الكلمات المفتاحية للرد التلقائي
const KEYWORD_REPLIES = {
  'السعر': '💰 الأسعار متوفرة في القناة الرسمية: ' + CHANNEL_URL,
  'الدعم': '🆘 للدعم يرجى التواصل مع الإدارة عبر الخاص',
  'التسجيل': '📝 للتسجيل يرجى زيارة الرابط: ' + CHANNEL_URL,
  'مرحباً': '👋 أهلاً وسهلاً بك! كيف يمكنني مساعدتك؟'
};

// القصص التحفيزية
const QUOTES = [
  '💪 "النجاح ليس غياب الفشل، بل هو الإصرار بعد الفشل"',
  '🌟 "كن التغيير الذي تريد رؤيته في العالم"',
  '🚀 "المستقبل يبدأ من حيث الإرادة تنتهي"',
  '🎯 "التركيز هو مفتاح الإبداع"'
];

// ============================================================
// 🤖 إنشاء البوت مع جلسات الذاكرة
// ============================================================
const bot = new Bot(BOT_TOKEN);

// تفعيل الجلسات للتخزين المؤقت
bot.use(session({
  initial: () => ({
    userMessages: {},
    lastReply: {}
  })
}));

// ============================================================
// 🛠️ أدوات مساعدة
// ============================================================

// دالة الحصول على اسم المستخدم
function getUserName(user) {
  if (!user) return 'العميل';
  return user.first_name || user.username || 'العميل';
}

// دالة إنشاء الزر الشفاف
function createChannelButton() {
  return new InlineKeyboard().url('📢 تحديثات نيرد', CHANNEL_URL);
}

// دالة التحقق من الكلمات البذيئة
function containsBadWords(text) {
  if (!text) return false;
  const lowerText = text.toLowerCase();
  return BAD_WORDS.some(word => lowerText.includes(word.toLowerCase()));
}

// دالة التحقق من السبام
function isSpam(text) {
  if (!text) return false;
  return SPAM_PATTERNS.some(pattern => pattern.test(text));
}

// دالة الحصول على اقتباس عشوائي
function getRandomQuote() {
  return QUOTES[Math.floor(Math.random() * QUOTES.length)];
}

// ============================================================
// 🌐 دالة الترجمة باستخدام Google Translate API
// ============================================================
async function detectAndTranslate(text, targetLang = 'ar') {
  try {
    // اكتشاف اللغة
    const detected = await translate(text, { to: targetLang });
    return {
      original: text,
      translated: detected.text,
      from: detected.from.language.iso,
      to: targetLang
    };
  } catch (error) {
    console.error('خطأ في الترجمة:', error);
    return { original: text, translated: text, from: 'unknown', to: targetLang };
  }
}

// ============================================================
// 📨 2. نظام الرد التلقائي الخاص (Business Private Auto-Reply)
// ============================================================

// معالج رسائل الأعمال
bot.on('business_message', async (ctx) => {
  try {
    const msg = ctx.businessMessage;
    if (!msg || !msg.text) return;

    const userId = msg.from.id;
    const userName = getUserName(msg.from);
    const userText = msg.text;

    // 🔍 اكتشاف لغة المستخدم
    const detectedLang = await detectAndTranslate(userText, 'ar');
    
    // إذا كانت اللغة غير عربية، إرسال تنبيه للأدمن
    if (detectedLang.from !== 'ar') {
      const adminAlert = `
🔔 **تنبيه: رسالة بلغة أجنبية**
👤 المستخدم: ${userName}
🆔 المعرف: ${userId}
🌐 اللغة المكتشفة: ${detectedLang.from}

📝 النص الأصلي:
${detectedLang.original}

📝 الترجمة للعربية:
${detectedLang.translated}
      `;
      await bot.api.sendMessage(ADMIN_ID, adminAlert, { parse_mode: 'Markdown' });
    }

    // 📝 بناء الرد المخصص
    const userFirstName = msg.from.first_name || userName;
    let replyText = `👋 مرحباً ${userFirstName}،\n\n`;
    replyText += `شكراً لتواصلك معنا! سنرد عليك في أقرب وقت.\n\n`;
    replyText += `💡 ${getRandomQuote()}`;

    // 🌐 ترجمة الرد إلى لغة المستخدم إذا كانت غير عربية
    let finalReply = replyText;
    if (detectedLang.from !== 'ar') {
      const translatedReply = await detectAndTranslate(replyText, detectedLang.from);
      finalReply = translatedReply.translated;
    }

    // 📤 إرسال الرد مع الزر
    const keyboard = createChannelButton();
    await ctx.reply(finalReply, {
      reply_parameters: { message_id: msg.message_id },
      reply_markup: keyboard
    });

    // 📊 تسجيل التفاعل
    console.log(`✅ رد تلقائي مرسل إلى ${userName} (${userId})`);

  } catch (error) {
    console.error('❌ خطأ في معالج الرسائل التجارية:', error);
  }
});

// ============================================================
// 🛡️ 3. الإشراف والحماية المتقدمة للمجموعات
// ============================================================

// معالج رسائل المجموعات
bot.on('message', async (ctx) => {
  try {
    const msg = ctx.message;
    if (!msg || !msg.text || !msg.chat) return;

    const chatId = msg.chat.id;
    const userId = msg.from.id;
    const userName = getUserName(msg.from);
    const text = msg.text;

    // التحقق من أن الرسالة في مجموعة
    if (msg.chat.type === 'private') return;

    // ==========================================================
    // 🚫 حظر الألفاظ البذيئة
    // ==========================================================
    if (containsBadWords(text)) {
      await ctx.deleteMessage();
      await ctx.reply(`⚠️ ${userName}، تم حذف رسالتك لاحتوائها على كلمات غير لائقة.`);
      
      // تسجيل المخالفة
      await bot.api.sendMessage(ADMIN_ID, 
        `🚫 **مخالفة: ألفاظ بذيئة**\n👤 ${userName}\n📝 ${text}`
      );
      return;
    }

    // ==========================================================
    // 🛑 منع السبام والإعلانات
    // ==========================================================
    if (isSpam(text)) {
      await ctx.deleteMessage();
      await ctx.reply(`🚫 ${userName}، ممنوع نشر الروابط والإعلانات في المجموعة.`);
      
      // طرد العضو
      try {
        await ctx.banChatMember(userId);
        await bot.api.sendMessage(ADMIN_ID,
          `🚨 **تم طرد عضو بسبب السبام**\n👤 ${userName}\n📝 ${text}`
        );
      } catch (banError) {
        console.error('خطأ في طرد العضو:', banError);
      }
      return;
    }

    // ==========================================================
    // 🤖 الرد الآلي الذكي (Smart Auto-Reply)
    // ==========================================================
    for (const [keyword, reply] of Object.entries(KEYWORD_REPLIES)) {
      if (text.toLowerCase().includes(keyword.toLowerCase())) {
        const keyboard = createChannelButton();
        await ctx.reply(`${reply}`, { reply_markup: keyboard });
        break;
      }
    }

    // ==========================================================
    // 📊 إحصائيات تفاعل الأعضاء
    // ==========================================================
    const sessionData = await ctx.session;
    if (!sessionData.userMessages[userId]) {
      sessionData.userMessages[userId] = 0;
    }
    sessionData.userMessages[userId]++;

    // تحديث الإحصائيات كل 10 رسائل
    if (sessionData.userMessages[userId] % 10 === 0) {
      await bot.api.sendMessage(ADMIN_ID,
        `📊 **إحصائية عضو نشط**\n👤 ${userName}\n📝 عدد الرسائل: ${sessionData.userMessages[userId]}`
      );
    }

  } catch (error) {
    console.error('❌ خطأ في معالج رسائل المجموعة:', error);
  }
});

// ============================================================
// 🔄 4. إدارة الملف الشخصي والقصص
// ============================================================

// أوامر إدارة الملف الشخصي (للاستخدام الداخلي)
bot.command('setname', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  const args = ctx.message.text.split(' ');
  if (args.length < 2) {
    await ctx.reply('⚠️ الاستخدام: /setname الاسم الجديد');
    return;
  }

  const newName = args.slice(1).join(' ');
  try {
    await ctx.api.setMyName(newName);
    await ctx.reply(`✅ تم تغيير الاسم إلى: ${newName}`);
  } catch (error) {
    await ctx.reply('❌ فشل تغيير الاسم: ' + error.message);
  }
});

bot.command('setbio', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  const bio = ctx.message.text.replace('/setbio', '').trim();
  if (!bio) {
    await ctx.reply('⚠️ الاستخدام: /setbio النبذة الجديدة');
    return;
  }

  try {
    await ctx.api.setMyDescription(bio);
    await ctx.reply(`✅ تم تغيير النبذة إلى: ${bio}`);
  } catch (error) {
    await ctx.reply('❌ فشل تغيير النبذة: ' + error.message);
  }
});

// أمر نشر قصة (للاستخدام الداخلي)
bot.command('story', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  const args = ctx.message.text.split(' ');
  if (args.length < 2) {
    await ctx.reply('⚠️ الاستخدام: /story [صورة|فيديو] [المدة]');
    return;
  }

  // هذا جزء توضيحي - يتطلب تكامل مع API تليجرام للأعمال
  await ctx.reply('📸 ميزة نشر القصص قيد التطوير...');
});

// ============================================================
// 💾 5. استرجاع الوسائط والمحذوفات (Anti-Delete & TTL Backup)
// ============================================================

// معالج الرسائل المحذوفة (للمحادثات الخاصة)
bot.on('deleted_business_messages', async (ctx) => {
  try {
    const deleted = ctx.deletedBusinessMessages;
    if (!deleted || !deleted.messages) return;

    for (const msg of deleted.messages) {
      const alert = `
🗑️ **رسالة محذوفة من الخاص**
👤 المستخدم: ${getUserName(msg.from)}
🆔 المعرف: ${msg.from.id}
⏰ وقت الحذف: ${new Date().toLocaleString()}

📝 المحتوى المحذوف:
${msg.text || 'رسالة وسائط'}
      `;
      
      await bot.api.sendMessage(ADMIN_ID, alert);
      
      // إذا كانت الرسالة تحتوي على وسائط، حفظ نسخة
      if (msg.photo || msg.video || msg.voice || msg.video_note) {
        // حفظ الوسائط إلى حساب الأدمن
        if (msg.photo) {
          const fileId = msg.photo[msg.photo.length - 1].file_id;
          await ctx.api.sendPhoto(ADMIN_ID, fileId);
        } else if (msg.video) {
          await ctx.api.sendVideo(ADMIN_ID, msg.video.file_id);
        } else if (msg.voice) {
          await ctx.api.sendVoice(ADMIN_ID, msg.voice.file_id);
        } else if (msg.video_note) {
          await ctx.api.sendVideoNote(ADMIN_ID, msg.video_note.file_id);
        }
      }
    }
  } catch (error) {
    console.error('❌ خطأ في معالج الرسائل المحذوفة:', error);
  }
});

// ============================================================
// 🛡️ 6. معالج الأخطاء العالمي
// ============================================================

bot.catch((error) => {
  console.error('❌ خطأ عالمي:', error);
});

// ============================================================
// 🚀 7. إعداد خادم Express و Webhook
// ============================================================

const app = express();
app.use(cors());
app.use(express.json());

// نقطة نهاية Webhook
app.use('/api/webhook', webhookCallback(bot, 'express'));

// صفحة رئيسية للتحقق من صحة الخادم
app.get('/api', (req, res) => {
  res.json({
    status: 'online',
    bot: 'Telegram Business Bot',
    version: '1.0.0',
    timestamp: new Date().toISOString()
  });
});

// معالج الأخطاء العالمي
app.use((err, req, res, next) => {
  console.error('❌ خطأ في الخادم:', err);
  res.status(500).json({ error: 'حدث خطأ داخلي في الخادم' });
});

// ============================================================
// 📤 8. تصدير الوظيفة الرئيسية لـ Vercel
// ============================================================

module.exports = app;

// ============================================================
// 📡 9. تعيين Webhook (يتم تشغيله مرة واحدة)
// ============================================================

async function setupWebhook() {
  try {
    const WEBHOOK_URL = process.env.WEBHOOK_URL || 'https://your-app.vercel.app/api/webhook';
    
    await bot.api.setWebhook(WEBHOOK_URL, {
      allowed_updates: [
        'message',
        'edited_message',
        'callback_query',
        'business_connection',
        'business_message',
        'deleted_business_messages',
        'managed_bot'
      ]
    });
    
    console.log('✅ Webhook تم تعيينه بنجاح');
    console.log(`🌐 الرابط: ${WEBHOOK_URL}`);
    
    // جلب معلومات البوت
    const me = await bot.api.getMe();
    console.log(`🤖 البوت: @${me.username}`);
    
  } catch (error) {
    console.error('❌ فشل تعيين Webhook:', error);
  }
}

// تشغيل إعداد Webhook عند بدء التشغيل
if (require.main === module) {
  setupWebhook();
}

// ============================================================
// 📝 ملاحظات هامة للتشغيل
// ============================================================
/*
🔧 إعدادات BotFather المطلوبة:
1. /setprivacy -> Disable (لقراءة جميع الرسائل)
2. /setjoingroups -> Enable (للانضمام للمجموعات)
3. تفعيل Secretary Mode من @BotFather

📱 إعدادات تليجرام للأعمال:
1. تفعيل "الرد على الرسائل" في الإعدادات
2. تعيين البوت كمساعد في الحساب التجاري

⚠️ ملاحظات الأمان:
- احفظ المتغيرات البيئية بشكل آمن
- استخدم HTTPS للـ Webhook
- قم بتحديث قوائم الكلمات الممنوعة بانتظام
*/
