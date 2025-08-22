package config

import (
  "flag"
  "os"
  "strconv"
)

type Config struct {
  Source  string
  Target  string
  Port    int
  Host    string
  UA      string
  Refresh int
}

func Parse() *Config {
  var c Config
  flag.StringVar(&c.Source, "source", "", "")
  flag.StringVar(&c.Target, "target", "", "")
  flag.IntVar(&c.Port, "port", 0, "")
  flag.StringVar(&c.Host, "host", "", "")
  flag.StringVar(&c.UA, "ua", "", "")
  flag.IntVar(&c.Refresh, "refresh", 0, "")
  flag.Parse()
  
  c.Source = getStr(&c.Source, "IPTV_SOURCE", "input.m3u")
  c.Target = getStr(&c.Target, "IPTV_TARGET", "iptv.m3u")
  c.Port = getInt(&c.Port, "WEB_PORT", 8080)
  c.Host = getStr(&c.Host, "WEB_HOST", "")
  c.UA = getStr(&c.UA, "IPTV_UA", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
  c.Refresh = getInt(&c.Refresh, "IPTV_REFRESH", 3600)
  return &c
}

func getStr(flag *string, env, def string) string {
  if flag != nil && *flag != "" {
    return *flag
  }
  if v := os.Getenv(env); v != "" {
    return v
  }
  return def
}
func getInt(flag *int, env string, def int) int {
  if flag != nil && *flag != 0 {
    return *flag
  }
  if v := os.Getenv(env); v != "" {
    if i, err := strconv.Atoi(v); err == nil {
      return i
    }
  }
  return def
}
