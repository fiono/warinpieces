package views

import (
  "fmt"
  "log"
  "net/http"
  "text/template"
)

type TplRenderer struct {
  Tpl string
  Data interface{}
}

func tplPath(tplName string) string {
  return fmt.Sprintf("static/tpl/%s.tmpl", tplName)
}

func (tr *TplRenderer) RenderView(w http.ResponseWriter, r *http.Request) {
  allTpls := []string{tplPath("header"), tplPath(tr.Tpl), tplPath("footer")}

  t := template.Must(template.New(fmt.Sprintf("%s.tmpl", tr.Tpl)).ParseFiles(allTpls...))
  err := t.Execute(w, tr.Data)
  if err != nil {
    log.Println(err)
    fmt.Println(err)
  }
}
