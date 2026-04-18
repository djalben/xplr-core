# XPLR

Финтех-платформа для управления виртуальными картами.

## Структура проекта (Clean Architecture)

```
XPLR/
├── backend/                 # Go API Server (Clean Architecture)
│   ├── cmd/server/          # Entry point (main.go)
│   ├── domain/              # Data models / entities
│   ├── handler/             # HTTP handlers (API routes)
│   ├── middleware/          # Auth, admin, rate-limit middleware
│   ├── repository/          # Database access layer
│   ├── service/             # External integrations (email, rates, etc.)
│   ├── usecase/             # Business logic (transactions, auto-replenish)
│   ├── providers/           # Card/eSIM provider wrappers
│   ├── notification/        # Telegram notification helpers
│   ├── telegram/            # Telegram bot integration
│   ├── shop/                # Store fulfillment engine
│   ├── configs/             # App configuration
│   ├── pkg/utils/           # Shared utilities (JWT, password, etc.)
│   └── Dockerfile
├── frontend/                # React Web App (Vite + Tailwind)
│   ├── src/
│   │   ├── components/      # UI components
│   │   ├── pages/           # Page views
│   │   ├── services/        # API clients (axios)
│   │   ├── store/           # State management (contexts)
│   │   ├── i18n/            # Internationalization
│   │   └── styles.css       # Global styles
│   ├── package.json
│   └── Dockerfile
├── api/                     # Vercel serverless handlers
├── docs/                    # Documentation
├── docker-compose.yml
├── go.mod
└── vercel.json
```

## Запуск

### Backend
```bash
cd backend/cmd/server
go run .
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Docker
```bash
docker-compose up --build
```
