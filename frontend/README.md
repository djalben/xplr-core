# XPLR Frontend

Веб-интерфейс (фронтенд) финтех-платформы XPLR: React, Dashboard, карты, команды, рефералы, настройки (Telegram).

## Стек

- React, React Router
- Axios, Chart.js
- Vite / Create React App (см. package.json)

## Запуск

```bash
npm install
npm run dev
# или
npm run build
npm run preview
```

Для production-сборки с nginx см. `Dockerfile` в корне.

## API

Фронтенд обращается к бэкенду **XPLR Backend** по адресу из переменной окружения (например `VITE_API_URL` или proxy в dev). По умолчанию — тот же хост на порту 8080.

Подробная документация — в папке **docs/**.
