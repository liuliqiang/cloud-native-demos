package main

import (
    "io/ioutil"
    "net/http"
)

func main() {
    resp, err := http.Get("https://liqiang.io")
    if err != nil {
        panic(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    println(string(body))
}
