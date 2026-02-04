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
zid-proxy, zid-appid, zid-threatd, zid-geolocation, zid-logs
```

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

Fallbacks adicionais:
- parse do `/conf/config.xml` (estrito + loose)
- cache temporário (TTL curto) quando leitura falha
- para `zid-packages`, lê `/etc/rc.conf.local` ou `/etc/rc.conf` (aceita YES/NO)

Flags de debug de enable:
- `ZID_PACKAGES_ENABLE_DEBUG=1` — loga fonte/valor lido
Snapshot automático quando para por `enabled=false`.

---

## GUI pfSense

### Abas
**Packages**  
Lista pacotes, versão local/remota, update e status.

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
- Reinicia o daemon do `zid-packages` ao final (onerestart)
### update.sh
— Reinstala bundle (sem desinstalar)
### uninstall.sh
— Remove GUI/priv/incs e unregister

---

## Arquivos e paths relevantes (pfSense)

**Binários**
- `/usr/local/sbin/zid-packages`
- `/usr/local/sbin/zid-proxy`
- `/usr/local/sbin/zid-appid`
- `/usr/local/sbin/zid-threatd`
- `/usr/local/sbin/zid-geolocation`
- `/usr/local/sbin/zid-logs`

**Socket**
- `/var/run/zid-packages.sock`

**Logs**
- `/var/log/zid-packages.log`

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

---

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

   IPC (Unix socket /var/run/zid-packages.sock)
   ---------------------------------------------
   zid-proxy/zid-geolocation/zid-logs -> CHECK -> zid-packages
   (HMAC/HKDF)                             (valida e responde)
```

---

## Próximos passos
- Validar em pfSense real (prod) com carga e ciclos longos.
