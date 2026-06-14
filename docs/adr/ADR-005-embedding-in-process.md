# ADR-005 — Embedding model in-process nel servizio `api`

- **Stato**: Accettato
- **Data**: 2026-06-14
- **Decisori**: Massimiliano Nicosia
- **Contesto correlato**: SPEC §4 (stack RAG), §6 (output), Sprint 2

## Contesto

Il modello di embedding del progetto è `intfloat/multilingual-e5-large`
(~560M parametri, ~2.2 GB di pesi in fp32 in RAM, più overhead di runtime).
Serve in due momenti distinti:

1. **Ingestion** (lato Python): embedding dei chunk dell'AI Act prima del push su Qdrant.
2. **Query time** (`/assess`): embedding della descrizione del sistema inserita dall'utente.

A query time il servizio `api` carica inoltre il reranker (cross-encoder
`ms-marco-MiniLM`). La domanda è quindi quanta RAM far risiedere stabilmente
nel container `api`, e se isolare l'embedding in un servizio dedicato.

L'MVP è single-instance, senza autenticazione né persistenza di sessione,
con target di latenza `/assess` < 10s (SPEC §11 metriche).

## Opzioni considerate

### A — In-process (modello dentro `api`)
- ✅ Più semplice: un container in meno, nessun hop di rete, latenza minima
- ✅ Dev locale immediato
- ❌ ~2.2 GB residenti nel container `api` (più reranker)
- ❌ Ogni eventuale replica orizzontale ricarica la propria copia del modello
- ❌ Accoppia il ciclo di vita del modello a quello dell'API (cold start più lento)

### B — Servizio di embedding dedicato (es. HuggingFace TEI)
- ✅ Un solo caricamento del modello, condiviso da ingestion e api
- ✅ `api` leggera; scaling indipendente; serving ottimizzabile (ONNX/GPU)
- ✅ Separazione pulita delle responsabilità
- ❌ Un container in più, hop di rete, più complessità nel compose (+ readiness)
- ❌ Over-engineering per un MVP single-user

## Decisione

Adottiamo l'**opzione A (in-process in `api`)** per l'MVP, con il vincolo di
**isolare la logica di embedding in un modulo Python condiviso**
(es. `app/services/embeddings.py`) anziché inlinearla nell'endpoint.

In questo modo la pipeline di ingestion riusa lo stesso modulo (carica il
modello, embedda, esce — job transitorio che rilascia la RAM), e l'estrazione
futura dietro un servizio dedicato non tocca i call site.

## Conseguenze

- Semplicità massima per l'MVP, coerente con il vincolo single-instance.
- Il container `api` ha un footprint RAM significativo (embedding + reranker):
  va dimensionato di conseguenza il VPS.
- **Trigger di rivalutazione** verso l'opzione B: RAM residente del container
  `api` oltre la soglia tollerata dal deploy target, oppure p95 della latenza
  di embedding fuori budget rispetto ai 10s end-to-end, oppure necessità di
  scaling orizzontale di `api`.

## Note operative

⚠️ `multilingual-e5-large` richiede i prefissi `query:` e `passage:` sul testo.
Vanno applicati **coerentemente** in ingestion e a query time: un mismatch
degrada la similarità in modo silenzioso. Il modulo condiviso è il punto
naturale dove centralizzare questa regola.
