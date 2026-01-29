# Plan

Vamos executar o projeto em fases, priorizando o padrão do `zid-proxy` para GUI e adotando daemon contínuo com cron obrigatório. Cada fase termina com validação básica (build/test) e, quando aplicável, smoke test no pfSense.

## Scope
- In: binário único Go, IPC/licenciamento, watchdog central, GUI pfSense padrão `zid-proxy`, scripts/install/update/uninstall, bundles e `licensing.md`.
- Out: mudanças no código dos pacotes `zid-proxy`, `zid-geolocation`, `zid-logs` (apenas documentação de integração).

## Action items
[x] Fase 1 — Descoberta: abrir repos `zid-proxy`, `zid-geolocation`, `zid-logs` via MCP GitHub; mapear enable/status/version, rc.d, update command, cron antigo e padrões GUI/packaging.
[x] Fase 2 — Núcleo Go: criar `cmd/zid-packages` + `internal/*` com CLI (`status`, `watchdog`, `license sync`, `package install/update`, `daemon`), logs e state assinado.
[x] Fase 3 — Licenciamento online: implementar cliente webhook, cache 7 dias, modos de estado, persistência e integração no `status --json`.
[x] Fase 4 — IPC seguro: socket UNIX com HMAC/HKDF, validação de timestamp/nonce, resposta assinada; gerar `licensing.md` PT-BR.
[x] Fase 5 — Watchdog central: daemon contínuo + cron `watchdog --once`, enforcement start/stop por pacote, remoção segura de watchdogs antigos.
[x] Fase 6 — GUI pfSense: páginas PHP (Packages/Services/Licensing/Logs) seguindo padrão `zid-proxy`, lendo `status --json` e disparando ações.
[x] Fase 7 — Scripts e empacotamento: `install.sh`, `update.sh`, `uninstall.sh`, `bundle-latest.sh`, pkg xml/inc/priv, rc.d, cron; validar reinício do webgui.
[x] Fase 8 — Validação por fase: `gofmt -w .`, `go test ./...`, `go build ./...`, e smoke test no pfSense.
[x] Fase 9 — Entrega: atualizar `CHANGELOG.md`, bump no `Makefile`, rodar `make bundle-latest`, garantir `.version` e `sha256.txt`.

## Open questions
- Nenhuma no momento.
