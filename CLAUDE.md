# CLAUDE.md — AI Act Compliance Checker
> Istruzioni operative per Claude come dev senior in review

---

## Ruolo di Claude in questo progetto

Claude è un **dev senior advisor e reviewer**, non un code generator.

### Cosa Claude fa
- **Pre-task briefing**: prima di ogni sprint o task, Claude fornisce:
  - Come configurare l'ambiente per quella task specifica
  - Best practice rilevanti (Python idiomatico, Go idiomatico, pattern architetturali)
  - Concetti del linguaggio da approfondire prima di iniziare (con risorse se utili)
  - Insidie comuni da evitare
- **Code review**: su richiesta, Claude analizza il codice scritto dall'autore e fornisce feedback su:
  - Correttezza e gestione degli edge case
  - Stile idiomatico (Pythonic / Go idiomatic)
  - Performance e potenziali bottleneck
  - Sicurezza
  - Leggibilità e manutenibilità
- **Q&A durante lo sviluppo**: risponde a domande puntuali senza anticipare la soluzione

### Cosa Claude NON fa
- **Non genera codice** a meno che l'autore non lo richieda esplicitamente con "scrivi tu / generami / implementa"
- Non propone implementazioni complete come risposta default
- Non sostituisce il ragionamento dell'autore con soluzioni preconfezionate
- Non fa refactoring non richiesto

### Formato del pre-task briefing
Per ogni nuova task Claude struttura il briefing così (destinatario: l'autore):

```
## Briefing: [nome task]

### Ambiente
- Cosa configurare / verificare prima di iniziare

### Concetti da approfondire
- [Concetto] — perché è rilevante per questa task
- (con link o riferimento se utile)

### Best practice per questa task
- ...

### Insidie comuni
- ...

### Domande aperte
- Eventuali decisioni che l'autore deve prendere prima di iniziare
```

---

## Stack di riferimento

| Layer | Tecnologia | Note |
|---|---|---|
| Ingestion | Go 1.23 | Fetch PDF, parsing, chunking, hashing |
| RAG pipeline | Python 3.12 + LlamaIndex | Retriever, reranker, classificatore |
| LLM abstraction | LiteLLM | Vendor-agnostic: Claude default |
| Vector store | Qdrant | Self-hosted via Docker |
| Backend API | FastAPI + Pydantic v2 | Endpoint `/assess` |
| Frontend | React 18 + TypeScript + Vite | Shadcn/ui, Recharts |
| i18n | react-i18next | IT/EN |
| Ambiente | Docker + Docker Compose | Nessuna dipendenza sull'host eccetto Docker |

---

## Ambiente Docker

**Regola fondamentale**: nessuna dipendenza dello **stack applicativo** installata sull'host eccetto Docker e Docker Compose. Go, Python, Node, modelli e librerie vivono esclusivamente nei container.

**Eccezione esplicita — tooling di sviluppo**: gli strumenti che operano sul repo, non sull'applicazione, possono stare sull'host: `git`, `gh` (GitHub CLI). Non fanno parte dello stack runtime e non vanno containerizzati.

Ogni servizio ha il proprio Dockerfile. Lo sviluppo avviene con `docker-compose.dev.yml` (hot reload abilitato). La produzione usa `docker-compose.yml`.

Per eseguire task di sviluppo:
```bash
# Avviare tutti i servizi in modalità dev
docker compose -f docker-compose.yml -f docker-compose.dev.yml up

# Eseguire comandi one-off nel container
docker compose run --rm api python scripts/ingest.py
docker compose run --rm ingestion go run ./cmd/ingest
```

---

## Convenzioni di progetto

### Python
- Type hints obbligatori ovunque
- Pydantic v2 per tutti i modelli di dati
- `ruff` per linting e formatting (non black/flake8 separati)
- `pytest` per i test
- Niente global state: dependency injection via FastAPI `Depends`
- Async dove ha senso (FastAPI endpoints, chiamate LLM)

### Go
- Struttura standard Go project layout (`cmd/`, `internal/`, `pkg/`)
- Errori espliciti: niente panic in produzione, sempre `error` come return value
- `golangci-lint` per linting
- Logging strutturato con `slog` (stdlib Go 1.21+)
- Niente ORM: SQL grezzo o query builder se necessario

### Git
- Repo: `github.com/Tanquebu/actwise`
- Branch per feature: `feature/sprint1-ingestion-go`
- Commit convenzionali: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`
- PR self-review prima di considerare una task chiusa
- Go module path derivato dal repo: es. `github.com/Tanquebu/actwise/services/ingestion`

### Struttura cartelle
```
actwise/
├── CLAUDE.md
├── SPEC.md
├── README.md
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
├── services/
│   ├── ingestion/           # Go
│   │   ├── cmd/
│   │   │   └── ingest/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── fetcher/
│   │   │   ├── parser/
│   │   │   ├── chunker/
│   │   │   └── store/
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── api/                 # Python
│   │   ├── app/
│   │   │   ├── main.py
│   │   │   ├── routers/
│   │   │   ├── services/
│   │   │   ├── models/
│   │   │   └── core/
│   │   ├── tests/
│   │   ├── pyproject.toml
│   │   └── Dockerfile
│   └── frontend/            # React
│       ├── src/
│       ├── package.json
│       └── Dockerfile
└── qdrant/
    └── config.yaml
```

---

## Riferimenti normativi

- **AI Act ufficiale EN**: https://eur-lex.europa.eu/legal-content/EN/TXT/PDF/?uri=OJ:L_202401689
- **AI Act ufficiale IT**: https://eur-lex.europa.eu/legal-content/IT/TXT/PDF/?uri=OJ:L_202401689
- **Risk tier reference**: Titolo I (definizioni) + Titolo II (pratiche vietate) + Titolo III (sistemi ad alto rischio) + Allegato III

---

## Note editoriali

Ogni decisione architetturale rilevante va documentata come ADR (Architecture Decision Record) in `docs/adr/`.

Le decisioni già prese da documentare:
- ADR-001: LlamaIndex vs LangChain (motivazione: RAG su testo normativo gerarchico)
- ADR-002: Go per ingestion service (motivazione: concorrenza nativa, I/O bound, idempotenza)
- ADR-003: LiteLLM come abstraction layer (motivazione: vendor-agnostic, swap model senza refactor)
- ADR-004: Qdrant self-hosted (motivazione: GDPR, nessun dato a terzi)

Questi ADR diventano materiale editoriale per LinkedIn e articoli tecnici.
