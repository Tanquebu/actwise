package main

import (
	"context"
	"log/slog"

	"github.com/Tanquebu/actwise/services/ingestion/internal/fetcher"
)

func main() {
	ctx := context.Background()
	sources := []struct{ Lang, URL string }{
		{"en", "https://eur-lex.europa.eu/legal-content/EN/TXT/PDF/?uri=OJ:L_202401689"},
		{"it", "https://eur-lex.europa.eu/legal-content/IT/TXT/PDF/?uri=OJ:L_202401689"},
	}
	for _, s := range sources {
		fetcher.Fetch(ctx, s.Lang, s.URL, "/tmp")
	}
	slog.Info("hello main.go")
}
