# zid-packages

## Status atual
- Estrutura inicial do projeto criada (Go module, CLI base e pacotes internos).
- Licenciamento online implementado com cache e state assinado.
- IPC local seguro implementado (socket + HMAC/HKDF + nonce/timestamp).
- Watchdog central (daemon + ciclo --once) implementado com enforcement basico.
- GUI pfSense criada (Packages/Services/Licensing/Logs).
- GUI pfSense padronizada com labels em ingles.
- Scripts de install/update/uninstall e empacotamento criados.
- Bundles gerados (tar.gz, .version e sha256).
- Sync de licenca a cada 2 horas com tolerancia a resposta vazia do webhook.
- Licenciamento aceita resposta do webhook com valores string ("true"/"false").
- IPC de licenca registra tentativas e possui flags de debug (inclui export da key).
- Watchdog loga motivo do start/stop (enabled/licensed/mode).
- Watchdog do zid-proxy le enable em /conf/config.xml (zidproxy/config/enable).
- Watchdog do zid-proxy/zid-geolocation le enable em /conf/config.xml (paths flexiveis).
- Aba Services permite start/stop/restart do zid-packages via GUI.
- Aba Services mostra botao Start ou Restart/Stop conforme status do daemon.
- Aba Logs mostra as 50 linhas mais recentes no topo, com refresh manual e auto refresh opcional.
- Aba Logs tem controles de refresh mais claros com badge de status.
- Aba Services mostra status do daemon do zid-packages corretamente.
- Leitura do enable com retry para evitar flapping durante gravações do config.xml.
- Cache temporario do enable para evitar flapping do watchdog.
- Log debug opcional para leitura do enable (ZID_PACKAGES_ENABLE_DEBUG=1).
- Watchdog loga snapshot bruto do enable quando para serviço por enabled=false.
- Leitura do enable via PHP config.inc (igual zid-geolocation) para evitar flapping.
- Aba Packages agora inclui o zid-packages (licenca N/A).
- Update desabilitado quando a versao remota estiver vazia.
- Install garante zid_packages_enable=YES para iniciar no boot.

## Proximos passos
- Validacao em pfSense real.
