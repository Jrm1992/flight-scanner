# PRD - Flight Price Monitor

**Versão:** 1.0  
**Data:** Março 2026  
**Autor:** Time de Desenvolvimento  
**Status:** Requisitos Consolidados

---

## 1. Executive Summary

O **Flight Price Monitor** é uma aplicação web que permite aos usuários monitorar preços de passagens aéreas internacionais em tempo real. O sistema busca continuamente preços em múltiplas rotas configuráveis, armazena histórico de preços e dispara alertas quando preços caem abaixo de um limite definido pelo usuário.

---

## 2. Visão do Produto

### 2.1 Visão Geral
Uma plataforma que automatiza o monitoramento de preços de voos, permitindo que usuários:
- Definam rotas aéreas para monitorar (origem → destino)
- Configurem alertas de preço personalizados
- Visualizem histórico de preços em gráficos
- Recebam notificações quando os preços atingem seu alvo

### 2.2 Público-Alvo
- Viajantes frequentes que buscam melhor custo-benefício
- Agentes de viagem
- Pesquisadores de preços
- Usuários que planejam viagens com antecedência

### 2.3 Objetivos Principais
1. ✅ Automatizar busca de preços de voos internacionais
2. ✅ Persistir histórico de preços para análise
3. ✅ Alertar usuários sobre oportunidades de preço baixo
4. ✅ Oferecer visualização clara de tendências de preço
5. ✅ Permitir configuração flexível de monitoramento

---

## 3. Requisitos Funcionais

### 3.1 Gerenciamento de Rotas Monitoradas

#### RF-01: Criar Rota de Monitoramento
**Descrição:** Usuário pode criar uma nova rota para monitorar

**Critérios de Aceitação:**
- [ ] Usuário fornece: origem, destino, preço alerta, frequência de busca
- [ ] Sistema valida aeroportos (IATA codes: JFK, GIG, GRU, etc)
- [ ] Sistema persiste rota no PostgreSQL
- [ ] Sistema retorna confirmação com ID da rota

**Entrada:**
```json
{
  "origin": "GIG",
  "destination": "JFK",
  "alert_price": 450.00,
  "check_frequency_minutes": 60
}
```

**Saída:**
```json
{
  "id": "uuid",
  "origin": "GIG",
  "destination": "JFK",
  "alert_price": 450.00,
  "check_frequency_minutes": 60,
  "created_at": "2026-03-06T10:00:00Z",
  "status": "active"
}
```

---

#### RF-02: Listar Rotas Monitoradas
**Descrição:** Usuário visualiza todas as rotas que está monitorando

**Critérios de Aceitação:**
- [ ] Retorna todas as rotas ativas
- [ ] Mostra preço atual, preço alerta, última atualização
- [ ] Mostra status do monitoramento (ativo/pausado)
- [ ] Ordenável por preço, data de criação

**Saída:**
```json
{
  "routes": [
    {
      "id": "uuid",
      "origin": "GIG",
      "destination": "JFK",
      "current_price": 480.00,
      "alert_price": 450.00,
      "last_check": "2026-03-06T10:30:00Z",
      "status": "active",
      "price_trend": "down"
    }
  ]
}
```

---

#### RF-03: Atualizar Configuração de Rota
**Descrição:** Usuário modifica preço alerta ou frequência de busca

**Critérios de Aceitação:**
- [ ] Permite atualizar: `alert_price`, `check_frequency_minutes`
- [ ] Valida novos valores
- [ ] Persiste mudanças imediatamente
- [ ] Retorna rota atualizada

---

#### RF-04: Deletar Rota
**Descrição:** Usuário para de monitorar uma rota

**Critérios de Aceitação:**
- [ ] Remove rota do monitoramento ativo
- [ ] Mantém histórico de preços para referência
- [ ] Para goroutine de monitoramento dessa rota

---

### 3.2 Busca e Integração com API de Voos

#### RF-05: Buscar Preços em Tempo Real
**Descrição:** Sistema busca preços atuais para uma rota

**Critérios de Aceitação:**
- [ ] Integra com API Kiwi.com
- [ ] Retorna múltiplas opções de voos (diferentes airlines)
- [ ] Captura: preço, aerolinha, data de saída, horários
- [ ] Trata erros de timeout/indisponibilidade da API
- [ ] Implementa retry com backoff exponencial

**Integração:**
- API: Kiwi.com Flight Search API
- Rate limit: Respeitar limites da API
- Timeout: 10 segundos por requisição

---

#### RF-06: Buscar Voos Interativamente
**Descrição:** Usuário busca voos sem criar monitoramento

**Critérios de Aceitação:**
- [ ] Permite busca ad-hoc: origem, destino, data (opcional)
- [ ] Retorna preços imediatos
- [ ] Exibe múltiplas opções com detalhes completos
- [ ] Oferece opção de "monitorar esta rota"

---

### 3.3 Monitoramento em Background

#### RF-07: Monitoramento Contínuo de Preços
**Descrição:** Sistema monitora preços automaticamente em goroutines

**Critérios de Aceitação:**
- [ ] Cada rota ativa tem sua própria goroutine
- [ ] Respeita frequência configurada (`check_frequency_minutes`)
- [ ] Usa `time.Ticker` para agendamento
- [ ] Registra preço atual no banco sempre que busca
- [ ] Trata falhas gracefully (não derruba sistema)
- [ ] Implementa context para cancelamento limpo

**Comportamento:**
- Sistema inicia goroutines para todas as rotas ao iniciar
- A cada intervalo configurado, busca preço atual
- Se preço < alert_price: cria alerta

---

#### RF-08: Pausar/Retomar Monitoramento
**Descrição:** Usuário pode pausar temporariamente o monitoramento

**Critérios de Aceitação:**
- [ ] Pause para goroutine específica
- [ ] Resume recria goroutine
- [ ] Status atualizado no banco

---

### 3.4 Sistema de Alertas

#### RF-09: Criar Alerta de Preço Baixo
**Descrição:** Quando preço cai abaixo do alerta, sistema cria registro

**Critérios de Aceitação:**
- [ ] Compara preço atual com `alert_price`
- [ ] Se preço < alert_price: cria alerta
- [ ] Armazena: rota_id, preço_alerta, preço_acionado, timestamp
- [ ] Evita alertas duplicados (máximo 1 alerta por rota por dia)

**Dados Armazenados:**
```json
{
  "id": "uuid",
  "route_id": "uuid",
  "alert_price": 450.00,
  "triggered_price": 420.00,
  "triggered_at": "2026-03-06T10:30:00Z",
  "notified": false
}
```

---

#### RF-10: Listar Alertas
**Descrição:** Usuário visualiza todos os alertas disparados

**Critérios de Aceitação:**
- [ ] Lista alertas por data descrescente
- [ ] Mostra: rota, preço alerta, preço acionado, economía
- [ ] Filtrável por rota, data, status (notificado/não-notificado)

---

### 3.5 Histórico de Preços

#### RF-11: Armazenar Histórico de Preços
**Descrição:** Sistema registra todos os preços buscados

**Critérios de Aceitação:**
- [ ] Cada busca insere registro em `price_history`
- [ ] Armazena: rota_id, preço_mínimo, preço_máximo, preço_médio, timestamp
- [ ] Não se limita quantidade de registros (crescimento contínuo)

---

#### RF-12: Visualizar Gráfico de Histórico
**Descrição:** Usuário vê tendência de preços ao longo do tempo

**Critérios de Aceitação:**
- [ ] Frontend exibe gráfico de linha com histórico
- [ ] Eixo X: tempo, Eixo Y: preço
- [ ] Mostra últimos 30 dias por padrão (configurável)
- [ ] Exibe preço alerta como linha horizontal
- [ ] Interativo (hover mostra detalhes)

---

#### RF-13: Exportar Histórico
**Descrição:** Usuário pode exportar dados de histórico

**Critérios de Aceitação:**
- [ ] Exporta como CSV ou JSON
- [ ] Inclui período selecionável
- [ ] Nomeia arquivo com rota e data

---

### 3.6 Interface de Usuário

#### RF-14: Tela de Busca de Voos
**Componente:** `SearchFlight.tsx`

**Funcionalidades:**
- [ ] Inputs: origem (IATA), destino (IATA), data (opcional)
- [ ] Autocomplete para aeroportos
- [ ] Botão "Buscar Voos"
- [ ] Loading state durante busca
- [ ] Exibe resultados com preços

---

#### RF-15: Tela de Rotas Monitoradas
**Componente:** `RouteList.tsx`

**Funcionalidades:**
- [ ] Lista todas as rotas
- [ ] Card por rota mostrando: origem, destino, preço atual, alerta
- [ ] Indicador visual de status (verde = ativo, cinza = pausado)
- [ ] Botões: editar, pausar/retomar, deletar, ver histórico
- [ ] Ícone de alerta quando preço < threshold

---

#### RF-16: Tela de Histórico de Preços
**Componente:** `PriceChart.tsx`

**Funcionalidades:**
- [ ] Gráfico de linha (recharts)
- [ ] Seletor de período (7 dias, 30 dias, 90 dias, custom)
- [ ] Linha horizontal mostrando preço alerta
- [ ] Estatísticas: mín, máx, médio, trend
- [ ] Botão de exportar

---

#### RF-17: Tela de Alertas
**Componente:** `AlertsList.tsx`

**Funcionalidades:**
- [ ] Lista alertas disparados
- [ ] Mostra economia (diferença entre alerta e preço acionado)
- [ ] Filtros: rota, data, status
- [ ] Botão de marcar como "visualizado"
- [ ] Sugestão de link para compra

---

#### RF-18: Modal de Criar/Editar Rota
**Componente:** Modal customizável

**Funcionalidades:**
- [ ] Inputs validados
- [ ] Previewpré-visualização de economia estimada
- [ ] Salvar cria rota e começa monitoramento

---

## 4. Requisitos Não-Funcionais

### 4.1 Performance
- **RNF-01:** API deve responder em < 500ms
- **RNF-02:** Página React carrega em < 2s
- **RNF-03:** Gráficos renderizam com até 1 ano de histórico sem lag
- **RNF-04:** Sistema suporta monitoramento de até 100 rotas simultâneas

### 4.2 Confiabilidade
- **RNF-05:** 99% uptime para monitoramento em background
- **RNF-06:** Recuperação automática de falhas de API (retry com backoff)
- **RNF-07:** Nenhum alerta deve ser perdido mesmo com queda

### 4.3 Escalabilidade
- **RNF-08:** Goroutines escaláveis (adicionar rotas sem redeploy)
- **RNF-09:** Banco de dados otimizado para queries de histórico
- **RNF-10:** Connection pooling no PostgreSQL

### 4.4 Segurança
- **RNF-11:** API keys de terceiros armazenadas em variáveis de ambiente
- **RNF-12:** CORS configurado corretamente
- **RNF-13:** Validação de input em todas as requisições
- **RNF-14:** Sem exposição de erro interno para cliente

### 4.5 Maintainability
- **RNF-15:** Código documentado com comments explicativos
- **RNF-16:** Estrutura de pastas clara e modular
- **RNF-17:** Usar dependency injection where applicable
- **RNF-18:** Logs estruturados com níveis (debug, info, warn, error)

### 4.6 Observability
- **RNF-19:** Logs de cada busca de preço
- **RNF-20:** Métricas de falhas de API
- **RNF-21:** Timestamps precisos para debugging

---

## 5. Stack Técnico

### Backend
- **Linguagem:** Go 1.21+
- **Framework HTTP:** `net/http` padrão + routing customizado
- **Database:** PostgreSQL (Render)
- **API de Voos:** Kiwi.com API
- **Dependências:**
  - `github.com/lib/pq` - Driver PostgreSQL
  - `github.com/joho/godotenv` - Variáveis de ambiente

### Frontend
- **Framework:** React 18 + TypeScript
- **Build Tool:** Vite
- **Charts:** Recharts ou Chart.js
- **UI Components:** Tailwind CSS
- **HTTP Client:** Axios ou Fetch API

### Infraestrutura
- **Database:** PostgreSQL no Render.com (plano free)
- **Backend Hosting:** (a definir - Railway, Render, Fly.io)
- **Frontend Hosting:** Vercel, Netlify ou mesmo Render

---

## 6. Estrutura de Dados

### 6.1 Schema PostgreSQL

#### Tabela: `routes`
```sql
CREATE TABLE routes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  origin VARCHAR(3) NOT NULL,           -- IATA code
  destination VARCHAR(3) NOT NULL,      -- IATA code
  alert_price DECIMAL(10, 2) NOT NULL,
  check_frequency_minutes INT DEFAULT 60,
  status VARCHAR(20) DEFAULT 'active',  -- active, paused
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

#### Tabela: `price_history`
```sql
CREATE TABLE price_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
  min_price DECIMAL(10, 2) NOT NULL,
  max_price DECIMAL(10, 2) NOT NULL,
  avg_price DECIMAL(10, 2) NOT NULL,
  airline VARCHAR(50),
  checked_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_route_checked ON price_history(route_id, checked_at);
```

#### Tabela: `alerts`
```sql
CREATE TABLE alerts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
  alert_price DECIMAL(10, 2) NOT NULL,
  triggered_price DECIMAL(10, 2) NOT NULL,
  triggered_at TIMESTAMP DEFAULT NOW(),
  notified BOOLEAN DEFAULT FALSE,
  notified_at TIMESTAMP
);

CREATE INDEX idx_route_alerts ON alerts(route_id, triggered_at);
```

---

## 7. Fluxos Principais

### 7.1 Fluxo: Criar Rota e Começar Monitoramento

```
Usuário clica em "Adicionar Rota"
    ↓
Frontend abre modal com inputs (origem, destino, preço alerta, frequência)
    ↓
Usuário clica "Monitorar"
    ↓
POST /api/routes → Backend
    ↓
Backend valida inputs
    ↓
Backend persiste em `routes` table
    ↓
Backend retorna route_id
    ↓
Frontend exibe sucesso
    ↓
Backend inicia goroutine para essa rota
    ↓
Goroutine começa ticker com frequência configurada
```

---

### 7.2 Fluxo: Monitoramento Contínuo

```
Sistema inicializa
    ↓
SELECT * FROM routes WHERE status = 'active'
    ↓
Para cada rota: cria goroutine
    ↓
Goroutine: time.Ticker({frequência})
    ↓
Tick → Busca preço na API Kiwi.com
    ↓
Insere resultado em `price_history`
    ↓
Se preço < alert_price:
    ├─ Verifica se já tem alerta hoje
    ├─ Se não: cria novo alerta em `alerts`
    └─ Se sim: atualiza existente
    ↓
Aguarda próximo tick
```

---

### 7.3 Fluxo: Usuário Visualiza Gráfico

```
Usuário navega para rota específica → clica "Ver Histórico"
    ↓
Frontend GET /api/routes/{id}/history?days=30
    ↓
Backend query:
  SELECT avg_price, checked_at FROM price_history
  WHERE route_id = ? AND checked_at >= NOW() - INTERVAL '30 days'
  ORDER BY checked_at ASC
    ↓
Backend retorna JSON com série de preços
    ↓
Frontend renderiza gráfico com Recharts
    ↓
Usuário vê linha de preços + linha de alerta
```

---

## 8. Endpoints da API

### Routes Management
- `POST /api/routes` - Criar rota
- `GET /api/routes` - Listar rotas
- `PUT /api/routes/:id` - Atualizar rota
- `DELETE /api/routes/:id` - Deletar rota
- `PATCH /api/routes/:id/pause` - Pausar
- `PATCH /api/routes/:id/resume` - Retomar

### Search
- `POST /api/search/flights` - Buscar voos

### History
- `GET /api/routes/:id/history` - Histórico de preços
- `GET /api/routes/:id/history/export` - Exportar histórico

### Alerts
- `GET /api/alerts` - Listar alertas
- `PATCH /api/alerts/:id/mark-read` - Marcar como visualizado

---

## 9. Roadmap de Desenvolvimento

### Sprint 1: Setup e Estrutura Base
- [x] Estrutura de pastas Go
- [x] Configuração PostgreSQL Render
- [x] Models (Flight, Route, Alert, PriceHistory)
- [x] Conexão com banco
- [ ] Migrations SQL

### Sprint 2: Integração com API de Voos
- [ ] Integração Kiwi.com API
- [ ] Busca de voos
- [ ] Error handling e retry

### Sprint 3: Monitoramento Background
- [ ] Goroutines de monitoramento
- [ ] Context e cancelamento
- [ ] Time.Ticker
- [ ] Persistência de preços

### Sprint 4: Sistema de Alertas
- [ ] Lógica de alertas
- [ ] Evitar duplicatas
- [ ] Notificações (email opcional futura)

### Sprint 5: API REST
- [ ] Handlers para todos endpoints
- [ ] CORS
- [ ] Error handling

### Sprint 6: Frontend React
- [ ] Setup React + TypeScript
- [ ] Componentes principais
- [ ] Integração com API
- [ ] Gráficos

### Sprint 7: Refinamentos
- [ ] Testes (unit e integration)
- [ ] Documentação
- [ ] Deploy

---

## 10. Conceitos de Go a Aprender

Durante este projeto, você vai dominar:

1. **Structs e Tipos**
   - Definição de structs
   - JSON marshaling/unmarshaling
   - Interfaces vazias

2. **Goroutines** ⭐⭐⭐
   - Criar goroutines
   - Sincronização com WaitGroup
   - Comunicação com channels

3. **Channels** ⭐⭐⭐
   - Send/receive
   - Buffered vs unbuffered
   - Close and range over channels

4. **Context** ⭐⭐
   - Context.WithCancel
   - Context propagation
   - Cancelamento de operações

5. **Time e Scheduling**
   - time.Ticker
   - time.Sleep
   - time.Time e parsing

6. **Database/SQL**
   - Prepared statements
   - Query vs QueryRow
   - Transactions
   - Error handling

7. **HTTP Server**
   - net/http.Server
   - Routing
   - Handlers
   - Middleware

8. **Error Handling**
   - Conventional Go error returns
   - Error wrapping (fmt.Errorf)
   - Panic vs error

9. **Logging e Debugging**
   - log package
   - Printf patterns

10. **Dependency Management**
    - go.mod e go.sum
    - Imports

---

## 11. Métricas de Sucesso

- ✅ Sistema monitora com sucesso 10+ rotas simultaneamente
- ✅ Alertas disparam < 1 minuto após preço cair
- ✅ Frontend carrega em < 2s
- ✅ 0 preços perdidos mesmo com falhas de API
- ✅ Código bem estruturado e documentado
- ✅ Deploy funcional em produção

---

## 12. Próximas Fases Futuras (Out of Scope)

- Notificações por email quando alerta dispara
- Mobile app
- Histórico de compras realizadas
- Integração com múltiplas APIs de voos
- Machine learning para previsão de preços
- Autenticação de usuário
- Dashboard de múltiplos usuários

---

## Apêndice A: Variáveis de Ambiente

```env
# Database
DATABASE_URL=postgres://user:pass@host:port/dbname

# API de Voos
KIWI_API_KEY=seu_api_key_aqui

# Server
SERVER_PORT=8080
ENV=development

# CORS
FRONTEND_URL=http://localhost:3000
```

---

**Fim do PRD**
