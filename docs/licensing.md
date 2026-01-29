# Licenciamento ZID via `zid-packages` (pfSense)

## Objetivo
Centralizar a validacao de licenca dos pacotes:
- `zid-proxy` (inclui `zid-appid`)
- `zid-geolocation`
- `zid-logs`

O pacote `zid-packages` e o "servidor local" de licencas no pfSense.
Cada pacote ZID **deve consultar** o `zid-packages` para saber se pode executar.
Se o `zid-packages` estiver ausente (desinstalado) ou negar a licenca, o pacote **deve parar imediatamente**.

## Regras de execucao (obrigatorias)
1) Ao iniciar o servico do pacote, antes de processar qualquer logica, o pacote deve consultar a licenca.
2) Se a licenca for **invalida**:
   - o pacote deve encerrar o processo com exit code != 0
   - e, se aplicavel, solicitar stop no rc.d (ou simplesmente finalizar e deixar o watchdog central garantir o stop)
3) Se nao for possivel consultar o `zid-packages` (socket nao existe / erro de comunicacao):
   - tratar como **SEM LICENCA** e encerrar/parar
4) Mesmo com licenca valida no start, o pacote deve revalidar periodicamente (ex: a cada 60s ou 5min) consultando o `zid-packages`.
   - Se a licenca passar a ser invalida, parar imediatamente.

## Canal de consulta (IPC local)
O `zid-packages` expoe um Unix Domain Socket:
- Path: `/var/run/zid-packages.sock`
- Permissoes: acesso restrito (root), socket local

### Modelo de seguranca
A mensagem e autenticada com HMAC (assinatura) para evitar spoofing local.
A chave e derivada no runtime usando informacoes do proprio sistema (ex: `/var/db/uniqueid` e `hostname -s`) e um segredo embutido no binario.

### Request (exemplo)
JSON:
{
  "op": "CHECK",
  "package": "zid-proxy",
  "ts": 1700000000,
  "nonce": "random-hex",
  "sig": "hmac-hex"
}

Campos:
- `op`: sempre `CHECK`
- `package`: `zid-proxy` | `zid-geolocation` | `zid-logs`
- `ts`: timestamp unix (segundos)
- `nonce`: string aleatoria para evitar replay
- `sig`: HMAC-SHA256 do payload (sem o campo sig)

### Response (exemplo)
{
  "ok": true,
  "licensed": true,
  "mode": "OK",
  "valid_until": 1700600000,
  "reason": "valid",
  "ts": 1700000001,
  "sig": "hmac-hex"
}

Campos:
- `licensed`: true/false
- `mode`:
  - `OK` (online recente)
  - `OFFLINE_GRACE` (sem internet mas dentro da janela de 7 dias)
  - `EXPIRED` (passou de 7 dias sem sucesso)
  - `NEVER_OK` (nunca validou online e nao conseguiu validar agora)
- `valid_until`: timestamp unix do limite (last_success + 7 dias) quando aplicavel
- `reason`: texto curto para log/diagnostico
- `sig`: assinatura da resposta

## Comportamento esperado no pacote (pseudocodigo)
- Na inicializacao:
  - if !LicenseCheck("zid-proxy") => exit(2)
- Em loop/intervalo:
  - if !LicenseCheck("zid-proxy") => stop imediatamente (exit / shutdown)

## Observacoes
- O watchdog central do `zid-packages` tambem aplica enforcement (stop/start), mas o pacote deve se auto-proteger.
- Se o cliente desinstalar o `zid-packages`, o socket desaparece e todos os pacotes devem parar.
- Nao persistir chaves de licenca em arquivos. Toda validacao e por consulta ao `zid-packages`.
