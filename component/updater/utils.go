package updater

import (
	"context"
	"io"
	"os"
	"time"

	clashHttp "github.com/metacubex/clash/component/http"

	"github.com/metacubex/http"
)

const defaultHttpTimeout = time.Second * 90

func downloadForBytes(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultHttpTimeout)
	defer cancel()
	resp, err := clashHttp.HttpRequest(ctx, url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func saveFile(bytes []byte, path string) error {
	return os.WriteFile(path, bytes, 0o644)
}
