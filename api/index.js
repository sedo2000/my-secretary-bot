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
if (!token) {
  console.error("❌ تحذير: TELEGRAM_BOT_TOKEN غير متوفر في متغيرات البيئة!");
}

const bot = new Bot(token || "DUMMY_TOKEN");

// الذاكرة المؤقتة لإدارة الجلسات والإعدادات (In-Memory State & Cache)
const memoryCache = new Map();

// معالج الأخطاء العالمي لمنع الانهيار
bot.catch((err) => {
  console.error("❌ حدث خطأ داخلي في معالجة البوت:", err.error || err);
});

// 2️⃣ القواميس والقواعد الأساسية
const badWords = [
  "كحبة", "مطي", "قندرة", "ساقط", "فرخ", "عير", "كس", "طيز", "زنيّم",
  "نگوة", "جحش", "منيوك", "قواد", "عاهرة", "شرموطة"
];

const smartAnswers = {
  السعر: "ℹ️ لمعرفة الأسعار والتفاصيل الكاملة، يمكنك زيارة القناة الرسمية أو مراسلة الدعم الفني المباشر.",
  الدعم: "🛠️ للتواصل مع فريق الدعم الفني، يرجى مراسلة الحساب التجاري المباشر وسيتم الرد قريباً.",
  التسجيل: "📝 يمكنك التسجيل والاشتراك عبر فتح المحادثة الخاصة واتباع التعليمات البرمجية بدقة.",
  الاشتراك: "💳 لمعرفة تفاصيل الاشتراكات المتاحة، يرجى مراجعة القناة الرسمية للتحديثات.",
  الموقع: "🌐 تابع كافة مستجداتنا وروابطنا عبر قناة التحديثات الرسمية المدرجة أدناه."
};

const quotes = [
  "قاوم ما تكره لتصل الى ما تحب",
  "الحرب بين أنت ضد أنت",
  "أبنِ نفسك بنفسك لنفسك",
  "ميخالف، عابر سبيل ستمر كل الصعاب",
  "حتى لو متأخر تگدر تبدأ من جديد..!",
  "من يعيش في خوف لن يكون حراً ابداً في حياته",
  "لا أبرح حتى أبلغ المبتغى أو أموت دونه",
  "أنه مبرمج فحسب، يصنع واقعه بيديه",
  "المرء نتاج خلواته وتأملاته العميقة",
  "لا مزيد من الأصدقاء المزيفين، الجودة تكمن في القلة"
];

// 3️⃣ نظام اللغات المتكامل (Localization System)
const i18n = {
  ar: {
    welcome: "أهلاً بك في لوحة تحكم سكرتير الحساب التجاري الشاملة 🤖\nاختر من الأزرار أدناه للتحكم بكافة الميزات:",
    main_menu: "القائمة الرئيسية للتحكم 🤖:",
    stop_btn: "🛑 إيقاف الرد الخاص",
    start_btn: "🟢 تشغيل الرد الخاص",
    edit_text_btn: "📝 تعديل نص الرد الخاص",
    exclude_btn: "👤 استثناء حساب محدد",
    list_excluded_btn: "📋 عرض الحسابات المستثناة",
    clear_excluded_btn: "🧹 مسح قائمة الاستثناءات",
    profile_btn: "🧑 إدارة الملف الشخصي التجاري",
    story_btn: "📖 نشر قصة (Story)",
    lang_ar_btn: "🇮🇶 العربية",
    lang_en_btn: "🇺🇸 English",
    back_btn: "🔙 رجوع للقائمة الرئيسية",
    stopped: "🛑 تم إيقاف الرد التلقائي الخاص بنجاح.",
    started: "🟢 تم تشغيل الرد التلقائي الخاص بنجاح.",
    edit_prompt: "📝 أرسل الآن نص الرد التلقائي الجديد للخاص (يمكنك استخدام متغير {name} لاسم العميل):",
    saved_text: "✅ تم حفظ نص الرد التلقائي الجديد وتطبيقه بنجاح!",
    exclude_prompt: "👤 أرسل آيدي (ID) الحساب المراد استثناؤه من الرد التلقائي:",
    id_added: "✅ تم إضافة الايدي المخصص إلى قائمة الاستثناءات بنجاح.",
    no_excluded: "لا توجد أي حسابات مستثناة حالياً في القائمة.",
    cleared_excluded: "🧹 تم مسح قائمة الاستثناءات بالكامل.",
    profile_menu: "🧑 إدارة الملف الشخصي التجاري - اختر الإعداد المطلوب تعديله:",
    edit_name: "✏️ تعديل اسم الحساب التجاري",
    edit_bio: "📝 تعديل النبذة التعريفية (Bio)",
    edit_photo: "🖼️ تعديل صورة الملف الشخصي",
    edit_username: "🔗 تعديل اسم المستخدم (Username)",
    story_dur_title: "⏱️ اختر مدة ظهور القصة المطلوب نشرها:",
    dur_6h: "6 ساعات",
    dur_12h: "12 ساعة",
    dur_24h: "24 ساعة",
    dur_48h: "48 ساعة",
    story_prompt: "📖 أرسل الآن صورة أو فيديو لنشره كقصة عبر الحساب التجاري (المدة المحددة: %s):",
    story_success: "✅ تم نشر القصة بنجاح عبر Telegram Business Stories!",
    no_biz_conn: "❌ لم يتم ربط حساب تجاري نشط بعد بالبوت. تأكد من إعدادات تليجرام للأعمال.",
    ttl_media_alert: "🔥 *تم حفظ نسخة احتياطية من وسائط واردة (مؤقتة/ذاتية التدمير)*\n👤 المرسل: %s (`%d`)",
    deleted_alert: "🗑️ *تنبيه أمني: قام العميل بحذف رسالة أو وسائط!*\n👤 العميل: %s (`%d`)"
  },
  en: {
    welcome: "Welcome to Business Secretary Control Panel 🤖\nChoose an option below to manage all features:",
    main_menu: "Main Control Menu 🤖:",
    stop_btn: "🛑 Stop Auto-Reply",
    start_btn: "🟢 Start Auto-Reply",
    edit_text_btn: "📝 Edit Auto-Reply Text",
    exclude_btn: "👤 Exclude Account ID",
    list_excluded_btn: "📋 View Excluded List",
    clear_excluded_btn: "🧹 Clear Exclusions",
    profile_btn: "🧑 Manage Business Profile",
    story_btn: "📖 Post Business Story",
    lang_ar_btn: "🇮🇶 العربية",
    lang_en_btn: "🇺🇸 English",
    back_btn: "🔙 Back to Main Menu",
    stopped: "🛑 Auto-reply has been stopped successfully.",
    started: "🟢 Auto-reply has been started successfully.",
    edit_prompt: "📝 Send the new auto-reply text now (you can use {name} variable):",
    saved_text: "✅ New auto-reply text saved successfully!",
    exclude_prompt: "👤 Send the Account ID to exclude from auto-replies:",
    id_added: "✅ Account ID added to exclusion list successfully.",
    no_excluded: "No excluded accounts currently found.",
    cleared_excluded: "🧹 All exclusion lists cleared successfully.",
    profile_menu: "🧑 Manage Business Profile - Select setting to modify:",
    edit_name: "✏️ Edit Business Name",
    edit_bio: "📝 Edit Business Bio",
    edit_photo: "🖼️ Edit Profile Photo",
    edit_username: "🔗 Edit Username",
    story_dur_title: "⏱️ Select Story Display Duration:",
    dur_6h: "6 Hours",
    dur_12h: "12 Hours",
    dur_24h: "24 Hours",
    dur_48h: "48 Hours",
    story_prompt: "📖 Send a photo or video to post as a story (Duration: %s):",
    story_success: "✅ Story posted successfully via Telegram Business Stories!",
    no_biz_conn: "❌ No connected business account found. Please check your Telegram Business settings.",
    ttl_media_alert: "🔥 *Backup copy of incoming media saved (TTL/Sensitive)*\n👤 From: %s (`%d`)",
    deleted_alert: "🗑️ *Security Alert: Customer deleted a message or media!*\n👤 Customer: %s (`%d`)"
  }
};

function t(lang, key) {
  const l = lang === "en" ? "en" : "ar";
  return i18n[l][key] || key;
}

// 4️⃣ دوال المساعدة للترجمة، الأزرار، والقصص
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

async function postBusinessStory(bizConnId, fileId, mediaType, activePeriod) {
  const fileInfo = await bot.api.getFile(fileId);
  const fileUrl = `https://api.telegram.org/file/bot${token}/${fileInfo.file_path}`;
  const inputFile = new InputFile({ url: fileUrl });

  if (mediaType === "photo") {
    return await bot.api.raw.postStory({
      business_connection_id: bizConnId,
      content: { type: "photo", photo: inputFile },
      active_period: activePeriod,
    });
  } else {
    return await bot.api.raw.postStory({
      business_connection_id: bizConnId,
      content: { type: "video", video: inputFile },
      active_period: activePeriod,
    });
  }
}

function extractMediaInfo(msg) {
  const info = {
    fromId: msg.from.id,
    fromName: `${msg.from.first_name || ""} ${msg.from.last_name || ""}`.trim() || msg.from.username || "Customer",
    caption: msg.caption || "",
    text: msg.text || "",
    type: "text",
    fileId: null,
  };

  if (msg.photo) {
    info.type = "photo";
    info.fileId = msg.photo[msg.photo.length - 1].file_id;
  } else if (msg.video) {
    info.type = "video";
    info.fileId = msg.video.file_id;
  } else if (msg.animation) {
    info.type = "animation";
    info.fileId = msg.animation.file_id;
  } else if (msg.sticker) {
    info.type = "sticker";
    info.fileId = msg.sticker.file_id;
  } else if (msg.voice) {
    info.type = "voice";
    info.fileId = msg.voice.file_id;
  } else if (msg.audio) {
    info.type = "audio";
    info.fileId = msg.audio.file_id;
  } else if (msg.document) {
    info.type = "document";
    info.fileId = msg.document.file_id;
  } else if (msg.video_note) {
    info.type = "video_note";
    info.fileId = msg.video_note.file_id;
  }

  return info;
}

async function sendMediaToAdmin(adminId, media, headerText) {
  const caption = `${headerText}\n\n` + (media.caption ? `💬 الشرح: ${media.caption}` : "");
  try {
    switch (media.type) {
      case "photo":
        await bot.api.sendPhoto(adminId, media.fileId, { caption, parse_mode: "Markdown" });
        break;
      case "video":
        await bot.api.sendVideo(adminId, media.fileId, { caption, parse_mode: "Markdown" });
        break;
      case "animation":
        await bot.api.sendAnimation(adminId, media.fileId, { caption, parse_mode: "Markdown" });
        break;
      case "sticker":
        await bot.api.sendMessage(adminId, headerText, { parse_mode: "Markdown" });
        await bot.api.sendSticker(adminId, media.fileId);
        break;
      case "voice":
        await bot.api.sendVoice(adminId, media.fileId, { caption, parse_mode: "Markdown" });
        break;
      case "audio":
        await bot.api.sendAudio(adminId, media.fileId, { caption, parse_mode: "Markdown" });
        break;
      case "document":
        await bot.api.sendDocument(adminId, media.fileId, { caption, parse_mode: "Markdown" });
        break;
      case "video_note":
        await bot.api.sendMessage(adminId, headerText, { parse_mode: "Markdown" });
        await bot.api.sendVideoNote(adminId, media.fileId);
        break;
      case "text":
        await bot.api.sendMessage(adminId, `${headerText}\n\n💬 *المحتوى المحذوف:*\n${media.text}`, { parse_mode: "Markdown" });
        break;
    }
  } catch (err) {
    console.error("خطأ في إعادة إرسال الوسائط للأدمن:", err);
  }
}

// 5️⃣ لوحات التحكم التفاعلية
function getMainMenuKeyboard(lang) {
  return new InlineKeyboard()
    .text(t(lang, "stop_btn"), "action_stop").text(t(lang, "start_btn"), "action_start").row()
    .text(t(lang, "edit_text_btn"), "action_edit_text").row()
    .text(t(lang, "exclude_btn"), "action_exclude").text(t(lang, "list_excluded_btn"), "action_list_excluded").row()
    .text(t(lang, "clear_excluded_btn"), "action_clear_excluded").row()
    .text(t(lang, "profile_btn"), "menu_profile").text(t(lang, "story_btn"), "menu_story").row()
    .text(t(lang, "lang_ar_btn"), "lang_ar").text(t(lang, "lang_en_btn"), "lang_en").row()
    .url("تحديثات نيرد 📢", "https://t.me/Xhwe2");
}

function getProfileKeyboard(lang) {
  return new InlineKeyboard()
    .text(t(lang, "edit_name"), "prof_name").row()
    .text(t(lang, "edit_bio"), "prof_bio").row()
    .text(t(lang, "edit_photo"), "prof_photo").row()
    .text(t(lang, "edit_username"), "prof_username").row()
    .text(t(lang, "back_btn"), "menu_main");
}

function getStoryDurationKeyboard(lang) {
  return new InlineKeyboard()
    .text(t(lang, "dur_6h"), "sdur_21600").text(t(lang, "dur_12h"), "sdur_43200").row()
    .text(t(lang, "dur_24h"), "sdur_86400").text(t(lang, "dur_48h"), "sdur_172800").row()
    .text(t(lang, "back_btn"), "menu_main");
}

function getBackKeyboard(lang) {
  return new InlineKeyboard().text(t(lang, "back_btn"), "menu_main");
}

function getAdminConfig(adminId) {
  let cfg = memoryCache.get(`config:${adminId}`);
  if (!cfg) {
    cfg = { isStopped: false, autoReply: "", excluded: [], state: "", lang: "ar", businessConnId: "" };
    memoryCache.set(`config:${adminId}`, cfg);
  }
  return cfg;
}

function saveAdminConfig(adminId, config) {
  memoryCache.set(`config:${adminId}`, config);
}

// 6️⃣ الحماية وإشراف المجموعات
bot.on(["message:group", "message:supergroup"], async (ctx) => {
  const messageText = ctx.message.text || ctx.message.caption || "";
  const userId = ctx.from?.id;

  const containsBadWord = badWords.some((word) => messageText.toLowerCase().includes(word));
  if (containsBadWord) {
    try { await ctx.deleteMessage(); return; } catch (e) {}
  }

  const hasLink = /(https?:\/\/[^\s]+)|(t\.me\/[^\s]+)|(telegram\.me\/[^\s]+)/i.test(messageText);
  if (hasLink && userId) {
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

  if (ctx.message.reply_to_message) {
    await ctx.reply("تابع جديدنا عبر قناة التحديثات الرسمية:", {
      reply_to_message_id: ctx.message.message_id,
      reply_markup: getNerdChannelKeyboard(),
    });
  }
});

// 7️⃣ الأوامر والأزرار التفاعلية الخاصة
bot.command("start", async (ctx) => {
  if (ctx.chat.type !== "private") return;
  const cfg = getAdminConfig(ctx.from.id);
  await ctx.reply(t(cfg.lang, "welcome"), {
    reply_markup: getMainMenuKeyboard(cfg.lang),
  });
});

bot.command("id", async (ctx) => {
  await ctx.reply(`آيدي حسابك الشخصي هو:\n\`${ctx.from.id}\``, { parse_mode: "Markdown" });
});

bot.on("callback_query:data", async (ctx) => {
  const data = ctx.callbackQuery.data;
  const adminId = ctx.from.id;
  const cfg = getAdminConfig(adminId);

  await ctx.answerCallbackQuery();

  if (data === "change_quote") {
    const randomQuote = quotes[Math.floor(Math.random() * quotes.length)];
    await ctx.editMessageReplyMarkup({
      reply_markup: new InlineKeyboard()
        .text(`✨ ${randomQuote}`, "change_quote").row()
        .url("تحديثات نيرد 📢", "https://t.me/Xhwe2"),
    });
    return;
  }

  if (data === "menu_main") {
    cfg.state = "";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "main_menu"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "action_stop") {
    cfg.isStopped = true;
    cfg.state = "";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "stopped"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "action_start") {
    cfg.isStopped = false;
    cfg.state = "";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "started"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "action_edit_text") {
    cfg.state = "waiting_text";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "edit_prompt"), {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "action_exclude") {
    cfg.state = "waiting_exclude_id";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "exclude_prompt"), {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "action_list_excluded") {
    let msg = `📋 قائمة الحسابات المستثناة:\n`;
    if (!cfg.excluded || cfg.excluded.length === 0) {
      msg += t(cfg.lang, "no_excluded");
    } else {
      msg += cfg.excluded.map((id) => `- \`${id}\``).join("\n");
    }
    await ctx.editMessageText(msg, {
      parse_mode: "Markdown",
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "action_clear_excluded") {
    cfg.excluded = [];
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "cleared_excluded"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "menu_profile") {
    await ctx.editMessageText(t(cfg.lang, "profile_menu"), {
      reply_markup: getProfileKeyboard(cfg.lang),
    });
  } else if (data.startsWith("prof_")) {
    if (!cfg.businessConnId) {
      await ctx.editMessageText(t(cfg.lang, "no_biz_conn"), {
        reply_markup: getProfileKeyboard(cfg.lang),
      });
      return;
    }
    const field = data.replace("prof_", "");
    cfg.state = `waiting_prof_${field}`;
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(`أرسل الآن القيمة الجديدة لـ (${field}):`, {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "menu_story") {
    if (!cfg.businessConnId) {
      await ctx.editMessageText(t(cfg.lang, "no_biz_conn"), {
        reply_markup: getMainMenuKeyboard(cfg.lang),
      });
      return;
    }
    await ctx.editMessageText(t(cfg.lang, "story_dur_title"), {
      reply_markup: getStoryDurationKeyboard(cfg.lang),
    });
  } else if (data.startsWith("sdur_")) {
    const period = data.replace("sdur_", "");
    cfg.state = `waiting_story_${period}`;
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "story_prompt").replace("%s", period + "s"), {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "lang_ar") {
    cfg.lang = "ar";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t("ar", "main_menu"), {
      reply_markup: getMainMenuKeyboard("ar"),
    });
  } else if (data === "lang_en") {
    cfg.lang = "en";
    saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t("en", "main_menu"), {
      reply_markup: getMainMenuKeyboard("en"),
    });
  }
});

// 8️⃣ معالجة نصوص الإدخال الخاصة بالأدمن
bot.on("message:private", async (ctx, next) => {
  const adminId = ctx.from.id;
  const cfg = getAdminConfig(adminId);

  if (!cfg.state) return next();

  if (cfg.state === "waiting_text") {
    cfg.autoReply = ctx.message.text;
    cfg.state = "";
    saveAdminConfig(adminId, cfg);
    await ctx.reply(t(cfg.lang, "saved_text"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (cfg.state === "waiting_exclude_id") {
    const exId = parseInt(ctx.message.text.trim());
    if (!isNaN(exId)) {
      if (!cfg.excluded.includes(exId)) cfg.excluded.push(exId);
      cfg.state = "";
      saveAdminConfig(adminId, cfg);
      await ctx.reply(t(cfg.lang, "id_added"), {
        reply_markup: getMainMenuKeyboard(cfg.lang),
      });
    }
  } else if (cfg.state.startsWith("waiting_prof_")) {
    const field = cfg.state.replace("waiting_prof_", "");
    try {
      if (field === "name") {
        const parts = ctx.message.text.trim().split(" ");
        await bot.api.raw.setBusinessAccountName({
          business_connection_id: cfg.businessConnId,
          first_name: parts[0],
          last_name: parts.slice(1).join(" ") || undefined,
        });
      } else if (field === "bio") {
        await bot.api.raw.setBusinessAccountBio({
          business_connection_id: cfg.businessConnId,
          bio: ctx.message.text.trim(),
        });
      } else if (field === "username") {
        await bot.api.raw.setBusinessAccountUsername({
          business_connection_id: cfg.businessConnId,
          username: ctx.message.text.trim().replace("@", ""),
        });
      }
      cfg.state = "";
      saveAdminConfig(adminId, cfg);
      await ctx.reply("✅ تم تحديث بيانات الملف الشخصي بنجاح!", {
        reply_markup: getMainMenuKeyboard(cfg.lang),
      });
    } catch (err) {
      await ctx.reply(`❌ فشل تحديث البيانات عبر API: ${err.message}`);
    }
  } else if (cfg.state.startsWith("waiting_story_")) {
    const activePeriod = parseInt(cfg.state.replace("waiting_story_", ""));
    const photo = ctx.message.photo;
    const video = ctx.message.video;

    if (!photo && !video) {
      await ctx.reply("❌ يرجى إرسال صورة أو فيديو صالح لنشره كقصة.");
      return;
    }

    try {
      if (photo) {
        const fileId = photo[photo.length - 1].file_id;
        await postBusinessStory(cfg.businessConnId, fileId, "photo", activePeriod);
      } else if (video) {
        if (video.duration > 60) {
          await ctx.reply("❌ عذراً، لا يمكن نشر فيديو أطول من 60 ثانية كقصة.");
          return;
        }
        await postBusinessStory(cfg.businessConnId, video.file_id, "video", activePeriod);
      }
      cfg.state = "";
      saveAdminConfig(adminId, cfg);
      await ctx.reply(t(cfg.lang, "story_success"), {
        reply_markup: getMainMenuKeyboard(cfg.lang),
      });
    } catch (err) {
      await ctx.reply(`❌ فشل نشر القصة: ${err.message}`);
    }
  }
});

// 9️⃣ إدارة الاتصالات التجارية
bot.on("business_connection", async (ctx) => {
  const bc = ctx.update.business_connection;
  if (bc.is_enabled && bc.user_chat_id) {
    const cfg = getAdminConfig(bc.user_chat_id);
    cfg.businessConnId = bc.id;
    saveAdminConfig(bc.user_chat_id, cfg);
    memoryCache.set(`conn:${bc.id}`, bc.user_chat_id);
  }
});

// 🔟 نظام الرد الآلي والترجمة والنسخ الاحتياطي للوسائط
bot.on("business_message", async (ctx) => {
  const msg = ctx.update.business_message;
  const connId = msg.business_connection_id;

  if (msg.from.is_bot) return;

  const mediaInfo = extractMediaInfo(msg);
  memoryCache.set(`msg:${msg.chat.id}:${msg.message_id}`, mediaInfo);

  const adminId = memoryCache.get(`conn:${connId}`);
  if (!adminId) return;

  const cfg = getAdminConfig(adminId);

  if (!msg.is_outgoing && mediaInfo.type !== "text") {
    const header = t(cfg.lang, "ttl_media_alert")
      .replace("%s", mediaInfo.fromName)
      .replace("%d", msg.from.id);
    await sendMediaToAdmin(adminId, mediaInfo, header);
  }

  if (msg.is_outgoing || cfg.isStopped || (cfg.excluded && cfg.excluded.includes(msg.from.id))) {
    return;
  }

  let detectedLang = "";
  let translatedText = "";
  if (msg.text) {
    const trRes = await translateText(msg.text, "ar");
    detectedLang = trRes.detectedLang;
    translatedText = trRes.text;

    if (detectedLang && detectedLang !== "ar") {
      const alertMsg = `🌐 *رسالة بلغة أجنبية مترجمة (${detectedLang})*\n👤 *العميل:* ${mediaInfo.fromName} (\`${msg.from.id}\`)\n\n💬 *النص الأصلي:*\n${msg.text}\n\n✨ *الترجمة العربية:*\n${translatedText}`;
      await bot.api.sendMessage(adminId, alertMsg, { parse_mode: "Markdown" }).catch(() => {});
    }
  }

  const customerName = msg.from.first_name || "عميلنا العزيز";
  let replyText = cfg.autoReply || "أهلاً بك يا {name} 🌸\nاستلمت رسالتك وسأرد عليك في أقرب وقت ممكن.";
  replyText = replyText.replace("{name}", customerName).replace("{الاسم}", customerName);

  if (detectedLang && detectedLang !== "ar") {
    const trReply = await translateText(replyText, detectedLang);
    if (trReply.text) replyText = trReply.text;
  }

  const initialQuote = quotes[Math.floor(Math.random() * quotes.length)];

  await bot.api.sendMessage(msg.chat.id, replyText, {
    business_connection_id: connId,
    reply_markup: new InlineKeyboard()
      .text(`✨ ${initialQuote}`, "change_quote").row()
      .url("تحديثات نيرد 📢", "https://t.me/Xhwe2"),
  });
});

// 1️⃣1️⃣ استرجاع الرسائل والوسائط المحذوفة
bot.on("deleted_business_messages", async (ctx) => {
  const dbm = ctx.update.deleted_business_messages;
  const adminId = memoryCache.get(`conn:${dbm.business_connection_id}`);
  if (!adminId) return;

  const cfg = getAdminConfig(adminId);

  for (const msgId of dbm.message_ids) {
    const cacheKey = `msg:${dbm.chat.id}:${msgId}`;
    const cached = memoryCache.get(cacheKey);
    if (cached) {
      const alertHeader = t(cfg.lang, "deleted_alert")
        .replace("%s", cached.fromName)
        .replace("%d", cached.fromId);

      await sendMediaToAdmin(adminId, cached, alertHeader);
    }
  }
});

// 1️⃣2️⃣ غلاف الحماية للتشغيل على Vercel
const handleUpdate = webhookCallback(bot, "http");

module.exports = async (req, res) => {
  try {
    return await handleUpdate(req, res);
  } catch (error) {
    console.error("❌ خطأ فادح في خادم Vercel Serverless:", error);
    res.statusCode = 500;
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify({ error: error.message }));
  }
};
