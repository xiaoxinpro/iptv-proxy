package main

import (
  "context"
  "fmt"
  "iptv-proxy/src/config"
  "iptv-proxy/src/loader"
  "iptv-proxy/src/proxy"
  "iptv-proxy/src/utils"
  "log"
  "net/http"
  "os"
  "os/signal"
  "syscall"
)

func main() {
  cfg := config.Parse()
  if cfg.Source == "" {
    log.Fatal("source cannot be empty")
  }
  if cfg.Host == "" {
    ip, err := utils.GetLANIP()
    if err != nil {
      log.Fatalf("get LAN ip: %v", err)
    }
    cfg.Host = ip
  }
  l := loader.New(cfg.Source, cfg.Refresh)
  ctx, cancel := context.WithCancel(context.Background())
  go l.Start(ctx)
  
  p := proxy.New(cfg.Host, cfg.Port, cfg.UA, l)
  mux := http.NewServeMux()
  mux.HandleFunc("/"+cfg.Target, p.ServeM3U)
  mux.HandleFunc("/iptv/", p.ServeHTTP)
  
  addr := fmt.Sprintf(":%d", cfg.Port)
  srv := &http.Server{Addr: addr, Handler: mux}
  go func() {
    log.Printf("load m3u: %s", cfg.Source)
    log.Printf("listening on http://%s:%d", cfg.Host, cfg.Port)
    log.Printf("output m3u: http://%s:%d/%s", cfg.Host, cfg.Port, cfg.Target)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
      log.Fatalf("listen: %v", err)
    }
  }()
  
  sig := make(chan os.Signal, 1)
  signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
  <-sig
  cancel()
  srv.Shutdown(ctx)
}
