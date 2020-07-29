package myTemplates

import (
	"github.com/paul-nelson-baker/randomstd"
	"github.com/shurcooL/httpfs/text/vfstemplate"
	"io"
	"strings"
	"text/template"
)

var (
	myRandom = randomstd.NaiveConstructor()
)

func ExecuteIndividualTemplate(wr io.Writer, name string, data interface{}) error {
	return IndividualTemplates().ExecuteTemplate(wr, name, data)
}

func ExecuteIterableTemplates(handler func(currentTemplate *template.Template) error) error {
	for _, currentTemplate := range IterableTemplates().Templates() {
		if err := handler(currentTemplate); err != nil {
			return err
		}
	}
	return nil
}

func IndividualTemplates() *template.Template {
	parentTemplate := newTemplate()
	templates, err := vfstemplate.ParseGlob(assets, parentTemplate, "individual/*.sql.tmpl")
	if err != nil {
		panic(err)
	}
	return templates
}

func IterableTemplates() *template.Template {
	parentTemplate := newTemplate()
	templates, err := vfstemplate.ParseGlob(assets, parentTemplate, "iterable/*.sql.tmpl")
	if err != nil {
		panic(err)
	}
	return templates
}

func newTemplate() *template.Template {
	parentTemplate := template.New("").Funcs(template.FuncMap{
		"RandomPassword": randomPassword,
	})
	return parentTemplate
}

func randomPassword() string {
	builder := strings.Builder{}
	for i := 0; i < 30; i++ {
		nextRune := rune(myRandom.Intn('z'-'a') + 'a')
		builder.WriteRune(nextRune)
	}
	return builder.String()
}
