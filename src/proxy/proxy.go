package proxy

import (
  "bufio"
  "fmt"
  "io"
  "iptv-proxy/src/loader"
  "log"
  "net/http"
  "net/http/httputil"
  "net/url"
  "os"
  "strings"
)

type Proxy struct {
  host   string
  port   int
  ua     string
  loader *loader.Loader
}

func New(host string, port int, ua string, loader *loader.Loader) *Proxy {
  return &Proxy{host, port, ua, loader}
}

// ServeHTTP 处理 /iptv/{token} 反向代理
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  token := strings.TrimPrefix(r.URL.Path, "/iptv/")
  log.Printf("[proxy] %s %s token=%s", r.Method, r.URL.Path, token)
  
  target, ok := p.loader.Get(token)
  if !ok {
    log.Printf("[proxy] token=%s not found", token)
    http.NotFound(w, r)
    return
  }
  
  targetURL, err := url.Parse(target)
  if err != nil {
    log.Printf("[proxy] parse upstream url fail: %v", err)
    http.Error(w, "invalid upstream url", http.StatusInternalServerError)
    return
  }
  
  // 记录目标URL信息
  log.Printf("[proxy] target url: %s", targetURL.String())
  
  proxy := &httputil.ReverseProxy{
    Director: func(req *http.Request) {
      // 完全替换请求的URL为targetURL，忽略原始请求路径
      req.URL = targetURL
      req.Host = targetURL.Host
      if p.ua != "" {
        req.Header.Set("User-Agent", p.ua)
      }
      log.Printf("[proxy] forwarding to: %s", req.URL.String())
    },
  }
  
  proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
    log.Printf("[proxy] upstream error: %v", err)
    http.Error(rw, "upstream unavailable", http.StatusBadGateway)
  }
  
  // 添加ModifyResponse来记录响应信息
  proxy.ModifyResponse = func(resp *http.Response) error {
    log.Printf("[proxy] response status: %s", resp.Status)
    log.Printf("[proxy] response content-type: %s", resp.Header.Get("Content-Type"))
    return nil
  }
  
  proxy.ServeHTTP(w, r)
}

// ServeM3U 提供 /xxx.m3u 下载
func (p *Proxy) ServeM3U(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
  
  rc, err := openSource(p.loader.GetSource())
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer rc.Close()
  
  sc := bufio.NewScanner(rc)
  for sc.Scan() {
    line := strings.TrimSpace(sc.Text())
    if strings.HasPrefix(line, "http") {
      token := p.loader.MakeToken(line)
      line = fmt.Sprintf("http://%s:%d/iptv/%s", p.host, p.port, token)
    }
    fmt.Fprintln(w, line)
  }
  if err := sc.Err(); err != nil {
    log.Printf("[proxy] scan m3u error: %v", err)
  }
}

// openSource 统一打开本地或远程文件
func openSource(src string) (io.ReadCloser, error) {
  if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
    resp, err := http.Get(src)
    if err != nil {
      return nil, err
    }
    return resp.Body, nil
  }
  return os.Open(src)
}
