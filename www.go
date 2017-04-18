package hc

import (
    "fmt"
    "html/template"
    "net/http"
    "github.com/julienschmidt/httprouter"
    "log"
)

type Context map[string]interface{}

var templates = template.Must(template.ParseGlob("templates/*"))

func showChecks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    conn := pool.Get()
    defer conn.Close()

    ctx := Context{"Checks": checksByTeam()}
    w.Header().Set("Content-Type", "text/html")
    templates.ExecuteTemplate(w, "my_checks", ctx)
}

func addCheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    dsAddCheck()
    http.Redirect(w, r, "/", 302)
}

func Serve() {
    router := httprouter.New()
    router.GET("/", showChecks)
    router.POST("/checks/add/", addCheck)

    fmt.Println("Running on :8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}