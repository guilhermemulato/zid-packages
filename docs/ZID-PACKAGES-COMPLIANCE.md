# Guia de Conformidade com `zid-packages` (Template para Novos Projetos)

## Objetivo
Este documento define um padrao unico para qualquer novo pacote que precise ser gerenciado pelo `zid-packages` (GUI, watchdog, auto-update, licenciamento e operacoes).

Use este guia como checklist obrigatorio antes de publicar um pacote novo.

---

## 1) Contrato de Integracao (Obrigatorio)

Todo pacote novo deve expor estes contratos:

1. `BundleURL` publico e estavel.
2. `VersionURL` publico e estavel.
3. updater bootstrap em `/usr/local/sbin/<pkg-key>-update`.
4. bundle contendo `scripts/install.sh` e `scripts/update.sh`.
5. binario principal em `/usr/local/sbin/<service-bin>`.
6. comando de versao funcional (`-version` e preferencialmente `--version`).
7. servico `rc.d` funcional com start/stop previsiveis.
8. fonte de `enable` previsivel para watchdog/status.
9. registro no `config.xml` (installedpackages/package e menu).
10. compatibilidade de licenciamento via socket do `zid-packages` (quando aplicavel).

---

## 2) Nomenclatura Canonica

Defina e mantenha consistencia entre:

1. `pkg-key` (chave usada no `zid-packages`), exemplo: `zid-orchestrator`.
2. nome tecnico interno do servico/binario, exemplo: `zid-orchestration`.
3. nome de package no `config.xml`, exemplo: `zid-orchestration`.

Regra:
1. Se houver diferenca entre `pkg-key` e nome tecnico, documente alias.
2. Garanta fallback de licenciamento para ambos os nomes.

Exemplo de alias no lado cliente:

```go
// pkg-key usado no zid-packages
key := "zid-orchestrator"
// nome tecnico aceito pelo webhook/licensing
alias := "zid-orchestration"
```

---

## 3) Estrutura Minima Recomendada do Repositorio

```text
.
├── Makefile
├── scripts/
│   ├── install.sh
│   ├── update.sh
│   └── update-bootstrap.sh
├── build/
│   └── <binario-principal>
├── pkg/
│   └── pfSense-pkg-<nome>/
│       ├── files/usr/local/etc/rc.d/<servico>
│       ├── files/usr/local/pkg/<arquivo>.inc
│       ├── files/usr/local/pkg/<arquivo>.xml
│       ├── files/usr/local/www/<telas>.php
│       ├── scripts/post-install
│       └── scripts/post-deinstall
├── dist/
└── <pkg-key>-latest.version
```

---

## 4) Contrato de Bundle (Obrigatorio)

O `.tar.gz` final deve conter apenas runtime/distribuicao. Nao empacote o repo inteiro.

Deve conter:

1. `scripts/install.sh`
2. `scripts/update.sh`
3. `scripts/update-bootstrap.sh`
4. binario compilado (`build/<binario>`)
5. arquivos pfSense (`pkg/pfSense-pkg-.../files/...`)

Nao deve conter:

1. `.git/`
2. artefatos antigos (`*-latest.tar.gz` dentro do bundle)
3. arquivos de desenvolvimento sem uso em runtime

Comando de validacao:

```sh
tar -tzf <pkg-key>-latest.tar.gz | sed -n '1,200p'
```

---

## 5) Makefile Base (Exemplo)

```makefile
VERSION ?= 1.0.0
DIST_DIR ?= dist
PKG_KEY ?= zid-exemplo
BUNDLE_NAME ?= $(PKG_KEY)-latest.tar.gz

.PHONY: build bundle-latest clean

build:
	go build -o build/zid-exemplo ./cmd/zid-exemplo

bundle-latest: build
	@mkdir -p $(DIST_DIR)
	@tmp_dir=$$(mktemp -d /tmp/$(PKG_KEY)-bundle.XXXXXX); \
	mkdir -p $$tmp_dir/$(PKG_KEY)/scripts $$tmp_dir/$(PKG_KEY)/build $$tmp_dir/$(PKG_KEY)/pkg; \
	cp -f scripts/install.sh scripts/update.sh scripts/update-bootstrap.sh $$tmp_dir/$(PKG_KEY)/scripts/; \
	cp -f build/zid-exemplo $$tmp_dir/$(PKG_KEY)/build/; \
	cp -R pkg/pfSense-pkg-zid-exemplo $$tmp_dir/$(PKG_KEY)/pkg/; \
	tar -C $$tmp_dir -czf $(DIST_DIR)/$(BUNDLE_NAME) $(PKG_KEY); \
	rm -rf $$tmp_dir
	@printf "$(VERSION)\n" > $(DIST_DIR)/$(PKG_KEY)-latest.version
	@sha256 -q $(DIST_DIR)/$(BUNDLE_NAME) > $(DIST_DIR)/sha256.txt || sha256sum $(DIST_DIR)/$(BUNDLE_NAME) | awk '{print $$1}' > $(DIST_DIR)/sha256.txt
```

---

## 6) `install.sh` (Obrigatorio)

Responsabilidades minimas:

1. validar root.
2. copiar binario para `/usr/local/sbin`.
3. copiar `rc.d` para `/usr/local/etc/rc.d`.
4. copiar `.inc`, `.xml` e paginas da GUI.
5. criar/atualizar config runtime.
6. registrar pacote/menu no `config.xml` via PHP (`write_config()`).
7. instalar bootstrap updater em `/usr/local/sbin/<pkg-key>-update`.
8. habilitar servico (`sysrc ..._enable=YES`) e iniciar.

Exemplo simplificado:

```sh
#!/bin/sh
set -eu

[ "$(id -u)" = "0" ] || { echo "Execute como root"; exit 1; }

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

install -d /usr/local/sbin /usr/local/pkg /usr/local/www/services /usr/local/etc/rc.d
install -m 755 "$ROOT_DIR/build/zid-exemplo" /usr/local/sbin/zid-exemplo
install -m 755 "$ROOT_DIR/pkg/pfSense-pkg-zid-exemplo/files/usr/local/etc/rc.d/zid_exemplo" /usr/local/etc/rc.d/zid_exemplo
install -m 644 "$ROOT_DIR/pkg/pfSense-pkg-zid-exemplo/files/usr/local/pkg/zid-exemplo.inc" /usr/local/pkg/zid-exemplo.inc
install -m 644 "$ROOT_DIR/pkg/pfSense-pkg-zid-exemplo/files/usr/local/pkg/zid-exemplo.xml" /usr/local/pkg/zid-exemplo.xml
install -m 644 "$ROOT_DIR/pkg/pfSense-pkg-zid-exemplo/files/usr/local/www/zid_exemplo_settings.php" /usr/local/www/services/zid_exemplo_settings.php

install -m 755 "$ROOT_DIR/scripts/update-bootstrap.sh" /usr/local/sbin/zid-exemplo-update

/usr/sbin/sysrc zid_exemplo_enable=YES >/dev/null 2>&1 || true
service zid_exemplo onestart >/dev/null 2>&1 || true
```

---

## 7) `update.sh` (Obrigatorio)

Responsabilidades minimas:

1. baixar bundle mais recente.
2. extrair em `/tmp`.
3. localizar `*/scripts/install.sh` dentro do bundle.
4. executar `install.sh`.
5. suportar `-u`, `-f`, `-k`.

---

## 8) `update-bootstrap.sh` (Obrigatorio)

Responsabilidades minimas:

1. permanecer pequeno e estavel.
2. baixar bundle novo.
3. executar `update.sh` embarcado no bundle.
4. ser instalado como `/usr/local/sbin/<pkg-key>-update`.

Este arquivo e o comando que o `zid-packages` chama no update manual/auto-update.

---

## 9) Contrato de Versao

O binario deve responder em pelo menos um formato:

1. `<nome> -version` ou
2. `<nome> --version`

Recomendacao:

1. suportar ambos para compatibilidade.
2. retornar string simples `X.Y.Z`.

Exemplo:

```go
if *showVersion {
	fmt.Println(version)
	return
}
```

---

## 10) Contrato de Servico (`rc.d`)

Minimo esperado:

1. `PROVIDE`, `REQUIRE`, `KEYWORD`.
2. `name`, `rcvar`, `command`, `command_args`, `pidfile`.
3. `run_rc_command "$1"`.

Exemplo:

```sh
name="zid_exemplo"
rcvar="${name}_enable"
command="/usr/local/sbin/zid-exemplo"
command_args="--config /usr/local/etc/zid-exemplo/config.json"
pidfile="/var/run/zid-exemplo.pid"
```

---

## 11) Fonte de `Enabled` (Watchdog/Status)

Defina uma fonte padrao, de preferencia:

1. `/etc/rc.conf.local` e `/etc/rc.conf` via `<rcvar>_enable`
2. ou campo fixo em `config.xml` se o pacote usa GUI pfSense tradicional

Regra:

1. documentar exatamente onde o `zid-packages` deve ler.
2. evitar formatos ambiguos.

---

## 12) Registro no `config.xml` (Obrigatorio)

No install, garantir:

1. entrada em `installedpackages.package`.
2. entrada em `installedpackages.menu`.
3. `include_file` e `configurationfile` corretos.

Sempre usar PHP + `write_config()`. Nao editar XML por `sed`.

---

## 13) Licenciamento com `zid-packages` (Quando Aplicavel)

Se o pacote consulta licenca via socket:

1. socket: `/var/run/zid-packages.sock`
2. `package_name` deve bater com chave esperada no licensing.
3. se houver alias de nome, suportar fallback.

Checklist:

1. pacote bloqueia funcionalidade quando sem licenca.
2. watchdog nao entra em loop por mismatch de chave.

---

## 14) Como Cadastrar no `zid-packages`

No `zid-packages`, cadastrar:

1. `Key`
2. `BundleURL`
3. `VersionURL`
4. `UpdateCommand`
5. `InstallScriptGlob`

Exemplo:

```go
{
	Key:               "zid-exemplo",
	Name:              "ZID Exemplo",
	BundleURL:         "https://s3.../zid-exemplo-latest.tar.gz",
	VersionURL:        "https://s3.../zid-exemplo-latest.version",
	UpdateCommand:     "/usr/local/sbin/zid-exemplo-update",
	InstallScriptGlob: "*/scripts/install.sh",
}
```

---

## 15) Testes de Aceite Minimos (Antes de Release)

1. `go test ./...`
2. `go build ./...`
3. `make bundle-latest`
4. validar conteudo do tar:
   `tar -tzf <pkg>-latest.tar.gz`
5. install limpo em pfSense:
   `sh scripts/install.sh`
6. update manual:
   `sh /usr/local/sbin/<pkg-key>-update -f`
7. update via `zid-packages`:
   `zid-packages package update <pkg-key>`
8. watchdog:
   `zid-packages watchdog --once`
9. status:
   `zid-packages status --json`

---

## 16) Erros Comuns que Quebram Integracao

1. nao existir `/usr/local/sbin/<pkg-key>-update`.
2. bundle nao conter `scripts/install.sh`.
3. install nao copiar binario para `/usr/local/sbin`.
4. versao nao detectavel (`-version`/`--version` ausente).
5. `pkg-key` diferente do nome de licenca sem alias.
6. install nao registrar package/menu no `config.xml`.
7. usar `post-install` como instalador principal do bundle (insuficiente na maioria dos casos).

---

## 17) Checklist Final (Copiar e Marcar)

1. [ ] `scripts/install.sh` existe e instala runtime completo.
2. [ ] `scripts/update.sh` existe e reaplica install do bundle.
3. [ ] `scripts/update-bootstrap.sh` existe.
4. [ ] bootstrap instalado como `/usr/local/sbin/<pkg-key>-update`.
5. [ ] binario em `/usr/local/sbin/<service-bin>`.
6. [ ] comando de versao funcional.
7. [ ] `rc.d` funcional com `onestart/onestop`.
8. [ ] `enable` lido de fonte documentada e estavel.
9. [ ] package/menu registrados no `config.xml`.
10. [ ] bundle contem apenas runtime.
11. [ ] `*-latest.version` e `sha256.txt` atualizados.
12. [ ] testado via `zid-packages` (status/watchdog/update).

