## GitOps-Flavor

### Repo-Kontext beim Start laden

Beim Start in einem GitOps-Repo MUSS Claude zuerst den Kontext laden:

1. **ArgoCD Applications lesen** — suche nach `kind: Application` und `kind: AppProject` in allen YAML-Dateien:
   ```bash
   grep -rl "kind: Application\|kind: AppProject" . --include="*.yaml" --include="*.yml"
   ```
   Daraus ableiten: welche Apps existieren, welche Cluster/Namespaces sie targeten, welche Repos sie referenzieren.

2. **Clusterweite Regeln lesen** — zwei Quellen prüfen, beide gelten gleichwertig als harte Constraints:

   *Maschinenlesbar (YAML):*
   ```bash
   grep -rl "kind: ClusterPolicy\|kind: Policy\|kind: ConstraintTemplate\|kind: Constraint" . --include="*.yaml" --include="*.yml"
   ```

   *Menschenlesbar (Markdown):*
   ```bash
   find . -name "RULES.md" 2>/dev/null
   ```
   Falls `RULES.md` gefunden: vollständig lesen. Die dort dokumentierten Regeln gelten genauso wie YAML-Policies — kein Manifest darf gegen sie verstoßen. Claude hält beide Quellen im Kontext.

3. **Separates GitOps-Repo** — wenn das aktuelle Repo kein GitOps-Repo ist (keine `Application`-Manifeste gefunden):

   Zuerst prüfen ob bereits konfiguriert:
   ```bash
   grep "GITOPS_REPO" .env 2>/dev/null
   ```

   Falls `GITOPS_REPO` in `.env` gesetzt: diesen Wert verwenden, nicht erneut fragen.

   Falls nicht gesetzt: einmalig fragen "Gibt es ein separates GitOps-Repo? Bitte Pfad oder URL angeben." — dann den Wert in `.env` schreiben:
   ```
   GITOPS_REPO=<antwort>
   ```
   `.env` in `.gitignore` prüfen — falls nicht eingetragen, Hinweis ausgeben dass lokale Pfade nicht committed werden sollten.

   Bearbeite keine Deployment-Manifeste ohne diesen Kontext.

### Verhaltensregeln

- **Jedes Deployment läuft über ArgoCD.** Direkte schreibende Cluster-Kommandos (`kubectl apply`, `kubectl delete`, `kubectl patch`, `helm upgrade`, `helm install`, `helm uninstall`) sind grundsätzlich verboten — auch wenn sie technisch funktionieren würden. Der einzige valide Deployment-Weg ist ein Commit + Merge in das GitOps-Repo, ArgoCD synchronisiert danach automatisch.

- **Ausnahmen nur mit expliziter Bestätigung.** Falls ein schreibendes Kommando ausnahmsweise notwendig ist (z.B. Notfall-Rollback, Bootstrap-Situation), MUSS Claude vor der Ausführung stoppen und fragen:
  > "Das ist ein direktes Cluster-Kommando außerhalb von ArgoCD: `<kommando>`. Soll ich es ausführen?"
  Ohne explizites Okay des Nutzers wird das Kommando nicht ausgeführt. Lesende Kommandos (`kubectl get`, `kubectl describe`, `kubectl logs`, `helm list`, `argocd app get`) sind ohne Bestätigung erlaubt.

- Secrets niemals im Repository speichern (SOPS, External Secrets Operator, Vault)
- Vor jedem Dry-Run oder Plan: `kubectl diff`, `helm diff upgrade` — niemals direkt apply ohne vorherige Prüfung
- Keine `latest`-Image-Tags — immer versionierte Tags oder Digests
- Manifeste die gegen clusterweite Regeln (Kyverno ClusterPolicy, OPA Gatekeeper) oder `RULES.md` verstoßen werden nicht vorgeschlagen — auch nicht mit dem Hinweis "das Policy-Check ignorieren"
- Drift zwischen Git-Zustand und laufendem Cluster regelmäßig prüfen: `/claude-setup-gitops-status`
