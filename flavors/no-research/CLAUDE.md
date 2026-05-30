## No-Research-Flavor (Opt-out)

Die Recherche-Pflicht aus dem base layer ist für dieses Repo **deaktiviert**.

- Alle Rollen dürfen ohne WebSearch/context7/fetch arbeiten
- Deaktiviert zusätzlich den **harten PreToolUse-Block**: Edit/Write/MultiEdit
  werden nicht mehr an eine vorherige Recherche gebunden. Der `require-research`-Hook
  erlaubt bei aktivem `no-research` immer (greift über das
  `HasFlavor("no-research")`-Gate).
- Verwende diesen Flavor nur für Repos mit eingeschränktem Netzwerk-Zugang
  (Air-gapped, strikte Compliance-Anforderungen) oder rein interne Logik
  ohne externe Abhängigkeiten
