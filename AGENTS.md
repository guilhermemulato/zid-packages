## Padroes e Regras de Desenvolvimento

- Sempre leia o arquivo **specs.md**, nela vai encontrar tudo que precisa referente ao projeto.
- Mantenha informacoes como funcionalidades sempre atualizada no arquivo **specs.md**
- Sempre responda em portugues PT-BR
- Use o MCP Context7 para buscar informaoes sobre documentacoes, se manter atualizado
- User o MCP Playwright para acessar o navegador e testar voce mesmo quando necessario.
- Use o MCP github para acessar os repositorios que tem relacao com esse projeto, como o **zid-proxy(guilhermemulato/zid-proxy)** e **zid-geolocation(guilhermemulato/zid-geolocation)**. Neles tambem ja foram implementadas diversas funcionalidades na WEB Gui do pfsense, entao voce pode consulta la quando tiver duvidas para ver como foi feito
- Go: sempre rodar `gofmt -w .` antes de entregar alteracoes.
- Testes: preferir testes determinísticos em `internal/*/*_test.go`.
- Mudou codigo? Atualize o `CHANGELOG.md` e **bump de versao** no `Makefile`.
- Alteracao pequena: use sufixo incremental (ex.: `1.6.1`).
- Alteracao grande: use sufixo incremental (ex.: `1.6`).
- Ao final, gere novamente os bundles (`make bundle-latest`) e garanta:
    - `zid-packages-latest.version` atualizado
    - `sha256.txt` atualizado
    - **bundles obrigatorios** (zid-packages-latest.tar.gz)
- Sempre executar ao final de cada implementacao: `go test ./...` e `go build ./...`

