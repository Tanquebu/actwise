# Architecture Decision Records — actwise

Registro delle decisioni architetturali. Ogni ADR documenta una scelta, il suo
contesto, le alternative considerate e le conseguenze. Le decisioni reversibili
indicano un *trigger di rivalutazione*.

Formato: Nygard-style (Contesto → Opzioni → Decisione → Conseguenze).
Stati possibili: `Proposto` · `Accettato` · `Superato da ADR-XXX` · `Deprecato`.

## Indice

| ADR | Titolo | Stato |
|---|---|---|
| ADR-001 | LlamaIndex vs LangChain | Da scrivere |
| ADR-002 | Go per ingestion service | Da scrivere |
| ADR-003 | LiteLLM come abstraction layer | Da scrivere |
| ADR-004 | Qdrant self-hosted | Da scrivere |
| [ADR-005](ADR-005-embedding-in-process.md) | Embedding model in-process nel servizio `api` | Accettato |
| [ADR-006](ADR-006-ingestion-job-via-profiles.md) | Job `ingestion` (Go) esposto via Compose profiles | Accettato |

> ADR-001…004 sono decisioni già prese (vedi CLAUDE.md §"Note editoriali")
> ma non ancora formalizzate. Diventano anche materiale editoriale.
