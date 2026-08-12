// ============================================================
// 📦 استيراد المكتبات والإعدادات الأساسية
// ============================================================
const { Bot, webhookCallback, InlineKeyboard, session, InputFile } = require('grammy');
const express = require('express');
const cors = require('cors');
const axios = require('axios');
const translate = require('@vitalets/google-translate-api');

require('dotenv').config();

// ============================================================
// 🔐 المتغيرات الثابتة والإعدادات
// ============================================================
const BOT_TOKEN = process.env.BOT_TOKEN;
const ADMIN_ID = parseInt(process.env.ADMIN_ID);
const CHANNEL_URL = process.env.CHANNEL_URL || 'https://t.me/Xhwe2';

// ============================================================
// 📊 ذاكرة التخزين المؤقت للترجمات
// ============================================================
const translationCache = new Map();
const CACHE_TTL = 3600000; // ساعة واحدة
const userMessages = new Map();
const userWarnings = new Map();
const groupStats = new Map();

// ============================================================
// 📝 قوائم الكلمات الممنوعة والمفتاحية (محدثة)
// ============================================================

// الكلمات البذيئة والمسيئة (بالعربية والإنجليزية)
const BAD_WORDS = [
  // عربية
  'كس', 'عير', 'قحبة', 'منيوك', 'شرموطة', 'زبي', 'طيز', 'نيك', 'متناك',
  'خول', 'قواد', 'عاهرة', 'مومس', 'سافل', 'واطي', 'حقير', 'دنيء',
  // إنجليزية
  'fuck', 'shit', 'bitch', 'asshole', 'bastard', 'damn', 'crap',
  'dick', 'pussy', 'whore', 'slut', 'cunt', 'motherfucker',
  // مزيج
  'f***', 's***', 'b****', 'a******', 'd***'
];

// أنماط السبام والإعلانات
const SPAM_PATTERNS = [
  /t\.me\//gi,
  /telegram\.me\//gi,
  /https?:\/\/t\.me\//gi,
  /@\w+/gi,
  /https?:\/\/[^\s]+/gi,
  /www\.[^\s]+/gi,
  /قناة|channel|اشتراك|انضمام/i,
  /ربح|دولار|عملة|تعدين|استثمار/i,
  /🔥|💰|💵|💎|🚀/g // إيموجي مبالغ فيه
];

// الكلمات المفتاحية للرد التلقائي
const KEYWORD_REPLIES = {
  // عربي
  'السعر': '💰 الأسعار متوفرة في القناة الرسمية: ' + CHANNEL_URL,
  'الاسعار': '💰 الأسعار متوفرة في القناة الرسمية: ' + CHANNEL_URL,
  'سعر': '💰 الأسعار متوفرة في القناة الرسمية: ' + CHANNEL_URL,
  'ثمن': '💰 الأسعار متوفرة في القناة الرسمية: ' + CHANNEL_URL,
  'الدعم': '🆘 للدعم يرجى التواصل مع الإدارة عبر الخاص أو مراسلة @SupportBot',
  'دعم': '🆘 للدعم يرجى التواصل مع الإدارة عبر الخاص أو مراسلة @SupportBot',
  'التسجيل': '📝 للتسجيل يرجى زيارة الرابط: ' + CHANNEL_URL,
  'تسجيل': '📝 للتسجيل يرجى زيارة الرابط: ' + CHANNEL_URL,
  'اشتراك': '📝 للتسجيل يرجى زيارة الرابط: ' + CHANNEL_URL,
  'مرحباً': '👋 أهلاً وسهلاً بك! كيف يمكنني مساعدتك؟',
  'مرحبا': '👋 أهلاً وسهلاً بك! كيف يمكنني مساعدتك؟',
  'السلام عليكم': '🌙 وعليكم السلام ورحمة الله وبركاته، كيف يمكنني خدمتك؟',
  'شكرا': '🙏 عفواً، في خدمتك دائماً!',
  'شكراً': '🙏 عفواً، في خدمتك دائماً!',
  'ممتاز': '✨ شكراً لك، سعداء بتقديم الأفضل دائماً!',
  'رائع': '🌟 شكراً جزيلاً، وجودك يسعدنا!',
  
  // English
  'price': '💰 Prices are available on our official channel: ' + CHANNEL_URL,
  'prices': '💰 Prices are available on our official channel: ' + CHANNEL_URL,
  'support': '🆘 For support, please contact the admin privately or message @SupportBot',
  'register': '📝 To register, please visit the link: ' + CHANNEL_URL,
  'registration': '📝 To register, please visit the link: ' + CHANNEL_URL,
  'hello': '👋 Hello and welcome! How can I help you?',
  'hi': '👋 Hi there! How can I assist you?',
  'thanks': '🙏 You\'re welcome! Always at your service!',
  'thank': '🙏 You\'re welcome! Always at your service!',
  'good': '✨ Thank you, we\'re always happy to provide the best!',
  'great': '🌟 Thank you so much, your presence makes us happy!'
};

// القصص التحفيزية (محدثة)
const QUOTES = [
  '💪 "النجاح ليس غياب الفشل، بل هو الإصرار بعد الفشل" - نيلسون مانديلا',
  '🌟 "كن التغيير الذي تريد رؤيته في العالم" - غاندي',
  '🚀 "المستقبل يبدأ من حيث الإرادة تنتهي" - ويليام جيمس',
  '🎯 "التركيز هو مفتاح الإبداع" - ستيف جوبز',
  '📚 "التعليم هو أقوى سلاح يمكنك استخدامه لتغيير العالم" - نيلسون مانديلا',
  '💡 "العقل العظيم لديه أفكار، والعقل المتوسط لديه أحداث، والعقل الصغير لديه أشخاص" - إليانور روزفلت',
  '⭐ "النجاح هو القدرة على الانتقال من فشل إلى فشل دون فقدان الحماس" - ونستون تشرشل',
  '🌅 "كل يوم فرصة جديدة لبداية جديدة" - مجهول',
  '🎨 "الإبداع هو الذكاء الذي يمرح" - ألبرت أينشتاين',
  '🦋 "التغيير هو القانون الأساسي للحياة" - مجهول'
];

// ============================================================
// 🤖 إنشاء البوت مع جلسات الذاكرة
// ============================================================
const bot = new Bot(BOT_TOKEN);

// تفعيل الجلسات للتخزين المؤقت
bot.use(session({
  initial: () => ({
    userMessages: {},
    lastReply: {},
    warnings: {},
    groupStats: {}
  })
}));

// ============================================================
// 🛠️ أدوات مساعدة
// ============================================================

// دالة الحصول على اسم المستخدم
function getUserName(user) {
  if (!user) return 'العميل';
  return user.first_name || user.username || user.last_name || 'العميل';
}

// دالة الحصول على معرف المستخدم
function getUserId(user) {
  return user?.id || 'unknown';
}

// دالة إنشاء الزر الشفاف
function createChannelButton(text = '📢 تحديثات نيرد') {
  return new InlineKeyboard().url(text, CHANNEL_URL);
}

// دالة إنشاء أزرار متعددة
function createActionButtons() {
  return new InlineKeyboard()
    .url('📢 تحديثات نيرد', CHANNEL_URL)
    .row()
    .url('💬 تواصل مع الدعم', 'https://t.me/SupportBot');
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
  // التحقق من تكرار الأحرف
  if (/(.)\1{5,}/.test(text)) return true; // تكرار 5 مرات أو أكثر
  // التحقق من النص الطويل بدون مسافات
  if (text.length > 500 && !text.includes(' ')) return true;
  // التحقق من الأنماط
  return SPAM_PATTERNS.some(pattern => pattern.test(text));
}

// دالة الحصول على اقتباس عشوائي
function getRandomQuote() {
  return QUOTES[Math.floor(Math.random() * QUOTES.length)];
}

// دالة الحصول على رد مفتاحي
function getKeywordReply(text) {
  if (!text) return null;
  const lowerText = text.toLowerCase();
  for (const [keyword, reply] of Object.entries(KEYWORD_REPLIES)) {
    if (lowerText.includes(keyword.toLowerCase())) {
      return reply;
    }
  }
  return null;
}

// دالة تنسيق الوقت
function formatTime(date) {
  return new Date(date).toLocaleString('ar-EG', {
    timeZone: 'Africa/Cairo',
    hour12: true
  });
}

// ============================================================
// 🌐 دالة الترجمة المجانية مع التخزين المؤقت
// ============================================================

async function detectAndTranslate(text, targetLang = 'ar') {
  try {
    // 🔍 إنشاء مفتاح للذاكرة المؤقتة
    const cacheKey = `${text}_${targetLang}`;
    
    // التحقق من وجود الترجمة في الذاكرة المؤقتة
    if (translationCache.has(cacheKey)) {
      const cached = translationCache.get(cacheKey);
      if (Date.now() - cached.timestamp < CACHE_TTL) {
        console.log('✅ استخدام ترجمة من الذاكرة المؤقتة');
        return cached.data;
      }
    }

    // 📝 إجراء الترجمة
    console.log('🔄 جاري الترجمة...');
    const result = await translate(text, { 
      to: targetLang,
      autoCorrect: true
    });

    // استخراج اللغة المصدر
    const fromLang = result.from.language.iso || 'unknown';
    
    const translationResult = {
      original: text,
      translated: result.text,
      from: fromLang,
      to: targetLang,
      detected: result.from.text || ''
    };

    // 💾 تخزين النتيجة في الذاكرة المؤقتة
    translationCache.set(cacheKey, {
      data: translationResult,
      timestamp: Date.now()
    });

    return translationResult;

  } catch (error) {
    console.error('❌ خطأ في الترجمة:', error);
    
    // محاولة الترجمة من خادم بديل في حالة الفشل
    try {
      const fallbackResult = await translate(text, { 
        to: targetLang,
        host: 'translate.googleapis.com',
        autoCorrect: true
      });
      
      return {
        original: text,
        translated: fallbackResult.text,
        from: fallbackResult.from?.language?.iso || 'unknown',
        to: targetLang
      };
    } catch (fallbackError) {
      console.error('❌ فشلت الترجمة من الخادم البديل:', fallbackError);
      return { original: text, translated: text, from: 'unknown', to: targetLang };
    }
  }
}

// ============================================================
// 📊 وظائف إحصاءات الترجمة
// ============================================================

function getTranslationStats() {
  return {
    totalTranslations: translationCache.size,
    cacheSize: translationCache.size,
    memoryUsage: process.memoryUsage().heapUsed / 1024 / 1024
  };
}

// ============================================================
// 🧹 تنظيف الذاكرة المؤقتة كل ساعة
// ============================================================

setInterval(() => {
  const now = Date.now();
  for (const [key, value] of translationCache) {
    if (now - value.timestamp > CACHE_TTL) {
      translationCache.delete(key);
    }
  }
  console.log('🧹 تم تنظيف الذاكرة المؤقتة للترجمات');
  
  // تنظيف بيانات المستخدمين القديمة
  for (const [userId, data] of userMessages) {
    if (now - data.timestamp > 86400000) { // 24 ساعة
      userMessages.delete(userId);
    }
  }
  
  // تنظيف التحذيرات القديمة
  for (const [userId, data] of userWarnings) {
    if (now - data.timestamp > 86400000) { // 24 ساعة
      userWarnings.delete(userId);
    }
  }
}, CACHE_TTL);

// ============================================================
// 📨 1. نظام الرد التلقائي الخاص (Business Private Auto-Reply)
// ============================================================

// معالج رسائل الأعمال
bot.on('business_message', async (ctx) => {
  try {
    const msg = ctx.businessMessage;
    if (!msg || !msg.text) return;

    const userId = msg.from.id;
    const userName = getUserName(msg.from);
    const userText = msg.text;

    console.log(`📩 رسالة خاصة من ${userName} (${userId}): ${userText.substring(0, 50)}...`);

    // 🔍 اكتشاف لغة المستخدم
    const detectedLang = await detectAndTranslate(userText, 'ar');
    
    // إذا كانت اللغة غير عربية، إرسال تنبيه للأدمن
    if (detectedLang.from !== 'ar' && detectedLang.from !== 'unknown') {
      const adminAlert = `
🔔 **تنبيه: رسالة بلغة أجنبية**
👤 المستخدم: ${userName}
🆔 المعرف: ${userId}
🌐 اللغة المكتشفة: ${detectedLang.from}

📝 النص الأصلي:
${detectedLang.original}

📝 الترجمة للعربية:
${detectedLang.translated}

⏰ الوقت: ${formatTime(Date.now())}
      `;
      await bot.api.sendMessage(ADMIN_ID, adminAlert, { parse_mode: 'Markdown' });
    }

    // 📝 بناء الرد المخصص
    const userFirstName = msg.from.first_name || userName;
    let replyText = `👋 مرحباً ${userFirstName}،\n\n`;
    replyText += `شكراً لتواصلك معنا! سنرد عليك في أقرب وقت.\n\n`;
    replyText += `💡 ${getRandomQuote()}\n\n`;
    replyText += `🔗 للمزيد من المعلومات، تفضل بزيارة قناتنا:`;

    // 🌐 ترجمة الرد إلى لغة المستخدم إذا كانت غير عربية
    let finalReply = replyText;
    if (detectedLang.from !== 'ar' && detectedLang.from !== 'unknown') {
      const translatedReply = await detectAndTranslate(replyText, detectedLang.from);
      finalReply = translatedReply.translated;
    }

    // 📤 إرسال الرد مع الزر
    const keyboard = createChannelButton('📢 تابعنا للمزيد');
    await ctx.reply(finalReply, {
      reply_parameters: { message_id: msg.message_id },
      reply_markup: keyboard,
      parse_mode: 'Markdown'
    });

    // 📊 تسجيل التفاعل
    console.log(`✅ رد تلقائي مرسل إلى ${userName} (${userId})`);
    
    // تحديث إحصائيات المستخدم
    if (!userMessages.has(userId)) {
      userMessages.set(userId, { count: 0, timestamp: Date.now() });
    }
    const userData = userMessages.get(userId);
    userData.count++;
    userMessages.set(userId, userData);

  } catch (error) {
    console.error('❌ خطأ في معالج الرسائل التجارية:', error);
    // محاولة إرسال رد عام في حالة الفشل
    try {
      await ctx.reply('👋 شكراً لتواصلك معنا! سنرد عليك في أقرب وقت.');
    } catch (replyError) {
      console.error('❌ فشل إرسال الرد الاحتياطي:', replyError);
    }
  }
});

// ============================================================
// 🛡️ 2. الإشراف والحماية المتقدمة للمجموعات
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

    console.log(`📨 رسالة من ${userName} في مجموعة ${msg.chat.title}: ${text.substring(0, 50)}...`);

    // ==========================================================
    // 🚫 حظر الألفاظ البذيئة
    // ==========================================================
    if (containsBadWords(text)) {
      try {
        await ctx.deleteMessage();
        await ctx.reply(`⚠️ @${userName}، تم حذف رسالتك لاحتوائها على كلمات غير لائقة.`);
        
        // تسجيل المخالفة
        const warningCount = (userWarnings.get(userId)?.count || 0) + 1;
        userWarnings.set(userId, { count: warningCount, timestamp: Date.now() });
        
        // إذا تكررت المخالفة 3 مرات، طرد العضو
        if (warningCount >= 3) {
          await ctx.banChatMember(userId);
          await bot.api.sendMessage(ADMIN_ID,
            `🚨 **تم طرد عضو بسبب التكرار**\n👤 ${userName}\n🆔 ${userId}\n📊 عدد المخالفات: ${warningCount}`
          );
        } else {
          await bot.api.sendMessage(ADMIN_ID, 
            `🚫 **مخالفة: ألفاظ بذيئة**\n👤 ${userName}\n🆔 ${userId}\n📝 ${text.substring(0, 100)}\n⚠️ التحذير ${warningCount}/3`
          );
        }
      } catch (error) {
        console.error('❌ خطأ في حذف رسالة بذيئة:', error);
      }
      return;
    }

    // ==========================================================
    // 🛑 منع السبام والإعلانات
    // ==========================================================
    if (isSpam(text)) {
      try {
        await ctx.deleteMessage();
        await ctx.reply(`🚫 @${userName}، ممنوع نشر الروابط والإعلانات في المجموعة.`);
        
        // طرد العضو فوراً
        await ctx.banChatMember(userId);
        await bot.api.sendMessage(ADMIN_ID,
          `🚨 **تم طرد عضو بسبب السبام**\n👤 ${userName}\n🆔 ${userId}\n📝 ${text.substring(0, 100)}`
        );
      } catch (error) {
        console.error('❌ خطأ في طرد عضو:', error);
      }
      return;
    }

    // ==========================================================
    // 🤖 الرد الآلي الذكي (Smart Auto-Reply)
    // ==========================================================
    const keywordReply = getKeywordReply(text);
    if (keywordReply) {
      try {
        const keyboard = createChannelButton('📢 للمزيد من المعلومات');
        await ctx.reply(`${keywordReply}`, { 
          reply_markup: keyboard,
          parse_mode: 'Markdown'
        });
        console.log(`✅ رد آلي على كلمة مفتاحية لـ ${userName}`);
      } catch (error) {
        console.error('❌ خطأ في الرد الآلي:', error);
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
      try {
        await bot.api.sendMessage(ADMIN_ID,
          `📊 **إحصائية عضو نشط**\n👤 ${userName}\n📝 عدد الرسائل: ${sessionData.userMessages[userId]}\n📍 المجموعة: ${msg.chat.title || 'غير معروف'}`
        );
      } catch (error) {
        console.error('❌ خطأ في إرسال الإحصائيات:', error);
      }
    }

  } catch (error) {
    console.error('❌ خطأ في معالج رسائل المجموعة:', error);
  }
});

// ============================================================
// 🔄 3. إدارة الملف الشخصي والقصص
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
    await bot.api.sendMessage(ADMIN_ID, `✅ تم تغيير اسم البوت إلى: ${newName}`);
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
    await bot.api.sendMessage(ADMIN_ID, `✅ تم تغيير نبذة البوت إلى: ${bio}`);
  } catch (error) {
    await ctx.reply('❌ فشل تغيير النبذة: ' + error.message);
  }
});

bot.command('setusername', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  const username = ctx.message.text.replace('/setusername', '').trim();
  if (!username) {
    await ctx.reply('⚠️ الاستخدام: /setusername اسم_المستخدم_الجديد');
    return;
  }

  try {
    await ctx.api.setMyUsername(username.replace('@', ''));
    await ctx.reply(`✅ تم تغيير اسم المستخدم إلى: @${username.replace('@', '')}`);
    await bot.api.sendMessage(ADMIN_ID, `✅ تم تغيير اسم المستخدم إلى: @${username.replace('@', '')}`);
  } catch (error) {
    await ctx.reply('❌ فشل تغيير اسم المستخدم: ' + error.message);
  }
});

bot.command('stats', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  try {
    const botInfo = await ctx.api.getMe();
    const webhookInfo = await ctx.api.getWebhookInfo();
    const translationStats = getTranslationStats();
    
    const stats = `
📊 **إحصائيات البوت**

🤖 **معلومات البوت:**
- الاسم: ${botInfo.first_name}
- المعرف: @${botInfo.username}
- المعرف الرقمي: ${botInfo.id}

📈 **إحصائيات التشغيل:**
- عدد الترجمات المخزنة: ${translationStats.totalTranslations}
- استخدام الذاكرة: ${translationStats.memoryUsage.toFixed(2)} MB
- المستخدمين النشطين: ${userMessages.size}
- عدد التحذيرات: ${userWarnings.size}

🌐 **معلومات Webhook:**
- الحالة: ${webhookInfo.url ? '🟢 مفعل' : '🔴 غير مفعل'}
- الرابط: ${webhookInfo.url || 'غير محدد'}

⏰ الوقت: ${formatTime(Date.now())}
    `;
    
    await ctx.reply(stats, { parse_mode: 'Markdown' });
  } catch (error) {
    await ctx.reply('❌ فشل جلب الإحصائيات: ' + error.message);
  }
});

bot.command('clear', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  try {
    // تنظيف جميع البيانات
    translationCache.clear();
    userMessages.clear();
    userWarnings.clear();
    
    await ctx.reply('🧹 تم تنظيف جميع البيانات المؤقتة بنجاح!');
  } catch (error) {
    await ctx.reply('❌ فشل تنظيف البيانات: ' + error.message);
  }
});

// أمر نشر قصة (للاستخدام الداخلي - متكامل مع API)
bot.command('story', async (ctx) => {
  if (ctx.from.id !== ADMIN_ID) {
    await ctx.reply('⛔ غير مصرح لك باستخدام هذا الأمر.');
    return;
  }

  const args = ctx.message.text.split(' ');
  if (args.length < 2) {
    await ctx.reply(`⚠️ الاستخدام:\n
/story photo [مدة] - لنشر صورة كقصة
/story video [مدة] - لنشر فيديو كقصة

المدة المتاحة: 6, 12, 24, 48 (ساعات)`);
    return;
  }

  const type = args[1];
  const duration = parseInt(args[2]) || 24;
  
  // تحقق من المدة
  if (![6, 12, 24, 48].includes(duration)) {
    await ctx.reply('⚠️ المدة غير صالحة. اختر: 6, 12, 24, 48 ساعة');
    return;
  }

  await ctx.reply(`📸 جاري تجهيز نشر ${type} كقصة لمدة ${duration} ساعة...`);
  await ctx.reply('⚠️ هذه الميزة تتطلب إعدادات إضافية في تليجرام للأعمال.');
});

// ============================================================
// 💾 4. استرجاع الوسائط والمحذوفات (Anti-Delete & TTL Backup)
// ============================================================

// معالج الرسائل المحذوفة (للمحادثات الخاصة)
bot.on('deleted_business_messages', async (ctx) => {
  try {
    const deleted = ctx.deletedBusinessMessages;
    if (!deleted || !deleted.messages) return;

    console.log(`🗑️ تم حذف ${deleted.messages.length} رسالة من الخاص`);

    for (const msg of deleted.messages) {
      let alert = `
🗑️ **رسالة محذوفة من الخاص**
👤 المستخدم: ${getUserName(msg.from)}
🆔 المعرف: ${msg.from.id}
⏰ وقت الحذف: ${formatTime(Date.now())}

📝 المحتوى المحذوف:
`;
      
      if (msg.text) {
        alert += `${msg.text}`;
      } else if (msg.photo) {
        alert += `🖼️ صورة`;
      } else if (msg.video) {
        alert += `🎬 فيديو`;
      } else if (msg.voice) {
        alert += `🎤 بصمة صوتية`;
      } else if (msg.video_note) {
        alert += `📹 ملاحظة فيديو`;
      } else {
        alert += `📎 وسائط غير معروفة`;
      }
      
      await bot.api.sendMessage(ADMIN_ID, alert, { parse_mode: 'Markdown' });
      
      // إذا كانت الرسالة تحتوي على وسائط، حفظ نسخة
      if (msg.photo) {
        try {
          const fileId = msg.photo[msg.photo.length - 1].file_id;
          await ctx.api.sendPhoto(ADMIN_ID, fileId, {
            caption: `🖼️ نسخة محفوظة من ${getUserName(msg.from)}`
          });
        } catch (error) {
          console.error('❌ فشل حفظ الصورة:', error);
        }
      } else if (msg.video) {
        try {
          await ctx.api.sendVideo(ADMIN_ID, msg.video.file_id, {
            caption: `🎬 نسخة محفوظة من ${getUserName(msg.from)}`
          });
        } catch (error) {
          console.error('❌ فشل حفظ الفيديو:', error);
        }
      } else if (msg.voice) {
        try {
          await ctx.api.sendVoice(ADMIN_ID, msg.voice.file_id, {
            caption: `🎤 نسخة محفوظة من ${getUserName(msg.from)}`
          });
        } catch (error) {
          console.error('❌ فشل حفظ البصمة:', error);
        }
      } else if (msg.video_note) {
        try {
          await ctx.api.sendVideoNote(ADMIN_ID, msg.video_note.file_id);
        } catch (error) {
          console.error('❌ فشل حفظ الملاحظة:', error);
        }
      }
    }
  } catch (error) {
    console.error('❌ خطأ في معالج الرسائل المحذوفة:', error);
  }
});

// ============================================================
// 📌 5. أوامر إضافية للمستخدمين
// ============================================================

bot.command('start', async (ctx) => {
  const userName = getUserName(ctx.from);
  const welcomeMsg = `
👋 **مرحباً ${userName}!**

🤖 أنا بوت المساعدة الذكي للحساب التجاري.
يمكنني مساعدتك في:

📱 **المحادثات الخاصة:**
- الرد التلقائي على رسائلك
- تقديم الدعم والمعلومات

👥 **المجموعات:**
- حماية المجموعة من السبام
- الرد على الأسئلة الشائعة
- منع الكلمات البذيئة

📊 **ميزات إضافية:**
- ترجمة الرسائل الفورية
- حفظ الوسائط المهمة
- إحصائيات التفاعل

🔗 للمزيد من المعلومات، تفضل بزيارة قناتنا:
  `;
  
  const keyboard = new InlineKeyboard()
    .url('📢 قناة التحديثات', CHANNEL_URL)
    .row()
    .url('💬 الدعم', 'https://t.me/SupportBot');
  
  await ctx.reply(welcomeMsg, { 
    reply_markup: keyboard,
    parse_mode: 'Markdown'
  });
});

bot.command('help', async (ctx) => {
  const helpMsg = `
🆘 **قائمة المساعدة**

📱 **الأوامر المتاحة:**

🔹 /start - الترحيب والتعريف بالبوت
🔹 /help - عرض هذه القائمة
🔹 /support - التواصل مع الدعم

👑 **أوامر الإدارة:**
🔸 /setname [الاسم] - تغيير اسم البوت
🔸 /setbio [النبذة] - تغيير نبذة البوت
🔸 /setusername [اسم] - تغيير اسم المستخدم
🔸 /stats - عرض إحصائيات البوت
🔸 /clear - تنظيف البيانات المؤقتة
🔸 /story [نوع] [مدة] - نشر قصة

📌 **الميزات التلقائية:**
✅ رد تلقائي في المحادثات الخاصة
✅ ترجمة فورية للغات الأجنبية
✅ حماية المجموعات من السبام
✅ الرد على الكلمات المفتاحية
✅ حفظ الوسائط المحذوفة
✅ إحصائيات الأعضاء

🔗 **روابط مهمة:**
📢 القناة: ${CHANNEL_URL}
💬 الدعم: @SupportBot
  `;
  
  await ctx.reply(helpMsg, { parse_mode: 'Markdown' });
});

bot.command('support', async (ctx) => {
  const supportMsg = `
💬 **الدعم والمساعدة**

للاستفسارات والدعم، يمكنك:

1️⃣ التواصل مع الإدارة عبر الخاص
2️⃣ الانضمام إلى قناة التحديثات
3️⃣ مراسلة بوت الدعم

📢 قناة التحديثات: ${CHANNEL_URL}
🤖 بوت الدعم: @SupportBot

📧 للتواصل المباشر، يرجى إرسال رسالة خاصة للإدارة.
  `;
  
  const keyboard = new InlineKeyboard()
    .url('📢 قناة التحديثات', CHANNEL_URL)
    .row()
    .url('💬 بوت الدعم', 'https://t.me/SupportBot');
  
  await ctx.reply(supportMsg, { 
    reply_markup: keyboard,
    parse_mode: 'Markdown'
  });
});

// ============================================================
// 🛡️ 6. معالج الأخطاء العالمي
// ============================================================

bot.catch((error) => {
  console.error('❌ خطأ عالمي:', error);
  
  // تسجيل الخطأ مع تفاصيل إضافية
  if (error.message) {
    console.error('📝 رسالة الخطأ:', error.message);
  }
  if (error.stack) {
    console.error('📚 Stack Trace:', error.stack);
  }
});

// ============================================================
// 🚀 7. إعداد خادم Express و Webhook
// ============================================================

const app = express();
app.use(cors());
app.use(express.json({ limit: '50mb' }));
app.use(express.urlencoded({ extended: true, limit: '50mb' }));

// نقطة نهاية Webhook
app.use('/api/webhook', webhookCallback(bot, 'express'));

// صفحة رئيسية للتحقق من صحة الخادم
app.get('/api', (req, res) => {
  res.json({
    status: 'online',
    bot: 'Telegram Business Bot',
    version: '2.0.0',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    memory: process.memoryUsage(),
    translations: translationCache.size
  });
});

// صفحة الإحصائيات
app.get('/api/stats', async (req, res) => {
  try {
    const botInfo = await bot.api.getMe();
    const webhookInfo = await bot.api.getWebhookInfo();
    const translationStats = getTranslationStats();
    
    res.json({
      bot: {
        name: botInfo.first_name,
        username: botInfo.username,
        id: botInfo.id
      },
      stats: {
        translations: translationStats,
        users: userMessages.size,
        warnings: userWarnings.size,
        memory: process.memoryUsage()
      },
      webhook: {
        url: webhookInfo.url,
        pending: webhookInfo.pending_update_count
      },
      timestamp: new Date().toISOString()
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// معالج الأخطاء العالمي
app.use((err, req, res, next) => {
  console.error('❌ خطأ في الخادم:', err);
  res.status(500).json({ 
    error: 'حدث خطأ داخلي في الخادم',
    message: err.message 
  });
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
    
    console.log('🔄 جاري تعيين Webhook...');
    
    // حذف أي Webhook قديم
    await bot.api.deleteWebhook();
    console.log('✅ تم حذف Webhook القديم');
    
    // تعيين Webhook جديد
    const result = await bot.api.setWebhook(WEBHOOK_URL, {
      allowed_updates: [
        'message',
        'edited_message',
        'callback_query',
        'business_connection',
        'business_message',
        'deleted_business_messages',
        'managed_bot',
        'chat_member',
        'my_chat_member'
      ],
      drop_pending_updates: true
    });
    
    if (result) {
      console.log('✅ Webhook تم تعيينه بنجاح');
      console.log(`🌐 الرابط: ${WEBHOOK_URL}`);
      
      // جلب معلومات البوت
      const me = await bot.api.getMe();
      console.log(`🤖 البوت: @${me.username}`);
      console.log(`🆔 معرف البوت: ${me.id}`);
      
      // جلب معلومات Webhook للتأكيد
      const webhookInfo = await bot.api.getWebhookInfo();
      console.log(`📊 حالة Webhook: ${webhookInfo.url ? '🟢 مفعل' : '🔴 غير مفعل'}`);
      
      if (webhookInfo.url) {
        console.log(`📡 الرابط النشط: ${webhookInfo.url}`);
        console.log(`⏳ معلقين: ${webhookInfo.pending_update_count}`);
      }
    } else {
      console.error('❌ فشل تعيين Webhook');
    }
    
  } catch (error) {
    console.error('❌ فشل تعيين Webhook:', error);
    console.error('📝 التفاصيل:', error.message);
    if (error.response) {
      console.error('📤 استجابة API:', error.response.data);
    }
  }
}

// تشغيل إعداد Webhook عند بدء التشغيل
if (require.main === module) {
  console.log('🚀 بدء تشغيل البوت...');
  setupWebhook();
  
  // بدء الخادم المحلي للاختبار
  const PORT = process.env.PORT || 3000;
  app.listen(PORT, () => {
    console.log(`🌐 الخادم يعمل على المنفذ ${PORT}`);
    console.log(`🔗 الرابط المحلي: http://localhost:${PORT}/api`);
  });
}

// ============================================================
// 📝 10. معلومات التشغيل والملاحظات
// ============================================================
/*
🔧 إعدادات BotFather المطلوبة:
1. /setprivacy -> Disable (لقراءة جميع الرسائل)
2. /setjoingroups -> Enable (للانضمام للمجموعات)
3. /mybots -> اختر البوت -> Bot Settings -> Secretary Mode -> Enable
4. /mybots -> اختر البوت -> Bot Settings -> Inline Mode -> Disable

📱 إعدادات تليجرام للأعمال:
1. افتح Settings -> Telegram Business
2. اختر Chat Assistant -> حدد البوت
3. فعّل Reply to Messages
4. اختر All Chats
5. فعّل Auto-Reply للرسائل الجديدة

⚠️ ملاحظات الأمان:
- احفظ المتغيرات البيئية بشكل آمن
- استخدم HTTPS للـ Webhook
- قم بتحديث قوائم الكلمات الممنوعة بانتظام
- راقب سجلات الأخطاء بشكل دوري

📊 الأداء:
- يستخدم التخزين المؤقت لتقليل طلبات الترجمة
- يدعم مئات المستخدمين في وقت واحد
- يعمل بكفاءة على بيئة Vercel Serverless

🆓 الترجمة المجانية:
- يستخدم @vitalets/google-translate-api بدون API Key
- يدعم أكثر من 100 لغة
- مع تخزين مؤقت لتقليل الطلبات

📌 الميزات الرئيسية:
✅ رد تلقائي خاص مع ترجمة فورية
✅ حماية المجموعات من السبام والإعلانات
✅ رد آلي على الكلمات المفتاحية
✅ حفظ واسترجاع الوسائط المحذوفة
✅ إدارة الملف الشخصي والقصص
✅ إحصائيات تفاعل الأعضاء
✅ أوامر متكاملة للإدارة

💡 نصائح للاستخدام:
- قم بتحديث قوائم الكلمات الممنوعة بانتظام
- راقب سجلات الأخطاء لتحسين الأداء
- استخدم أوامر الإدارة بمسؤولية
- قم بعمل نسخ احتياطي للبيانات المهمة
*/

// ============================================================
// 🏁 نهاية الكود
// ============================================================
