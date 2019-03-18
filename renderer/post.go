package renderer

import (
	"html/template"
	"os"

	"hawx.me/code/tally-ho/config"
)

func New(conf config.Config, glob string) (*Renderer, error) {
	tmpls, err := template.ParseGlob(glob)

	return &Renderer{conf: conf, tmpls: tmpls}, err
}

type Renderer struct {
	conf  config.Config
	tmpls *template.Template
}

func (r *Renderer) RenderPost(id string, properties map[string][]interface{}) error {
	file, err := os.Create(r.conf.PostPath(id))
	if err != nil {
		return err
	}

	return r.tmpls.ExecuteTemplate(file, "post.gotmpl", properties)
}
