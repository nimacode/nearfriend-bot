package bot

import "fmt"

// UI language codes supported by the bot interface.
const (
	UILangEn = "en"
	UILangRu = "ru"
	UILangFa = "fa"
)

// ValidUILangs is the ordered list of supported UI languages.
var ValidUILangs = []string{UILangEn, UILangRu, UILangFa}

// UILangNames maps a UI language code → native display name + flag.
var UILangNames = map[string]string{
	UILangEn: "🇬🇧 English",
	UILangRu: "🇷🇺 Русский",
	UILangFa: "🇮🇷 فارسی",
}

// isValidUILang reports whether code is a supported UI language.
func isValidUILang(code string) bool {
	_, ok := UILangNames[code]
	return ok
}

// t returns the translated string for the given language and key.
// Falls back to English, then to the key itself.
func t(lang, key string) string {
	if m, ok := i18n[key]; ok {
		if s, ok := m[lang]; ok && s != "" {
			return s
		}
		if s, ok := m[UILangEn]; ok && s != "" {
			return s
		}
	}
	return key
}

// tf is like t but formats the result with fmt.Sprintf.
func tf(lang, key string, args ...interface{}) string {
	return fmt.Sprintf(t(lang, key), args...)
}

// -----------------------------------------------------------------------
// All UI strings keyed by message id, then by language code.
// -----------------------------------------------------------------------

var i18n = map[string]map[string]string{
	// -- Language picker --------------------------------------------------
	"lang_select": {
		"en": "🌐 Please select your language:",
		"ru": "🌐 Пожалуйста, выберите ваш язык:",
		"fa": "🌐 لطفاً زبان خود را انتخاب کنید:",
	},
	"lang_select_short": {
		"en": "🌐 Select language:",
		"ru": "🌐 Выберите язык:",
		"fa": "🌐 انتخاب زبان:",
	},

	// -- Menu buttons -----------------------------------------------------
	"btn_profile": {
		"en": "👤 My profile",
		"ru": "👤 Мой профиль",
		"fa": "👤 پروفایل من",
	},
	"btn_set_gender": {
		"en": "🚻 Set my gender",
		"ru": "🚻 Указать пол",
		"fa": "🚻 تنظیم جنسیت",
	},
	"btn_update_location": {
		"en": "📍 Update my location",
		"ru": "📍 Обновить геолокацию",
		"fa": "📍 به‌روزرسانی موقعیت",
	},
	"btn_find_nearby": {
		"en": "🔍 Find nearby friends",
		"ru": "🔍 Найти рядом",
		"fa": "🔍 جستجوی دوستان نزدیک",
	},
	"btn_coffee_chat": {
		"en": "☕ Coffee chat (15 min)",
		"ru": "☕ Кофе-чат (15 мин)",
		"fa": "☕ گفتگوی کوتاه (۱۵ دقیقه)",
	},
	"btn_language": {
		"en": "🌐 Language",
		"ru": "🌐 Язык",
		"fa": "🌐 زبان",
	},
	"btn_wake_hours": {
		"en": "⏰ Wake hours",
		"ru": "⏰ Часы бодрствования",
		"fa": "⏰ ساعات بیداری",
	},
	"btn_notify": {
		"en": "🔔 Notify me",
		"ru": "🔔 Уведомлять меня",
		"fa": "🔔 اعلان به من",
	},
	"btn_achievements": {
		"en": "🏆 Achievements",
		"ru": "🏆 Достижения",
		"fa": "🏆 دستاوردها",
	},
	"btn_end_chat": {
		"en": "❌ End chat",
		"ru": "❌ Завершить чат",
		"fa": "❌ پایان گفتگو",
	},

	// -- Chat keyboard buttons -------------------------------------------
	"btn_block": {
		"en": "🚫 Block",
		"ru": "🚫 Заблокировать",
		"fa": "🚫 مسدود کردن",
	},
	"btn_report": {
		"en": "🚩 Report",
		"ru": "🚩 Пожаловаться",
		"fa": "🚩 گزارش دادن",
	},

	// -- Location button --------------------------------------------------
	"btn_share_location": {
		"en": "📎 Share my location",
		"ru": "📎 Поделиться геолокацией",
		"fa": "📎 اشتراک‌گذاری موقعیت من",
	},

	// -- Gender buttons ---------------------------------------------------
	"btn_male": {
		"en": "👨 Male",
		"ru": "👨 Мужской",
		"fa": "👨 مرد",
	},
	"btn_female": {
		"en": "👩 Female",
		"ru": "👩 Женский",
		"fa": "👩 زن",
	},
	"btn_both": {
		"en": "🌐 Both / Any",
		"ru": "🌐 Все",
		"fa": "🌐 هر دو / همه",
	},

	// -- Profile buttons --------------------------------------------------
	"btn_edit_alias": {
		"en": "✏️ Edit alias",
		"ru": "✏️ Изменить псевдоним",
		"fa": "✏️ ویرایش نام مستعار",
	},
	"btn_edit_bio": {
		"en": "📝 Edit bio",
		"ru": "📝 Изменить описание",
		"fa": "📝 ویرایش توضیحات",
	},
	"btn_edit_interests": {
		"en": "🏷️ Edit interests",
		"ru": "🏷️ Изменить интересы",
		"fa": "🏷️ ویرایش علایق",
	},
	"btn_set_photo": {
		"en": "🖼️ Set photo",
		"ru": "🖼️ Установить фото",
		"fa": "🖼️ تنظیم عکس",
	},

	// -- Common buttons ---------------------------------------------------
	"btn_done": {
		"en": "✅ Done",
		"ru": "✅ Готово",
		"fa": "✅ تمام",
	},
	"btn_clear_all": {
		"en": "🗑️ Clear all",
		"ru": "🗑️ Очистить всё",
		"fa": "🗑️ پاک کردن همه",
	},
	"btn_yes": {
		"en": "✅ Yes",
		"ru": "✅ Да",
		"fa": "✅ بله",
	},
	"btn_no": {
		"en": "❌ No",
		"ru": "❌ Нет",
		"fa": "❌ خیر",
	},
	"btn_cancel": {
		"en": "↩️ Cancel",
		"ru": "↩️ Отмена",
		"fa": "↩️ لغو",
	},
	"btn_skip_rating": {
		"en": "⏭ Skip rating",
		"ru": "⏭ Пропустить оценку",
		"fa": "⏭ رد کردن امتیاز",
	},
	"btn_keep_chatting": {
		"en": "✅ Keep chatting",
		"ru": "✅ Продолжить чат",
		"fa": "✅ ادامه گفتگو",
	},
	"btn_end_here": {
		"en": "❌ End here",
		"ru": "❌ Завершить",
		"fa": "❌ پایان در اینجا",
	},

	// -- Common messages --------------------------------------------------
	"msg_send_start": {
		"en": "Send /start to begin.",
		"ru": "Отправьте /start для начала.",
		"fa": "برای شروع /start را ارسال کنید.",
	},
	"msg_main_menu": {
		"en": "Main menu 👇",
		"ru": "Главное меню 👇",
		"fa": "منوی اصلی 👇",
	},
	"msg_use_menu": {
		"en": "Use the menu below 👇",
		"ru": "Воспользуйтесь меню ниже 👇",
		"fa": "از منوی زیر استفاده کنید 👇",
	},
	"msg_unknown_cmd": {
		"en": "Unknown command. Try /start or /menu.",
		"ru": "Неизвестная команда. Попробуйте /start или /menu.",
		"fa": "دستور ناشناخته. /start یا /menu را امتحان کنید.",
	},
	"msg_button_unavailable": {
		"en": "This button is no longer available.",
		"ru": "Эта кнопка больше недоступна.",
		"fa": "این دکمه دیگر در دسترس نیست.",
	},
	"msg_skipped": {
		"en": "Skipped.",
		"ru": "Пропущено.",
		"fa": "رد شد.",
	},
	"msg_nothing_to_skip": {
		"en": "Nothing to skip.",
		"ru": "Нечего пропускать.",
		"fa": "چیزی برای رد کردن نیست.",
	},
	"msg_saved": {
		"en": "Saved!",
		"ru": "Сохранено!",
		"fa": "ذخیره شد!",
	},

	// -- Start / registration --------------------------------------------
	"msg_welcome_back": {
		"en": "Welcome back! 👋",
		"ru": "С возвращением! 👋",
		"fa": "خوش آمدید! 👋",
	},
	"msg_start_intro": {
		"en": "Hey %s! 👋\n\n" +
			"I'm *NearFriend* — I help you meet people nearby for a chat.\n\n" +
			"1️⃣ Share your location with the button below.\n" +
			"2️⃣ Tell me your gender.\n" +
			"3️⃣ Tap *Find nearby friends* whenever you want to match.\n\n" +
			"_Tip: type /skip at any prompt to skip._",
		"ru": "Привет, %s! 👋\n\n" +
			"Я *NearFriend* — помогу найти собеседников поблизости.\n\n" +
			"1️⃣ Поделитесь геолокацией кнопкой ниже.\n" +
			"2️⃣ Укажите свой пол.\n" +
			"3️⃣ Нажимайте *Найти рядом* когда захотите найти собеседника.\n\n" +
			"_Подсказка: введите /skip чтобы пропустить шаг._",
		"fa": "سلام %s! 👋\n\n" +
			"من *NearFriend* هستم — کمک می‌کنم افراد نزدیک خودت رو برای گپ پیدا کنی.\n\n" +
			"۱️⃣ موقعیت خود را با دکمه زیر به اشتراک بگذار.\n" +
			"۲️⃣ جنسیت خود را مشخص کن.\n" +
			"۳️⃣ هر زمان خواستی روی *جستجوی دوستان نزدیک* بزن.\n\n" +
			"_نکته: برای رد کردن هر مرحله /skip را تایپ کن._",
	},
	"msg_got_location": {
		"en": "📍 Got your location! Now tell me your gender:",
		"ru": "📍 Геолокация получена! Теперь укажите свой пол:",
		"fa": "📍 موقعیت شما دریافت شد! حالا جنسیت خود را مشخص کنید:",
	},
	"msg_location_updated": {
		"en": "📍 Location updated!",
		"ru": "📍 Геолокация обновлена!",
		"fa": "📍 موقعیت به‌روزرسانی شد!",
	},
	"msg_live_location_updated": {
		"en": "📍 Live location updated! (This only affects your profile — share live location with a chat partner from inside a chat.)",
		"ru": "📍 Геолокация в реальном времени обновлена! (Это влияет только на ваш профиль — живую геолокацию с собеседником делите из чата.)",
		"fa": "📍 موقعیت زنده به‌روزرسانی شد! (این فقط روی پروفایل شما تأثیر دارد — موقعیت زنده را با هم‌صحبت از داخل چت به اشتراک بگذارید.)",
	},
	"msg_share_location_prompt": {
		"en": "Tap the button to share your current location:",
		"ru": "Нажмите кнопку, чтобы поделиться геолокацией:",
		"fa": "برای اشتراک‌گذاری موقعیت فعلی خود روی دکمه بزنید:",
	},

	// -- Gender -----------------------------------------------------------
	"msg_whats_gender": {
		"en": "What's your gender?",
		"ru": "Какой у вас пол?",
		"fa": "جنسیت شما چیست؟",
	},
	"msg_pick_gender_below": {
		"en": "Please pick your gender with the buttons below 👇",
		"ru": "Пожалуйста, выберите пол кнопками ниже 👇",
		"fa": "لطفاً جنسیت خود را با دکمه‌های زیر انتخاب کنید 👇",
	},
	"fmt_gender_set": {
		"en": "✅ Your gender is set to *%s*.",
		"ru": "✅ Ваш пол установлен: *%s*.",
		"fa": "✅ جنسیت شما تنظیم شد: *%s*.",
	},
	"msg_gender_all_set": {
		"en": "You're all set! Tap *Find nearby friends* when you're ready.",
		"ru": "Всё готово! Нажмите *Найти рядом* когда будете готовы.",
		"fa": "همه چیز آماده است! هر وقت آماده بودی روی *جستجوی دوستان نزدیک* بزن.",
	},

	// -- Search -----------------------------------------------------------
	"msg_share_location_first": {
		"en": "Please share your location first.",
		"ru": "Сначала поделитесь геолокацией.",
		"fa": "لطفاً ابتدا موقعیت خود را به اشتراک بگذارید.",
	},
	"msg_set_gender_first": {
		"en": "Set your gender first:",
		"ru": "Сначала укажите свой пол:",
		"fa": "ابتدا جنسیت خود را مشخص کنید:",
	},
	"msg_who_chat_with": {
		"en": "Who do you want to chat with?",
		"ru": "С кем вы хотите пообщаться?",
		"fa": "می‌خواهی با چه کسی گپ بزنی؟",
	},
	"msg_coffee_who_chat_with": {
		"en": "☕ Coffee chat (15 min) — who do you want to chat with?",
		"ru": "☕ Кофе-чат (15 мин) — с кем вы хотите пообщаться?",
		"fa": "☕ گفتگوی کوتاه (۱۵ دقیقه) — می‌خواهی با چه کسی گپ بزنی؟",
	},
	"msg_pick_search_gender": {
		"en": "Pick who you want to chat with 👇",
		"ru": "Выберите, с кем хотите пообщаться 👇",
		"fa": "انتخاب کنید با چه کسی می‌خواهید گپ بزنید 👇",
	},
	"msg_pick_radius": {
		"en": "Pick a search radius 👇",
		"ru": "Выберите радиус поиска 👇",
		"fa": "شعاع جستجو را انتخاب کنید 👇",
	},
	"msg_search_within": {
		"en": "Search within…",
		"ru": "Искать в радиусе…",
		"fa": "جستجو در محدوده…",
	},
	"fmt_searching_within": {
		"en": "Searching within %.0f km…",
		"ru": "Поиск в радиусе %.0f км…",
		"fa": "جستجو در شعاع %.0f کیلومتر…",
	},
	"fmt_looking_for_radius": {
		"en": "Looking for: *%s*. Now pick a radius 👇",
		"ru": "Ищете: *%s*. Теперь выберите радиус 👇",
		"fa": "جستجو برای: *%s*. حالا شعاع را انتخاب کنید 👇",
	},
	"msg_no_matches": {
		"en": "😕 No nearby matches. Try a bigger radius, or update your preferences.",
		"ru": "😕 Нет людей поблизости. Попробуйте увеличить радиус или измените настройки.",
		"fa": "😕 هیچ شخصی نزدیک شما نیست. شعاع را افزایش دهید یا تنظیمات را تغییر دهید.",
	},
	"msg_no_coffee_matches": {
		"en": "😕 No nearby matches for a coffee chat. Try a bigger radius.",
		"ru": "😕 Нет людей поблизости для кофе-чата. Попробуйте увеличить радиус.",
		"fa": "😕 هیچ شخصی برای گفتگوی کوتاه نزدیک شما نیست. شعاع را افزایش دهید.",
	},
	"fmt_found_nearby": {
		"en": "Found *%d* nearby %s. Pick one 👇",
		"ru": "Найдено *%d* человек поблизости: %s. Выберите 👇",
		"fa": "*%d* نفر نزدیک شما پیدا شد: %s. یکی را انتخاب کنید 👇",
	},
	"fmt_found_coffee_nearby": {
		"en": "☕ Found *%d* nearby %s for a coffee chat. Pick one 👇",
		"ru": "☕ Найдено *%d* человек поблизости для кофе-чата: %s. Выберите 👇",
		"fa": "☕ *%d* نفر برای گفتگوی کوتاه نزدیک شما پیدا شد: %s. یکی را انتخاب کنید 👇",
	},
	"msg_cancelled_dot": {
		"en": "Cancelled.",
		"ru": "Отменено.",
		"fa": "لغو شد.",
	},

	// -- Chat -------------------------------------------------------------
	"msg_in_chat_hint": {
		"en": "You're in a chat — send messages to your partner, or /end to disconnect.",
		"ru": "Вы в чате — отправляйте сообщения собеседнику или /end для выхода.",
		"fa": "شما در حال گپ هستید — پیام بفرستید یا /end را بزنید برای خروج.",
	},
	"msg_not_in_chat": {
		"en": "You're not in a chat right now.",
		"ru": "Сейчас вы не в чате.",
		"fa": "شما در حال حاضر در گپ نیستید.",
	},
	"msg_connected": {
		"en": "Connected!",
		"ru": "Соединено!",
		"fa": "متصل شد!",
	},
	"fmt_now_chatting_with": {
		"en": "✅ You're now chatting with %s. Say hi!",
		"ru": "✅ Вы теперь общаетесь с %s. Поздоровайтесь!",
		"fa": "✅ شما الان با %s در حال گپ هستید. سلام کنید!",
	},
	"fmt_partner_wants_chat": {
		"en": "💬 %s wants to chat with you. Say hi!",
		"ru": "💬 %s хочет пообщаться с вами. Поздоровайтесь!",
		"fa": "💬 %s می‌خواهد با شما گپ بزند. سلام کنید!",
	},
	"msg_live_loc_tip": {
		"en": "📍 _Tip: share your live location with your match!_\n" +
			"_Tap the 📎 next to the text field → Location → Share My Live Location._",
		"ru": "📍 _Подсказка: поделитесь живой геолокацией с собеседником!_\n" +
			"_Нажмите 📎 рядом с полем ввода → Геолокация → Поделиться живой геолокацией._",
		"fa": "📍 _نکته: موقعیت زنده خود را با هم‌صحبت به اشتراک بگذارید!_\n" +
			"_روی 📎 کنار فیلد متن بزنید → Location → Share My Live Location._",
	},
	"msg_icebreaker_prefix": {
		"en": "💡 *Icebreaker:* ",
		"ru": "💡 *Лёд сломан:* ",
		"fa": "💡 *سوال آشنایی:* ",
	},
	"msg_partner_left": {
		"en": "Your partner left. Chat ended.",
		"ru": "Собеседник вышел. Чат завершён.",
		"fa": "هم‌صحبت شما خارج شد. گفتگو پایان یافت.",
	},
	"msg_msg_not_delivered": {
		"en": "⚠️ Could not deliver your message.",
		"ru": "⚠️ Не удалось доставить сообщение.",
		"fa": "⚠️ پیام شما تحویل داده نشد.",
	},
	"msg_chat_already_ended": {
		"en": "Chat already ended",
		"ru": "Чат уже завершён",
		"fa": "گفتگو قبلاً پایان یافته است",
	},
	"msg_partner_left_chat": {
		"en": "💔 Your chat partner has left the conversation.",
		"ru": "💔 Ваш собеседник покинул чат.",
		"fa": "💔 هم‌صحبت شما گفتگو را ترک کرد.",
	},
	"fmt_chat_with_ended": {
		"en": "Chat with %s ended. How was it? Tap a star:",
		"ru": "Чат с %s завершён. Как всё прошло? Оцените:",
		"fa": "گفتگو با %s پایان یافت. چطور بود؟ امتیاز بدهید:",
	},
	"msg_chat_ended_find_new": {
		"en": "Chat ended. Tap *Find nearby friends* to start a new one.",
		"ru": "Чат завершён. Нажмите *Найти рядом* чтобы начать новый.",
		"fa": "گفتگو پایان یافت. برای شروع گفتگوی جدید روی *جستجوی دوستان نزدیک* بزن.",
	},
	"msg_continuing": {
		"en": "Continuing!",
		"ru": "Продолжаем!",
		"fa": "ادامه می‌دهیم!",
	},
	"msg_continuing_no_timer": {
		"en": "✅ Continuing the chat — no more timer.",
		"ru": "✅ Продолжаем чат — без таймера.",
		"fa": "✅ گفتگو ادامه دارد — دیگر تایمر وجود ندارد.",
	},
	"msg_coffee_time_up": {
		"en": "⏰ *Coffee time's up!*\nWould you like to keep chatting?",
		"ru": "⏰ *Время кофе-чата истекло!*\nХотите продолжить общение?",
		"fa": "⏰ *زمان گفتگوی کوتاه تمام شد!*\nمی‌خواهید گفتگو را ادامه دهید؟",
	},

	// -- Rating -----------------------------------------------------------
	"msg_rate_chat": {
		"en": "Tap a star to rate your chat 👇",
		"ru": "Нажмите звезду чтобы оценить чат 👇",
		"fa": "برای امتیاز دادن به گفتگو ستاره بزنید 👇",
	},
	"fmt_rated_stars": {
		"en": "Rated %d ⭐",
		"ru": "Оценка: %d ⭐",
		"fa": "امتیاز: %d ⭐",
	},
	"fmt_you_rated": {
		"en": "✅ You rated your partner %d ⭐",
		"ru": "✅ Вы оценили собеседника на %d ⭐",
		"fa": "✅ شما به هم‌صحبت خود %d ⭐ دادید",
	},
	"msg_thanks_feedback": {
		"en": "Thanks for the feedback!",
		"ru": "Спасибо за отзыв!",
		"fa": "از بازخورد شما سپاسگزاریم!",
	},
	"msg_skipped_rating": {
		"en": "Skipped rating.",
		"ru": "Оценка пропущена.",
		"fa": "امتیازدهی رد شد.",
	},
	"msg_no_problem_menu": {
		"en": "No problem — back to the main menu.",
		"ru": "Без проблем — возврат в главное меню.",
		"fa": "اشکالی ندارد — بازگشت به منوی اصلی.",
	},

	// -- Block / Report ---------------------------------------------------
	"fmt_report_confirm": {
		"en": "🚩 Report %s for bad behavior?",
		"ru": "🚩 Пожаловаться на %s за плохое поведение?",
		"fa": "🚩 از %s گزارش بدهید؟",
	},
	"fmt_block_confirm": {
		"en": "🚫 Block %s? You won't see them in future searches.",
		"ru": "🚫 Заблокировать %s? Они не будут появляться в будущем поиске.",
		"fa": "🚫 %s مسدود شود؟ در جستجوهای آینده دیده نخواهد شد.",
	},
	"msg_reported": {
		"en": "🚩 Thanks — we'll review it.",
		"ru": "🚩 Спасибо — мы рассмотрим это.",
		"fa": "🚩 سپاس — بررسی خواهیم کرد.",
	},
	"msg_blocked": {
		"en": "🚫 User blocked.",
		"ru": "🚫 Пользователь заблокирован.",
		"fa": "🚫 کاربر مسدود شد.",
	},
	"msg_suspended": {
		"en": "⏸ You've been temporarily suspended for 24h due to multiple reports. " +
			"You can still browse, but you won't appear in others' results.",
		"ru": "⏸ Вы временно отстранены на 24 часа из-за нескольких жалоб. " +
			"Вы можете просматривать, но не будете появляться в результатах других.",
		"fa": "⏸ شما به دلیل گزارش‌های متعدد به مدت ۲۴ ساعت معلق شده‌اید. " +
			"می‌توانید مرور کنید اما در نتایج دیگران نمایش داده نخواهید شد.",
	},

	// -- Profile ----------------------------------------------------------
	"msg_alias_prompt": {
		"en": "📛 Send me your alias (max 32 chars), or /skip:",
		"ru": "📛 Введите псевдоним (макс. 32 символа) или /skip:",
		"fa": "📛 نام مستعار خود را بفرستید (حداکثر ۳۲ کاراکتر) یا /skip:",
	},
	"msg_bio_prompt": {
		"en": "📝 Send me a short bio (max 200 chars), or /skip:",
		"ru": "📝 Введите краткое описание (макс. 200 символов) или /skip:",
		"fa": "📝 توضیحات کوتاه بنویسید (حداکثر ۲۰۰ کاراکتر) یا /skip:",
	},
	"msg_interests_prompt": {
		"en": "🏷️ Tap tags to toggle, then press *Done*:",
		"ru": "🏷️ Нажимайте теги чтобы выбрать, затем *Готово*:",
		"fa": "🏷️ روی تگ‌ها بزنید تا انتخاب شوند، سپس *تمام* را بزنید:",
	},
	"msg_interests_prompt_2": {
		"en": "Tap the tags below to toggle them, then press *Done*.",
		"ru": "Нажимайте теги ниже чтобы выбрать их, затем нажмите *Готово*.",
		"fa": "روی تگ‌های زیر بزنید تا انتخاب شوند، سپس *تمام* را بزنید.",
	},
	"msg_photo_prompt": {
		"en": "🖼️ Send me a profile photo, or /skip:",
		"ru": "🖼️ Отправьте фото профиля или /skip:",
		"fa": "🖼️ عکس پروفایل خود را بفرستید یا /skip:",
	},
	"msg_alias_empty": {
		"en": "Alias can't be empty. Try again or /skip.",
		"ru": "Псевдоним не может быть пустым. Попробуйте снова или /skip.",
		"fa": "نام مستعار نمی‌تواند خالی باشد. دوباره تلاش کنید یا /skip.",
	},
	"fmt_alias_set": {
		"en": "✅ Alias set to *%s*",
		"ru": "✅ Псевдоним установлен: *%s*",
		"fa": "✅ نام مستعار تنظیم شد: *%s*",
	},
	"msg_bio_saved": {
		"en": "✅ Bio saved.",
		"ru": "✅ Описание сохранено.",
		"fa": "✅ توضیحات ذخیره شد.",
	},
	"msg_interests_saved": {
		"en": "✅ Interests saved.",
		"ru": "✅ Интересы сохранены.",
		"fa": "✅ علایق ذخیره شد.",
	},
	"msg_photo_updated": {
		"en": "✅ Profile photo updated.",
		"ru": "✅ Фото профиля обновлено.",
		"fa": "✅ عکس پروفایل به‌روزرسانی شد.",
	},
	"msg_send_photo_or_skip": {
		"en": "Please send a photo, or /skip.",
		"ru": "Пожалуйста, отправьте фото или /skip.",
		"fa": "لطفاً عکس بفرستید یا /skip.",
	},

	// Profile labels
	"p_your_profile": {
		"en": "👤 *Your profile*",
		"ru": "👤 *Ваш профиль*",
		"fa": "👤 *پروفایل شما*",
	},
	"p_name": {
		"en": "📛 *Name:*",
		"ru": "📛 *Имя:*",
		"fa": "📛 *نام:*",
	},
	"p_alias": {
		"en": "🎭 *Alias:*",
		"ru": "🎭 *Псевдоним:*",
		"fa": "🎭 *نام مستعار:*",
	},
	"p_bio": {
		"en": "📝 *Bio:*",
		"ru": "📝 *Описание:*",
		"fa": "📝 *توضیحات:*",
	},
	"p_interests": {
		"en": "🏷️ *Interests:*",
		"ru": "🏷️ *Интересы:*",
		"fa": "🏷️ *علایق:*",
	},
	"p_language": {
		"en": "🌐 *Language:*",
		"ru": "🌐 *Язык:*",
		"fa": "🌐 *زبان:*",
	},
	"p_wake_hours": {
		"en": "⏰ *Wake hours:*",
		"ru": "⏰ *Часы бодрствования:*",
		"fa": "⏰ *ساعات بیداری:*",
	},
	"p_rating": {
		"en": "⭐ *Rating:*",
		"ru": "⭐ *Рейтинг:*",
		"fa": "⭐ *امتیاز:*",
	},
	"p_notifications": {
		"en": "🔔 *Notifications:*",
		"ru": "🔔 *Уведомления:*",
		"fa": "🔔 *اعلان‌ها:*",
	},
	"p_not_set": {
		"en": "_(not set)_",
		"ru": "_(не задано)_",
		"fa": "_(تنظیم نشده)_",
	},
	"p_none": {
		"en": "_(none)_",
		"ru": "_(нет)_",
		"fa": "_(هیچ)_",
	},
	"p_not_set_lang": {
		"en": "_(not set — chat will not be translated)_",
		"ru": "_(не задано — чат не будет переводиться)_",
		"fa": "_(تنظیم نشده — گفتگو ترجمه نخواهد شد)_",
	},
	"p_not_set_wake": {
		"en": "_(not set — always shown)_",
		"ru": "_(не задано — всегда видны)_",
		"fa": "_(تنظیم نشده — همیشه نمایش داده می‌شوید)_",
	},
	"p_no_ratings": {
		"en": "_(no ratings yet)_",
		"ru": "_(пока без оценок)_",
		"fa": "_(هنوز امتیازی ندارد)_",
	},
	"p_off": {
		"en": "off",
		"ru": "выкл",
		"fa": "خاموش",
	},
	"fmt_p_on": {
		"en": "on, within %.0f km",
		"ru": "вкл, в радиусе %.0f км",
		"fa": "روشن، در شعاع %.0f کیلومتر",
	},
	"fmt_p_rating_detail": {
		"en": "%.1f ⭐ (%d reviews)",
		"ru": "%.1f ⭐ (%d отзывов)",
		"fa": "%.1f ⭐ (%d نظر)",
	},
	"fmt_p_wake_detail": {
		"en": "`%02d:00 - %02d:00` (%s)",
		"ru": "`%02d:00 - %02d:00` (%s)",
		"fa": "`%02d:00 - %02d:00` (%s)",
	},

	// -- Language ---------------------------------------------------------
	"fmt_lang_set": {
		"en": "🌐 Language set to *%s*.",
		"ru": "🌐 Язык установлен: *%s*.",
		"fa": "🌐 زبان تنظیم شد: *%s*.",
	},
	"msg_lang_translation_info": {
		"en": "From now on, text chats with someone speaking a different language will be auto-translated.",
		"ru": "Теперь текстовые сообщения собеседнику с другим языком будут автоматически переводиться.",
		"fa": "از این پس، پیام‌های متنی با شخصی که زبان متفاوتی دارد به‌طور خودکار ترجمه می‌شوند.",
	},

	// -- Wake hours -------------------------------------------------------
	"msg_pick_timezone": {
		"en": "🌍 First, pick your timezone:",
		"ru": "🌍 Сначала выберите часовой пояс:",
		"fa": "🌍 ابتدا منطقه زمانی خود را انتخاب کنید:",
	},
	"fmt_timezone_set": {
		"en": "🌍 Timezone: *%s*\n\nSend the hour you usually wake up (0-23), or /skip to clear wake hours:",
		"ru": "🌍 Часовой пояс: *%s*\n\nВведите час пробуждения (0-23) или /skip чтобы очистить:",
		"fa": "🌍 منطقه زمانی: *%s*\n\nساعت بیداری خود را بفرستید (۰-۲۳) یا /skip برای پاک کردن:",
	},
	"msg_send_wake_from": {
		"en": "Now send the hour you usually wake up (0-23):",
		"ru": "Теперь введите час пробуждения (0-23):",
		"fa": "حالا ساعت بیداری خود را بفرستید (۰-۲۳):",
	},
	"fmt_wake_from_set": {
		"en": "Wake from *%d:00*. Now send the hour you usually go to sleep (0-23):",
		"ru": "Пробуждение с *%d:00*. Теперь введите час сна (0-23):",
		"fa": "بیداری از *%d:00*. حالا ساعت خواب را بفرستید (۰-۲۳):",
	},
	"fmt_wake_from_set_named": {
		"en": "Wake from *%02d:00*. Now send the hour you usually go to sleep (0-23):",
		"ru": "Пробуждение с *%02d:00*. Теперь введите час сна (0-23):",
		"fa": "بیداری از *%02d:00*. حالا ساعت خواب را بفرستید (۰-۲۳):",
	},
	"msg_send_hour_0_23": {
		"en": "Send a number 0-23, or /skip.",
		"ru": "Введите число 0-23 или /skip.",
		"fa": "یک عدد ۰-۲۳ بفرستید یا /skip.",
	},
	"fmt_wake_hours_set": {
		"en": "✅ Wake hours: *%02d:00 - %02d:00*",
		"ru": "✅ Часы бодрствования: *%02d:00 - %02d:00*",
		"fa": "✅ ساعات بیداری: *%02d:00 - %02d:00*",
	},

	// -- Notifications ----------------------------------------------------
	"msg_share_location_for_notify": {
		"en": "Share your location first.",
		"ru": "Сначала поделитесь геолокацией.",
		"fa": "ابتدا موقعیت خود را به اشتراک بگذارید.",
	},
	"msg_notifications_disabled": {
		"en": "🔕 Notifications disabled.",
		"ru": "🔕 Уведомления отключены.",
		"fa": "🔕 اعلان‌ها غیرفعال شد.",
	},
	"msg_notify_prompt": {
		"en": "🔔 *Get notified when someone new joins nearby.*\nPick a radius:",
		"ru": "🔔 *Получайте уведомления о новых людях поблизости.*\nВыберите радиус:",
		"fa": "🔔 *از افراد جدید در نزدیکی شما مطلع شوید.*\nشعاع را انتخاب کنید:",
	},
	"msg_pick_notify_radius": {
		"en": "Pick a notification radius 👇",
		"ru": "Выберите радиус уведомлений 👇",
		"fa": "شعاع اعلان را انتخاب کنید 👇",
	},
	"fmt_notify_set": {
		"en": "🔔 You'll be notified when someone joins within %.0f km.",
		"ru": "🔔 Вы будете получать уведомления когда кто-то появится в радиусе %.0f км.",
		"fa": "🔔 وقتی کسی در شعاع %.0f کیلومتری ظاهر شود به شما اعلان می‌شود.",
	},
	"msg_notify_turn_off": {
		"en": "You can turn this off any time from the menu.",
		"ru": "Вы можете отключить это в любое время из меню.",
		"fa": "می‌توانید هر زمان از منو آن را خاموش کنید.",
	},

	// -- Notify worker ----------------------------------------------------
	"fmt_notify_someone_nearby": {
		"en": "🔔 *Someone new is nearby!*\n👤 %s · %s away\n\nTap *Find nearby friends* to connect!",
		"ru": "🔔 *Кто-то новый поблизости!*\n👤 %s · %s от вас\n\nНажмите *Найти рядом* чтобы связаться!",
		"fa": "🔔 *فرد جدیدی نزدیک شماست!*\n👤 %s · %s دور\n\nبرای اتصال روی *جستجوی دوستان نزدیک* بزنید!",
	},

	// -- Gender labels ----------------------------------------------------
	"label_male": {
		"en": "Male",
		"ru": "Мужской",
		"fa": "مرد",
	},
	"label_female": {
		"en": "Female",
		"ru": "Женский",
		"fa": "زن",
	},
	"label_both": {
		"en": "Both / Any",
		"ru": "Все",
		"fa": "هر دو / همه",
	},
	"label_unknown_gender": {
		"en": "Unknown",
		"ru": "Неизвестно",
		"fa": "نامشخص",
	},
	"label_stranger": {
		"en": "Stranger",
		"ru": "Незнакомец",
		"fa": "غریبه",
	},
	"label_got_it": {
		"en": "Got it",
		"ru": "Принято",
		"fa": "باشه",
	},
	"label_cancelled": {
		"en": "Cancelled",
		"ru": "Отменено",
		"fa": "لغو شد",
	},

	// -- Misc errors ------------------------------------------------------
	"err_please_start": {
		"en": "Please /start first",
		"ru": "Сначала /start",
		"fa": "ابتدا /start کنید",
	},
	"err_bad_value": {
		"en": "Bad value",
		"ru": "Неверное значение",
		"fa": "مقدار نامعتبر",
	},
	"err_bad_radius": {
		"en": "Bad radius",
		"ru": "Неверный радиус",
		"fa": "شعاع نامعتبر",
	},
	"err_bad_language": {
		"en": "Bad language",
		"ru": "Неверный язык",
		"fa": "زبان نامعتبر",
	},
	"err_bad_timezone": {
		"en": "Bad timezone",
		"ru": "Неверный часовой пояс",
		"fa": "منطقه زمانی نامعتبر",
	},
	"err_bad_hour": {
		"en": "Bad hour",
		"ru": "Неверный час",
		"fa": "ساعت نامعتبر",
	},
	"err_bad_rating": {
		"en": "Bad rating",
		"ru": "Неверная оценка",
		"fa": "امتیاز نامعتبر",
	},
	"err_bad_user_id": {
		"en": "Bad user id",
		"ru": "Неверный ID",
		"fa": "شناسه کاربر نامعتبر",
	},
	"err_bad_id": {
		"en": "Bad id",
		"ru": "Неверный ID",
		"fa": "شناسه نامعتبر",
	},
	"err_bad_request": {
		"en": "Bad request",
		"ru": "Неверный запрос",
		"fa": "درخواست نامعتبر",
	},
	"err_bad_tag": {
		"en": "Bad tag",
		"ru": "Неверный тег",
		"fa": "تگ نامعتبر",
	},
	"err_bad_page": {
		"en": "Bad page",
		"ru": "Неверная страница",
		"fa": "صفحه نامعتبر",
	},
	"err_not_waiting_radius": {
		"en": "Not waiting for radius",
		"ru": "Не ожидается радиус",
		"fa": "در انتظار شعاع نیست",
	},
	"err_user_unavailable": {
		"en": "User unavailable",
		"ru": "Пользователь недоступен",
		"fa": "کاربر در دسترس نیست",
	},
	"err_already_chatting": {
		"en": "Already chatting",
		"ru": "Уже в чате",
		"fa": "در حال گپ است",
	},
	"err_already_rated": {
		"en": "Already rated or chat ended",
		"ru": "Уже оценено или чат завершён",
		"fa": "قبلاً امتیاز داده شده یا گفتگو پایان یافته",
	},
	"err_pick_fresh_search": {
		"en": "Pick a fresh search first",
		"ru": "Сначала начните новый поиск",
		"fa": "ابتدا یک جستجوی جدید شروع کنید",
	},
	"err_unknown_action": {
		"en": "Unknown action",
		"ru": "Неизвестное действие",
		"fa": "عملیات ناشناخته",
	},
	"err_unknown_gender_scope": {
		"en": "Unknown gender scope",
		"ru": "Неизвестная область пола",
		"fa": "محدوده جنسیت ناشناخته",
	},
	"err_unknown_profile_action": {
		"en": "Unknown profile action",
		"ru": "Неизвестное действие профиля",
		"fa": "عملیات پروفایل ناشناخته",
	},
	"err_unknown_interest_action": {
		"en": "Unknown interest action",
		"ru": "Неизвестное действие интереса",
		"fa": "عملیات علاقه ناشناخته",
	},
	"err_unknown": {
		"en": "Unknown",
		"ru": "Неизвестно",
		"fa": "نامشخص",
	},
	"err_sleeping": {
		"en": "They're sleeping right now 😴",
		"ru": "Они сейчас спят 😴",
		"fa": "آن‌ها الان خواب هستند 😴",
	},
	"err_unknown_lang_scope": {
		"en": "Unknown language scope",
		"ru": "Неизвестная область языка",
		"fa": "محدوده زبان ناشناخته",
	},

	// -- Achievements -----------------------------------------------------
	"msg_achievements_title": {
		"en": "🏆 *Your achievements*\n\n",
		"ru": "🏆 *Ваши достижения*\n\n",
		"fa": "🏆 *دستاوردهای شما*\n\n",
	},
	"fmt_achievements_count": {
		"en": "_%d unlocked · %d to go_",
		"ru": "_%d получено · %d осталось_",
		"fa": "_%d کسب شد · %d باقی مانده_",
	},
	"msg_achievement_unlocked": {
		"en": "🎉 Achievement unlocked!\n%s *%s* — %s",
		"ru": "🎉 Достижение разблокировано!\n%s *%s* — %s",
		"fa": "🎉 دستاورد جدید!\n%s *%s* — %s",
	},

	// Achievement titles & descriptions
	"ach.first_chat.title": {
		"en": "First Chat", "ru": "Первый чат", "fa": "اولین گفتگو",
	},
	"ach.first_chat.desc": {
		"en": "You had your first conversation!",
		"ru": "Вы провели первую беседу!",
		"fa": "اولین گفتگوی خود را انجام دادید!",
	},
	"ach.multi_city.title": {
		"en": "Globetrotter", "ru": "Глобус", "fa": "جهانگرد",
	},
	"ach.multi_city.desc": {
		"en": "Chatted with people from 3+ cities",
		"ru": "Общались с людьми из 3+ городов",
		"fa": "با افراد ۳ شهر مختلف گپ زدید",
	},
	"ach.night_owl.title": {
		"en": "Night Owl", "ru": "Сова", "fa": "شب‌بیدار",
	},
	"ach.night_owl.desc": {
		"en": "Chatted between midnight and 5 AM",
		"ru": "Общались с полуночи до 5 утра",
		"fa": "بین نیمه‌شب تا ۵ صبح گپ زدید",
	},
	"ach.five_star.title": {
		"en": "Five Stars", "ru": "Пять звёзд", "fa": "پنج ستاره",
	},
	"ach.five_star.desc": {
		"en": "Received a perfect 5/5 rating",
		"ru": "Получили идеальную оценку 5/5",
		"fa": "امتیاز کامل ۵/۵ دریافت کردید",
	},
	"ach.chatterbox.title": {
		"en": "Chatterbox", "ru": "Болтун", "fa": "پرحرف",
	},
	"ach.chatterbox.desc": {
		"en": "Had 10 conversations",
		"ru": "Провели 10 бесед",
		"fa": "۱۰ گفتگو انجام دادید",
	},
	"ach.polyglot.title": {
		"en": "Polyglot", "ru": "Полиглот", "fa": "چندزبانه",
	},
	"ach.polyglot.desc": {
		"en": "Chatted with 3+ different languages",
		"ru": "Общались на 3+ разных языках",
		"fa": "با ۳+ زبان مختلف گپ زدید",
	},
	"ach.early_bird.title": {
		"en": "Early Bird", "ru": "Жаворонок", "fa": "سحرخیز",
	},
	"ach.early_bird.desc": {
		"en": "Chatted before 8 AM",
		"ru": "Общались до 8 утра",
		"fa": "قبل از ساعت ۸ صبح گپ زدید",
	},
	"ach.explorer.title": {
		"en": "Explorer", "ru": "Исследователь", "fa": "کاوشگر",
	},
	"ach.explorer.desc": {
		"en": "Chatted in 5+ different cities",
		"ru": "Общалиись в 5+ разных городах",
		"fa": "در ۵+ شهر مختلف گپ زدید",
	},
	"ach.well_liked.title": {
		"en": "Well Liked", "ru": "Всеобщий любимец", "fa": "محبوب",
	},
	"ach.well_liked.desc": {
		"en": "Average rating >= 4.5 with 5+ reviews",
		"ru": "Средний рейтинг >= 4.5 при 5+ отзывах",
		"fa": "میانگین امتیاز ۴.۵+ با ۵+ نظر",
	},
	"ach.coffee_lover.title": {
		"en": "Coffee Lover", "ru": "Любитель кофе", "fa": "عاشق قهوه",
	},
	"ach.coffee_lover.desc": {
		"en": "Completed a coffee chat",
		"ru": "Завершили кофе-чат",
		"fa": "یک گفتگوی کوتاه را کامل کردید",
	},

	// -- Icebreakers ------------------------------------------------------
	"ice_0": {
		"en": "If you could travel anywhere right now, where would you go? ✈️",
		"ru": "Если бы вы могли поехать куда угодно прямо сейчас, куда бы вы отправились? ✈️",
		"fa": "اگر الان می‌توانستید به هر جایی سفر کنید، کجا می‌رفتید؟ ✈️",
	},
	"ice_1": {
		"en": "What's the last book you read? 📚",
		"ru": "Какую книгу вы прочитали последней? 📚",
		"fa": "آخرین کتابی که خواندید چه بود؟ 📚",
	},
	"ice_2": {
		"en": "Coffee or tea? ☕",
		"ru": "Кофе или чай? ☕",
		"fa": "قهوه یا چای؟ ☕",
	},
	"ice_3": {
		"en": "What's your favorite way to spend a weekend?",
		"ru": "Как вы любите проводить выходные?",
		"fa": "بهترین راه برای گذراندن آخر هفته برایتان چیست؟",
	},
	"ice_4": {
		"en": "If you could have dinner with anyone (alive or dead), who would it be? 🍽️",
		"ru": "С кем бы вы хотели поужинать (живым или умершим)? 🍽️",
		"fa": "اگر می‌توانستید با هر کسی شام بخورید (زنده یا درگذشته)، با کی بود؟ 🍽️",
	},
	"ice_5": {
		"en": "What's a skill you'd love to learn? 🎯",
		"ru": "Какой навык вы бы хотели освоить? 🎯",
		"fa": "چه مهارتی دوست دارید یاد بگیرید؟ 🎯",
	},
	"ice_6": {
		"en": "What's the best concert or festival you've been to? 🎵",
		"ru": "На лучшем концерте или фестивале, где вы были? 🎵",
		"fa": "بهترین کنسرت یا جشنواره‌ای که شرکت کرده‌اید چه بود؟ 🎵",
	},
	"ice_7": {
		"en": "Cats or dogs? 🐱🐶",
		"ru": "Кошки или собаки? 🐱🐶",
		"fa": "گربه یا سگ؟ 🐱🐶",
	},
	"ice_8": {
		"en": "What's something you're proud of from the past year? 🌟",
		"ru": "Чем вы гордитесь за прошедший год? 🌟",
		"fa": "از چه چیزی در سال گذشته افتخار می‌کنید؟ 🌟",
	},
	"ice_9": {
		"en": "If you won the lottery tomorrow, what's the first thing you'd do? 💰",
		"ru": "Если бы вы завтра выиграли в лотерею, что бы вы сделали в первую очередь? 💰",
		"fa": "اگر فردا در لاتاری بُردید، اولین کاری که می‌کنید چه بود؟ 💰",
	},
	"ice_10": {
		"en": "Three things you can't live without?",
		"ru": "Три вещи, без которых вы не можете жить?",
		"fa": "سه چیزی که بدون آن‌ها نمی‌توانید زندگی کنید؟",
	},
	"ice_11": {
		"en": "Most embarrassing song on your playlist? 🎧",
		"ru": "Самая неловкая песня в вашем плейлисте? 🎧",
		"fa": "خجالت‌آورترین آهنگ در لیست پخش شما؟ 🎧",
	},
	"ice_12": {
		"en": "Morning person or night owl?",
		"ru": "Жаворонок или сова?",
		"fa": "سحرخیز هستید یا شب‌بیدار؟",
	},
	"ice_13": {
		"en": "What's your hidden talent?",
		"ru": "Какой у вас скрытый талант?",
		"fa": "استعداد پنهان شما چیست؟",
	},
	"ice_14": {
		"en": "Best meal you've ever had? 🍕",
		"ru": "Лучшая еда, которую вы когда-либо ели? 🍕",
		"fa": "بهترین غذایی که تا به حال خورده‌اید؟ 🍕",
	},
}

// achTitle returns the translated achievement title.
func achTitle(lang, id string) string {
	return t(lang, "ach."+id+".title")
}

// achDesc returns the translated achievement description.
func achDesc(lang, id string) string {
	return t(lang, "ach."+id+".desc")
}

// icebreaker returns a translated icebreaker question by index.
func icebreaker(lang string, idx int) string {
	return t(lang, fmt.Sprintf("ice_%d", idx))
}
