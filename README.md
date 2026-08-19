# Chatwoot Lite - WhatsApp Gateway

Sistema backend modular e de alta performance desenvolvido em **Go**, **PostgreSQL** e **Redis**, focado exclusivamente no recebimento, persistência e encaminhamento assíncrono de eventos da **Meta WhatsApp Cloud API**, mantendo **100% de compatibilidade com a estrutura de dados do Chatwoot**.

---

## 🚀 Arquitetura e Fluxo

```text
Meta WhatsApp Cloud API
        │
        ▼
   [GET / POST /webhooks/whatsapp]
        │
        ├─► Validação de Token e Assinatura
        ├─► Identificação/Criação do Contato (contacts)
        ├─► Associação do Contato à Caixa de Entrada (contact_inboxes)
        ├─► Identificação/Criação da Conversa (conversations)
        ├─► Persistência da Mensagem e Anexos (messages, attachments)
        ├─► Idempotência garantida via wamid (source_id)
        │
        ▼
   [Redis Queue: incoming_messages & outgoing_webhooks]
        │
        ▼
   [Worker Go Assíncrono]
        │
        ├─► Carrega assinantes ativos (webhooks)
        └─► Disparo HTTP com Retentativas Exponenciais (5s -> 15s -> 1m -> 5m -> Dead-Letter)
```

---

## 📦 Stack Tecnológica

* **Go 1.22+**: HTTP API REST, Handlers, Services, Workers e Dispatchers.
* **PostgreSQL 16**: Persistência utilizando o schema nativo do Chatwoot.
* **Redis 7**: Filas de eventos (`queue:incoming_messages`, `queue:outgoing_webhooks`, retries ZSET e dead-letter).
* **Docker & Docker Compose**: Orquestração simplificada para desenvolvimento e produção.

---

## 🗄️ Compatibilidade com Chatwoot

O sistema foi desenhado para operar diretamente sobre um banco de dados PostgreSQL existente do Chatwoot:

* **Contatos:** `contacts` (telefone normalizado em formato E.164 `+55...`, nome de perfil, atributos).
* **Caixas de Entrada:** `inboxes` e `channel_whatsapp`.
* **Vínculos:** `contact_inboxes` com chave única `(inbox_id, source_id)`.
* **Conversas:** `conversations` (com `display_id` sequencial por conta, `uuid`, `last_activity_at`).
* **Mensagens:** `messages` com tipos nativos (`incoming = 0`, `outgoing = 1`), payload bruto em JSONB e identificador Meta em `source_id` para idempotência.
* **Mídias:** `attachments` (imagens, áudios, vídeos, documentos, localizações e contatos).
* **Webhooks Externos:** `webhooks` e histórico em `webhook_delivery_attempts`.

Consulte o documento [CHATWOOT_SCHEMA.md](CHATWOOT_SCHEMA.md) para detalhes exatos do schema.

---

## ⚙️ Como Executar

### 1. Usando Docker Compose (Recomendado)

```bash
docker compose up --build -d
```

Serviços iniciados:
* **API REST & UI:** `http://localhost:8080`
* **Documentação e ambiente de testes:** `http://localhost:8080/docs`
* **Worker Assíncrono:** background process
* **PostgreSQL:** `localhost:5432`
* **Redis:** `localhost:6379`

### 2. Executando Localmente (Go)

```bash
# 1. Configurar variáveis de ambiente
cp .env.example .env

# 2. Baixar dependências
go mod tidy

# 3. Executar a API
go run ./cmd/api

# 4. Executar o Worker (em outro terminal)
go run ./cmd/worker
```

---

## 📡 Endpoints da API

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | Verificação de liveness |
| `GET` | `/ready` | Verificação de prontidão (PostgreSQL + Redis) |
| `GET` | `/webhooks/whatsapp` | Verificação do Webhook Meta (`hub.challenge`) |
| `POST` | `/webhooks/whatsapp` | Recebimento rápido de eventos da Meta |
| `GET` | `/api/conversations` | Lista conversas ordenadas por última atividade |
| `GET` | `/api/conversations/:id` | Detalhes de uma conversa específica |
| `GET` | `/api/conversations/:id/messages` | Mensagens e anexos de uma conversa |
| `POST` | `/api/conversations/:id/messages` | Envio de mensagem para o contato |
| `GET` | `/api/webhooks` | Lista webhooks externos cadastrados |
| `POST` | `/api/webhooks` | Cadastra novo webhook externo com secret HMAC |

### Envio de mensagem de voz

O parâmetro `is_voice` é opcional e pertence ao anexo de áudio. Use `true` somente para mensagem de voz. O arquivo final enviado à Meta precisa ser Ogg codificado com Opus. Gravações WebM feitas pela interface são convertidas automaticamente.

```json
{
  "attachments": [
    {
      "file_type": 1,
      "external_url": "https://exemplo.com/audio.ogg",
      "fallback_title": "audio.ogg",
      "is_voice": true
    }
  ]
}
```

Para áudio básico, defina `is_voice` como `false` ou omita o campo. Em requisições `multipart/form-data`, envie `is_voice=true` junto com o arquivo em `attachments[]`.

A especificação OpenAPI está disponível em `/openapi.yaml`. A interface Swagger em `/docs` permite autorizar com `API_ACCESS_TOKEN` e executar requisições no ambiente atual.

---

## 🧪 Executando os Testes

```bash
go test -v -race ./...
```

---

## 🔁 Política de Retentativa de Webhooks

Falhas transitórias em webhooks externos são reprocessadas automaticamente com backoff exponencial:

1. **Tentativa 1:** Imediata
2. **Tentativa 2:** +5 segundos
3. **Tentativa 3:** +15 segundos
4. **Tentativa 4:** +1 minuto
5. **Tentativa 5:** +5 minutos
6. **Dead-Letter Queue:** Movido para `queue:dead_letter_webhooks` e registrado na tabela `webhook_delivery_attempts`.
