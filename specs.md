# zid-packages — Documentação completa

## Visão geral
O `zid-packages` é o pacote central de licenciamento e orquestração para os serviços ZID no pfSense.  
Ele fornece:
- **Licenciamento online** (webhook) com state local assinado.
- **IPC local seguro** via Unix socket para que outros pacotes validem licença.
- **Watchdog central** que inicia/para serviços conforme `enable` + licença.
- **GUI no pfSense** com abas: Packages, Services, Licensing e Logs.
- **Empacotamento/bundles** para distribuição via S3.

Pacotes controlados:
- `zid-proxy` (inclui `zid-appid` e `zid-threatd`)
- `zid-geolocation`
- `zid-logs`
- `zid-access`
- `zid-orchestrator` (serviço `zid-orchestration`)
- `zid-packages` (self, com licença N/A)

---

## Estrutura de diretórios (principais)

### Go (core)
- `cmd/zid-packages/` — CLI principal (status, license sync, watchdog, daemon).
- `internal/licensing/` — sincronização com webhook + state local.
- `internal/ipc/` — servidor IPC (Unix socket) com HMAC/HKDF.
- `internal/watchdog/` — watchdog central para serviços.
- `internal/status/` — status JSON consumido pela GUI.
- `internal/packages/` — inventário de pacotes, enable/running, start/stop, versões.
- `internal/state/` — state assinado de licenças.
- `internal/secure/` — derivação de chave, HMAC, HKDF.
- `internal/logx/` — logger simples.
- `internal/s3/` — fetch de versão remota.

### pfSense (GUI + scripts)
- `packaging/pfsense/files/usr/local/www/`
  - `zid-packages_packages.php`
  - `zid-packages_services.php`
  - `zid-packages_licensing.php`
  - `zid-packages_logs.php`
- `packaging/pfsense/files/usr/local/pkg/`
  - `zid-packages.inc`
  - `zid-packages.xml`
- `packaging/pfsense/files/etc/inc/priv/`
  - `zid-packages.priv.inc`
- `packaging/pfsense/scripts/`
  - `install.sh`, `update.sh`, `uninstall.sh`
  - `update-bootstrap.sh` — helper de update; oculta URL do bundle no output
  - `register-package.php`, `unregister-package.php`
  - `bundle-latest.sh`

### Build/output
- `build/` — binário Go local.
- `dist/` — estrutura do bundle pfSense.
- `zid-packages-latest.tar.gz` — bundle final.
- `zid-packages-latest.version` — versão do bundle.
- `sha256.txt` — hash dos bundles.

---

## CLI (cmd/zid-packages)

Comandos principais:
- `zid-packages status --json`
  - Gera JSON usado pela GUI.
- `zid-packages license sync`
  - Faz sync online do licenciamento.
- `zid-packages watchdog --once`
  - Executa um ciclo do watchdog.
- `zid-packages daemon`
  - Roda watchdog em loop.
- `zid-packages package install <pkg>`
- `zid-packages package update <pkg>`
- `zid-packages auto-update --once`
- `zid-packages -version`

Binário instalado no pfSense:
- `/usr/local/sbin/zid-packages`

Log principal:
- `/var/log/zid-packages.log`
  - Rotação automática: 1 MB, mantém 7 arquivos (`.1` a `.7`) e envia SIGHUP para reabrir o log.

---

## Licenciamento (internal/licensing)

### Webhook
URL e auth header estão hardcoded em `internal/licensing/licensing.go`.  
Resposta aceita:
- `{"pkg": true}` (bool)
- `{"pkg": "true"}` (string)
- `{"pkg": 1}` (número)

### State local
Gravado em:
- `/var/db/zid-packages/state.json` (via `internal/state`)
Assinado com HMAC para evitar alteração local.

### Modos
Retorna para clientes:
- `OK`
- `OFFLINE_GRACE`
- `EXPIRED`
- `NEVER_OK`

---

## IPC local (internal/ipc)

Socket:
- `/var/run/zid-packages.sock`

Segurança:
HMAC-SHA256 com chave derivada via HKDF:
- `masterSecret = "zid-packages-master-secret-2026-01"`
- `salt = uniqueid + ":" + shortHostname`
- `info = "zid-packages-hkdf"`

Request:
```json
{"op":"CHECK","package":"zid-proxy","ts":1700000000,"nonce":"hex","sig":"hmac"}
```

Response:
```json
{"ok":true,"licensed":true,"mode":"OK","valid_until":1700600000,"reason":"OK","ts":1700000001,"sig":"hmac"}
```

Flags de debug IPC:
- `ZID_PACKAGES_IPC_DEBUG=1` — loga request/response no `/var/log/zid-packages.log`
- `ZID_PACKAGES_IPC_LOG_KEYS=1` — loga key derivada (hex)

---

## Watchdog central (internal/watchdog)

### Função
Avalia:
```
shouldRun = enabled && licensed
```
E inicia/para:
```
zid-proxy, zid-appid, zid-threatd, zid-geolocation, zid-logs, zid-access, zid-orchestrator
```

### Cleanup de regras de firewall no stop
- Ao parar `zid-proxy`, o fluxo do watchdog executa o stop do serviço e também aciona o hook de pós-stop do pacote para remover a NAT auto-rule usada pelo aplicativo.
- Ao parar `zid-geolocation`, o fluxo do watchdog executa o stop do serviço e também chama a limpeza de floating rules/aliases (`zid_geolocation_clear_floating_rules`) para remover as regras de firewall aplicadas pelo pacote.

### Recriação de regras no start
- Ao iniciar `zid-geolocation`, o fluxo do watchdog executa o start do serviço e dispara `zid_geolocation_apply_async` para recriar floating rules/aliases quando ainda não estiverem presentes.

### Logs
Exemplo:
```
watchdog start: zid-proxy (enabled=true licensed=true mode=OK)
watchdog stop: zid-proxy (enabled=false licensed=true mode=OK)
```

### Leitura do enable (pfSense)
Para evitar flapping, usa o mesmo método do zid-geolocation:
- **PHP + config.inc** (fonte principal)
  - `require_once("/etc/inc/config.inc");`
  - lê `$config['installedpackages'][...]['config'][0]['enable']`
  - para `zid-access`, aceita também seção legada `installedpackages/zid-access` (e `zid_access`) quando existir
  - para `zid-access`, suporta config legada “quebrada” (lista escalar) onde `config[0]` é o enable (gera `<config>on</config>` repetido no XML)
  - para `zid-orchestrator`, lê `zid_orchestration_enable` de `/etc/rc.conf.local` e `/etc/rc.conf`

#### zid-access: formatos legados de config
O `zid-access` pode aparecer no `config.xml`/`$config` em formatos diferentes:
- **Canônico** (preferido):
  - `$config['installedpackages']['zidaccess']['config'][0]['enable'] = "on|off"`
- **Legado** (associativo direto ou `item`):
  - `$config['installedpackages']['zidaccess']['config']['enable']`
  - `$config['installedpackages']['zidaccess']['config']['item'][0]['enable']`
- **Quebrado (lista escalar)**: ocorre quando a config foi serializada como lista sem chaves.
  - Exemplo no `config.xml`:
    - `<zidaccess><config>on</config><config>wan</config>...</zidaccess>`
  - Exemplo no `config.inc` (PHP):
    - `$config['installedpackages']['zidaccess']['config'][0] === "on"`
  - Nesse caso o primeiro `config` é tratado como **enable**.

Recomendação: abrir a tela **Zid Access > Settings** e salvar uma vez para migrar para o formato canônico.

Fallbacks adicionais:
- parse do `/conf/config.xml` (estrito + loose)
- cache temporário (TTL curto) quando leitura falha
- para `zid-packages`, lê `/etc/rc.conf.local` ou `/etc/rc.conf` (aceita YES/NO)

#### Matcher do config.xml (estrito vs loose)
- **Estrito**: compara o *sufixo contíguo* do path, ignorando o root (ex.: `<pfsense>`). Ex.: path `installedpackages/zidaccess/config/enable` faz match com stack `pfsense/installedpackages/zidaccess/config/enable`.
- **Loose**: compara *subsequência* (ordem preservada, mas permite pular níveis). É mais permissivo e serve como fallback para cenários de XML fora do esperado.

Flags de debug de enable:
- `ZID_PACKAGES_ENABLE_DEBUG=1` — loga fonte/valor lido
Snapshot automático quando para por `enabled=false`.

---

## GUI pfSense

### Abas
**Packages**  
Lista pacotes, versão local/remota, update e status.
Coluna **Auto Update** mostra contador de dias desde a versão ficar disponível e indica quando está **Due** para atualização automática (23:59, via daemon), com ETA exibido.
Update manual na GUI roda em background e registra output no `/var/log/zid-packages-update-<pkg>.log`, com polling incremental (stream completo), fallback para submit normal e log exibido sob demanda na aba Packages (com botao fechar).
Para o `zid-packages`, o update é **seguro**: quando o daemon já está rodando, o installer não reinicia automaticamente e a GUI exibe badge **Restart pending** até o admin aplicar um restart em janela de manutenção.

**Services**  
Tabela com serviços, status e licença.  
Botões para Start/Stop/Restart do `zid-packages`:
- Se rodando: mostra Restart/Stop
- Se parado: mostra Start

**Licensing**  
Botão **Force sync** + estado de licenças.

**Logs**  
Mostra **últimas 50 linhas**, mais recentes no topo.  
Controles:
- Refresh
- Auto refresh (badge Auto/Manual)

---

## Empacotamento e bundles

Arquivos gerados:
- `zid-packages-latest.tar.gz`
- `zid-packages-latest.version`
- `sha256.txt`

Comando:
```
make bundle-latest
```

---

## Scripts pfSense

### install.sh
- Copia GUI/priv/incs
- Registra pacote no pfSense
- Garante `localpkg_enable=YES` e `local_startup` com `/usr/local/etc/rc.d`
- Instala `zid_packages.sh` para execucao via localpkg (localpkg executa apenas scripts *.sh)
- Mantem `zid_packages` como wrapper para uso manual/compatibilidade
- Update seguro: **se o daemon do `zid-packages` já estiver rodando**, não reinicia automaticamente (evita downtime do IPC/licenciamento) e cria marcador `restart-pending` para aplicar na próxima janela de manutenção
  - marker: `/var/db/zid-packages/restart-pending` (conteúdo = versão instalada)
  - o marker é limpo automaticamente quando o daemon sobe novamente (após restart/reboot)
  - para forçar restart no install/update: `ZID_PACKAGES_UPDATE_RESTART=1`
### update.sh
— Reinstala bundle (sem desinstalar)
### zid-packages-update (bootstrap)
- Script instalado em `/usr/local/sbin/zid-packages-update`
- Fluxo:
  - baixa o bundle (`zid-packages-latest.tar.gz`)
  - extrai em `/tmp`
  - executa o `scripts/update.sh` de dentro do bundle
  - `update.sh` chama `install.sh` (reinstala/atualiza arquivos)
- Opções úteis:
  - `-u <url>` para bundle custom
  - `-f` para forçar update (mantido por compatibilidade; hoje o fluxo reinstala o bundle)
  - `-k` para manter diretório temporário (debug)
### auto-update (daemon)
— Verificado diariamente às 23:59 (hora local) pelo daemon e executa update automático quando a versão estiver disponível há **0 dias** (temporário para testes).
- Estado em `/var/db/zid-packages/auto-update.json`:
  - por pacote: `version`, `first_seen`, `last_seen`
  - global: `last_run_day` (evita rodar mais de 1x por dia)
- Threshold (dias) é definido por `internal/autoupdate.MinDays` (exposto em `auto_update_threshold_days` no status).
- Ordem: atualiza o `zid-packages` **por último** (reduz risco de interromper a rodada).
- Quando atualizar o `zid-packages` via auto-update, o bundle é aplicado mas fica **Restart pending** (sem reiniciar automaticamente).
### uninstall.sh
— Remove GUI/priv/incs e unregister

---

## Detecção de versão local (pfSense)
Heurística usada pelo `zid-packages status --json`:
- `zid-packages`: tenta `config.xml` (registro do package) e depois `zid-packages -version`.
- `zid-proxy`: `zid-proxy -version`.
- `zid-geolocation`: `zid-geolocation -version`.
- `zid-logs`: prioriza `/usr/local/pkg/zid-logs.xml` e `/usr/local/share/pfSense-pkg-zid-logs/VERSION`; se o `config.xml` tiver versão **não numérica** (ex.: `"zid-logs version dev"`), ela é ignorada para comparação.
- `zid-access`: tenta `config.xml` (registro do package) e depois `/usr/local/share/pfSense-pkg-zid-access/VERSION`.
- `zid-orchestrator`: tenta `config.xml` (nome do package `zid-orchestration`), depois `/usr/local/share/pfSense-pkg-zid-orchestration/VERSION` e por fim `zid-orchestration --version`.

## Arquivos e paths relevantes (pfSense)

**Binários**
- `/usr/local/sbin/zid-packages`
- `/usr/local/sbin/zid-proxy`
- `/usr/local/sbin/zid-appid`
- `/usr/local/sbin/zid-threatd`
- `/usr/local/sbin/zid-geolocation`
- `/usr/local/sbin/zid-logs`
- `/usr/local/sbin/zidaccess`
- `/usr/local/sbin/zid-orchestration`

**Socket**
- `/var/run/zid-packages.sock`

**Logs**
- `/var/log/zid-packages.log`

**Estado**
- `/var/db/zid-packages/auto-update.json`
- `/var/db/zid-packages/restart-pending`

**Config pfSense**
- `/conf/config.xml`

**RC.d**
- `/usr/local/etc/rc.d/zid_packages`
- `/usr/local/etc/rc.d/zid_packages.sh`

---

## Padrões de atualização
- Mudou código: atualizar `CHANGELOG.md` + bump no `Makefile`.
- Rodar:
  - `gofmt -w .`
  - `go test ./...`
  - `go build ./...`
  - `make bundle-latest`

---

## Troubleshooting rápido
1) Ver versão:
```
/usr/local/sbin/zid-packages -version
```
2) Logs:
```
tail -n 50 /var/log/zid-packages.log
```
3) Licença:
```
/usr/local/sbin/zid-packages license sync
```
4) IPC:
Habilitar debug com:
```
env ZID_PACKAGES_IPC_DEBUG=1 /usr/local/sbin/zid-packages daemon
```

5) Enable (debug):
```
env ZID_PACKAGES_ENABLE_DEBUG=1 /usr/local/sbin/zid-packages watchdog --once
tail -n 200 /var/log/zid-packages.log | egrep "enable read:|watchdog enable snapshot"
```

6) Restart pendente (self-update seguro):
- Ver marcador:
```
ls -l /var/db/zid-packages/restart-pending
cat /var/db/zid-packages/restart-pending
```
- Aplicar em janela de manutenção:
```
/usr/local/etc/rc.d/zid_packages onerestart
```

---

## Notas de compatibilidade (2026-02-13)
- **zid-access Enabled falso na aba Packages/watchdog**:
  - Causa: config legada serializada como lista escalar (`<zidaccess><config>on</config>...</zidaccess>`) sem chave `<enable>`.
  - Solução: leitura do enable passou a aceitar `config[0]` (primeiro item) como enable; recomendado salvar Settings do `zid-access` para migrar ao formato canônico.
- **zid-logs versão local aparecendo como "dev"**:
  - Causa: o registro do pacote no `config.xml` podia conter string não numérica (ex.: `zid-logs version dev`).
  - Solução: versão local agora prioriza `VERSION`/`zid-logs.xml` e ignora valor não numérico do `config.xml`.
- **Self-update seguro do `zid-packages`**:
  - Objetivo: evitar downtime do IPC (`/var/run/zid-packages.sock`) e derrubar os outros serviços por licenciamento.
  - Implementação: installer não reinicia o daemon automaticamente quando já estiver rodando; cria `restart-pending` e a GUI exibe badge até o admin aplicar restart em manutenção.

## Diagrama de fluxo (ASCII)

```
           +---------------------------+
           |     pfSense GUI           |
           |  (Packages/Services/etc)  |
           +-------------+-------------+
                         |
                         | status --json
                         v
                 +-------+--------+
                 | zid-packages   |
                 |  (CLI/daemon)  |
                 +---+--------+---+
                     |        |
                     |        | license sync (webhook)
                     |        v
                     |   +----+-----+
                     |   | Webhook  |
                     |   +----+-----+
                     |        |
                     |   state.json (assinado)
                     v
         +-----------+-----------+
         |  Watchdog central     |
         | enabled && licensed?  |
         +-----+-----------+-----+
               |           |
               | start/stop|
               v           v
        zid-proxy     zid-geolocation
        zid-appid     zid-logs
        zid-access
        zid-orchestrator

   IPC (Unix socket /var/run/zid-packages.sock)
   ---------------------------------------------
   zid-proxy/zid-geolocation/zid-logs/zid-access/zid-orchestrator -> CHECK -> zid-packages
   (HMAC/HKDF)                             (valida e responde)
```

---

## Próximos passos
- Validar em pfSense real (prod) com carga e ciclos longos.
