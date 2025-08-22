package utils

import (
  "net"
  "strings"
)

func GetLANIP() (string, error) {
  conn, err := net.Dial("udp", "8.8.8.8:80")
  if err != nil {
    return "", err
  }
  defer conn.Close()
  return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}
