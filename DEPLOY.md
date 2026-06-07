# Cognita.uz — Serverга yuklash (VPS + FastPanel + Docker)

## Arxitektura
- **frontend** (SvelteKit/Node) → ichki port `127.0.0.1:3000`
- **backend** (Go/Fiber)        → ichki port `127.0.0.1:8080`  (yo'llar: `/api/...`, `/ws/...`)
- **postgres** (PostgreSQL 16)  → ichki port `127.0.0.1:5432`
- FastPanel'ning **nginx**'i `cognita.uz` (80/443, SSL) ni qabul qilib, ichki portlarga uzatadi.

Hamma narsa **bitta domen** (`cognita.uz`) orqali ishlaydi.

---

## 1. Docker o'rnatish (bir marta)
```bash
curl -fsSL https://get.docker.com | sh
docker compose version
```

## 2. Fayllarni joylash
```bash
mkdir -p /root/cognita && unzip cognita_deploy.zip -d /root/cognita
cd /root/cognita
```

## 3. .env yaratish
```bash
cp .env.example .env
nano .env
```
```ini
POSTGRES_PASSWORD=KUCHLI_PAROL
JWT_SECRET=<openssl rand -base64 64>
FRONTEND_URL=https://cognita.uz
VITE_API_URL=https://cognita.uz      # MUHIM: port yo'q, /api yo'q, localhost emas!
```

## 4. Ishga tushirish
```bash
docker compose up -d --build
docker compose ps          # 3 ta servis "running" bo'lsin
curl -I http://127.0.0.1:3000   # frontend javob bersa OK
curl -I http://127.0.0.1:8080/api/health 2>/dev/null || true
```

## 5. FastPanel'da domen + SSL + proxy
1. FastPanel → **Sites** → yangi sayt `cognita.uz`, turini **Reverse proxy** qilib,
   manzil: `http://127.0.0.1:3000`.
2. FastPanel → sayt → **SSL** → Let's Encrypt (bepul) yoqing.
3. Sayt sozlamalari → **Qo'shimcha nginx direktivalari**ga `nginx-cognita.conf` dagi
   `location /api/` va `location /ws/` bloklarini joylang. Saqlang.
4. Tekshirish: `nginx -t && systemctl reload nginx`

Tayyor → https://cognita.uz

---

## Yangilanish (kod o'zgarganda)
```bash
cd /root/cognita        # yangi fayllarni qo'ying
docker compose up -d --build
```

## Foydali buyruqlar
```bash
docker compose logs -f backend     # backend loglari
docker compose logs -f frontend
docker compose down                # to'xtatish
docker compose restart backend
```
