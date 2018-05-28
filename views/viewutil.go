package views

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"
)

type TplRenderer struct {
	Tpl   string
	Data  interface{}
	IsWeb bool
}

func tplPath(tplName string) string {
	return fmt.Sprintf("static/tpl/%s.tmpl", tplName)
}

func (tr *TplRenderer) RenderView(w io.Writer) error {
	allTpls := []string{tplPath(tr.Tpl)}
	if tr.IsWeb {
		allTpls = append(allTpls, tplPath("header"))
		allTpls = append(allTpls, tplPath("footer"))
	}
	t := template.Must(template.New(fmt.Sprintf("%s.tmpl", tr.Tpl)).ParseFiles(allTpls...))
	return t.Execute(w, tr.Data)
}

func (tr *TplRenderer) ServeView(w http.ResponseWriter, r *http.Request) {
	err := tr.RenderView(w)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
}

func (tr *TplRenderer) GetView() (html string, err error) {
	buf := new(bytes.Buffer)
	tr.RenderView(buf)
	return buf.String(), err
}
