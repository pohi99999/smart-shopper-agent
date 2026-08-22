# Smart Shopper Agent

Többügynökös AI backend (Go) + platformfüggetlen mobil/web kliens (React Native / Expo)
a legolcsóbb bevásárlóút és -lista megtervezéséhez. A felhasználó szabad szöveges
bevásárlólistáját és GPS koordinátáit egy Google Gemini-alapú Parser ügynök
strukturálja, a Pricer lekéri a boltláncok árait, az Optimizer pedig a valós
útvonaltervező (OSRM) adatai alapján kiszámolja a legköltséghatékonyabb útvonalat.

A fejlesztés teljes története (fázisonként, magyarul) a [GEMINI.md](GEMINI.md)
fájlban található.

## Architektúra

```
cmd/app/            HTTP szerver belépési pont
internal/agents/    Parser, Pricer, Optimizer AI ügynökök
internal/mcp/       PriceScraper és RoutePlanner külső adatforrás-eszközök
internal/api/       HTTP handlerek, middleware (rate limit, CORS, admin auth)
internal/models/    Megosztott adatstruktúrák
internal/data/      prices.json - boltláncok ár- és koordináta-adatbázisa
internal/automation/ n8n automatizációs workflow blueprint
mobile/             React Native (Expo) mobil + web kliens
scripts/            n8n deploy és webhook szimulációs segédszkriptek
```

## Gyors indítás - Backend

```bash
cp .env.example .env   # töltsd ki: ADMIN_TOKEN, GEMINI_API_KEY
go run ./cmd/app
```

A szerver a `:8080` porton indul. Végpontok:

| Végpont | Metódus | Leírás |
|---|---|---|
| `/api/v1/optimize` | POST | Bevásárlólista + GPS -> optimalizált útvonalterv |
| `/api/v1/admin/prices` | GET/POST | Áradatok lekérdezése / frissítése (`X-Admin-Token` védett) |
| `/swagger/` | GET | Interaktív API dokumentáció |

Tesztek: `go test ./...`

## Gyors indítás - Mobil/Web kliens

```bash
cd mobile
cp .env.example .env   # opcionális: EXPO_PUBLIC_API_URL, Sentry, RevenueCat kulcsok
npm install --legacy-peer-deps
npm start               # Expo dev szerver (i / a / w a platformválasztáshoz)
```

Tesztek: `npm test` · Típusellenőrzés: `npx tsc --noEmit`

> [!warning] react-native-maps nincs web-támogatással
> A `MapSection.tsx` csak natív (iOS/Android) platformon jelenít meg
> interaktív térképet. Web build esetén Metro automatikusan a
> `MapSection.web.tsx` fallbackot tölti be (a boltokat egyszerű listaként
> jeleníti meg koordinátákkal).

## Docker

```bash
docker compose up --build -d
```

## CI

- [`.github/workflows/backend-ci.yml`](.github/workflows/backend-ci.yml) - Go build, teszt, Docker build ellenőrzés minden `main`-t érintő push/PR-nál.
- [`.github/workflows/mobile-ci.yml`](.github/workflows/mobile-ci.yml) - `npm ci`, típusellenőrzés (`tsc --noEmit`), Jest tesztek minden `mobile/**`-t érintő push/PR-nál.

## Monetizáció - jelenlegi állapot

- **Pro előfizetés (RevenueCat):** teljes körűen implementálva (Paywall UI,
  `SubscriptionContext`, App Store/Play Store IAP integráció). Éles
  bevetéshez lásd a `mobile/.env.example` RevenueCat kulcsokra vonatkozó
  megjegyzését.
- **Ingyenes, hirdetés-támogatott tier:** a [GEMINI.md](GEMINI.md) 21. fázisában
  rögzített üzleti tervben szerepel, de **még nincs implementálva** - nincs
  `AdBanner` komponens vagy AdMob/hirdetési SDK integráció a kódbázisban.
  Ehhez egy AdMob (vagy más ad hálózat) fiók létrehozása és a platform-natív
  konfiguráció szükséges, mielőtt implementálható.
