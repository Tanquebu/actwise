# Repository Guidelines

## Struttura del Progetto e Organizzazione dei Moduli

ActWise e' un compliance checker per l'AI Act. L'implementazione attuale e' centrata sul servizio di ingestion in Go sotto `services/ingestion/`, con entry point eseguibili in `cmd/` e package privati in `internal/`. Il package attivo oggi e' `internal/fetcher`.

La documentazione vive nella root del repository e in `docs/adr/`: `SPEC.md` descrive obiettivi di prodotto e architettura, `CLAUDE.md` registra le convenzioni di sviluppo, e gli ADR documentano le decisioni architetturali. La configurazione Docker di sviluppo e' in `docker-compose.dev.yml`.

I servizi pianificati nella specifica includono `services/api/` per FastAPI, `services/frontend/` per React/Vite e `qdrant/` per la configurazione del vector store. Aggiungi queste directory solo quando implementi il servizio corrispondente.

## Comandi di Build, Test e Sviluppo

Usa workflow Docker-first; evita di installare sull'host dipendenze dello stack applicativo.

```bash
docker compose -f docker-compose.dev.yml run --rm ingestion go run ./cmd/ingest
```

Esegue il comando di ingestion dentro il container Go.

```bash
docker compose -f docker-compose.dev.yml run --rm ingestion go test ./...
```

Esegue tutti i test Go del modulo di ingestion.

```bash
docker compose -f docker-compose.dev.yml run --rm ingestion gofmt -w .
```

Formatta il codice Go nel servizio di ingestion. Esegui i comandi dalla root del repository, salvo indicazioni esplicite diverse.

## Stile di Codice e Convenzioni di Naming

Per Go, segui il layout standard: comandi in `cmd/<name>/`, implementazione privata in `internal/<package>/` e identificatori esportati solo quando fanno parte di una vera API di package. Usa `gofmt`, ritorni `error` espliciti e logging strutturato con `log/slog`. Evita `panic` nei percorsi di produzione.

Per i futuri servizi Python, usa type hints, modelli Pydantic v2, `ruff` e `pytest`. Per il futuro frontend, usa React 18, TypeScript, Vite e lo stack di design gia' pianificato in `SPEC.md`.

## Linee Guida per i Test

Colloca i test Go accanto al package testato con file `_test.go`, per esempio `services/ingestion/internal/fetcher/fetcher_test.go`. Preferisci test table-driven per parsing, hashing ed edge case di fetch. Isola il comportamento di rete esterna dietro test double, cosi' i test restano deterministici.

Esegui `go test ./...` nel container di ingestion prima di aprire una PR. Aggiungi test quando modifichi comportamento di fetch, logica di parsing, chunking o contratti dei data model.

## Linee Guida per Commit e Pull Request

La cronologia Git usa prefissi in stile Conventional Commits, inclusi `chore:` e `docs:`. Continua con `feat:`, `fix:`, `refactor:`, `test:` e `docs:` quando appropriato. Mantieni i messaggi brevi e all'imperativo, per esempio `feat: add ingestion PDF fetcher`.

Le pull request devono includere una descrizione concisa, issue o task collegati quando disponibili, comandi eseguiti e screenshot solo per modifiche UI. Documenta come ADR in `docs/adr/` le decisioni con impatto architetturale prima del merge.

## Sicurezza e Configurazione

Non committare segreti, artefatti locali di processo, briefing generati o PDF sorgente scaricati. Mantieni la configurazione runtime in variabili d'ambiente e fornisci esempi sicuri tramite `.env.example` quando aggiungi file di configurazione.
