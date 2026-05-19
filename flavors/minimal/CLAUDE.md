## Minimal-Flavor

Dieser Flavor fügt keine zusätzliche Konfiguration hinzu. Er dient als explizite Entscheidung, keine weiteren Pflicht-Skills oder Workflows zu aktivieren.

**Wann nutzen:** `--flavors minimal` ist sinnvoll wenn du forgecrate ohne TDD- oder Strict-Review-Overhead initialisieren willst — z. B. für Prototypen, Solo-Projekte oder Projekte in der frühen Explorationsphase.

**Was er NICHT tut:** Er deaktiviert keine Base-Konfiguration. Das Compose-System ist additiv — Base-Layer-Skills und -Workflows bleiben vollständig aktiv. `minimal` kombiniert sich problemlos mit anderen Flavors.

- Keine zusätzlichen Pflicht-Skills außer dem Base-Layer
- Kein TDD-Zwang — Tests wo sinnvoll
- Hooks aktiv aber keine zusätzliche Blockierung
