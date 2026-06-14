# SPEC.md — AI Act Compliance Checker
> Versione 0.1 — Draft iniziale  
> Autore: Massimiliano Nicosia  
> Data: Giugno 2026

---

## 1. Vision

Uno strumento che permette a un IT Manager o responsabile tecnico di una PMI italiana di inserire la descrizione di un sistema AI in uso (o in valutazione) e ottenere:

- la classificazione del risk tier secondo l'AI Act (UE 2024/1689)
- gli obblighi applicabili per quel tier
- una gap analysis rispetto alla situazione dichiarata
- le azioni prioritarie con riferimenti normativi tracciabili

**Non è un sostituto della consulenza legale.** È uno strumento operativo per orientarsi prima di coinvolgere un DPO o un consulente.

---

## 2. Target utente

**Primario**: IT Manager / Tech Lead di PMI (50–500 dipendenti), senza team legale interno dedicato, che deve capire se un sistema AI che sta adottando ricade sotto l'AI Act e cosa comporta.

**Secondario**: consulenti IT che vogliono uno strumento di pre-assessment da usare con i clienti.

---

## 3. Perimetro MVP

### In scope
- Classificazione risk tier: Unacceptable / High / Limited / Minimal
- Articoli AI Act pertinenti (fonte: testo ufficiale EN + IT)
- Gap analysis su 5 aree: trasparenza, supervisione umana, data governance, documentazione tecnica, gestione del rischio
- Output strutturato con confidence score e tracciabilità fonti
- UI bilingue IT/EN
- Modalità "explain": l'utente può chiedere spiegazioni su ogni voce dell'output

### Out of scope (v1)
- Integrazione con sistemi aziendali reali
- Autenticazione utenti / persistenza sessioni
- Supporto a sistemi AI General Purpose (GPAI) — trattazione separata nell'AI Act
- Conformità ad altri regolamenti (GDPR, NIS2) — possibile v2

---

## 4. Stack tecnico

### Ingestion service — Go
- Fetch del PDF ufficiale AI Act (EN + IT) dal sito EUR-Lex
- Parsing e chunking semantico rispettando la struttura gerarchica (Titolo → Capo → Articolo → Paragrafo → Lettera)
- Hash SHA-256 per idempotenza (non re-indicizza chunk già presenti)
- Output: JSONL strutturato → push a vector store

**Motivazione Go**: task CPU-bound e I/O intensivo (fetch, parsing PDF, hashing batch). Concorrenza nativa con goroutine. Giustificabile architetturalmente, non forzato.

### RAG pipeline — Python
- **Framework**: LlamaIndex (vs LangChain — vedi sezione 8)
- **LLM abstraction**: LiteLLM (vendor-agnostic: Claude, OpenAI, Mistral switchabili)
- **Vector store**: Qdrant (self-hostable, rilevante per GDPR)
- **Embedding model**: `intfloat/multilingual-e5-large` (multilingua EN/IT, open source)
- **Backend API**: FastAPI
- **Validazione output**: Pydantic v2

### Frontend — React + TypeScript
- Shadcn/ui come component library
- Recharts / D3 per visualizzazioni (gauge risk tier, radar gap analysis)
- i18n: react-i18next (IT/EN)
- Vite come bundler

---

## 5. Architettura

```
┌─────────────────────────────────────────────────────┐
│                   INGESTION (Go)                    │
│  EUR-Lex PDF → Parser → Chunker → SHA256 → Qdrant  │
└─────────────────────────┬───────────────────────────┘
                          │ vettori + metadati
┌─────────────────────────▼───────────────────────────┐
│                  RAG PIPELINE (Python)               │
│                                                     │
│  Input utente                                       │
│       │                                             │
│       ▼                                             │
│  Retriever (LlamaIndex)                             │
│  → top-k chunk rilevanti con metadati articolo      │
│       │                                             │
│       ▼                                             │
│  Reranker (cross-encoder)                           │
│       │                                             │
│       ▼                                             │
│  Classificatore risk tier (LLM via LiteLLM)         │
│       │                                             │
│       ▼                                             │
│  Gap analyzer (LLM via LiteLLM)                     │
│       │                                             │
│       ▼                                             │
│  Output strutturato + confidence + fonti            │
│  (Pydantic v2)                                      │
└─────────────────────────┬───────────────────────────┘
                          │ JSON
┌─────────────────────────▼───────────────────────────┐
│                  FRONTEND (React)                   │
│  Form input → Risk gauge → Radar gap → Fonti trace  │
└─────────────────────────────────────────────────────┘
```

---

## 6. Data model output (Pydantic)

```python
class RiskTier(str, Enum):
    UNACCEPTABLE = "unacceptable"
    HIGH = "high"
    LIMITED = "limited"
    MINIMAL = "minimal"

class GapArea(BaseModel):
    area: str                    # es. "Trasparenza"
    status: str                  # "compliant" | "partial" | "gap"
    description: str
    articles_referenced: list[str]  # es. ["Art. 13", "Art. 14"]
    priority: int                # 1 (alta) → 3 (bassa)

class ComplianceReport(BaseModel):
    system_description: str
    risk_tier: RiskTier
    risk_tier_rationale: str
    applicable_articles: list[str]
    gap_analysis: list[GapArea]
    priority_actions: list[str]
    confidence_score: float      # 0.0–1.0
    sources: list[SourceChunk]   # chunk recuperati con riferimento articolo
    disclaimer: str
    language: str                # "it" | "en"
    generated_at: datetime
```

---

## 7. Chunking strategy — AI Act

Il testo normativo ha struttura gerarchica con riferimenti incrociati. Un chunking naive (fixed-size) perde il contesto.

**Strategia adottata**:
- **Unità base**: paragrafo di articolo (es. "Art. 13, par. 2")
- **Metadati per chunk**: titolo, capo, articolo, paragrafo, lettera (se presente), lingua
- **Overlap semantico**: i considerando (recitals) rilevanti vengono collegati agli articoli corrispondenti come chunk satellite
- **Cross-reference resolution**: quando un articolo rimanda ad un altro ("ai sensi dell'articolo 6"), il chunk include un riferimento esplicito nel metadato

Questo permette al retriever di recuperare non solo il testo pertinente ma anche il contesto normativo corretto.

---

## 8. Decisione architetturale: LlamaIndex vs LangChain

| Criterio | LlamaIndex | LangChain |
|---|---|---|
| RAG su documenti strutturati | Nativo (NodeParser, QueryEngine) | Possibile ma più verboso |
| Testo normativo gerarchico | SubQuestionQueryEngine naturale | Richiede customizzazione |
| Reranking integrato | Sì | Parziale |
| Curva di apprendimento | Media | Bassa (già usato) |
| Ecosistema agent | In crescita | Più maturo |

**Scelta**: LlamaIndex per questo progetto. Il testo normativo con struttura gerarchica e cross-reference è il suo caso d'uso ideale. LangChain rimane nel portfolio per structured output extraction (concorsi-qualifier).

**Nota editoriale**: questa scelta sarà documentata in un post dedicato — "perché ho scelto LlamaIndex su LangChain per RAG su testo normativo".

---

## 9. Considerazioni GDPR

- **Nessun dato utente persistito** in v1: le descrizioni dei sistemi AI inserite dall'utente non vengono salvate
- **Vector store self-hosted** (Qdrant locale o VPS): nessun dato inviato a terzi eccetto la chiamata LLM
- **LiteLLM + Claude API**: i prompt contengono solo la descrizione del sistema AI (dati tecnici, non personali in senso stretto)
- **Disclaimer esplicito** in output: lo strumento non costituisce parere legale

---

## 10. Workflow di sviluppo

### Filosofia
Il codice è scritto dall'autore. Claude agisce come dev senior in review e come advisor pre-task.

**Ruolo di Claude per ogni sprint/task**:
1. **Pre-task briefing**: ambiente da configurare, best practice da seguire, concetti Python/Go da approfondire prima di iniziare
2. **Review del codice**: feedback su correttezza, idiomatic style, performance, sicurezza
3. **Risposta a domande puntuali** durante lo sviluppo

Claude non genera codice a meno che l'autore non lo richieda esplicitamente.

### Ambiente: Docker-first
Nessuna dipendenza installata sulla macchina host eccetto Docker e Docker Compose.

```
actwise/
├── docker-compose.yml          # orchestrazione completa
├── docker-compose.dev.yml      # override per sviluppo (hot reload)
├── services/
│   ├── ingestion/              # Go service
│   │   └── Dockerfile
│   ├── api/                    # Python FastAPI
│   │   └── Dockerfile
│   └── frontend/               # React + Vite
│       └── Dockerfile
├── qdrant/                     # config Qdrant
└── .env.example
```

**Servizi docker-compose**:
| Servizio | Immagine base | Porta |
|---|---|---|
| `ingestion` | `golang:1.23-alpine` | — (CLI, no HTTP) |
| `api` | `python:3.12-slim` | 8000 |
| `frontend` | `node:22-alpine` | 5173 (dev) / 80 (prod) |
| `qdrant` | `qdrant/qdrant` | 6333 |

---

## 11. Sprint plan

### Sprint 1 — Fondamenta (settimane 1–2)
- [ ] Setup repo monorepo (Go service + Python backend + React frontend)
- [ ] Go: fetch + parsing PDF AI Act EN e IT da EUR-Lex
- [ ] Go: chunking gerarchico + hash idempotente
- [ ] Python: setup Qdrant locale + ingestion pipeline
- [ ] Python: primo retriever LlamaIndex funzionante (smoke test)

### Sprint 2 — RAG core (settimane 3–4)
- [ ] Embedding multilingua (`multilingual-e5-large`)
- [ ] Reranker (cross-encoder `ms-marco-MiniLM`)
- [ ] Classificatore risk tier via LiteLLM + Claude
- [ ] Output Pydantic v2 + confidence score
- [ ] FastAPI endpoint `/assess`

### Sprint 3 — Gap analysis + UI (settimane 5–6)
- [ ] Gap analyzer per 5 aree
- [ ] Tracciabilità fonti (chunk → articolo → link EUR-Lex)
- [ ] React: form input + risk gauge + radar chart
- [ ] React: modalità "explain" per ogni voce
- [ ] i18n IT/EN

### Sprint 4 — Polish + documentazione (settimane 7–8)
- [ ] Evaluation del RAG (hit rate, MRR su test set manuale)
- [ ] Gestione casi edge (sistemi GPAI, sistemi esclusi dal perimetro)
- [ ] README tecnico + ADR (Architecture Decision Records)
- [ ] Preparazione contenuto editoriale (3 post LinkedIn + articolo lungo)
- [ ] Demo pubblica (Vercel frontend + VPS backend)

---

## 11. Metriche di successo MVP

| Metrica | Target |
|---|---|
| Hit rate retriever (test set 20 query) | > 80% |
| Classificazione risk tier corretta (validazione manuale) | > 85% |
| Tempo risposta endpoint `/assess` | < 10s |
| Copertura articoli AI Act indicizzati | 100% Titoli I–IX |
| Contenuto editoriale prodotto | ≥ 3 post LinkedIn + 1 articolo |

---

## 12. Rischi

| Rischio | Probabilità | Mitigazione |
|---|---|---|
| Chunking AI Act perde riferimenti incrociati | Alta | Metadati espliciti + overlap semantico |
| LLM allucinazione su articoli specifici | Media | Tracciabilità fonti obbligatoria + confidence score |
| Go parsing PDF complesso | Media | Fallback: pdftotext + processing Python |
| Scope creep (GDPR, NIS2...) | Alta | Perimetro hard: solo AI Act v1 |
| Demo pubblica con input malevoli | Bassa | Rate limiting + sanitizzazione input |
