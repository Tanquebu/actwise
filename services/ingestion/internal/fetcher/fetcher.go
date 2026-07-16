package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Result struct {
	Lang   string
	SHA256 string
	Bytes  int64
}

func Fetch(
	ctx context.Context,
	lang string,
	url string,
	destDir string,
) (Result, error) {
	result := Result{Lang: lang}
	slog.Info("hello 2 from fetcher", "lang", lang, "url", url)
	client := http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("Errore in linea 36", "error", err)
		return result, err
	}
	req.Header.Set("User-Agent", "ActWise-Ingestion/1.0")
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Errore in linea 37", "error", err)
		return result, err
	}

	slog.Info("hello 4 from fetcher", "lang", lang, "url", url, "status", resp.StatusCode)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/pdf" {
		return result, fmt.Errorf("unexpected content type: %s", resp.Header.Get("Content-Type"))
	}

	tempFile, err := os.CreateTemp(destDir, fmt.Sprintf("%s_*.pdf.tmp", lang))
	if err != nil {
		slog.Error("Errore in linea 46", "error", err)
		return result, err
	}

	h := sha256.New()
	reader := io.TeeReader(resp.Body, h)

	result.Bytes, err = io.Copy(tempFile, reader)

	if err != nil {
		slog.Error("Errore in linea 58", "error", err)
		return result, err
	}

	if err := tempFile.Close(); err != nil {
		return result, err
	}

	err = os.Rename(tempFile.Name(), fmt.Sprintf("%s/%s.pdf", destDir, lang))
	if err != nil {
		slog.Error("Errore in linea 75", "error", err)
		return result, err
	}

	result.SHA256 = hex.EncodeToString(h.Sum(nil))

	return result, nil

}
