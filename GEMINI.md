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

## 13. Fázis: Jules Aszinkron Tesztelés és Optimalizálás
- **Ellenőrzés és Beolvasztás:** A fejlesztések tesztelése sikeresen megtörtént (a `go test ./...` hibátlanul lefutott), majd a változtatások beolvasztásra kerültek a `main` ágba és feltöltésre kerültek a távoli [smart-shopper-agent GitHub](https://github.com/pohi99999/smart-shopper-agent.git) tárolóba.
- **Backend Hibakezelés:** Bevezetésre került egy strukturált `ErrorResponse` és egy `SendJSONError` segédfüggvény a [handlers.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers.go) fájlban, amely biztosítja, hogy minden HTTP hiba JSON formátumban kerüljön visszaküldésre a mobil kliensnek a plain text helyett.
- **Backend Timeout Kezelés:** A [route_planner_mcp.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/mcp/route_planner_mcp.go) és a [parser.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/parser.go) (Gemini API) hívások mostantól 10 másodperces timeouttal rendelkező beépített `http.Client`-et használnak, elkerülve a szerver panic-ot a timeoutokból származó egyértelmű hibajelzésekkel.
- **Backend Távolság Korlát:** Az [optimizer.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/optimizer.go) kiegészült egy logikával, ami az OSRM adatok alapján figyelmen kívül hagyja a kiindulási ponttól (GPS) 50 km-nél messzebb lévő boltokat.
- **Backend Admin Végpont:** Létrehozásra került az `/api/v1/admin/prices` GET végpont, ami `X-Admin-Token` védelemmel van ellátva, és sikeres hitelesítés esetén teszt áradatokat szolgáltat.
- **Backend Tesztek:** Megírásra kerültek a [parser_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/parser_test.go), [pricer_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/pricer_test.go), [optimizer_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/agents/optimizer_test.go) és a [handlers_test.go](file:///Z:/001_Workspace/smart-shopper-agent/internal/api/handlers_test.go) egység- és integrációs tesztek, elérve a 70% feletti teszt lefedettséget a kritikus csomagokon.
- **Frontend Custom Hook Refaktorálás:** A [ShoppingListScreen.tsx](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/screens/ShoppingListScreen.tsx) logikája, beleértve az API hívásokat és a GPS lokáció lekérést, leválasztásra került és ki lett szervezve a [useShoppingOptimizer.ts](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/hooks/useShoppingOptimizer.ts) custom hookba a tiszta kód alapelvek (clean code) jegyében.
- **Frontend Típusdefiníciók:** A [api.ts](file:///Z:/001_Workspace/smart-shopper-agent/mobile/src/services/api.ts) kiegészült teljes és részletes TypeScript interfészekkel a bejövő kérések, válaszok és a strukturált JSON hibaüzenetek típusbiztos kezelésére.
