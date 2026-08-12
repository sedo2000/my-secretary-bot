const { Bot, webhookCallback, InlineKeyboard } = require("grammy");
const { Redis } = require("@upstash/redis");

// 1️⃣ إعداد قاعدة البيانات Upstash Redis
const redis =
  process.env.UPSTASH_REDIS_REST_URL && process.env.UPSTASH_REDIS_REST_TOKEN
    ? new Redis({
        url: process.env.UPSTASH_REDIS_REST_URL,
        token: process.env.UPSTASH_REDIS_REST_TOKEN,
      })
    : null;

const memoryCache = new Map();

async function setCache(key, value, ttlSeconds = 86400) {
  if (redis) {
    await redis.set(key, JSON.stringify(value), { ex: ttlSeconds });
  } else {
    memoryCache.set(key, value);
  }
}

async function getCache(key) {
  if (redis) {
    const val = await redis.get(key);
    return typeof val === "string" ? JSON.parse(val) : val;
  }
  return memoryCache.get(key);
}

// 2️⃣ القواميس والبيانات الأساسية (الكلمات المسيئة، الردود الذكية، والاقتباسات)
const badWords = ["كحبة", "مطي", "قندرة", "ساقط", "فرخ", "عير", "كس", "طيز", "زنيّم"];

const smartAnswers = {
  السعر: "ℹ️ لمعرفة الأسعار والتفاصيل الكاملة، يمكنك زيارة القناة الرسمية أو مراسلة الدعم.",
  الدعم: "🛠️ للتواصل مع الدعم الفني، يرجى مراسلة الحساب التجاري المباشر.",
  التسجيل: "📝 يمكنك التسجيل والاشتراك عبر فتح المحادثة الخاصة وتتبع التعليمات.",
};

const quotes = [
  "قاوم ما تكره لتصل الى ما تحب",
  "الحرب بين أنت ضد أنت",
  "أبنِ نفسك بنفسك لنفسك",
  "ميخالف",
  "حتى لو متأخر تگدر..!",
  "من يعيش في خوف لن يكون حراً ابداً",
  "لا أبرح حتى أبلغ",
  "أنه مبرمج فحسب",
  "المرء نتاج خلواته",
  "لا مزيد من الأصدقاء المزيفين",
];

// قاموس اللغات (العربية والإنكليزية)
const i18n = {
  ar: {
    welcome: "أهلاً بك في لوحة تحكم سكرتير الحساب التجاري 🤖\nاختر من الأزرار أدناه للتحكم الكامل:",
    main_menu: "القائمة الرئيسية 🤖:",
    stop_btn: "🛑 إيقاف الرد الخاص",
    start_btn: "🟢 تشغيل الرد الخاص",
    edit_text_btn: "📝 تعديل نص الرد الخاص",
    exclude_btn: "👤 استثناء حساب",
    list_excluded_btn: "📋 عرض المستثنين",
    clear_excluded_btn: "🧹 مسح المستثنين",
    profile_btn: "🧑 إدارة الملف الشخصي",
    story_btn: "📖 نشر قصة",
    lang_ar_btn: "🇮🇶 العربية",
    lang_en_btn: "🇺🇸 English",
    back_btn: "🔙 رجوع",
    stopped: "🛑 تم إيقاف الرد التلقائي الخاص بنجاح.",
    started: "🟢 تم تشغيل الرد التلقائي الخاص بنجاح.",
    edit_prompt: "📝 أرسل الآن نص الرد التلقائي الجديد للخاص:",
    saved_text: "✅ تم حفظ نص الرد التلقائي الجديد بنجاح!",
    exclude_prompt: "👤 أرسل ايدي (ID) الحساب المراد استثناؤه:",
    id_added: "✅ تم إضافة الايدي إلى قائمة الاستثناء.",
    no_excluded: "لا يوجد حسابات مستثناة حالياً.",
    cleared_excluded: "🧹 تم مسح قائمة الاستثناءات بالكامل.",
    profile_menu: "🧑 إدارة الملف الشخصي - اختر ما تريد تعديله:",
    edit_name: "✏️ تعديل الاسم",
    edit_bio: "📝 تعديل النبذة",
    edit_photo: "🖼️ تعديل الصورة",
    edit_username: "🔗 تعديل اليوزر",
    story_dur_title: "⏱️ اختر مدة ظهور القصة المطلوب نشرها:",
    dur_6h: "6 ساعات",
    dur_12h: "12 ساعة",
    dur_24h: "24 ساعة",
    dur_48h: "48 ساعة",
    story_prompt: "📖 أرسل الآن صورة أو فيديو لنشره كقصة (مدة الظهور: %s):",
    story_success: "✅ تم نشر القصة بنجاح!",
    no_biz_conn: "❌ لم يتم ربط حساب تجاري بعد بالبوت.",
    ttl_media_alert: "🔥 *تم حفظ نسخة احتياطية من وسائط واردة!*\n👤 من: %s (`%d`)",
    deleted_alert: "🗑️ *تنبيه: تم حذف رسالة/وسائط!*\n👤 العميل: %s (`%d`)",
  },
  en: {
    welcome: "Welcome to Business Secretary Control Panel 🤖\nChoose an option below:",
    main_menu: "Main Menu 🤖:",
    stop_btn: "🛑 Stop Auto-Reply",
    start_btn: "🟢 Start Auto-Reply",
    edit_text_btn: "📝 Edit Reply Text",
    exclude_btn: "👤 Exclude Account ID",
    list_excluded_btn: "📋 View Excluded List",
    clear_excluded_btn: "🧹 Clear Excluded List",
    profile_btn: "🧑 Manage Profile",
    story_btn: "📖 Post Story",
    lang_ar_btn: "🇮🇶 العربية",
    lang_en_btn: "🇺🇸 English",
    back_btn: "🔙 Back",
    stopped: "🛑 Auto-reply stopped successfully.",
    started: "🟢 Auto-reply started successfully.",
    edit_prompt: "📝 Send the new auto-reply text now:",
    saved_text: "✅ Auto-reply text saved successfully!",
    exclude_prompt: "👤 Send the Account ID to exclude:",
    id_added: "✅ Account ID added to exclusions.",
    no_excluded: "No excluded accounts currently.",
    cleared_excluded: "🧹 All exclusions cleared successfully.",
    profile_menu: "🧑 Manage Profile - Select setting to edit:",
    edit_name: "✏️ Edit Name",
    edit_bio: "📝 Edit Bio",
    edit_photo: "🖼️ Edit Photo",
    edit_username: "🔗 Edit Username",
    story_dur_title: "⏱️ Select Story Duration:",
    dur_6h: "6 Hours",
    dur_12h: "12 Hours",
    dur_24h: "24 Hours",
    dur_48h: "48 Hours",
    story_prompt: "📖 Send a photo/video to post as story (Duration: %s):",
    story_success: "✅ Story posted successfully!",
    no_biz_conn: "❌ No connected business account found.",
    ttl_media_alert: "🔥 *Backup of incoming media saved!*\n👤 From: %s (`%d`)",
    deleted_alert: "🗑️ *Alert: Deleted Message/Media!*\n👤 Customer: %s (`%d`)",
  },
};

function t(lang, key) {
  const l = lang === "en" ? "en" : "ar";
  return i18n[l][key] || key;
}

// 3️⃣ محرك الترجمة الفورية عبر Google Translate API
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

const token = process.env.TELEGRAM_BOT_TOKEN;
const bot = new Bot(token);

async function getAdminConfig(adminId) {
  const cfg = await getCache(`config:${adminId}`);
  return cfg || { isStopped: false, autoReply: "", excluded: [], state: "", lang: "ar", businessConnId: "" };
}

async function saveAdminConfig(adminId, config) {
  await setCache(`config:${adminId}`, config, 30 * 86400);
}

// زر الشفاف لقناة تحديثات نيرد
function getNerdChannelKeyboard() {
  return new InlineKeyboard().url("تحديثات نيرد 📢", "https://t.me/Xhwe2");
}

// دالة نشر القصص عبر API
async function postBusinessStory(bizConnId, fileId, mediaType, activePeriod, duration = 0) {
  const fileInfo = await bot.api.getFile(fileId);
  const fileUrl = `https://api.telegram.org/file/bot${token}/${fileInfo.file_path}`;
  const res = await fetch(fileUrl);
  const buffer = await res.arrayBuffer();

  const formData = new FormData();
  formData.append("business_connection_id", bizConnId);
  formData.append("active_period", activePeriod.toString());

  const contentJSON = mediaType === "video"
    ? JSON.stringify({ type: "video", video: "attach://media", duration })
    : JSON.stringify({ type: "photo", photo: "attach://media" });

  formData.append("content", contentJSON);
  formData.append("media", new Blob([buffer]), mediaType === "video" ? "story.mp4" : "story.jpg");

  const apiRes = await fetch(`https://api.telegram.org/bot${token}/postStory`, {
    method: "POST",
    body: formData,
  });
  return await apiRes.json();
}

// دالة تحديث صورة البروفايل عبر API
async function setBusinessProfilePhoto(bizConnId, fileId) {
  const fileInfo = await bot.api.getFile(fileId);
  const fileUrl = `https://api.telegram.org/file/bot${token}/${fileInfo.file_path}`;
  const res = await fetch(fileUrl);
  const buffer = await res.arrayBuffer();

  const formData = new FormData();
  formData.append("business_connection_id", bizConnId);
  formData.append("photo", JSON.stringify({ type: "static", photo: "attach://photo_file" }));
  formData.append("photo_file", new Blob([buffer]), "profile.jpg");

  const apiRes = await fetch(`https://api.telegram.org/bot${token}/setBusinessAccountProfilePhoto`, {
    method: "POST",
    body: formData,
  });
  return await apiRes.json();
}

// استخراج وسائط الرسائل
function extractMediaInfo(msg) {
  const info = {
    fromId: msg.from.id,
    fromName: `${msg.from.first_name || ""} ${msg.from.last_name || ""}`.trim() || msg.from.username || "Customer",
    isOutgoing: msg.is_outgoing,
    isBot: msg.from.is_bot,
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
    info.duration = msg.video.duration;
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

// إعادة إرسال الوسائط للأدمن عند الحذف أو النسخ الاحتياطي
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
    console.error("خطأ إعادة إرسال الوسائط:", err);
  }
}

// لوحات التحكم بالأزرار (Inline Keyboards)
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

// 4️⃣ الإشراف التلقائي، منع السبام، والردود في المجموعات (Group Moderation)
bot.on(["message:group", "message:supergroup"], async (ctx) => {
  const messageText = ctx.message.text || ctx.message.caption || "";
  const chatId = ctx.chat.id;
  const userId = ctx.from.id;

  // أ) جمع البيانات وإحصائيات رسائل الأعضاء (Analytics)
  if (redis && userId) {
    await redis.incr(`stats:msg_count:${chatId}:${userId}`);
  }

  // ب) رصد وحظر الكلمات البذيئة
  const containsBadWord = badWords.some((word) => messageText.toLowerCase().includes(word));
  if (containsBadWord) {
    try {
      await ctx.deleteMessage();
      return;
    } catch (e) {
      console.error("خطأ في حذف الكلمة البذيئة:", e);
    }
  }

  // ج) كشف ومنع الإعلانات والروابط (Spam Protection)
  const hasLink = /(https?:\/\/[^\s]+)|(t\.me\/[^\s]+)|(telegram\.me\/[^\s]+)/i.test(messageText);
  if (hasLink) {
    try {
      await ctx.deleteMessage();
      await ctx.banChatMember(userId);
      return;
    } catch (e) {
      console.error("خطأ حظر مرسل السبام:", e);
    }
  }

  // د) الرد الآلي الذكي على الكلمات المفتاحية
  for (const [key, answer] of Object.entries(smartAnswers)) {
    if (messageText.includes(key)) {
      await ctx.reply(answer, {
        reply_to_message_id: ctx.message.message_id,
        reply_markup: getNerdChannelKeyboard(),
      });
      return;
    }
  }

  // هـ) عند الرد على رسالة في المجموعات -> إرفاق زر شفاف يحتوي على "تحديثات نيرد"
  if (ctx.message.reply_to_message) {
    await ctx.reply("تابع جديدنا عبر قناة التحديثات الرسمية:", {
      reply_to_message_id: ctx.message.message_id,
      reply_markup: getNerdChannelKeyboard(),
    });
  }
});

// 5️⃣ أوامر وتفاعلات لوحة التحكم للأدمن في الخاص
bot.command("start", async (ctx) => {
  if (ctx.chat.type !== "private") return;
  const cfg = await getAdminConfig(ctx.from.id);
  await ctx.reply(t(cfg.lang, "welcome"), {
    reply_markup: getMainMenuKeyboard(cfg.lang),
  });
});

bot.command("id", async (ctx) => {
  await ctx.reply(`آيدي حسابك هو:\n\`${ctx.from.id}\``, { parse_mode: "Markdown" });
});

bot.on("callback_query:data", async (ctx) => {
  const data = ctx.callbackQuery.data;
  const adminId = ctx.from.id;
  const cfg = await getAdminConfig(adminId);

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
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "main_menu"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "action_stop") {
    cfg.isStopped = true;
    cfg.state = "";
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "stopped"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "action_start") {
    cfg.isStopped = false;
    cfg.state = "";
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "started"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (data === "action_edit_text") {
    cfg.state = "waiting_text";
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "edit_prompt"), {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "action_exclude") {
    cfg.state = "waiting_exclude_id";
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "exclude_prompt"), {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "action_list_excluded") {
    let msg = `📋 قائمة المستثنين:\n`;
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
    await saveAdminConfig(adminId, cfg);
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
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(`أرسل القيمة الجديدة لـ (${field}):`, {
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
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t(cfg.lang, "story_prompt").replace("%s", period + "s"), {
      reply_markup: getBackKeyboard(cfg.lang),
    });
  } else if (data === "lang_ar") {
    cfg.lang = "ar";
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t("ar", "main_menu"), {
      reply_markup: getMainMenuKeyboard("ar"),
    });
  } else if (data === "lang_en") {
    cfg.lang = "en";
    await saveAdminConfig(adminId, cfg);
    await ctx.editMessageText(t("en", "main_menu"), {
      reply_markup: getMainMenuKeyboard("en"),
    });
  }
});

// التعامل مع مدخلات الأدمن في المحادثة الخاصة
bot.on("message", async (ctx, next) => {
  if (ctx.update.business_message) return next();

  const adminId = ctx.from.id;
  const cfg = await getAdminConfig(adminId);

  if (!cfg.state) return next();

  if (cfg.state === "waiting_text") {
    cfg.autoReply = ctx.message.text;
    cfg.state = "";
    await saveAdminConfig(adminId, cfg);
    await ctx.reply(t(cfg.lang, "saved_text"), {
      reply_markup: getMainMenuKeyboard(cfg.lang),
    });
  } else if (cfg.state === "waiting_exclude_id") {
    const exId = parseInt(ctx.message.text.trim());
    if (!isNaN(exId)) {
      if (!cfg.excluded.includes(exId)) cfg.excluded.push(exId);
      cfg.state = "";
      await saveAdminConfig(adminId, cfg);
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
      } else if (field === "photo") {
        if (!ctx.message.photo) {
          await ctx.reply("❌ أرسل صورة لتحديث صورة الملف الشخصي.");
          return;
        }
        const fileId = ctx.message.photo[ctx.message.photo.length - 1].file_id;
        await setBusinessProfilePhoto(cfg.businessConnId, fileId);
      }
      cfg.state = "";
      await saveAdminConfig(adminId, cfg);
      await ctx.reply("✅ تم تحديث بيانات الملف الشخصي بنجاح!", {
        reply_markup: getMainMenuKeyboard(cfg.lang),
      });
    } catch (err) {
      await ctx.reply(`❌ فشل تحديث البيانات: ${err.message}`);
    }
  } else if (cfg.state.startsWith("waiting_story_")) {
    const activePeriod = parseInt(cfg.state.replace("waiting_story_", ""));
    const photo = ctx.message.photo;
    const video = ctx.message.video;

    if (!photo && !video) {
      await ctx.reply("❌ يرجى إرسال صورة أو فيديو لنشره كقصة.");
      return;
    }

    try {
      if (photo) {
        const fileId = photo[photo.length - 1].file_id;
        await postBusinessStory(cfg.businessConnId, fileId, "photo", activePeriod);
      } else if (video) {
        if (video.duration > 60) {
          await ctx.reply("❌ الفيديو أطول من 60 ثانية.");
          return;
        }
        await postBusinessStory(cfg.businessConnId, video.file_id, "video", activePeriod, video.duration);
      }
      cfg.state = "";
      await saveAdminConfig(adminId, cfg);
      await ctx.reply(t(cfg.lang, "story_success"), {
        reply_markup: getMainMenuKeyboard(cfg.lang),
      });
    } catch (err) {
      await ctx.reply(`❌ فشل نشر القصة: ${err.message}`);
    }
  }
});

// 6️⃣ التعامل مع البوتات المُدارة (Managed Bots)
bot.on("managed_bot", async (ctx) => {
  const mb = ctx.update.managed_bot;
  try {
    const tokenRes = await bot.api.raw.getManagedBotToken({ managed_bot_id: mb.bot.id });
    const adminId = mb.user.id;
    await bot.api.sendMessage(
      adminId,
      `🤖 *تم إنشاء بوت فرعي جديد!*\n\n` +
      `👤 الاسم: ${mb.bot.first_name}\n` +
      `🔗 اليوزر: @${mb.bot.username}\n` +
      `🔑 التوكن: \`${tokenRes.token}\``,
      { parse_mode: "Markdown" }
    );
  } catch (err) {
    console.error("خطأ جلب توكن البوت المدار:", err);
  }
});

// 7️⃣ ربط الحساب التجاري
bot.on("business_connection", async (ctx) => {
  const bc = ctx.update.business_connection;
  if (bc.is_enabled && bc.user_chat_id) {
    const cfg = await getAdminConfig(bc.user_chat_id);
    cfg.businessConnId = bc.id;
    await saveAdminConfig(bc.user_chat_id, cfg);
    await setCache(`conn:${bc.id}`, bc.user_chat_id, 365 * 86400);

    const devId = process.env.DEVELOPER_CHAT_ID;
    if (devId) {
      await bot.api.sendMessage(
        devId,
        `🔔 *تفعيل جديد لسكرتير الأعمال!*\n\n👤 المستخدم: ${bc.user.first_name}\n🆔 الايدي: \`${bc.user.id}\``,
        { parse_mode: "Markdown" }
      );
    }
  }
});

// 8️⃣ الرد الخاص في محادثات العملاء والتراسل التجاري (Business Message)
bot.on("business_message", async (ctx) => {
  const msg = ctx.update.business_message;
  const connId = msg.business_connection_id;

  if (msg.from.is_bot) return;

  const mediaInfo = extractMediaInfo(msg);
  const cacheKey = `msg:${msg.chat.id}:${msg.message_id}`;
  await setCache(cacheKey, mediaInfo, 7 * 86400);

  const adminId = await getCache(`conn:${connId}`);
  if (!adminId) return;

  const cfg = await getAdminConfig(adminId);

  // نسخ الاحتياط للوسائط ذاتية التدمير
  if (!msg.is_outgoing && mediaInfo.type !== "text") {
    const header = t(cfg.lang, "ttl_media_alert")
      .replace("%s", mediaInfo.fromName)
      .replace("%d", msg.from.id);
    await sendMediaToAdmin(adminId, mediaInfo, header);
  }

  // تجنب الرد التلقائي على الرسائل الصادرة أو عند الإيقاف
  if (msg.is_outgoing || cfg.isStopped || cfg.excluded.includes(msg.from.id)) {
    return;
  }

  // كشف اللغة والترجمة الفورية
  let detectedLang = "";
  let translatedText = "";
  if (msg.text) {
    const trRes = await translateText(msg.text, "ar");
    detectedLang = trRes.detectedLang;
    translatedText = trRes.text;

    if (detectedLang && detectedLang !== "ar") {
      const alertMsg = `🌐 *رسالة بلغة مترجمة (` + detectedLang + `)*\n👤 *العميل:* ` + mediaInfo.fromName + ` (\`${msg.from.id}\`)\n\n💬 *الأصل:*\n` + msg.text + `\n\n✨ *الترجمة:*\n` + translatedText;
      await bot.api.sendMessage(adminId, alertMsg, { parse_mode: "Markdown" });
    }
  }

  // صياغة الرد الآلي للعميل في الخاص
  const customerName = msg.from.first_name || "عميلنا العزيز";
  let replyText = cfg.autoReply || "أهلاً بك يا {name} 🌸\nاستلمت رسالتك وسأرد عليك في أقرب وقت.";
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

// 9️⃣ كشف واسترجاع المحذوفات للأدمن
bot.on("deleted_business_messages", async (ctx) => {
  const dbm = ctx.update.deleted_business_messages;
  const adminId = await getCache(`conn:${dbm.business_connection_id}`);
  if (!adminId) return;

  const cfg = await getAdminConfig(adminId);

  for (const msgId of dbm.message_ids) {
    const cacheKey = `msg:${dbm.chat.id}:${msgId}`;
    const cached = await getCache(cacheKey);
    if (cached) {
      const alertHeader = t(cfg.lang, "deleted_alert")
        .replace("%s", cached.fromName)
        .replace("%d", cached.fromId);

      await sendMediaToAdmin(adminId, cached, alertHeader);
    }
  }
});

// تصدير المعالج لبيئة Vercel
module.exports = webhookCallback(bot, "http");
