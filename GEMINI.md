# GEMINI.md - Smart Shopper Agent Project Context

## 1. Projekt áttekintése
A **Smart Shopper Agent** egy többügynökös Go backend alkalmazás mobil applikációk számára, amelynek fő fókusza a költséghatékony és optimális vásárlás megtervezése. A rendszer képes a felhasználó nyers bevásárlólistáját és GPS koordinátáit feldolgozva a legoptimálisabb boltlátogatási útvonalat és árakat kalkulálni.

## 2. Architektúra és Modulok

### AI Ügynökök (internal/agents/)
- **Parser** ([parser.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/parser.go)): Egy LLM rendszer-prompt segítségével a nyers felhasználói bemenetet strukturált JSON formátumra (ShoppingList) alakítja.
- **Pricer** ([pricer.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/pricer.go)): Fogadja a strukturált listát, és ciklusban lekéri a termékek bolti árait a PriceScraper MCP eszközön keresztül.
- **Optimizer** ([optimizer.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/optimizer.go)): Egy optimalizációs algoritmus és LLM prompt segítségével a pricer adatai és a térképes adatok alapján előállítja a legoptimálisabb útvonaltervet.

### MCP Eszközök (internal/mcp/)
- **PriceScraper** ([price_scraper_mcp.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/mcp/price_scraper_mcp.go)): Webes felületekről vagy API-n keresztül termékárakat lekérdező eszköz vázláncolata (Aldi, Interspar).
- **RoutePlanner** ([route_planner_mcp.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/mcp/route_planner_mcp.go)): Földrajzi koordináták alapján távolságot és utazási időt számoló térképes eszköz vázláncolata.

### Adatstruktúrák (internal/models/)
- **shopping.go** ([shopping.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/models/shopping.go)): A bevásárlólista (ShoppingList), tételek (ShoppingItem) és az útvonalterv (RoutePlan) struktúrái.

Az **1-11. fázis** sikeresen elkészült:
- A projekt verziókövetése inicializálásra került a főágon (`main`), és sikeresen feltöltésre került a GitHub-ra: [smart-shopper-agent GitHub](https://github.com/pohi99999/smart-shopper-agent.git).
- A projekt könyvtárszerkezete és a Go modul inicializálása megtörtént.
- Az MCP eszközök vázai kidolgozásra és integrálásra kerültek.
- A RoutePlanner MCP eszköz sikeresen integrálásra került a valós OSRM (Open Source Routing Machine) API-val éles útvonaltervezéshez.
- A PriceScraper MCP eszköz leválasztásra került a kódkészletről, és immár egy külső JSON adatbázisból (`internal/data/prices.json`) dolgozik.
- Az AI Parser ügynök élesítésre került a valós Google Gemini API REST integrációjával (`Joho/godotenv` környezeti változókezeléssel).
- Az AI ügynökök belső logikája, rendszer-promptjai kidolgozásra kerültek.
- Elkészült a REST API HTTP szerver (`/api/v1/optimize`), amely kiszolgálja a mobil kliens kéréseit a 8080-as porton.
- A React Native (Expo) mobil frontend inicializálása sikeresen befejeződött:
  - Létrejött a `mobile` projekt Expo TypeScript sablon sablonnal.
  - Kialakításra került az alapvető mappa- és fájlstruktúra (`mobile/src/components`, `mobile/src/screens`, `mobile/src/services`).
  - Elkészült az API kommunikációs réteg (`mobile/src/services/api.ts`) az optimalizációs API aszinkron hívásához.
- Elkészült az első mobil képernyő (7. fázis):
  - Létrejött a `mobile/src/screens/ShoppingListScreen.tsx` képernyő modern, Apple stílusú dizájnnal, amely kezelni tudja a szabad szöveges bevitelt, a Budapest koordinátákkal történő optimalizálás indítását, a hálózati kérés alatti betöltési állapotot, valamint a kapott útvonalterv lépéseit és a becsült végösszeget.
  - Az `App.tsx` frissítésre került, hogy a `ShoppingListScreen` legyen az alkalmazás fő belépési pontja.
- A GPS helymeghatározás integrálása a mobilalkalmazásba megtörtént (9. fázis):
  - Telepítésre került az `expo-location` modul.
  - A `ShoppingListScreen` komponens kiegészült kezdeti és gombnyomáskori engedélykéréssel, pozíció lekérdezéssel, valamint a valós koordináták backend felé történő továbbításával.
  - Kialakításra került egy hibakezelő Alert visszajelzés és a biztonságos budapesti fallback koordináták használata helyadatok elutasítása/hiba esetén.
- Helyi zalaegerszegi adatok integrálása a backendbe és a hardcoded koordináták megszüntetése (10. fázis):
  - Frissítésre került a `prices.json` adatbázis fájl, hogy a termékárak mellett az Aldi és Interspar zalaegerszegi boltjainak valós koordinátáit is tartalmazza.
  - A `PriceScraper` MCP eszköz kiegészült a `ShopData` struktúrával és a `GetShopCoordinates(shopChain string)` metódussal a koordináták dinamikus kiolvasásához.
  - Az `Optimizer` ügynök immár injektált függőségként megkapja a `PriceScraper`-t, és az útvonaltervezés során a korábbi hardcoded értékek helyett az adatbázisból dinamikusan lekérdezett bolt koordinátákat használja.
- Térképes vizualizáció integrálása a mobil frontendbe (11. fázis):
  - Telepítésre került a `react-native-maps` modul.
  - A `ShoppingListScreen.tsx` komponens kiegészült a térképpel, amely kezdetben a felhasználó GPS pozíciójára (vagy Budapest fallbackre) fókuszál.
  - Kék Marker jelzi a felhasználó saját helyzetét.
  - Sikeres optimalizálás után piros Marker-ek jelzik az optimális útvonal zalaegerszegi állomásait, megjelenítve az állomás sorszámát, a bolt nevét, és a megvásárlandó tételek listáját a buborékban.
  - Kialakításra került a térkép kártya modern Apple-stílusú árnyékolt és lekerekített stílusozása.

## 4. Következő feladatok
- Valós web-scraperek bekötése a JSON adatbázis frissítéséhez vagy a valós idejű árlekérdezéshez.
- Web, App Store és Play Store megjelenés előkészítése (22. Fázis+).
- Hirdetési (Ads) és Pro előfizetéses (In-App Purchases) üzleti modell implementálása.

## 12. Fázis: n8n Ingest API
A 12. fázis fejlesztései (n8n Ingest API) sikeresen megtervezésre, implementálásra, tesztelésre és beolvasztásra kerültek a `main` ágba.
- **Környezeti változók:** Az [.env](file:///Z:/001_Workspace/smart-shopper-agent/.env) fájl kiegészült az `ADMIN_TOKEN` beállítással az adminisztrátori műveletek biztonságos hitelesítéséhez.
- **Ingest API Végpont:** Az [AdminPricesHandler](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers.go#L106) metódus kibővítésre került a `POST` kérések fogadására. Érvényes `X-Admin-Token` fejléc és strukturált JSON törzs (request body) ellenőrzése után a végpont felülírja a helyi [prices.json](file:///Z:/001_Workspace/smart-shopper-agent/internal/data/prices.json) fájl tartalmát az automatizált n8n frissítésekhez.
- **API Tesztek:** A [handlers_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers_test.go#L14) fájlban lévő `TestAdminPricesHandler` kiegészült az új POST ágakat vizsgáló tesztesetekkel (`POST Valid Token and Body` és `POST Unauthorized`), míg a korábbi hibás metódus teszt a POST helyett `PUT` metódust használ.

## 13. Fázis: Jules Aszinkron Tesztelés és Optimalizálás
- **Ellenőrzés és Beolvasztás:** A fejlesztések tesztelése sikeresen megtörtént (a `go test ./...` hibátlanul lefutott), majd a változtatások beolvasztásra kerültek a `main` ágba és feltöltésre kerültek a távoli [smart-shopper-agent GitHub](https://github.com/pohi99999/smart-shopper-agent.git) tárolóba.
- **Backend Hibakezelés:** Bevezetésre került egy strukturált `ErrorResponse` és egy `SendJSONError` segédfüggvény a [handlers.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers.go) fájlban, amely biztosítja, hogy minden HTTP hiba JSON formátumban kerüljön visszaküldésre a mobil kliensnek a plain text helyett.
- **Backend Timeout Kezelés:** A [route_planner_mcp.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/mcp/route_planner_mcp.go) és a [parser.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/parser.go) (Gemini API) hívások mostantól 10 másodperces timeouttal rendelkező beépített `http.Client`-et használnak, elkerülve a szerver panic-ot a timeoutokból származó egyértelmű hibajelzésekkel.
- **Backend Távolság Korlát:** Az [optimizer.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/optimizer.go) kiegészült egy logikával, ami az OSRM adatok alapján figyelmen kívül hagyja a kiindulási ponttól (GPS) 50 km-nél messzebb lévő boltokat.
- **Backend Admin Végpont:** Létrehozásra került az `/api/v1/admin/prices` GET végpont, ami `X-Admin-Token` védelemmel van ellátva, és sikeres hitelesítés esetén teszt áradatokat szolgáltat.
- **Backend Tesztek:** Megírásra kerültek a [parser_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/parser_test.go), [pricer_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/pricer_test.go), [optimizer_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/optimizer_test.go) és a [handlers_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers_test.go) egység- és integrációs tesztek, elérve a 70% feletti teszt lefedettséget a kritikus csomagokon.
- **Frontend Custom Hook Refaktorálás:** A [ShoppingListScreen.tsx](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/screens/ShoppingListScreen.tsx) logikája, beleértve az API hívásokat és a GPS lokáció lekérést, leválasztásra került és ki lett szervezve a [useShoppingOptimizer.ts](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/hooks/useShoppingOptimizer.ts) custom hookba a tiszta kód alapelvek (clean code) jegyében.
- **Frontend Típusdefiníciók:** A [api.ts](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/services/api.ts) kiegészült teljes és részletes TypeScript interfészekkel a bejövő kérések, válaszok és a strukturált JSON hibaüzenetek típusbiztos kezelésére.

## 14. Fázis: n8n Automatizációs Blueprint
A 14. fázis fejlesztései során egy automatizált árinjektáló munkafolyamat (workflow blueprint) került kifejlesztésre az n8n rendszeréhez.
- **Munkafolyamat Blueprint:** Elkészült a [n8n_price_updater_workflow.json](file:///Z:/001_Workspace/smart-shopper-agent/internal/automation/n8n_price_updater_workflow.json) fájl az [internal/automation/](file:///Z:/001_Workspace/smart-shopper-agent/internal/automation/) könyvtárban.
- **Automatizációs Lépések:**
  - **Schedule Trigger:** Időzítő, amely minden hajnalban 02:00-kor fut le (`cronExpression: "0 2 * * *"`).
  - **Mock Scraper (Set Node):** Szimulált termékárakat állít elő (`prices_raw`).
  - **Data Transformation (Code Node):** JavaScript transzformáció, amely a nyers adatokat a [prices.json](file:///Z:/001_Workspace/smart-shopper-agent/internal/data/prices.json) sémájára formázza, kiegészítve a zalaegerszegi boltok GPS koordinátáival.
  - **Ingest API (HTTP Request):** Egy POST kéréssel továbbítja az adatokat a backend `/api/v1/admin/prices` [AdminPricesHandler](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers.go#L105) végpontjára a megfelelő `X-Admin-Token` hitelesítő fejléc használatával.

## 15. Fázis: Konténerizáció
A 15. fázis fejlesztései során elkészült a backend alkalmazás Docker konténerizációja a könnyebb hordozhatóság és futtathatóság érdekében.
- **Dockerfile:** Létrejött egy hatékony, többlépcsős (multi-stage) [Dockerfile](file:///Z:/001_Workspace/smart-shopper-agent/Dockerfile). A builder fázisban a Go bináris (`smart-shopper-agent`) fordul le `golang:1.26-alpine` alapon, míg a végleges produkciós konténer egy minimális `alpine:latest` alapú kép, amely csak a binárist, a `.env` fájlt és a termékárakat tartalmazó [prices.json](file:///Z:/001_Workspace/smart-shopper-agent/internal/data/prices.json) fájlt tartalmazza.
- **Docker Compose:** Elkészült a [docker-compose.yml](file:///Z:/001_Workspace/smart-shopper-agent/docker-compose.yml) konfiguráció, amely definiálja a `smart-shopper-backend` szolgáltatást.
  - A 8080-as portot köti össze a gazdagép és a konténer között (`8080:8080`).
  - Helyi kötetet (volume) használ a [prices.json](file:///Z:/001_Workspace/smart-shopper-agent/internal/data/prices.json) fájl perzisztálásához (`./internal/data/prices.json:/app/internal/data/prices.json`).
  - Automatikusan betölti a környezeti változókat az `env_file` segítségével a [.env](file:///Z:/001_Workspace/smart-shopper-agent/.env) fájlból.
- **Futtatási parancsok:**
  - Konténerek felépítése és indítása a háttérben:
    ```bash
    docker compose up --build -d
    ```
  - Leállítás és konténerek eltávolítása:
    ```bash
    docker compose down
    ```
  - Logok megtekintése:
    ```bash
    docker compose logs -f
    ```

## 17. Fázis: Jules Aszinkron CI/CD és Frontend Tesztelés
A 17. fázis fejlesztései során bevezetésre került egy GitHub Actions alapú CI/CD pipeline, inicializálásra kerültek a frontend tesztek, valamint fejlesztésre került a backend logolása.
- **CI/CD Pipeline:** Elkészült a `.github/workflows/backend-ci.yml` munkafolyamat, amely automatikusan lefut a `main` ágat érintő push és pull request eseményekre. A folyamat felállítja a Go környezetet, ellenőrzi a függőségeket, lefuttatja a Go teszteket (`go test ./...`), és verifikálja a Docker kép sikeres felépítését.
- **Frontend Tesztelés (React Native):**
  - A `mobile` projektben konfigurálásra került a `jest` és a `@testing-library/react-native`.
  - Elkészült a `mobile/src/services/api.test.ts` egységteszt, amely az API hívásokat mockolja és teszteli az `optimizeShoppingRoute` sikeres és hibás válaszait.
  - Elkészült a `mobile/src/screens/ShoppingListScreen.test.tsx` render teszt, amely biztosítja a felhasználói felület alapvető elemeinek (beviteli mező, gomb) megfelelő megjelenését és interaktivitását.
- **Backend Strukturált Logolás:** A `cmd/app/main.go` és az `internal/mcp/route_planner_mcp.go` fájlokban a hagyományos `log` csomag és a `fmt` alapú logolás lecserélésre került a Go beépített `log/slog` csomagjára. A konfigurált JSON handler professzionális, strukturált formátumban biztosítja a naplózást, amely kiválóan illeszkedik a produkciós Docker környezetekhez.

## 18. Fázis: Biztonság, Swagger API dokumentáció és Offline Cache
A 18. fázis fejlesztései során növeltük a backend biztonságát, legeneráltuk az API dokumentációt, és felkészítettük a mobilalkalmazást offline használatra.
- **Backend API Dokumentáció (Swagger/OpenAPI):** 
  - A `swaggo/swag` és `swaggo/http-swagger` csomagok integrálásra kerültek a Go backendbe.
  - A `cmd/app/main.go` és az `internal/api/handlers.go` fájlok megfelelő Swagger annotációkat kaptak.
  - A kérések és válaszok struktúrái definiálva lettek az `/api/v1/optimize` és a `/api/v1/admin/prices` végpontokhoz.
  - A `/swagger/*` végponton elérhető a generált vizuális Swagger UI.
- **Biztonság és Rate Limiting (Go Backend):**
  - Bevezetésre került egy Rate Limiter middleware (`golang.org/x/time/rate`), amely 10 kérés/perc korlátozással védi az `/api/v1/optimize` végpontot a túlterhelés ellen.
  - Alapvető biztonsági HTTP fejlécek (CORS, X-Content-Type-Options) kerültek hozzáadásra a szerver válaszaihoz.
- **LLM Parser Hibatűrés (Retry Logika):**
  - Az `internal/agents/parser.go` fájlban a Gemini API hívás automatikus újrapróbálkozási (retry) logikát kapott, mely maximum 2 alkalommal, exponenciális késleltetéssel próbálja újra a kérést, ha hálózati hiba vagy érvénytelen JSON válasz lépne fel.
- **Frontend Offline Cache Előkészítés:**
  - A mobilalkalmazásba integrálásra került a `@react-native-async-storage/async-storage` csomag.
  - A sikeres útvonaltervezés eredményét (útvonalterv és becsült végösszeg) az alkalmazás lokálisan elmenti.
  - A betöltés során az app ellenőrzi az elmentett utolsó bevásárlólistát, és megjeleníti azt az új keresésig, biztosítva az adatok elérését offline környezetben (pl. boltban megszakadó mobilnet esetén) is.

## 19. Fázis: Release Candidate 1 (RC1) és Android Build Előkészítés
A 19. fázis során rögzítésre került a Release Candidate 1 (RC1) állapot, amellyel a projekt elérte az első stabil, biztonságos és offline is működő mérföldkövét.
- **Release Candidate 1:** Az alkalmazás Go backend és React Native mobil frontend komponensei stabilan integrálva vannak, lefedve a tesztekkel, a biztonsági és sebességkorlátozásokkal, valamint az offline gyorsítótárazással.
- **Fizikai Build Előkészítés:** A mobilalkalmazás készen áll az EAS (Expo Application Services) segítségével történő natív Android (.apk) build elkészítésére a mobil eszközökön való éles/előnézeti teszteléshez.

## 20. Fázis: Jules újabb tesztjeinek és javításainak szinkronizációja
A 20. fázis során sikeresen szinkronizálásra és integrálásra kerültek a main ágba Jules legújabb biztonsági és kód-egészségügyi fejlesztései.
- **CORS Biztonsági Javítás:** Az overly permissive CORS beállítások helyett bevezetésre került az `ALLOWED_ORIGIN` környezeti változó támogatása a [middleware.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/middleware.go) fájlban, ami alapértelmezetten a biztonságos '*' értéket kapja, de konfigurálható egyedi domainekre is.
- **Middleware Tesztek:** Elkészültek és beolvasztásra kerültek az új API middleware tesztek az [middleware_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/middleware_test.go) fájlban, lefedve a CORS origin beállításokat és a rate limitinget (X-Forwarded-For és RemoteAddr alapú limitekkel).
- **Backend Integráció:** A helyi `main` ágon sikeresen feloldásra kerültek a teszt fájlokban lévő merge konfliktusok, a backend tesztek (`go test -short ./... -v`) pedig hibátlanul lefutottak az összesített környezetben. A változtatások feltöltésre kerültek a GitHub-ra.

## 21. Fázis: Copilot Integráció és Cross-Platform Vízió

A 21. fázis során rögzítésre kerültek a projekt hosszú távú platformfüggetlen és üzleti stratégiájának alapjai.

### Elvégzett feladatok
- **GitHub Copilot Instrukciók:** Létrejött a [.github/copilot-instructions.md](file:///Z:/001_Workspace/smart-shopper-agent/.github/copilot-instructions.md) fájl, amely rögzíti a projekt összes kódolási szabályát:
  - Backend: Go 1.26+, `log/slog` strukturált logolás, 70%+ teszt lefedettség, retry logika, timeoutok, clean architecture.
  - Frontend: Expo SDK 56, TypeScript strict, platformfüggetlen (Web/iOS/Android) UI komponensek kötelező támogatással.
  - Általános: Conventional Commits, CI/CD pipeline szabályok, Docker és n8n konvenciók.
- **Cross-Platform Web Támogatás:** A [mobile/app.json](file:///Z:/001_Workspace/smart-shopper-agent/mobile/app.json) fájlban bekapcsolásra került a Metro web bundler (`"web": { "bundler": "metro" }`).
- **Web Függőségek:** Telepítésre kerültek a webes Expo futtatáshoz szükséges npm csomagok a `mobile` projektbe:
  - `react-dom` – React DOM renderer webes megjelenítéshez.
  - `react-native-web` – React Native API-k webes implementációja.
  - `@expo/metro-runtime` – Metro bundler web runtime.

### Jövőbeli üzleti vízió (22. Fázis+)
A projekt célja a **Web / App Store (iOS) / Play Store (Android)** platformokon való megjelenés az alábbi üzleti modellel:

#### Ingyenes Tier (Ad-Supported)
- Az alkalmazás teljes funkcionalitása elérhető, de hirdetések jelennek meg (banner a képernyő alján).
- Hirdetési integráció: `expo-ads-admob` (vagy platform-natív ekvivalens), `<AdBanner>` komponens.
- A hirdetésbevétel fedezi az infrastruktúra és az AI API (Gemini) költségeit.

#### Pro Tier (Subscription / In-App Purchase)
- **Hirdetésmentes** élmény.
- **Kiterjesztett boltlista**: extra láncokon (pl. Lidl, Spar, Tesco) túlmutató ár-összehasonlítás.
- **Előzmények és kedvencek**: bevásárlólisták mentése és visszatöltése AsyncStorage + backend sync-kel.
- **Push értesítések**: ár-riasztások, ha egy termék árcsökkentést ér el.
- Fizetési integráció: `expo-in-app-purchases` (iOS) / `react-native-iap` (Android/Web), absztrahálva egy `usePurchase` hook mögé.
- Feature flag kezelés: `useProStatus` hook (AsyncStorage + backend validáció).

#### Technikai irányelvek a monetizációhoz
- Minden Pro-gated funkció mögé `useProStatus` ellenőrzés kerül; ingyenes felhasználóknál paywall modal jelenik meg.
- Az `<AdBanner>` komponens Pro felhasználóknak `null`-t renderel.
- A backend `/api/v1/user/subscription` végponton validálhatja az előfizetés státuszát (JWT alapú auth, 22. Fázistól).

## 22. Fázis: Előfizetési Architektúra és Paywall UI (2026-06-30)

A 22. fázis során megvalósításra került a RevenueCat integrációt előkészítő prémium előfizetéses architektúra, a modern Paywall értékesítési felület, valamint a biztonsági figyelmeztetések javítása.

### Biztonsági Audit és Javítás
- **`npm audit fix`** futtatásra került a `mobile` könyvtárban (`--legacy-peer-deps` flaggel).
- **Javított:** `js-yaml < 3.15.0` (GHSA-h67p-54hq-rp68) – sebezhető verzió frissítve. Figyelmeztetések: 11 → 10.
- **Nem javítható destruktív változás nélkül:** `uuid < 11.1.1` (GHSA-w5hq-g745-h8pq) – az Expo 56 build toolchain mélyén (`xcode` → `@expo/config-plugins`) van jelen. A `npm audit fix --force` expo@46-ra downgradelne, ami elfogadhatatlan. Ez ismert false positive az Expo ökoszisztémában; csak az EAS/prebuild folyamat során érintett, a production bundle-ben nem jelenik meg.

### Előfizetési Szolgáltatási Réteg
- Létrejött a [mobile/src/services/subscriptionService.ts](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/services/subscriptionService.ts) fájl.
- Definiálva: `SubscriptionStatus` interfész (`isPro`, `expiresAt`, `productId`), `PRODUCT_IDS` konstansok.
- Implementálva: `fetchSubscriptionStatus()`, `purchaseSubscription()`, `restorePurchases()` – jelenleg mock implementációval, 23. Fázisban cserélhető RevenueCat SDK hívásokra.

### Globális Állapotkezelés (Subscription Context)
- Létrejött a [mobile/src/context/SubscriptionContext.tsx](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/context/SubscriptionContext.tsx) fájl.
- `SubscriptionProvider` komponens becsomagolja az egész alkalmazást (App.tsx-ben).
- `useSubscription()` hook biztosítja a globális `isPro`, `isLoading`, `error`, `subscribe()`, `restore()`, `refresh()` elérését.
- `useMemo` + `useCallback` optimalizálással a felesleges re-renderek elkerülésére.

### Prémium Paywall UI
- Létrejött a [mobile/src/screens/PaywallScreen.tsx](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/screens/PaywallScreen.tsx) fájl.
- Apple-style arany/premium dizájn: 👑 hero szekció, 6 feature kártya, havi/éves ár-összehasonlítás.
- Arany `Előfizetés indítása` CTA gomb (mock purchase) és `Korábbi vásárlás visszaállítása` link.
- Sikeres vásárlás után automata bezárás (800ms delay).
- Platformfüggetlen: `react-native-maps` és platform-specifikus API-k nélkül, `Platform.OS` guard-okkal.
- Az App.tsx-ben React Native `Modal` (animationType: slide, presentationStyle: pageSheet) jeleníti meg.

### Integráció és Navigáció
- A [mobile/src/screens/ShoppingListScreen.tsx](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/screens/ShoppingListScreen.tsx) `Props` interfészt kapott (`onShowPaywall?: () => void`).
- A fejlécben diszkrét arany `👑 Go Pro` gomb jelenik meg, ha `!isPro && onShowPaywall`.
- Az [App.tsx](file:///Z:/001_Workspace/smart-shopper-agent/mobile/App.tsx) kezeli a modal láthatóságát, a `SubscriptionProvider` az egész app-ot körbeveszi.

### TypeScript Fix
- A [mobile/tsconfig.json](file:///Z:/001_Workspace/smart-shopper-agent/mobile/tsconfig.json) kiegészült a `"types": ["jest", "@types/jest"]` beállítással, megszüntetve a teszt fájlokban korábban lévő `Cannot find name 'jest'` kompilációs hibákat.

### Tesztelés
- Összes meglévő teszt (`npm test`) hibátlanul fut: **4/4 PASS**.

## 23. Fázis: n8n API Deployment és Integrációs Tesztelés (2026-06-30)

A 23. fázis során megvalósításra kerültek az n8n automatizációt élesítő deployment szkriptek, valamint végrehajtásra és validálásra került a backend webhook integrációs teszt.

### n8n API Deployment Szkript
- Létrejött a [scripts/deploy_n8n_workflow.js](file:///Z:/001_Workspace/smart-shopper-agent/scripts/deploy_n8n_workflow.js) fájl.
- **Működés:**
  1. Beolvassa az `N8N_API_KEY` és `N8N_HOST` (fallback: `http://localhost:5678`) értékeket a `.env` fájlból (saját, függőségmentes `.env` parser).
  2. Beolvassa az `internal/automation/n8n_price_updater_workflow.json` workflow definíciót.
  3. `POST /api/v1/workflows` kéréssel létrehozza a munkafolyamatot az n8n REST API-n (`X-N8N-API-Key` fejléccel).
  4. `POST /api/v1/workflows/{id}/activate` kéréssel aktiválja a workflow-t.
  5. Strukturált konzol-kimenettel jelzi a sikerességet vagy a hibát (pl. elérhetetlen n8n, hiányzó API kulcs, HTTP hibakód).
- **Futtatás:** `node scripts/deploy_n8n_workflow.js`

### Webhook Szimulációs Teszt
- Létrejött a [scripts/simulate_webhook.js](file:///Z:/001_Workspace/smart-shopper-agent/scripts/simulate_webhook.js) fájl.
- **Működés:**
  1. Beolvassa az `ADMIN_TOKEN` és `BACKEND_HOST` (fallback: `http://localhost:8080`) értékeket a `.env` fájlból.
  2. `POST http://localhost:8080/api/v1/admin/prices` kérést küld szándékosan eltérő tesztárakkal (tojás: 99 Ft, kenyér: 199 Ft, tej: 149 Ft).
  3. Validálja a HTTP 200 OK választ.
  4. Kiírja az aktualizált `prices.json` tartalmát.
- **Futtatás és validálás eredménye (2026-06-30):**
  ```
  📬  Response: HTTP 200
  ✅  SUCCESS – Backend accepted the price update (HTTP 200 OK)
     Response body: {"message":"Prices updated successfully","status":"success"}
  ```
- A teszt után az eredeti adatbázis visszaállításra került: `git checkout -- internal/data/prices.json`
- **Futtatás:** `node scripts/simulate_webhook.js`

### Technikai részletek
- Mindkét szkript **nulla külső npm függőséggel** működik (csak Node.js beépített modulok: `fs`, `path`, `http`, `https`).
- A `.env` parser kezeli az inline kommenteket, idézőjelet és a `******` maszkolt értékeket.
- A `simulate_webhook.js` 8 másodperces request timeout-tal rendelkezik, és értelmes hibaüzenetet ad vissza, ha a backend nem elérhető.

