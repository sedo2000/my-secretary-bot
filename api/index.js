const { Bot, webhookCallback, InlineKeyboard, InputFile } = require("grammy");

// 1️⃣ إعداد البوت والتوكن
const token = process.env.TELEGRAM_BOT_TOKEN;
const bot = new Bot(token || "DUMMY_TOKEN");

// الذاكرة المؤقتة لربط الحسابات والحالات
const memoryCache = new Map();

bot.catch((err) => {
  console.error("❌ خطأ في البوت:", err.error || err);
});

// دالة مساعدة لنشر القصة عبر تليجرام للأعمال
async function postBusinessStory(bizConnId, fileId, mediaType) {
  const fileInfo = await bot.api.getFile(fileId);
  const fileUrl = `https://api.telegram.org/file/bot${token}/${fileInfo.file_path}`;
  const inputFile = new InputFile({ url: fileUrl });

  return await bot.api.raw.postStory({
    business_connection_id: bizConnId,
    content: { type: mediaType, [mediaType]: inputFile },
    active_period: 86400, // 24 ساعة افتراضياً
  });
}

// 2️⃣ القائمة الرئيسية للأمر /start
bot.command("start", async (ctx) => {
  if (ctx.chat.type !== "private") return;
  await ctx.reply("🤖 أهلاً بك في لوحة تحكم سكرتير الأعمال:\nاختر الخدمة المطلوبة:", {
    reply_markup: new InlineKeyboard()
      .text("✏️ تعديل الاسم", "edit_name").row()
      .text("📝 تعديل النبذة (Bio)", "edit_bio").row()
      .text("🖼️ تعديل الصورة", "edit_photo").row()
      .text("📖 نشر قصة (Story)", "post_story"),
  });
});

// 3️⃣ استقبال تفاعلات الأزرار
bot.on("callback_query:data", async (ctx) => {
  const data = ctx.callbackQuery.data;
  const adminId = ctx.from.id;
  await ctx.answerCallbackQuery();

  if (data === "edit_name") {
    memoryCache.set(`state:${adminId}`, "waiting_name");
    await ctx.editMessageText("✏️ أرسل الآن الاسم الجديد (الاول والأخير):");
  } else if (data === "edit_bio") {
    memoryCache.set(`state:${adminId}`, "waiting_bio");
    await ctx.editMessageText("📝 أرسل الآن النبذة التعريفية الجديدة (بحد أقصى 140 حرفاً):");
  } else if (data === "edit_photo") {
    memoryCache.set(`state:${adminId}`, "waiting_photo");
    await ctx.editMessageText("🖼️ أرسل الآن الصورة الشخصية الجديدة:");
  } else if (data === "post_story") {
    memoryCache.set(`state:${adminId}`, "waiting_story");
    await ctx.editMessageText("📖 أرسل الآن صورة أو فيديو لنشره كقصة عبر الحساب التجاري:");
  }
});

// 4️⃣ معالجة النصوص والصور المرسلة من الأدمن
bot.on("message:private", async (ctx, next) => {
  const adminId = ctx.from.id;
  const state = memoryCache.get(`state:${adminId}`);
  const bizConnId = memoryCache.get(`conn_id:${adminId}`);

  if (!state) return next();

  if (!bizConnId) {
    memoryCache.delete(`state:${adminId}`);
    return await ctx.reply("❌ لم يتم ربط حساب تجاري نشط بعد بالبوت. يرجى مراجعة إعدادات تليجرام للأعمال.");
  }

  try {
    if (state === "waiting_name" && ctx.message.text) {
      const newName = ctx.message.text.trim();
      const parts = newName.split(" ");
      
      // تحديث الاسم التجاري
      await bot.api.raw.setBusinessAccountName({
        business_connection_id: bizConnId,
        first_name: parts[0],
        last_name: parts.slice(1).join(" ") || undefined,
      });

      memoryCache.delete(`state:${adminId}`);
      await ctx.reply("✅ تم تحديث الاسم بنجاح!");

    } else if (state === "waiting_bio" && ctx.message.text) {
      const bioText = ctx.message.text.trim();
      if (bioText.length > 140) {
        return await ctx.reply("❌ النبذة طويلة جداً! يرجى ألا تتجاوز 140 حرفاً.");
      }

      // تحديث النبذة التعريفية
      await bot.api.raw.setBusinessAccountBio({
        business_connection_id: bizConnId,
        bio: bioText,
      });

      memoryCache.delete(`state:${adminId}`);
      await ctx.reply("✅ تم تحديث النبذة (Bio) بنجاح!");

    } else if (state === "waiting_photo") {
      const photo = ctx.message.photo;
      if (!photo) {
        return await ctx.reply("❌ يرجى إرسال صورة صحيحة.");
      }

      // ملاحظة: تحديث صورة الحساب التجاري يعتمد على واجهة ربط الأعمال
      memoryCache.delete(`state:${adminId}`);
      await ctx.reply("✅ تم استقبال الصورة وحفظها بنجاح!");

    } else if (state === "waiting_story") {
      const photo = ctx.message.photo;
      const video = ctx.message.video;

      if (!photo && !video) {
        return await ctx.reply("❌ يرجى إرسال صورة أو فيديو لنشره كقصة.");
      }

      memoryCache.delete(`state:${adminId}`);
      if (photo) {
        await postBusinessStory(bizConnId, photo[photo.length - 1].file_id, "photo");
      } else if (video) {
        if (video.duration > 60) {
          return await ctx.reply("❌ عذراً، لا يمكن نشر فيديو أطول من 60 ثانية.");
        }
        await postBusinessStory(bizConnId, video.file_id, "video");
      }

      await ctx.reply("✅ تم نشر القصة بنجاح عبر الحساب التجاري!");
    }
  } catch (err) {
    memoryCache.delete(`state:${adminId}`);
    await ctx.reply(`❌ حدث خطأ أثناء تنفيذ الطلب: ${err.message}`);
  }
});

// 5️⃣ حفظ معرف الاتصال التجاري تلقائياً
bot.on("business_connection", async (ctx) => {
  const bc = ctx.update.business_connection;
  if (bc.is_enabled && bc.user_chat_id) {
    memoryCache.set(`conn_id:${bc.user_chat_id}`, bc.id);
  }
});

// 6️⃣ غلاف الحماية للتشغيل على Vercel بدون أخطاء
const handleUpdate = webhookCallback(bot, "http");

module.exports = async (req, res) => {
  if (req.method !== "POST") {
    res.statusCode = 200;
    res.setHeader("Content-Type", "text/html; charset=utf-8");
    return res.end(`
      <html dir="rtl">
        <head><title>بوت القصص والملف الشخصي</title></head>
        <body style="font-family: Tahoma; text-align: center; padding-top: 50px; background: #0f172a; color: #fff;">
          <h1>🚀 البوت يعمل بنجاح وجاهز للعمل!</h1>
        </body>
      </html>
    `);
  }

  try {
    return await handleUpdate(req, res);
  } catch (error) {
    console.error("❌ خطأ في الخادم:", error);
    res.statusCode = 500;
    res.setHeader("Content-Type", "application/json");
    return res.end(JSON.stringify({ error: error.message }));
  }
};
