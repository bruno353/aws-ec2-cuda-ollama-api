package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "path/filepath"
)

const (
    ollamaURL = "http://localhost:11434"
)

var apiKey string

func main() {
    apiKey = os.Getenv("API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable not set")
    }

    webroot := "/var/www/html"

    http.Handle("/.well-known/", http.StripPrefix("/.well-known/", http.FileServer(http.Dir(filepath.Join(webroot, ".well-known")))))

    // Handler principal
    http.HandleFunc("/v1/", handleProxy)

    fmt.Println("Server is running on :443")

    certFile := "/etc/letsencrypt/live/api.seudominio.com/fullchain.pem"
    keyFile := "/etc/letsencrypt/live/api.seudominio.com/privkey.pem"

    if _, err := os.Stat(certFile); os.IsNotExist(err) {
        log.Fatalf("Cert not found: %s", certFile)
    }
    if _, err := os.Stat(keyFile); os.IsNotExist(err) {
        log.Fatalf("Cert pk not found: %s", keyFile)
    }

    log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, nil))
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        log.Printf("Error RemoteAddr: %v", err)
        ip = r.RemoteAddr
    }
    log.Printf("Server called from IP: %s", ip)

    if !validateAPIKey(r) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    log.Printf("Received requisition: %s %s", r.Method, r.URL.Path)

    logRequest(r)

    target, err := url.Parse(ollamaURL)
    if err != nil {
        http.Error(w, "Error parsing Ollama URL", http.StatusInternalServerError)
        return
    }

    proxy := httputil.NewSingleHostReverseProxy(target)

    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
        req.Host = target.Host

        if req.Header.Get("Accept") == "text/event-stream" {
            w.Header().Set("Content-Type", "text/event-stream")
            w.Header().Set("Cache-Control", "no-cache")
            w.Header().Set("Connection", "keep-alive")
        }
    }

    proxy.Transport = &streamTransport{http.DefaultTransport}

    proxy.ServeHTTP(w, r)
}

func validateAPIKey(r *http.Request) bool {
    authHeader := r.Header.Get("Authorization")
    return authHeader == "Bearer "+apiKey
}

type streamTransport struct {
    http.RoundTripper
}

func (t *streamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    resp, err := t.RoundTripper.RoundTrip(req)
    if err != nil {
        return nil, err
    }

    if req.Header.Get("Accept") == "text/event-stream" {
        resp.Header.Set("Content-Type", "text/event-stream")
    }

    return resp, nil
}

func logRequest(r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Printf("Erro ao ler o corpo da requisição: %v", err)
        return
    }

    log.Printf("Corpo da Requisição: %s", string(body))

    r.Body = io.NopCloser(bytes.NewBuffer(body))
}
