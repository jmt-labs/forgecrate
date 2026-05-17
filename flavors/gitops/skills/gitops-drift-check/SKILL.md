# GitOps Status

Drift-Check und Deployment-Status für ein GitOps-Repo. Kombiniert ArgoCD-App-Topologie, clusterweite Regeln und Live-Cluster-Zustand.

## Voraussetzung

Mindestens eines der folgenden Tools muss installiert und konfiguriert sein: `kubectl`, `argocd`, `helm`, `flux`.

## Ablauf

### 1. Repo-Kontext sicherstellen

Falls noch nicht geladen (Sessionbeginn oder nach Kompaktierung): Kontext aus dem Repo lesen.

**ArgoCD Applications:**
```bash
grep -rl "kind: Application\|kind: AppProject" . --include="*.yaml" --include="*.yml" 2>/dev/null
```
Für jede gefundene Datei: `name`, `destination.server`, `destination.namespace`, `source.repoURL` und `source.path` extrahieren.
Ergebnis: eine interne App-Karte `{app-name → cluster/namespace/source-path}`.

Falls keine Applications gefunden: zuerst `.env` prüfen:
```bash
grep "GITOPS_REPO" .env 2>/dev/null
```
Falls gesetzt: in das angegebene Verzeichnis wechseln und dort weitersuchen.
Falls nicht gesetzt: fragen "Ist dies das GitOps-Repo? Falls nicht, bitte Pfad oder URL angeben." — Antwort in `.env` als `GITOPS_REPO=<wert>` speichern, dann fortfahren. Skill-Ausführung pausieren bis Antwort vorliegt.

**Clusterweite Regeln — zwei Quellen:**

*YAML-Policies (Kyverno, Gatekeeper):*
```bash
grep -rl "kind: ClusterPolicy\|kind: Policy\|kind: ConstraintTemplate\|kind: Constraint" . --include="*.yaml" --include="*.yml" 2>/dev/null
```
Für jede gefundene Datei: `metadata.name` und `spec.rules[].validate`/`spec.rules[].mutate` lesen.

*RULES.md (menschenlesbare Teamregeln):*
```bash
find . -name "RULES.md" 2>/dev/null
```
Falls gefunden: vollständig lesen. Gilt gleichwertig zu YAML-Policies — auch Regeln die nicht als CRD durchgesetzt werden, sind verbindlich.

Beide Quellen zusammenführen und als aktiven Kontext halten.

### 2. ArgoCD Sync-Status

```bash
argocd app list -o json 2>/dev/null | jq -r '.[] | "\(.metadata.name): sync=\(.status.sync.status) health=\(.status.health.status)"'
```

Falls nicht eingeloggt:
```bash
argocd login <server-url>
```

Alternativ über kubectl:
```bash
kubectl get applications -A -o custom-columns="NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status,NAMESPACE:.spec.destination.namespace" 2>/dev/null
```

Markiere Apps mit `sync=OutOfSync` als Drift-Kandidaten.

### 3. Drift-Check

Für jede App aus der App-Karte: lokalen Manifest-Pfad mit Live-Cluster vergleichen.

**kubectl:**
```bash
kubectl diff -R -f <source-path> 2>&1
```
Exit-Code 0 = kein Drift; Exit-Code 1 = Diff-Output (kein Fehler).

**Helm-basierte Apps:**
```bash
helm diff upgrade <release-name> <chart-path> -f <values-file> -n <namespace> 2>/dev/null
```
Falls `helm-diff`-Plugin fehlt: `helm plugin install https://github.com/databus23/helm-diff` vorschlagen, Skill-Ausführung für diesen Release überspringen.

**Flux:**
```bash
flux get all -A 2>/dev/null
```

### 4. Policy-Validierung

Server-side Dry-Run für geänderte Manifeste — triggert Admission-Controller (Kyverno, Gatekeeper):

```bash
kubectl apply --dry-run=server -R -f <source-path> 2>&1
```

Policy-Verletzungen im Output erkennen (Kyverno: `admission webhook … denied`; Gatekeeper: `denied the request`).

Falls Verletzungen: Ausgabe mit Regelname und betroffener Ressource. Niemals `--validate=false` vorschlagen.

**Hinweis:** Clusterweite Regeln aus Schritt 1 sind der Referenzrahmen — Verletzungen im Dry-Run bestätigen, dass die Repo-Regeln aktiv durchgesetzt werden.

### 5. Deployment-Status

```bash
# Pods die nicht Running/Succeeded sind
kubectl get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded 2>/dev/null

# Deployments mit nicht vollständig verfügbaren Replicas
kubectl get deployments -A -o json 2>/dev/null | jq -r '.items[] | select(.status.availableReplicas < .spec.replicas) | "\(.metadata.namespace)/\(.metadata.name): \(.status.availableReplicas // 0)/\(.spec.replicas)"'
```

### 6. Image-Tags prüfen

```bash
grep -rn ":latest" . --include="*.yaml" --include="*.yml" 2>/dev/null
```

Falls Treffer: Datei, Zeile und betroffenes Image ausgeben.

### 7. Ergebnis ausgeben

```
## GitOps Status

### Apps (aus ArgoCD Application-Manifesten)
- <app-name>: sync=<status> health=<status> → <namespace>@<cluster>
- ...

### Drift
[kein Drift | N Ressourcen weichen ab — Details:]
<kubectl diff Output>

### Policy-Validierung
[alle Manifeste valide | N Verletzungen:]
<Regelname: Ressource>

### Deployment-Status
[alle Deployments Ready | N Probleme:]
<namespace/name: X/Y Replicas>

### Image-Tags
[keine latest-Tags | N latest-Tags:]
<datei:zeile image:latest>

### Nächste Schritte
<konkrete Empfehlungen>
```

## Hinweise

- App-Karte und Regeln aus Schritt 1 einmal pro Session laden und wiederverwenden — nicht bei jedem Skill-Aufruf neu scannen.
- Separates GitOps-Repo: Wenn App-Repo ≠ Infra-Repo, immer explizit fragen bevor Manifeste bearbeitet werden. Niemals annehmen, dass Deployment-Manifeste im App-Repo liegen.
- `kubectl diff` benötigt für manche Ressourcen `--server-side`.
- ArgoCD: `OutOfSync` = Drift (Git ≠ Cluster); `Degraded` = Laufzeitproblem (unabhängig von Sync).
