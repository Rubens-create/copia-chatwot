# Mapeamento do Schema Chatwoot para WhatsApp Gateway

Este documento descreve a estrutura oficial das tabelas do **Chatwoot** utilizadas para suportar canais WhatsApp e o fluxo exato de persistência e consulta adotado pelo nosso backend em Go.

---

## 1. Tabelas Relevantes do Chatwoot

### 1.1 `accounts` (Contas / Tenancy)
Representa a conta/tenant dona dos canais, conversas e contatos.
* `id` (serial/integer, PK): Identificador da conta (ex: `1`).
* `name` (varchar): Nome da organização (ex: "Default Account").
* `locale` (integer): Idioma padrão.
* `settings` (jsonb): Configurações globais da conta.
* `created_at`, `updated_at` (timestamp): Metadados temporais.

### 1.2 `channel_whatsapp` (Canal WhatsApp)
Representa a configuração da linha/número do WhatsApp (Meta WhatsApp Cloud API / On-Premises).
* `id` (bigserial/bigint, PK): Identificador único do canal.
* `account_id` (integer, FK -> `accounts.id`, NOT NULL): Conta associada.
* `phone_number` (varchar, NOT NULL, UNIQUE): Número do WhatsApp cadastrado (formato E.164, ex: `+5562999999999`).
* `provider` (varchar, default `'default'`): Provedor WhatsApp (`whatsapp_cloud`, `360dialog`, etc.).
* `provider_config` (jsonb): Contém credenciais (`phone_number_id`, `business_account_id`, `api_key`/`access_token`, `webhook_verify_token`).
* `business_management_token` (text): Token de gerenciamento da Meta.
* `message_templates` (jsonb): Modelos de mensagens aprovados na Meta.
* `created_at`, `updated_at` (timestamp).

### 1.3 `inboxes` (Caixas de Entrada)
Representa a caixa de entrada que conecta a conta ao canal de atendimento.
* `id` (serial/integer, PK): Identificador da caixa de entrada.
* `channel_id` (integer, NOT NULL): ID do canal correspondente (ex: `channel_whatsapp.id`).
* `channel_type` (varchar): Tipo do canal (`"Channel::Whatsapp"`).
* `account_id` (integer, NOT NULL, FK -> `accounts.id`).
* `name` (varchar, NOT NULL): Nome da caixa de entrada (ex: `"WhatsApp +5562999999999"`).
* `lock_to_single_conversation` (boolean, default false): Se permite apenas 1 conversa aberta simultânea por contato.
* `created_at`, `updated_at` (timestamp).

### 1.4 `contacts` (Contatos)
Representa o usuário final que envia ou recebe mensagens no WhatsApp.
* `id` (serial/integer, PK): Identificador único do contato.
* `account_id` (integer, NOT NULL, FK -> `accounts.id`): Conta associada.
* `name` (varchar, default `''`): Nome do perfil no WhatsApp ou número formatado.
* `phone_number` (varchar, nullable): Telefone em formato E.164 (`+5562999999999`).
* `identifier` (varchar, nullable): Identificador externo opcional.
* `additional_attributes` (jsonb, default `{}`): Metadados adicionais (avatar, pushname, etc.).
* `custom_attributes` (jsonb, default `{}`): Atributos personalizados.
* `last_activity_at` (timestamp): Última interação do contato.
* `created_at`, `updated_at` (timestamp).
* **Índices:** `index_contacts_on_phone_number_and_account_id`, `uniq_identifier_per_account_contact`.

### 1.5 `contact_inboxes` (Vínculo Contato <-> Caixa de Entrada)
Mapeia o contato a uma caixa de entrada específica por meio de um `source_id`.
* `id` (bigserial/bigint, PK).
* `contact_id` (bigint, NOT NULL, FK -> `contacts.id`).
* `inbox_id` (bigint, NOT NULL, FK -> `inboxes.id`).
* `source_id` (text, NOT NULL): Identificador único do contato naquele canal (o número de telefone E.164 ou ID WhatsApp, ex: `+5562999999999`).
* `pubsub_token` (varchar, unique): Token de pubsub gerado para o cliente.
* `hmac_verified` (boolean, default false).
* `created_at`, `updated_at` (timestamp).
* **Constraint / Índice Único:** `index_contact_inboxes_on_inbox_id_and_source_id` `UNIQUE(inbox_id, source_id)`.

### 1.6 `conversations` (Conversas)
Representa a sessão ou histórico de atendimento entre o contato e a caixa de entrada.
* `id` (serial/integer, PK): ID interno da conversa.
* `account_id` (integer, NOT NULL, FK -> `accounts.id`).
* `inbox_id` (integer, NOT NULL, FK -> `inboxes.id`).
* `contact_id` (bigint, NOT NULL, FK -> `contacts.id`).
* `contact_inbox_id` (bigint, NOT NULL, FK -> `contact_inboxes.id`).
* `status` (integer, NOT NULL, default `0`):
  * `0`: `open`
  * `1`: `resolved`
  * `2`: `pending`
  * `3`: `snoozed`
* `display_id` (integer, NOT NULL): Número sequencial visível da conversa por conta (ex: 1, 2, 3...).
* `uuid` (uuid, NOT NULL, default `gen_random_uuid()`, UNIQUE): UUID da conversa.
* `identifier` (varchar, nullable): Identificador único de canal.
* `last_activity_at` (timestamp, NOT NULL, default `CURRENT_TIMESTAMP`).
* `additional_attributes` (jsonb, default `{}`): Guarda contexto da Meta (ex: `whatsapp_phone_number_id`).
* `custom_attributes` (jsonb, default `{}`): Atributos customizados.
* `created_at`, `updated_at` (timestamp).

### 1.7 `messages` (Mensagens)
Armazena cada mensagem trocada na conversa.
* `id` (serial/integer, PK): ID interno da mensagem.
* `account_id` (integer, NOT NULL, FK -> `accounts.id`).
* `inbox_id` (integer, NOT NULL, FK -> `inboxes.id`).
* `conversation_id` (integer, NOT NULL, FK -> `conversations.id`).
* `message_type` (integer, NOT NULL):
  * `0`: `incoming` (recebida do contato)
  * `1`: `outgoing` (enviada pelo sistema/agente)
  * `2`: `activity` (evento de sistema)
  * `3`: `template` (mensagem de modelo Meta)
* `content` (text): Texto da mensagem ou legenda da mídia.
* `content_type` (integer, default `0`):
  * `0`: `text`
  * `1`: `input_text`
  * `2`: `input_textarea`
  * `3`: `input_email`
  * `4`: `input_select`
  * `5`: `cards`
  * `6`: `form`
  * `7`: `article`
  * `8`: `incoming_email`
* `content_attributes` (jsonb, default `{}`): Metadados de conteúdo, replies (`in_reply_to`, etc.).
* `sender_type` (varchar): `"Contact"` (incoming) ou `"User"` / `"AgentBot"` (outgoing).
* `sender_id` (bigint): ID do contato ou usuário que enviou.
* `source_id` (text): Identificador externo da Meta (`wamid.HBg...`).
* `external_source_ids` (jsonb, default `{}`): Ex: `{"whatsapp": "wamid.HBg..."}`.
* `additional_attributes` (jsonb, default `{}`): Payload bruto, IDs de contexto, etc.
* `processed_message_content` (text, nullable).
* `status` (integer, default `0`):
  * `0`: `sent`
  * `1`: `delivered`
  * `2`: `read`
  * `3`: `failed`
* `private` (boolean, default false): Se é nota privada interna.
* `created_at`, `updated_at` (timestamp).
* **Índices Relevantes:** `index_messages_on_source_id`, `index_messages_on_conversation_id`, `index_messages_on_account_id`.

### 1.8 `attachments` (Anexos e Mídias)
Armazena mídias associadas a uma mensagem.
* `id` (serial/integer, PK).
* `account_id` (integer, NOT NULL).
* `message_id` (integer, NOT NULL, FK -> `messages.id`).
* `file_type` (integer, default `0`):
  * `0`: `image`
  * `1`: `audio`
  * `2`: `video`
  * `3`: `file` / `document`
  * `4`: `location`
  * `5`: `fallback`
  * `7`: `contact`
* `external_url` (varchar): URL da mídia na Meta ou storage.
* `coordinates_lat`, `coordinates_long` (float): Coordenadas geográficas para localização.
* `fallback_title` (varchar): Nome do arquivo, legenda ou nome do vCard.
* `extension` (varchar): Extensão do arquivo (ex: `"png"`, `"mp4"`, `"pdf"`).
* `meta` (jsonb, default `{}`): MIME type, tamanho em bytes, sha256, etc.
* `created_at`, `updated_at` (timestamp).

### 1.9 `webhooks` (Assinantes Externos de Webhook)
Representa destinos HTTP externos configurados para receber eventos do Chatwoot / Gateway.
* `id` (bigserial/bigint, PK).
* `account_id` (integer, FK -> `accounts.id`).
* `inbox_id` (integer, nullable).
* `url` (text, NOT NULL): URL HTTP de destino.
* `webhook_type` (integer, default `0`): `0` (account webhook), `1` (inbox webhook).
* `subscriptions` (jsonb): Ex: `["message_created", "message_updated", "conversation_created", "conversation_updated"]`.
* `name` (varchar).
* `secret` (varchar).
* `created_at`, `updated_at` (timestamp).

---

## 2. Fluxo Completo de uma Mensagem WhatsApp Recebida

```text
               Meta WhatsApp Cloud API
                          │
                          ▼
            [POST /webhooks/whatsapp]
                          │
             (1) Validação da assinatura/token
             (2) Parsing do JSON da Meta
                          │
                          ▼
           Localizar Canal WhatsApp / Inbox
             (Query: channel_whatsapp onde phone_number = display_phone_number
                     ou provider_config->>'phone_number_id' = phone_number_id)
                          │
                          ▼
            Localizar ou Criar Contato (contacts)
             (Query: contacts por phone_number + account_id)
                          │
                          ▼
       Localizar ou Criar Vínculo (contact_inboxes)
             (Query: contact_inboxes por inbox_id + source_id)
                          │
                          ▼
        Localizar ou Criar Conversa (conversations)
             (Buscar conversa aberta status=0 para contact_id + inbox_id,
              ou criar nova com display_id incremental)
                          │
                          ▼
          Verificação de Idempotência no PostgreSQL
             (Query: messages onde source_id = wamid)
             Se já existe -> ignora criação (HTTP 200)
             Se nova -> persiste em messages (+ attachments se houver)
                          │
                          ▼
            Publicação de Job na Fila Redis
             Queue: "incoming_messages" / "outgoing_webhooks"
                          │
                          ▼
               Worker Go Assíncrono
                          │
                          ▼
             Carrega Webhooks Ativos (`webhooks`)
                          │
                          ▼
          Disparo HTTP Externo com Retentativas (Exponential Backoff)
```

---

## 3. Compatibilidade e Idempotência

1. **Campos Reutilizados:** Todos os campos essenciais das tabelas `accounts`, `channel_whatsapp`, `inboxes`, `contacts`, `contact_inboxes`, `conversations`, `messages`, `attachments` e `webhooks` seguem estritamente a tipagem e os enums do Chatwoot.
2. **Garantia de Idempotência:**
   * Busca e trava por `source_id` (`wamid`).
   * Verificação e restrição única por `source_id` evitando duplicação em caso de reenvio pela Meta.
3. **Persistência do Payload Bruto:**
   * O payload bruto original da Meta é preservado integralmente dentro de `messages.additional_attributes -> 'raw_payload'`.
