# Changelog

## [0.4.13] - 2026-01-29
- Mostra badge \"Up to date\" na aba Packages quando a versao local = remota.

## [0.4.12] - 2026-01-29
- Forca zid_packages_enable=YES em /etc/rc.conf.local durante install.

## [0.4.11] - 2026-01-29
- Corrige erro de parse PHP no botao Update da aba Packages.

## [0.4.10] - 2026-01-29
- Mostra versao do zid-packages apenas com numero (via build com ldflags).
- Desabilita Update quando versao remota estiver vazia.

## [0.4.9] - 2026-01-29
- Exibe zid-packages na aba Packages e marca licenca como N/A.
- Suporte ao pacote zid-packages em status/update.

## [0.4.8] - 2026-01-29
- Padroniza a GUI do pfSense em ingles (menus, colunas e mensagens).

## [0.4.7] - 2026-01-29
- Trata resposta vazia do licensing sem erro (EOF) e cria diretório do state automaticamente.
- Cria /var/db/zid-packages no install.

## [0.4.6] - 2026-01-29
- Corrige erro de parse em zid-packages.inc que quebrava a GUI.

## [0.4.5] - 2026-01-29
- Fallback para php-cgi quando php-cli estiver ausente/zerado.

## [0.4.4] - 2026-01-29
- Alinha install/register ao padrão do zid-proxy e corrige paths da GUI para raiz.

## [0.4.3] - 2026-01-29
- Sincroniza versao do installer com a versao do bundle.
- Adiciona target make bundle-latest.

## [0.4.2] - 2026-01-29
- Registro inline agora usa script temporario e grava log dedicado /var/log/zid-packages-register.log.

## [0.4.1] - 2026-01-29
- Install agora grava log do PHP inline com codigo de saida e versao do installer.

## [0.4.0] - 2026-01-29
- Corrige install.sh para apontar binario no bundle e registrar config.xml com includes absolutos.
- Logs inline do registro com mensagens PHP.

## [0.3.9] - 2026-01-29
- Registro inline no install (mesmo padrão do zid-geolocation) para garantir menu.
- GUI agora em /usr/local/www/services e XML/privs ajustados.

## [0.3.8] - 2026-01-29
- Logs em arquivo no install e register para diagnostico no pfSense.

## [0.3.7] - 2026-01-29
- Logs detalhados no install.sh e register-package.php.

## [0.3.6] - 2026-01-29
- register-package.php agora igual ao padrao do zid-proxy e com logs detalhados.
- CLI expõe "-version" para detectar versao no registro.

## [0.3.5] - 2026-01-29
- Ajusta register/unregister para seguir o padrao do zid-proxy com mensagens e menu.

## [0.3.4] - 2026-01-29
- Adiciona unregister-package.php e instala scripts de registro/remoção no pfSense.

## [0.3.3] - 2026-01-29
- Registro automatico no config.xml para aparecer no menu do pfSense.
- Install usa onestart e corrige habilitacao persistente.

## [0.3.2] - 2026-01-29
- Fix no install script encontrar arquivos no bundle (paths corrigidos).

## [0.3.1] - 2026-01-29
- Ajuste no bundle-latest para rodar fora do pfSense (sha256sum fallback).
- Bundle e version file gerados.

## [0.3.0] - 2026-01-29
- GUI pfSense para gerenciamento de pacotes, servicos, licenciamento e logs.
- Scripts de install/update/uninstall e empacotamento do pacote.

## [0.2.0] - 2026-01-29
- IPC local seguro via socket + HMAC/HKDF com protecao de replay.
- Watchdog central com daemon continuo e enforcement de servicos.
- Estado de licenca assinado (HMAC) e validado ao carregar.
- Status JSON inclui servico zid-appid.

## [0.1.1] - 2026-01-29
- Licenciamento online via webhook com persistencia local e modos de validade.
- Status JSON inclui informacoes de licenca por pacote.

## [0.1.0] - 2026-01-29
- Estrutura inicial do projeto (Go module, CLI base e pacotes internos).
- Metadados hardcoded dos pacotes (URLs e comandos de update).
- Estado base de licenciamento e status em JSON.
