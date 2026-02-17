# Настройка трёх репозиториев на GitHub и приглашение vardump@inbox.ru

По правилам программиста созданы три структуры:

1. **epn-killer-mvp** — бэкенд/MVP: в корне `go.mod`, `go.sum`, `Makefile`, `README.md`, `schema.sql`, `Dockerfile`; папки `cmd/` (main.go) и `internal/` (остальной код).
2. **epn-killer-web** — веб-морда (фронтенд).
3. **epn-killer-docs** — документация, ТЗ, старт, отчёты, мануалы.

---

## Шаг 1. Перенести папки в основную папку проекта

- **README основной папки:** в **C:\Users\aalab\epn-killer-project** должен лежать актуальный **README.md** полного проекта (Docker, запуск, API). Он есть в воркспейсе yzg; при необходимости скопируйте: `Copy-Item "C:\Users\aalab\.cursor\worktrees\epn-killer-project\yzg\README.md" "C:\Users\aalab\epn-killer-project\README.md" -Force`.
- Из воркспейса (yzg) скопировать в **C:\Users\aalab\epn-killer-project** три готовые папки:

- `epn-killer-mvp` (уже со структурой cmd + internal)
- `epn-killer-web` (копия frontend + README)
- `epn-killer-docs` (README, START.md; папку ТЗ с docx при необходимости скопировать вручную из основной папки в epn-killer-docs)

Команда (PowerShell, из папки yzg):

```powershell
$src = "C:\Users\aalab\.cursor\worktrees\epn-killer-project\yzg"
$dst = "C:\Users\aalab\epn-killer-project"
Copy-Item -Path "$src\epn-killer-mvp" -Destination "$dst\epn-killer-mvp" -Recurse -Force
Copy-Item -Path "$src\epn-killer-web" -Destination "$dst\epn-killer-web" -Recurse -Force
Copy-Item -Path "$src\epn-killer-docs" -Destination "$dst\epn-killer-docs" -Recurse -Force
# ТЗ (docx) — при необходимости:
# Copy-Item -Path "$dst\ТЗ\*" -Destination "$dst\epn-killer-docs\" -Recurse -Force
```

---

## Шаг 2. Создать три репозитория на GitHub

1. Зайти на https://github.com/new (или через меню **Create repository**).
2. Создать репозитории (можно пустые, без README):
   - **epn-killer-mvp**
   - **epn-killer-web**
   - **epn-killer-docs**

Владелец: ваш аккаунт (например djalben). Видимость — по желанию (Private/Public).

---

## Шаг 3. Полностью почистить старый репозиторий epn-killer-backend-mvp (если используете его под epn-killer-mvp)

Если репозиторий **https://github.com/djalben/epn-killer-backend-mvp** будет заменён на **epn-killer-mvp**:

1. Клонировать его в отдельную папку (или уже есть клон).
2. Удалить всё содержимое (кроме `.git`), например:
   ```powershell
   cd путь\к\клону\epn-killer-backend-mvp
   Get-ChildItem -Force | Where-Object { $_.Name -ne ".git" } | Remove-Item -Recurse -Force
   ```
3. Скопировать в эту папку содержимое из **C:\Users\aalab\epn-killer-project\epn-killer-mvp** (все файлы и папки: go.mod, go.sum, Makefile, README.md, schema.sql, Dockerfile, cmd/, internal/).
4. Переименовать репозиторий на GitHub в **epn-killer-mvp** (Settings → Repository name) **или** создать новый репозиторий **epn-killer-mvp** и запушить туда содержимое этой папки, старый потом удалить или оставить архивом.

Рекомендация: проще создать **новый** репозиторий **epn-killer-mvp** и первый раз запушить туда папку epn-killer-mvp; старый epn-killer-backend-mvp потом удалить или не использовать.

---

## Шаг 4. Пригласить vardump@inbox.ru во все три репозитория

1. Открыть репозиторий на GitHub.
2. **Settings** → **Collaborators** (или **Manage access**).
3. **Add people** → ввести **vardump@inbox.ru**.
4. Выбрать роль (например **Write** или **Maintain**).
5. Приглашение уйдёт на почту vardump@inbox.ru; пользователь должен принять его.

Повторить для **epn-killer-mvp**, **epn-killer-web**, **epn-killer-docs**.

---

## Шаг 5. Первый пуш (из основной папки C:\Users\aalab\epn-killer-project)

Для каждого репозитория — отдельная папка и свой remote.

**epn-killer-mvp:**

```powershell
cd C:\Users\aalab\epn-killer-project\epn-killer-mvp
git init
git add .
git commit -m "Initial: epn-killer-mvp (cmd + internal)"
git branch -M main
git remote add origin https://github.com/djalben/epn-killer-mvp.git
git push -u origin main
```

**epn-killer-web:**

```powershell
cd C:\Users\aalab\epn-killer-project\epn-killer-web
git init
git add .
git commit -m "Initial: epn-killer-web (frontend)"
git branch -M main
git remote add origin https://github.com/djalben/epn-killer-web.git
git push -u origin main
```

**epn-killer-docs:**

```powershell
cd C:\Users\aalab\epn-killer-project\epn-killer-docs
git init
git add .
git commit -m "Initial: epn-killer-docs (docs, TZ, start)"
git branch -M main
git remote add origin https://github.com/djalben/epn-killer-docs.git
git push -u origin main
```

(Замените `djalben` на ваш логин GitHub, если другой.)

---

## Итог

- **epn-killer-mvp** — бэкенд, структура: корень (go.mod, go.sum, Makefile, README.md, schema.sql, Dockerfile), `cmd/main.go`, `internal/`.
- **epn-killer-web** — фронтенд.
- **epn-killer-docs** — документация, ТЗ, старт, отчёты, мануалы.
- Во все три репозитория приглашён **vardump@inbox.ru** (приглашение отправляется из GitHub по шагу 4).

После переноса папок в **C:\Users\aalab\epn-killer-project** там будет полный проект: можно продолжать запускать Docker и пушить только из этой папки (по правилу основной рабочей папки).
