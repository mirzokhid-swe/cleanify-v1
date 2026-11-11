# CLAUDE.md

Cleanify backend (Go, gin + sqlx, PostgreSQL 14). Bu repo `/api/v1` ni beradi; Laravel repo (`cleanify-telegram-bot-service`) faqat Telegram bot va eski hisobotlar.

## Qoidalar

- Telegram botga doir koddan tashqari **hamma yangi va qayta ishlanadigan kod shu repoda (Go)** yoziladi. Laravel'ga faqat mavjud hisobotga zarur shart qo'shiladi.
- Hujjatlar, speclar va rejalar `../cleanify-docs/` da (`implementation/<modul>.md` + `implementation/plans/`). Kod yozishdan oldin modul spec'ini o'qing; topilgan bug'lar `implementation/bugs-2026-11-06.md`.
- Push, deploy (`make prod-deploy`) va prod migratsiya — egasining ishi. So'ramasdan qilinmaydi.
- Har o'zgarish avval qizil test bilan (`go test`), keyin kod. Tekshirish: `go test ./... && go vet ./...`.
- `.env`, tokenlar, `tokens.json` commit qilinmaydi va chatga yozilmaydi.

## Buyruqlar

```bash
go run ./cmd/main.go        # 8000-port, .env joriy papkadan (godotenv)
go test ./... && go vet ./...
psql -d bwerp -f storage/migrations/postgres/NNN_x.up.sql   # migratsiya runner yo'q, qo'lda
```

Lokal stack (Go + Laravel + front) uchun `../Makefile` (`make`, `make link`, `make stop`) va `../cleanify-docs/infra/local.md`.

## Kod tuzilishi

- `api/api.go` — yo'llar. Yo'lga xos middleware **har yo'lga alohida** beriladi (`orderItemTypeRoutes`, `companyRoutes` yordamchilari namuna); `group.Use(...)` zanjiri to'planib ketadi, ishlatilmaydi.
- `api/middleware/` — `AuthMiddleware` (JWT, Laravel bilan bitta `API_SECRET`), `UserActiveMiddleware` (kontekstga `user`), `RequirePermission(slug)` (so'rovdagi `company_id` bo'yicha), `UserCan(c, companyID, slug)` (kompaniya buyurtmadan olinganda), `SuperAdminMiddleware`.
- `api/handlers/` — javob `h.handleResponse(c, http.<Status>, data)` (`api/http/status.go`): huquq yo'q → `http.PermissionDenied` (`{"status":"PERMISSION_DENIED"}`), topilmadi → `http.NOT_FOUND`, kompaniya cheklangan → 423. Front shu shartnomaga tayanadi.
- `models/` — so'rov/javob strukturalari (gin `binding`), sof mantiq (`MeasureForUnit`, `UnitIsCountable`...). `storage/repo/` interfeyslar, `storage/postgres/` SQL.
- Testlar: handler'lar `api/*_test.go` (httptest + fake storage, `api/order_item_type_test.go` namuna); middleware `api/middleware/*_test.go`; SQL `storage/postgres/*_db_test.go` (`.env` dan DB, yo'q bo'lsa skip).

## Tuzoqlar

- lib/pq: `CASE WHEN ... THEN $3` ichidagi parametrga tip kerak (`$3::bigint`).
- `helper.ReplaceQueryParams` har `:key` ni bir marta almashtiradi.
- `auth/me` **POST**. Buyurtma holati o'zgarganda `WEBHOOK_URL` ga (Laravel) so'rov ketadi — lokalda Laravel `.env` da dummy bot tokenlar turishi shart, aks holda haqiqiy Telegram guruhlariga xabar ketadi.
- Go migratsiyalari `storage/migrations/postgres/`, `.up.sql`/`.down.sql`, idempotent yoziladi.
