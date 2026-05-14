# Claude Setup Advisor

Analysiert ein Repo und empfiehlt das passende `claude-setup`-Profil und die passenden Flavors.

## Verfügbare Profile

| Profil | Wann |
|--------|------|
| `backend` | Go, Rust, Python, Java, C# — kein oder minimales Frontend |
| `frontend` | React, Vue, Svelte, Angular — UI-lastiges Projekt |
| `fullstack` | Frontend und Backend in einem Repo |

## Verfügbare Flavors

| Flavor | Wann |
|--------|------|
| `strict-review` | Team >1 Person, PRs, formale Review-Prozesse |
| `tdd` | Test-first-Kultur, Testabdeckung >70%, CI-Pflicht |
| `minimal` | Prototyp, Solo-Projekt, wenig Overhead gewünscht |

## Ablauf

1. **Sprache und Framework erkennen** — prüfe `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`. Schau ob Frontend-Abhängigkeiten (react, vue, …) vorhanden sind.

2. **Profil ableiten**:
   - Nur Backend-Sprache → `backend`
   - Nur Frontend-Framework → `frontend`
   - Beides → `fullstack`

3. **Test-Kultur erkennen** — prüfe ob Tests vorhanden sind, ob CI konfiguriert ist (`.github/workflows/`), ob Coverage-Reports erzeugt werden → Flavor `tdd` sinnvoll?

4. **Review-Anforderungen abfragen** — stelle eine Frage: "Arbeitest du alleine oder im Team mit PR-Reviews?" → Flavor `strict-review` sinnvoll?

5. **Empfehlung ausgeben**:

```
Empfehlung basierend auf diesem Repo:

Profil:  backend
Reasons: Go-Modul erkannt, kein Frontend-Framework

Flavors: strict-review, tdd
Reasons: .github/workflows/ vorhanden (CI), >50 Testdateien gefunden

Befehl:
  claude-setup run --profile backend --flavor strict-review,tdd

Alternativ ohne TDD-Disziplin:
  claude-setup run --profile backend --flavor strict-review
```

6. **Frage ob ausführen** — "Soll ich `claude-setup run` mit dieser Konfiguration jetzt ausführen?"
