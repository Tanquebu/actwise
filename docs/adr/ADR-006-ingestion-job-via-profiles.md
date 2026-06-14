# ADR-006 — Job `ingestion` (Go) esposto via Compose profiles

- **Stato**: Accettato
- **Data**: 2026-06-14
- **Decisori**: Massimiliano Nicosia
- **Contesto correlato**: SPEC §4 (ingestion Go), §5 (architettura), §10 (compose), Sprint 1

## Contesto

Nel progetto esistono due componenti distinti spesso entrambi chiamati "ingestion":

| Componente | Cosa fa | Dove | Come gira |
|---|---|---|---|
| `services/ingestion` (Go) | fetch PDF EUR-Lex → parse → chunk → hash → JSONL | container `ingestion` | job batch, on-demand |
| pipeline ingestion (Python) | legge JSONL → embedda → push su Qdrant | dentro `api` | `docker compose run --rm api python scripts/ingest.py` |

Il servizio Go **non** parla con Qdrant: non embedda (l'embedding è Python,
Sprint 2). Il diagramma in SPEC §5 ("… → Qdrant") è semplificato; il flusso
reale è **Go → JSONL → Python → Qdrant**.

Questo ADR riguarda **come modellare il container Go nel Compose**: è un job
batch CLI senza porta HTTP (SPEC §10), che non deve restare "always up".

## Opzioni considerate

### A — Stesso compose, con `profiles`
Servizio definito con `profiles: ["tools"]`: non parte con `docker compose up`
(solo i servizi senza profilo partono di default), si invoca esplicitamente.
- ✅ Un solo file, semantica chiara, idiomatico

### B — Stesso compose, senza profilo
- ❌ `docker compose up` proverebbe ad avviarlo; essendo un batch o esce subito
  o entra in restart-loop. Da evitare.

### C — Compose separato (`docker-compose.ingestion.yml`)
- ✅ Isolamento massimo dei job batch
- ❌ Più file, più carico cognitivo. Sovradimensionato per l'MVP.

## Decisione

Adottiamo l'**opzione A**: il servizio `ingestion` nel compose principale con
`profiles: ["tools"]`, invocato on-demand:

```bash
# dev
docker compose run --rm ingestion go run ./cmd/ingest
# prod: stesso servizio, binario compilato
```

Passaggio dati Go → Python tramite **volume condiviso** tra `ingestion` e `api`:
bind mount su host in dev (es. `./data/chunks/*.jsonl`, ispezionabile a occhio),
named volume in prod.

## Conseguenze

- Il container `ingestion` è "dormiente" finché non viene invocato: nessun
  processo inutile sempre attivo, nessun restart-loop.
- Va definito un volume condiviso `./data` (o named `actwise-chunks`) montato sia
  da `ingestion` sia da `api`.
- Il servizio Go **non** dipende da Qdrant. È lo script Python (in `api`) a dover
  attendere Qdrant `healthy` (`depends_on` con `condition: service_healthy`).
- Se i job batch si moltiplicano e il compose principale diventa rumoroso, si può
  rivalutare l'opzione C.
