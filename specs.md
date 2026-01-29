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
- Aba Packages agora inclui o zid-packages (licenca N/A).
- Update desabilitado quando a versao remota estiver vazia.
- Install garante zid_packages_enable=YES para iniciar no boot.

## Proximos passos
- Validacao em pfSense real.
