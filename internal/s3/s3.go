package s3

import (
	"bufio"
	"errors"
	"net/http"
	"strings"
	"time"
)

func FetchVersion(url string) (string, error) {
	if strings.TrimSpace(url) == "" {
		return "", errors.New("url vazio")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("http status invalido")
	}
	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return "", errors.New("version vazio")
	}
	return strings.TrimSpace(scanner.Text()), nil
}
