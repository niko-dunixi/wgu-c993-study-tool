package main

import (
	"fmt"
	"github.com/paul-nelson-baker/randomstd"
	"os"
	"sort"
	"strings"
	"text/template"
)

var random = randomstd.NaiveConstructor()

func main() {
	namedTemplates, err := templatesByName()
	if err != nil {
		panic(err)
	}
	data := map[string]interface{}{
		"AdminData": map[string]interface{}{
			"PrimarySID":     mustEnv("ORACLE_SID"),
			"PrimaryPDB":     mustEnv("ORACLE_PDB"),
			"OraclePassword": mustEnv("ORACLE_PASSWORD"),
			"UserPassword":   mustEnv("USER_PASSWORD"),
			"AdminPassword":  mustEnv("ADMIN_PASSWORD"),
		},
		"SimpleData": map[string]interface{}{
			"Password": "superrandom132;",
			"Items": []string{
				"Hi, I''m Paul",
				"I like to dance, lift kettlebells, and be a badass",
				"My most recent interest is mycology!",
			},
		},
	}
	for _, name := range namedTemplates.Names {
		fmt.Printf("-- FILE: %s\n", name)
		err := namedTemplates.Templates[name].Execute(os.Stdout, data)
		if err != nil {
			panic(err)
		}
		fmt.Println()
	}
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	value = strings.TrimSpace(value)
	if value == "" {
		err := fmt.Errorf("key '%s' had no value", key)
		panic(err)
	}
	return value
}

type namedTemplates struct {
	Names     []string
	Templates map[string]*template.Template
}

func templatesByName() (namedTemplates namedTemplates, err error) {
	templates, err := template.New("").Funcs(template.FuncMap{
		"RandomPassword": RandomPassword,
	}).ParseGlob("./templates/*.sql.tmpl")
	if err != nil {
		return
	}
	namedTemplates.Templates = map[string]*template.Template{}
	for _, currentTemplate := range templates.Templates() {
		currentName := currentTemplate.ParseName
		namedTemplates.Names = append(namedTemplates.Names, currentName)
		namedTemplates.Templates[currentName] = currentTemplate
	}
	sort.Strings(namedTemplates.Names)
	return
}

func RandomPassword() string {
	builder := strings.Builder{}
	for i := 0; i < 30; i++ {
		letter := random.Intn('z'-'a') + 'a'
		builder.WriteRune(rune(letter))
	}
	return builder.String()
}
