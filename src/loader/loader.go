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
  lastAccess time.Time
}

func New(source string, refresh int) *Loader {
  return &Loader{
    urlMap:  make(map[string]string),
    source:  source,
    refresh: refresh,
  }
}

func (l *Loader) Start(ctx context.Context) {
  l.load() // 初始加载
  <-ctx.Done()
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
  // 检查是否需要刷新
  if l.shouldRefresh() {
    // 需要释放读锁，获取写锁进行刷新
    l.mu.RUnlock()
    l.mu.Lock()
    // 双重检查，防止并发情况下重复刷新
    if l.shouldRefresh() {
      if err := l.load(); err != nil {
        log.Printf("[loader] refresh on access failed: %v", err)
      }
    }
    l.mu.Unlock()
    l.mu.RLock()
  }
  
  defer l.mu.RUnlock()
  t, ok := l.urlMap[token]
  l.lastAccess = time.Now() // 更新访问时间
  return t, ok
}

func (l *Loader) shouldRefresh() bool {
  if l.refresh <= 0 {
    return false
  }
  return time.Since(l.lastLoad) > time.Duration(l.refresh)*time.Second
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
