package loader

import (
  "bufio"
  "context"
  "crypto/sha1"
  "encoding/hex"
  "io"
  "log"
  "net/http"
  "os"
  "strings"
  "sync"
  "time"
)

type Loader struct {
  urlMap   map[string]string
  mu       sync.RWMutex
  source   string
  refresh  int
  lastLoad time.Time
}

func New(source string, refresh int) *Loader {
  return &Loader{
    urlMap:  make(map[string]string),
    source:  source,
    refresh: refresh,
  }
}

func (l *Loader) Start(ctx context.Context) {
  l.load()
  t := time.NewTicker(time.Duration(l.refresh) * time.Second)
  defer t.Stop()
  for {
    select {
    case <-t.C:
      if err := l.load(); err != nil {
        log.Printf("[loader] reload err: %v", err)
      }
    case <-ctx.Done():
      return
    }
  }
}

func (l *Loader) load() error {
  var rc io.ReadCloser
  var err error
  if strings.HasPrefix(l.source, "http") {
    resp, err := http.Get(l.source)
    if err != nil {
      return err
    }
    rc = resp.Body
  } else {
    rc, err = os.Open(l.source)
    if err != nil {
      return err
    }
  }
  defer rc.Close()
  
  newMap := make(map[string]string)
  sc := bufio.NewScanner(rc)
  for sc.Scan() {
    line := strings.TrimSpace(sc.Text())
    if strings.HasPrefix(line, "http") {
      newMap[makeToken(line)] = line
    }
  }
  l.mu.Lock()
  l.urlMap = newMap
  l.lastLoad = time.Now()
  l.mu.Unlock()
  return nil
}

func (l *Loader) Get(token string) (string, bool) {
  l.mu.RLock()
  defer l.mu.RUnlock()
  t, ok := l.urlMap[token]
  return t, ok
}

// ------------- 导出方法 -------------
func (l *Loader) GetSource() string {
  return l.source
}

func (l *Loader) MakeToken(raw string) string {
  return makeToken(raw)
}

// ------------- 内部工具 -------------
func makeToken(raw string) string {
  h := sha1.Sum([]byte(raw))
  return hex.EncodeToString(h[:8])
}
