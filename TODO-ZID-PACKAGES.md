# PROMPT PARA CODEX — IMPLEMENTAR "zid-packages" (pfSense / FreeBSD) — PT-BR

Você vai desenvolver um NOVO pacote para pfSense chamado **zid-packages**, que será o “gerenciador” dos pacotes:
- zid-proxy (inclui serviço adicional zid-appid)
- zid-geolocation
- zid-logs

REQUISITOS CRÍTICOS (NÃO NEGOCIÁVEIS)
1) Toda a lógica do pacote e do licenciamento deve ficar compilada dentro de UM ÚNICO BINÁRIO GO chamado **zid-packages**.
   - NÃO criar um binário de licença por pacote.
   - NÃO colocar lógica/segredos de licenciamento em arquivos de configuração.
   - Arquivos permitidos fora do binário: apenas LOGS e “STATE/CACHE” mínimo necessário (ex: último sucesso de consulta, timestamps), mas SEM revelar lógica/segredos. Se criar state, proteger contra adulteração (ex: HMAC).

2) NÃO “usar funções” do zid-proxy dentro do zid-packages.
   - Você deve APENAS usar os projetos existentes como referência de COMO fazem integração pfSense (scripts, layout das páginas PHP, config.xml, cron/watchdog, update remoto etc).
   - Copie o estilo e os padrões, mas implemente seu próprio código no zid-packages.

3) O WATCHDOG passa a ser CENTRALIZADO no zid-packages.
   - Os pacotes zid-proxy / zid-geolocation / zid-logs NÃO terão mais watchdog próprio.
   - Se existirem watchdogs antigos (cron/scripts) desses pacotes, o zid-packages deve desabilitar/remover de forma segura durante sua instalação (somente os itens reconhecidos), para evitar conflito.

REFERÊNCIAS (OBRIGATÓRIO CONSULTAR E SEGUIR O PADRÃO)
- Projeto zid-proxy (GitHub guilhermemulato/zid-proxy): estrutura de package pfSense, páginas WebGUI, scripts install/update/uninstall, bootstrap updater e padrão de cron watchdog via pfSense. (Use como modelo de “como fazer”.)
- Specs já existentes:
  - zid-geolocation: paths, rc.d, processos, GUI e padrão watchdog/scheduler. 
  - zid-logs: padrão de GUI, status, update e integração com inputs.d etc.
(Use como referência de estilo/estrutura e compatibilidade.) 

IMPORTANTE: O zid-proxy tem update via /usr/local/sbin/zid-proxy-update e compara versão remota via arquivo .version (ver scripts do projeto). Use isso como padrão de “não reinventar update”, chamando o update já existente do pacote. 
(Para geolocation e logs, localizar o update remoto pronto e chamar, sem reimplementar.) 

===========================================================
OBJETIVO 1) INSTALAÇÃO / UPDATE DOS PACOTES (GUI pfSense)
===========================================================

A WebGUI do zid-packages deve exibir os 3 pacotes e, para cada um:
- Se NÃO instalado: mostrar botão **Install**
- Se instalado: mostrar **versão instalada** + status do serviço + se existe versão nova no S3 -> botão **Update**

1.2 Instalação (Install)
- Ao clicar em Install, o zid-packages deve:
  a) baixar o bundle do S3 (base: https://s3.soulsolucoes.com.br/portal/<pacote> ou a URL exata que o pacote já usa hoje)
  b) extrair em diretório temporário
  c) executar o script de instalação do próprio pacote (install.sh) como o pacote já faz hoje
  d) reiniciar webgui se necessário (padrão igual aos pacotes existentes)
- IMPORTANTE: não inventar nomes de arquivos de bundle/versão. Você deve abrir os repositórios/scripts de update de cada pacote e pegar:
  - URL default do bundle
  - URL do arquivo de versão (latest.version)
  - ponto de entrada “update” (ex: /usr/local/sbin/<pkg>-update)

1.3 Update (Update)
- NÃO reinventar o update.
- O botão Update deve chamar o “update remoto pronto” do pacote:
  - Ex: zid-proxy usa /usr/local/sbin/zid-proxy-update (bootstrap updater)
  - Para zid-geolocation e zid-logs: localizar o equivalente e usar.
- A execução do update deve retornar saída para a GUI (em tempo real se possível, ou ao final), sem travar o webConfigurator.

1.4 Detecção de versão instalada e versão remota
- Versão instalada: usar o mesmo método que a GUI do pacote usa hoje (ler versão do binário, ou arquivo VERSION, ou status).
- Versão remota: baixar e ler o arquivo *.version do S3 (como os updaters existentes fazem).
- Nunca hardcodar suposições de nomes; derive dos scripts existentes.

===========================================================
OBJETIVO 2) WATCHDOG CENTRAL (CRON + MONITORAMENTO)
===========================================================

2.1 Cron criado durante instalação do zid-packages
- Durante a instalação do zid-packages, criar um item de cron no pfSense para rodar o watchdog do zid-packages.
- Frequência sugerida: a cada 1 minuto.
- O cron deve chamar o binário:
  - /usr/local/sbin/zid-packages watchdog --once
  (ou comando equivalente definido por você)
- Se preferir também rodar como serviço rc.d com daemon contínuo, OK, mas ainda assim CRIAR o cron (requisito).

2.2 Monitoramento automático (sem alertar usuário)
O watchdog deve:
- Descobrir todos os serviços relacionados aos 3 pacotes (incluindo zid-appid do zid-proxy)
- Verificar se cada serviço deveria estar rodando:
  - Condição A: o pacote está habilitado (enable) na configuração do próprio pacote no pfSense (config.xml / config.json conforme o pacote já usa)
  - Condição B: a licença do pacote está ATIVA (ver OBJETIVO 3)
- Se A e B forem verdade:
  - Se processo/serviço NÃO estiver rodando -> iniciar automaticamente (sem popup pro usuário)
- Se B for falso (sem licença):
  - Garantir que o serviço esteja PARADO (stop) mesmo que estivesse rodando
  - E o watchdog NÃO deve iniciar esse serviço

2.3 WebGUI mostra status de serviços
- Listar e mostrar status (Running/Stopped) de:
  - zid-proxy
  - zid-appid
  - zid-geolocation
  - zid-logs
- Mostrar também se cada um está habilitado (enable) e se está licenciado.

2.4 Remover watchdogs antigos dos pacotes (centralização)
- Durante instalação do zid-packages:
  - localizar e remover/desabilitar cron jobs antigos dos pacotes ZID (watchdog antigo), com critérios seguros:
    - remover SOMENTE itens reconhecidos (ex: comando contém “zid-proxy-watchdog”, “zid-geolocation-watchdog”, etc, de acordo com o que existe hoje)
  - objetivo: evitar “duplo watchdog” brigando.
- Durante uninstall do zid-packages:
  - não precisa reativar os watchdog antigos, mas deve deixar o sistema consistente.

===========================================================
OBJETIVO 3) LICENCIAMENTO (ONLINE + CACHE 7 DIAS + IPC LOCAL)
===========================================================

3.1 Comunicação segura/criptografada entre pacotes e zid-packages
- Implementar um canal IPC LOCAL seguro para que os pacotes consultem licença no zid-packages.
- Requisito: “seguro e criptografado”.
- Sugestão (implementar): Unix domain socket + mensagens assinadas com HMAC:
  - Socket: /var/run/zid-packages.sock (permissão 0600 root:wheel)
  - Protocolo JSON (request/response)
  - Assinatura HMAC-SHA256 usando uma chave derivada no runtime:
    - master secret compilado no binário (const)
    - derivar key = HKDF(master, uniqueid + hostname) (uniqueid vem de /var/db/uniqueid; hostname de “hostname -s”)
  - Incluir timestamp/nonce para replay protection
  - zid-packages valida assinatura e timestamp; responde com payload também assinado.

OBS: mesmo que o socket seja local/root, essa camada atende o “criptografado/seguro” sem criar arquivos de segredo.

3.2 Pacotes devem parar imediatamente sem licença
- O zid-packages watchdog SEMPRE força stop quando licença inválida.
- Além disso, cada pacote (zid-proxy, zid-geolocation, zid-logs) terá integração (documentada) para consultar licença via zid-packages:
  - Se não conseguir consultar (socket não existe, zid-packages removido, etc) => tratar como SEM LICENÇA e encerrar/parar.
- Isso garante: se desinstalar zid-packages, os pacotes param (requisito 3.4).

3.3 Gerar documento licensing.md (PT-BR) para os 3 pacotes implementarem
- Você deve criar um arquivo **docs/licensing.md** (ou licensing.md na raiz) em PORTUGUÊS BRASIL explicando:
  - Como o pacote deve consultar a licença no zid-packages (IPC)
  - Comandos/protocolo
  - Respostas possíveis
  - Como reagir (stop imediatamente)
  - Como tratar cache/expiração
  - Como lidar com “zid-packages ausente”
- Esse documento é para ser implementado dentro de CADA pacote depois.

3.4 Watchdog não inicia serviços sem licença
- Já descrito: se licença inválida, watchdog não inicia; se estiver rodando, watchdog para.

===========================================================
OBJETIVO 4) BUSCA DE LICENCIAMENTO ONLINE (N8N WEBHOOK)
===========================================================

4.1 Endpoint de licenciamento
- Fazer POST para:
  URL: https://webhook.c01.soulsolucoes.com.br/webhook/bf26a31e-11f4-4dfd-8659-94ce045b3323/soul/licensing
  Header:
    x-auth-n8n: 58ff7159c6d562c4d665de1d4d9a60f9546a0fcec885a15239f5bf5d25a48c80

4.2 Body do POST
- Deve enviar hostname (hostname -s) e id (conteúdo de /var/db/uniqueid)
- Body JSON (defina chaves consistentes, ex):
  { "hostname": "<hostname>", "id": "<uniqueid>" }

4.3 Resposta
{
  "zid-proxy": true,
  "zid-geolocation": true,
  "zid-logs": true
}

4.4 Se true: licenciado
4.5 Se o serviço de licença estiver fora e não houver retorno:
- A licença permanece válida localmente por 7 dias desde o ÚLTIMO SUCESSO.

4.6 Se ficar 7 dias sem comunicação bem sucedida:
- Marcar como SEM LICENÇA e aplicar enforcement (parar serviços).

4.7 WebGUI deve permitir acompanhar licenciamento:
- Mostrar para cada pacote:
  - Licenciado (sim/não)
  - Última consulta bem sucedida (data/hora)
  - Última tentativa (data/hora)
  - “Válido até” (last_success + 7 dias)
  - Motivo atual (OK / Offline grace / Expirado / Nunca validou)

4.8 Botão “Forçar atualização de licença”
- Na GUI: botão que executa o POST imediatamente e atualiza state

4.9 Processo normal: consulta a cada 2 horas
- Implementar scheduler interno (daemon) OU via cron adicional:
  - Recomendado: daemon interno + persistência de state; e watchdog cron garante.
- Sempre registrar data/hora da última tentativa e do último sucesso.

===========================================================
ARQUITETURA / IMPLEMENTAÇÃO (OBRIGATÓRIO)
===========================================================

A) Estrutura do repositório (sugestão, seguir padrão dos outros pacotes)
- cmd/zid-packages/                 (main + CLI subcommands)
- internal/
  - internal/licensing/             (cliente HTTP webhook + cache + validação 7 dias)
  - internal/ipc/                   (unix socket server + HMAC + protocolo)
  - internal/watchdog/              (checagem enable + serviço + start/stop)
  - internal/packages/              (metadados hardcoded dos 3 pacotes)
  - internal/s3/                    (check de version file + download bundle se precisar)
  - internal/status/                (gerar status JSON para GUI)
- packaging/pfsense/
  - files/usr/local/pkg/zid-packages.xml
  - files/usr/local/pkg/zid-packages.inc
  - files/usr/local/www/zid-packages_*.php (abas)
  - files/etc/inc/priv/zid-packages.priv.inc
  - scripts/install.sh, update.sh, uninstall.sh, bundle-latest.sh
  - pkg-plist

B) Metadados hardcoded dos pacotes (NÃO config externo)
Dentro de internal/packages, criar uma tabela hardcoded com:
- package_key: "zid-proxy" / "zid-geolocation" / "zid-logs"
- serviços a monitorar:
  - nome do rc.d, nome do processo (pgrep), comandos start/stop/status
  - zid-proxy inclui zid-appid também
- como detectar “enable” (onde ler no config.xml/config.json):
  - Você deve abrir cada pacote e replicar a forma correta.
- como obter versão instalada (método igual ao pacote)
- como checar versão remota (usar .version e URL igual ao update do pacote)
- qual comando de update (bootstrap updater existente) chamar

C) Binário único: /usr/local/sbin/zid-packages
Implementar subcomandos:
- zid-packages status --json
  - retorna JSON com:
    - pacotes: installed, enabled, licensed, service_status, version_installed, version_remote, update_available
    - licenciamento: last_attempt, last_success, valid_until, mode (OK/OFFLINE_GRACE/EXPIRED/NEVER_OK)
- zid-packages watchdog --once
  - executa 1 ciclo de enforcement:
    - carrega state local
    - se está na hora (>=2h) faz check online (ou deixa para daemon/licensing sync)
    - para cada pacote:
      - decide should_run = enabled && licensed
      - se should_run: start se necessário
      - se não: stop se estiver rodando
- zid-packages license sync
  - força POST e atualiza state
- zid-packages package install <pkg>
  - baixa bundle e roda install.sh do pacote
- zid-packages package update <pkg>
  - chama update pronto do pacote (sem reinventar)
- zid-packages daemon
  - opcional: roda IPC server + scheduler de licença (2h) + watchdog loop
  - mesmo assim criar cron (requisito)

D) Persistência de state (permitido)
- Criar /var/db/zid-packages/state.json (ou state.db), contendo SOMENTE:
  - last_attempt_ts
  - last_success_ts
  - map licensed bool por pacote (resultado da última resposta válida)
  - assinatura HMAC do arquivo para detectar alteração (HMAC com chave derivada)
- Se state for adulterado => considerar inválido (SEM LICENÇA).

E) Integração pfSense WebGUI (PHP)
- Criar menu em Services > ZID Packages
- Abas sugeridas:
  1) Packages (install/update/status)
  2) Services (status e estado do watchdog)
  3) Licensing (status por pacote + botões “Forçar atualização”)
  4) Logs (tail do log do zid-packages)
- GUI deve chamar o binário para obter status JSON e renderizar (não duplicar lógica em PHP).
- GUI deve acionar ações chamando o binário:
  - install, update, license sync
- Usar padrão de layout igual aos pacotes existentes (tabelas, badges, alerts, top tabs).

F) rc.d + cron
- Criar /usr/local/etc/rc.d/zid_packages (ou zid-packages) para iniciar “daemon” se optar.
- Criar item cron (config.xml) chamando watchdog --once.
- Garantir que, se o webConfigurator reiniciar, o item cron e menu persistem.

G) LOGS
- Log do zid-packages: /var/log/zid-packages.log
- Logar:
  - start/stop efetuado por watchdog
  - status de licença (sucesso/falha, motivo)
  - instalações/updates iniciados e resultado
- Não expor segredos no log (não logar header auth).

===========================================================
CONTEÚDO OBRIGATÓRIO DO ARQUIVO licensing.md (PT-BR)
===========================================================

Você DEVE criar o arquivo licensing.md em português Brasil com a especificação abaixo (pode adaptar texto, mas manter o conteúdo técnico). 
Arquivo: licensing.md (ou docs/licensing.md)

--- INÍCIO DO licensing.md ---

# Licenciamento ZID via `zid-packages` (pfSense)

## Objetivo
Centralizar a validação de licença dos pacotes:
- `zid-proxy` (inclui `zid-appid`)
- `zid-geolocation`
- `zid-logs`

O pacote `zid-packages` é o “servidor local” de licenças no pfSense.  
Cada pacote ZID **deve consultar** o `zid-packages` para saber se pode executar.  
Se o `zid-packages` estiver ausente (desinstalado) ou negar a licença, o pacote **deve parar imediatamente**.

## Regras de execução (obrigatórias)
1) Ao iniciar o serviço do pacote, antes de processar qualquer lógica, o pacote deve consultar a licença.
2) Se a licença for **inválida**:
   - o pacote deve encerrar o processo com exit code != 0
   - e, se aplicável, solicitar stop no rc.d (ou simplesmente finalizar e deixar o watchdog central garantir o stop)
3) Se não for possível consultar o `zid-packages` (socket não existe / erro de comunicação):
   - tratar como **SEM LICENÇA** e encerrar/parar
4) Mesmo com licença válida no start, o pacote deve revalidar periodicamente (ex: a cada 60s ou 5min) consultando o `zid-packages`.  
   - Se a licença passar a ser inválida, parar imediatamente.

## Canal de consulta (IPC local)
O `zid-packages` expõe um Unix Domain Socket:
- Path: `/var/run/zid-packages.sock`
- Permissões: acesso restrito (root), socket local

### Modelo de segurança
A mensagem é autenticada com HMAC (assinatura) para evitar spoofing local.  
A chave é derivada no runtime usando informações do próprio sistema (ex: `/var/db/uniqueid` e `hostname -s`) e um segredo embutido no binário.

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
- `nonce`: string aleatória para evitar replay
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
  - `NEVER_OK` (nunca validou online e não conseguiu validar agora)
- `valid_until`: timestamp unix do limite (last_success + 7 dias) quando aplicável
- `reason`: texto curto para log/diagnóstico
- `sig`: assinatura da resposta

## Comportamento esperado no pacote (pseudocódigo)
- Na inicialização:
  - if !LicenseCheck("zid-proxy") => exit(2)
- Em loop/intervalo:
  - if !LicenseCheck("zid-proxy") => stop imediatamente (exit / shutdown)

## Observações
- O watchdog central do `zid-packages` também aplica enforcement (stop/start), mas o pacote deve se auto-proteger.
- Se o cliente desinstalar o `zid-packages`, o socket desaparece e todos os pacotes devem parar.
- Não persistir chaves de licença em arquivos. Toda validação é por consulta ao `zid-packages`.

--- FIM DO licensing.md ---

===========================================================
CHECKLIST FINAL (obrigatório executar)
===========================================================

1) Implementar código + GUI + scripts (install/update/uninstall + pkg xml/inc)
2) Rodar:
   - gofmt -w .
   - go test ./...
3) Build FreeBSD/pfSense:
   - GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -o build/zid-packages ./cmd/zid-packages
4) Gerar bundle “latest” e version file (mesmo padrão dos outros pacotes)
5) Validar num pfSense:
   - ./scripts/install.sh
   - /etc/rc.restart_webgui
   - Services > ZID Packages aparece
   - Install/Update funcionam e mostram output
   - Licensing mostra status e botão forçar
   - Watchdog (cron) inicia/para serviços conforme enable + licença
   - Se remover zid-packages: pacotes não conseguem consultar e param

ENTREGA
- Repositório completo com:
  - binário zid-packages + internal/*
  - packaging/pfsense/* completo
  - GUI pages PHP completas
  - scripts completos
  - licensing.md (PT-BR) criado
  - README com instruções
