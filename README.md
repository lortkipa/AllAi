# Allai

One chat for OpenRouter's free text models. Pick a model, send a message, and switch models at any point in the same in-memory conversation.

## What is included

- React + TypeScript + Vite frontend
- Go backend built entirely on `net/http`
- Live OpenRouter free-model catalog with a 10-minute server cache
- Streaming responses over Server-Sent Events
- Markdown and code rendering
- Stop, retry, copy, clear-chat, responsive, and accessible interaction states
- No accounts, database, payments, saved chats, settings, or image generation

## Run locally

Requirements: Node.js 20+ and Go 1.24+.

1. Open `.env` and set your key:

   ```env
   OPENROUTER_API_KEY=your_key_here
   ```

2. Install the frontend tooling:

   ```bash
   npm install
   ```

3. Start both servers:

   ```bash
   npm run dev
   ```

4. Open [http://localhost:5173](http://localhost:5173).

The Go API runs on `http://localhost:8080`. Vite proxies `/api` requests to it during development.

## Commands

```bash
npm run dev    # React and Go development servers
npm run test   # Go tests and frontend lint
npm run build  # Production frontend and Go executable
```

## Environment variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `OPENROUTER_API_KEY` | empty | Required for model requests |
| `HOST` | `127.0.0.1` | API bind address; local-only by default |
| `PORT` | `8080` | Go API port |
| `FRONTEND_URL` | `http://localhost:5173` | Allowed browser origin and URL sent to OpenRouter |
| `OPENROUTER_APP_NAME` | `Allai` | App name sent to OpenRouter |

The `.env` file is ignored by Git. `.env.example` documents the expected values.

## API

- `GET /api/health`
- `GET /api/models`
- `POST /api/chat/stream`

The chat endpoint accepts only `openrouter/free` and model IDs ending in `:free`, preventing this version from accidentally using a paid model.
