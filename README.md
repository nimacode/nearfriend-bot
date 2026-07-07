# NearFriend Bot

> A Telegram bot in Go that helps users find chat partners nearby — mutual-consent matching, in-memory state, anonymous relay chat.

[![CI](https://img.shields.io/github/actions/workflow/status/nimacode/nearfriend-bot/ci.yml?branch=main&style=flat-square)](https://github.com/nimacode/nearfriend-bot/actions)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&style=flat-square)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)
[![Telegram Bot API](https://img.shields.io/badge/Telegram-Bot%20API-26A5E4?logo=telegram&style=flat-square)](https://core.telegram.org/bots/api)

---

## Features

- 📍 **Location-based matching** — Haversine distance with user-picked radius (1 → 500 km).
- 🚻 **Mutual-consent gender filter** — both sides' preferences must agree before a match is shown.
- 💬 **Anonymous chat relay** — forwards text, photos, voice, stickers, and documents via `copyMessage`.
- ⚡ **Tiny in-memory storage** — single binary, no DB needed for the demo.
- 🤖 **State machine** — every user has a tracked conversation state, so the bot always knows what to do next.

## Demo flow

```
/start
  ↓
[📎 Share my location]   ← Telegram location button
  ↓
[👨 Male | 👩 Female | 🌐 Both]   ← self-declared gender
  ↓
Main menu:
   📍 Update location  |  🚻 Set gender
   🔍 Find nearby friends
   ❌ End chat
       ↓
[👨 Male | 👩 Female | 🌐 Both]   ← who you're looking for
       ↓
[1km | 5km | 10km | 50km | 100km | 500km]
       ↓
Inline list of nearby matches (mutual-consent filtered)
   → tap a name to connect
       ↓
Bot relays messages between you and your match
   → /end or "❌ End chat" to disconnect
```

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

go run .
```

You should see:

```
[nearfriend] bot @your_bot_username is online
```

### 3. Try it out

You'll need **two Telegram accounts** (or a friend) to test matching:

1. From account A: `/start` → share location → set gender (e.g. Female).
2. From account B: same.
3. From account A: 🔍 Find nearby friends → 👩 Female → 1 km → pick B from the list.
4. Both accounts are now in an active chat — anything A sends is forwarded to B and vice versa.

## How matching works (mutual consent)

When user A searches for "Female, within 5 km", candidate B is shown **only if**:

- B is within 5 km of A.
- B's gender is Female (matches A's `LookingFor`).
- A's gender matches B's own `LookingFor` (so B isn't surprised by A's gender).
- B is not already in another chat.

Each user therefore has two fields: `Gender` (their own) and `LookingFor` (their preference).

## Project layout

```
nearfriend-bot/
├── main.go              # entry point — wires up the bot
├── go.mod / go.sum
├── bot/
│   ├── bot.go           # Bot struct + update router
│   ├── handlers.go      # all command / message / callback handlers
│   └── storage.go       # User model, in-memory storage, Haversine distance
├── .github/
│   ├── workflows/ci.yml # CI: build, vet, test
│   ├── ISSUE_TEMPLATE/
│   └── PULL_REQUEST_TEMPLATE.md
├── LICENSE              # MIT
└── README.md
```

## Commands

| Command   | What it does                       |
|-----------|------------------------------------|
| `/start`  | Begin registration                 |
| `/menu`   | Reshow the main menu               |
| `/end`    | Disconnect from current chat       |
| `/cancel` | Same as `/end`                     |

## Persistence

Storage is **in-memory only** — users are wiped on restart. That's deliberate for the demo; the `Storage` type is small and easy to swap. Drop in SQLite (`mattn/go-sqlite3`), Postgres (`pgx`), or Redis by replacing `bot/storage.go` while keeping the same method signatures.

## Configuration

The bot reads exactly one environment variable:

| Variable              | Required | Description                          |
|-----------------------|----------|--------------------------------------|
| `TELEGRAM_BOT_TOKEN`  | ✅        | Bot token from [@BotFather](https://t.me/BotFather) |

## Extending it

Common next steps:

- **Block / report** — add a `Blocked map[int64]map[int64]bool` to `Storage`.
- **Live location** — Telegram's `LiveLocation` messages arrive with `Location` set too; share every update as the user moves.
- **Profile fields** — age, bio, interests. Add to `User` and add a "Set bio" menu item.
- **Reputation** — rate your partner after a chat. Store a `Rating` on the partner.
- **Persistent storage** — swap `Storage` for Postgres or SQLite.
- **BotFather menu commands** — list `/start`, `/menu`, `/end` via `setMyCommands` so users discover them in the UI.

## Contributing

PRs welcome. Run `go build ./...` and `go vet ./...` before pushing — CI will yell at you otherwise.

## License

[MIT](LICENSE) — do whatever you want with it.