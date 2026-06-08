# 🛡️ Módulo B: Serviço de Análise de Risco (Antifraude)

Este microsserviço é o **Módulo B** da arquitetura distribuída desenvolvida para a avaliação de risco financeiro em transações de gateway de pagamento. Desenvolvido em **Go**, ele expõe simultaneamente endpoints **gRPC** e **REST HTTP/1.1** para permitir testes comparativos de desempenho e multiplexação.

O serviço foi projetado utilizando **Arquitetura Hexagonal (Ports and Adapters)**, garantindo alto isolamento entre a regra de negócio, os protocolos de comunicação e a persistência de dados.

## 📂 Estrutura do Projeto e Arquitetura

A organização das pastas segue o padrão de mercado para microsserviços em Go:

```text
.
|-- build/                    # Infraestrutura e Deploy
|   |-- Dockerfile            # Multi-stage build para gerar a imagem do contêiner
|   `-- backup.json           # Seed inicial do banco persistente baseado em arquivos
|-- cmd/
|   `-- server/
|       `-- main.go           # Ponto de entrada. Injeta dependências e sobe os servidores
|-- internal/                 # Código privado do microsserviço
|   |-- adapters/             # Camada de Infraestrutura (Comunicação com o mundo externo)
|   |   |-- grpc/
|   |   |   `-- handler.go    # Implementa Unary e Client Streaming via gRPC
|   |   |-- repository/
|   |   |   `-- in_memory.go  # DB Thread-safe em memória com persistência em JSON
|   |   `-- rest/
|   |       `-- handler.go    # Expõe a mesma lógica via REST para o benchmark
|   |-- config/               # Gerenciamento de Ambiente
|   |   `-- app.go            # Leitura do .env e fallback de variáveis de ambiente
|   `-- core/                 # Coração da Arquitetura Hexagonal
|       |-- domain/
|       |   `-- risk.go       # Entidades puras do negócio (Transaction, RiskResult)
|       `-- ports/
|           |-- repository.go # Contrato de acesso a dados (Driven Port)
|           `-- risk.go       # Contrato de entrada do caso de uso (Driving Port)
|   |-- usecase/              # Camada de Aplicação (Use Cases)
|   |   `-- risk_service.go   # Lógica antifraude, cruzamento de dados e geração de carga (Key Stretching)
`-- proto/                    # Contratos de Comunicação
    |-- antifraud.proto       # Definição agnóstica das mensagens e serviços
    `-- pb/                   # Código Go auto-gerado pelo compilador do Protocol Buffers

```

### 🧠 Decisões Arquiteturais Relevantes

1. **Banco de Dados em Memória Persistente:** Utilizamos `sync.RWMutex` para suportar milhares de requisições concorrentes. Os dados (histórico e blocklist) são persistidos automaticamente no `build/backup.json` a cada nova escrita.
2. **Key Stretching (Gargalo de CPU):** Para o benchmark gRPC vs REST, a aplicação implementa um loop de 50.000 iterações de `sha256` para gerar a Assinatura de Segurança da transação. Isso simula o processamento pesado e evidencia o poder do HTTP/2 no gRPC em ambientes de alto estresse.



---

## ⚙️ Variáveis de Ambiente

O projeto utiliza o arquivo `.env` para configuração local (veja o `.env.example`). No ambiente Kubernetes, essas variáveis devem ser injetadas nos *Pods*.

| Variável | Valor Padrão | Descrição |
| --- | --- | --- |
| `REST_PORT` | `8081` | Porta em que o servidor REST/JSON irá escutar. |
| `GRPC_PORT` | `50052` | Porta em que o servidor gRPC (HTTP/2) irá escutar. |
| `BACKUP_FILE` | `build/backup.json` | Caminho do arquivo para carregar e salvar o estado do DB. |

---

## 🚀 Como Executar

### 1. Rodando Localmente (Para Desenvolvimento)

Certifique-se de ter o **Go 1.25+** instalado.

```bash
# Baixe as dependências
go mod tidy

# Execute o servidor a partir da raiz do projeto
go run ./cmd/server/main.go

```

*Os servidores subirão nas portas `8081` (REST) e `50052` (gRPC) e carregarão a seed de dados do JSON.*

### 2. Executar via Docker (Para Kubernetes / Minikube)

O `Dockerfile` foi posicionado na pasta `build/`. Para gerar a imagem corretamente, **o comando de compilação (build) deve ser executado na raiz do repositório**:

```bash
# Gerar a imagem estática localmente
docker build -f build/Dockerfile -t antifraud-service:latest .

# Executar o contentor a mapear ambas as portas
docker run -p 8081:8081 -p 50052:50052 antifraud-service:latest

```

---

## 🧪 Como Testar e Gerar a Massa de Dados

Você pode testar a aplicação utilizando o **Postman** (que possui suporte nativo a gRPC e multiplexação).

### 📡 API REST (Porta 8081)

**POST** `http://localhost:8081/api/v1/risk/analyze`

```json
{
  "transactionId": "tx-12345",
  "userId": "user-888",
  "amount": 2500.50,
  "ipAddress": "192.168.0.10"
}

```

### 🚄 API gRPC (Porta 50052)

No Postman, crie uma nova requisição do tipo **gRPC** e importe o arquivo `proto/antifraud.proto` para carregar os contratos.

#### 1. Unary Streaming (`AnalyzeRiskScore`)

Validação imediata da transação. Selecione o método e envie o payload abaixo:

```json
{
  "transaction_id": "tx-grpc-404",
  "user_id": "user-999", 
  "amount": 900.00,
  "ip_address": "172.16.254.1"
}

```

*(Nota: O `user-999` e o IP `172.16.254.1` já estão na seed do `backup.json`, o que causará bloqueio automático provando que o banco em memória está funcionando sob a trava do Mutex).*



#### 2. Client Streaming (`BatchChargeback`)

Processamento em lote sob demanda. Para testar esse fluxo contínuo:

1. Selecione o método `BatchChargeback`.
2. Clique em **Invoke** para abrir o túnel HTTP/2.
3. Envie múltiplos payloads individuais (como o abaixo). Cada IP enviado é processado um a um e adicionado à *blocklist* com complexidade de memória $\mathcal{O}(1)$.
4. Clique em **End Stream** para que o servidor feche a conexão e retorne o sumário (`ChargebackSummary`).

```json
{
  "transaction_id": "tx-hacked-1",
  "reason": "Cloned Card",
  "ip_address": "10.0.0.155"
}

```



---