package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"

    server "github.com/liuliqiang/xds-demo/interface/http"
)

var (
    host       string
    port       int
    iden       string
    staticDir  = ""
    storageDir = ""
)

func init() {
    flag.StringVar(&host, "host", "0.0.0.0", "ping server listen interface ip")
    flag.IntVar(&port, "port", 9000, "ping server listen port")
    flag.StringVar(&iden, "iden", "111111", "identity for this ping http server")
    flag.StringVar(&staticDir, "static", staticDir, "admin static dir")
    flag.StringVar(&storageDir, "store", storageDir, "cache data storage dir")
    flag.Parse()
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    pingServer := server.NewPingService(iden)
    adminOpts := server.NewAdminServerOpts().
        WithStaticDir(staticDir).
        WithStorageDir(storageDir)
    adminServer, err := server.NewAdminServer(adminOpts)
    if err != nil {
        panic(fmt.Errorf("new admin server: %s", err))
    }

    http.HandleFunc("/ping/ping", pingServer.Ping)
    http.HandleFunc("/ping/change_health", pingServer.ChangeHealth)
    http.HandleFunc("/ping/reset_count", pingServer.ResetCount)
    http.HandleFunc("/ping/state", pingServer.State)

    http.HandleFunc("/admin/", adminServer.Index)
    http.HandleFunc("/admin/ws", adminServer.Websocket)

    addr := fmt.Sprintf("%s:%d", host, port)
    log.Println("http server ready to run at: " + addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
