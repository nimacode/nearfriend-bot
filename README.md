# NearFriend Bot

> A Telegram bot in Go that helps users find chat partners nearby — mutual-consent matching, anonymous relay chat, live translation, profiles, achievements, and more.

[![CI](https://img.shields.io/github/actions/workflow/status/nimacode/nearfriend-bot/ci.yml?branch=main&style=flat-square)](https://github.com/nimacode/nearfriend-bot/actions)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&style=flat-square)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)
[![Telegram Bot API](https://img.shields.io/badge/Telegram-Bot%20API-26A5E4?logo=telegram&style=flat-square)](https://core.telegram.org/bots/api)

---

## Features

### Matching
- 📍 **Location-based matching** — Haversine distance with user-picked radius (1 → 500 km).
- 🚻 **Mutual-consent gender filter** — both sides' preferences must agree before a match is shown.
- ⏰ **Wake hours** — only show people who are actually awake in their local timezone.
- ⭐ **Rating filter** — users with avg rating below 3.5 (over 3+ reviews) are hidden.
- 🚫 **Block list** — blocked users never appear in your matches.
- 🚩 **Reports** — 3+ reports suspend a user for 24h.

### Profiles
- 🎭 **Alias** — display name separate from your Telegram name.
- 📝 **Bio** — up to 200 chars.
- 🏷️ **Interests** — pick from 22 curated tags (books, code, travel, …).
- 🖼️ **Profile photo** — your photo, shown to matches.
- 🌐 **Language** — your language code, used for live chat translation.

### Chat
- 💬 **Anonymous relay** — forwards text, photos, voice, stickers, documents, locations, **and live locations** via `copyMessage`.
- 🌐 **Live translation** — when two users have different languages, text is auto-translated via LibreTranslate.
- 🧊 **Icebreaker** — a random question is sent to both sides at the start of every chat.
- ☕ **Coffee chat (15 min)** — a timed chat that asks "Keep chatting?" when the timer runs out.
- 🚫 **Block / 🚩 Report** — available from the in-chat keyboard.

### Smart features
- 🔔 **Smart notifications** — get a ping when a new match joins within your chosen radius (checked every 5 min).
- 🏆 **Achievements** — 10 unlockable badges, from "First Chat" to "Well Liked".

### Plumbing
- ⚡ **Tiny in-memory storage** — single binary, no DB needed for the demo.
- 🤖 **State machine** — every user has a tracked conversation state.
- 🛡️ **Thread-safe** — `sync.RWMutex` around all storage access; `WithUser` for atomic read-modify-write.

---

## Demo flow

```
/start
  ↓
[📎 Share my location]   ← Telegram location button
  ↓
[👨 Male | 👩 Female | 🌐 Both]   ← self-declared gender
  ↓
Main menu:
  👤 My profile | 🚻 Set my gender
  📍 Update my location
  🔍 Find nearby | ☕ Coffee chat (15 min)
  🌐 Language    | ⏰ Wake hours
  🔔 Notify me   | 🏆 Achievements
  ❌ End chat
        ↓
[👨 Male | 👩 Female | 🌐 Both]   ← who you're looking for
        ↓
[1km | 5km | 10km | 50km | 100km | 500km]
        ↓
Inline list of nearby matches (mutual-consent + awake + rating + not blocked):
   → Ali · 1.2 km · ♂ · books, code · 4.8⭐
   → Sara · 0.4 km · ♀ · travel, food · 4.2⭐
   ↩️ Cancel
        ↓
Bot connects you. Icebreaker + a hint to share your live location.
  ↓
Chat keyboard:
  ❌ End chat
  🚫 Block | 🚩 Report
        ↓
After the chat: rate your partner 1–5 ⭐ (or skip).
        ↓
🎉 Achievement unlocked: "🥇 First Chat"
```

---

## Quick start

### 1. Create the bot on Telegram

1. Talk to [@BotFather](https://t.me/BotFather).
2. `/newbot` → give it a name & username.
3. Copy the **token**.

### 2. Run

You need **Go 1.21+**.

```bash
git clone https://github.com/nimacode/nearfriend-bot.git
cd nearfriend-bot
go mod download

export TELEGRAM_BOT_TOKEN="123456789:ABCdef…"

# Optional: enable live chat translation via LibreTranslate.
# export TRANSLATION_API_URL="https://libretranslate.com/translate"
# export TRANSLATION_API_KEY="…"   # only if your instance requires it

go run .
```

You should see:

```
[nearfriend] chat translation enabled (https://libretranslate.com/translate)
[nearfriend] bot @your_bot_username is online
```

### 3. Try it out

You'll need **two Telegram accounts** (or a friend) to test matching:

1. From account A: `/start` → share location → set gender.
2. From account B: same.
3. From A: 🔍 Find nearby friends → 👩 Female → 1 km → pick B from the list.
4. Both accounts are now in an active chat — anything A sends is forwarded to B and vice versa. If they speak different languages, text is auto-translated.

---

## Deploy with Docker (recommended)

The fastest way to run the bot on a server. All you need is a token from [@BotFather](https://t.me/BotFather).

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + [Docker Compose](https://docs.docker.com/compose/install/) (v2)

### 1. Configure

```bash
git clone https://github.com/nimacode/nearfriend-bot.git
cd nearfriend-bot

cp .env.example .env
# edit .env and set TELEGRAM_BOT_TOKEN
nano .env
```

`.env` is git-ignored, so your token never gets committed.

### 2. Build & run

```bash
docker compose up -d --build
```

That's it. The bot is now running in the background. Check it:

```bash
docker compose logs -f
# → [nearfriend] bot @your_bot_username is online
```

### 3. Operate

| Action | Command |
|--------|---------|
| View logs | `docker compose logs -f` |
| Restart | `docker compose restart` |
| Stop | `docker compose down` |
| Update to latest code | `git pull && docker compose up -d --build` |

### Optional: enable chat translation

Uncomment and fill these in `.env`:

```env
TRANSLATION_API_URL=https://libretranslate.com/translate
TRANSLATION_API_KEY=your_key_if_required
```

Then `docker compose up -d --build`.

### What's in the image

- Multi-stage build → final image is **~15 MB** (Alpine + static binary).
- Runs as a **non-root** user.
- Ships `ca-certificates` (HTTPS to Telegram/translation API) and `tzdata` (the bot resolves per-user timezones like `Asia/Tehran` for wake-hours & time-based achievements).
- `restart: unless-stopped` — survives crashes and reboots.

### Deploy without Compose

```bash
docker build -t nearfriend-bot .
docker run -d --name nearfriend-bot --restart unless-stopped \
  -e TELEGRAM_BOT_TOKEN="123:ABC..." nearfriend-bot
```

> **Note on persistence:** storage is in-memory, so registered users are wiped whenever the container restarts. For a long-running deployment, swap `bot/storage.go` for a real DB (see [Persistence](#persistence)).

---

## How matching works (mutual consent)

When user A searches for "Female, within 5 km", candidate B is shown **only if**:

- B is within 5 km of A.
- B's gender is Female (matches A's `LookingFor`).
- A's gender matches B's own `LookingFor` (so B isn't surprised by A's gender).
- B is not already in another chat.
- B is not suspended (3+ reports).
- B's avg rating ≥ 3.5 (if they have 3+ reviews).
- B is not blocked by A.
- B is currently awake in their local timezone.

Each user therefore has two fields: `Gender` (their own) and `LookingFor` (their preference).

---

## Achievements

| ID | Title | How |
|----|-------|-----|
| 🥇 | First Chat | Complete your first chat |
| 💬 | Chatterbox | 10 chats |
| 🌍 | Globetrotter | 3+ distinct cities |
| 🗺️ | Explorer | 5+ distinct cities |
| 🌙 | Night Owl | Chat between 0–5 (your local time) |
| 🐦 | Early Bird | Chat before 8 AM (your local time) |
| ⭐ | Five Stars | Receive a 5/5 rating |
| ❤️ | Well Liked | avg ≥ 4.5 over 5+ reviews |
| 🗣️ | Polyglot | Chat with 3+ distinct languages |
| ☕ | Coffee Lover | Complete a coffee chat |

---

## Project layout

```
nearfriend-bot/
├── main.go              # entry point — wires up the bot + translation
├── go.mod / go.sum
├── bot/
│   ├── bot.go           # Bot struct, update router, worker startup
│   ├── handlers.go      # all command / message / callback handlers
│   ├── keyboards.go     # reply + inline keyboards
│   ├── storage.go       # User model, in-memory storage, Haversine
│   ├── translate.go     # LibreTranslate client
│   ├── workers.go       # background timers: coffee chat + notifications
│   ├── achievements.go  # achievement-grant logic
│   └── storage_test.go  # unit tests
├── Dockerfile           # multi-stage build → ~15 MB Alpine image
├── docker-compose.yml   # one-command deploy (reads .env)
├── .github/workflows/ci.yml
├── LICENSE              # MIT
└── README.md
```

---

## Commands

| Command         | What it does                                  |
|-----------------|-----------------------------------------------|
| `/start`        | Begin registration                            |
| `/menu`         | Reshow the main menu                          |
| `/profile`      | Show your profile                             |
| `/language`     | Set your language                             |
| `/hours`        | Set your wake hours                           |
| `/notify`       | Toggle "notify me when someone is nearby"     |
| `/achievements` | Show your achievements                        |
| `/end`          | Disconnect from current chat                  |
| `/cancel`       | Same as `/end`                                |
| `/skip`         | Skip the current input prompt                 |

---

## Configuration

| Variable              | Required | Description                                              |
|-----------------------|----------|----------------------------------------------------------|
| `TELEGRAM_BOT_TOKEN`  | ✅        | Bot token from @BotFather                                |
| `TRANSLATION_API_URL` | ❌        | LibreTranslate endpoint for live chat translation        |
| `TRANSLATION_API_KEY` | ❌        | API key for the translation endpoint (if required)       |

When `TRANSLATION_API_URL` is unset, text messages are relayed as-is and the bot logs a notice at startup.

---

## Persistence

Storage is **in-memory only** — users are wiped on restart. That's deliberate for the demo; the `Storage` type is small and easy to swap. Drop in SQLite (`mattn/go-sqlite3`), Postgres (`pgx`), or Redis by replacing `bot/storage.go` while keeping the same method signatures.

---

## Extending it

Common next steps:

- **Rate limiting** — wrap `handleMessage` in a token-bucket keyed by `msg.From.ID`.
- **Live location persistence** — store `PhotoID`, `Interests`, and a live-location snapshot in storage so the profile is richer.
- **More achievements** — add to `bot/achievements.go`'s `AllAchievements` and the `checkAchievements` switch.
- **Webhook mode** — replace `GetUpdatesChan` with a Gin/Chi HTTP server.
- **Graceful shutdown** — close a `stop` channel in `bot.Run` and let the workers drain.
- **BotFather menu commands** — list `/start`, `/menu`, `/end` via `setMyCommands`.

---

## Contributing

PRs welcome. Run `go build ./...`, `go vet ./...` and `go test ./...` before pushing — CI will yell at you otherwise.

---

## License

[MIT](LICENSE) — do whatever you want with it.
