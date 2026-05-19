# Profile und Flavors

## Profile (eines wählbar)

| Profil | Fokus | Zusätzliche Extensions |
|---|---|---|
| `backend` | API, Datenbank, Integrationstests | keine (Base-Layer reicht: memory, fetch, github, context7, context-mode) |
| `frontend` | Komponenten, State, Barrierefreiheit | Plugins: frontend-design, typescript-lsp, playwright; MCP: playwright |
| `fullstack` | Kombination beider, shared Types, E2E | MCP: playwright |

## Flavors (mehrere kombinierbar)

| Flavor | Fokus |
|---|---|
| `tdd` | Test-First, kein Produktionscode ohne Test |
| `strict-review` | Pflicht-Review vor jedem Commit |
| `minimal` | Nur Basis-Enforcement |
| `gitops` | Infrastruktur via Git, Drift-Checks, Policy-Enforcement |
