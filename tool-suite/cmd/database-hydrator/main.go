package main

import (
	"fmt"
	"io"
	"log"
	"text/template"
	toolSuite "tool-suite"
	"tool-suite/myTemplates"
)

func main() {
	allData := map[string]interface{}{
		"Admin": map[string]string{
			"AdminPassword": toolSuite.MustEnv("ADMIN_PASSWORD"),
			"UserPassword":  toolSuite.MustEnv("USER_PASSWORD"),
		},
		"SimpleData": map[string]interface{}{
			"Items": []string{
				"Hello! I''m Paul!",
				"I love to dance, swing kettlebells, and the iterative process of being a badass!",
				"Don''t mind me, I''m just making really cool stuff :)",
			},
		},
	}
	// Create initial administrative data
	initialConnectionString := fmt.Sprintf("sys/%s@//localhost:1521/%s as sysdba",
		//initialConnectionString := fmt.Sprintf("pdbadmin/%s@//localhost:1521/%s",
		toolSuite.MustEnv("ORACLE_PWD"),
		toolSuite.MustEnv("ORACLE_PDB"),
	)
	err := toolSuite.SqlPlus(initialConnectionString, func(stdin io.WriteCloser) error {
		defer stdin.Close()
		return myTemplates.ExecuteIndividualTemplate(stdin, "initial.sql.tmpl", allData)
	})
	if err != nil {
		log.Fatalln(err)
	}
	myAdminConnectionString := fmt.Sprintf("studentadmin/%s@//localhost:1521/%s",
		toolSuite.MustEnv("ADMIN_PASSWORD"),
		toolSuite.MustEnv("ORACLE_PDB"),
	)
	err = myTemplates.ExecuteIterableTemplates(func(currentTemplate *template.Template) error {
		log.Printf("Executing: %s", currentTemplate.ParseName)
		return toolSuite.SqlPlus(myAdminConnectionString, func(stdin io.WriteCloser) error {
			defer stdin.Close()
			return currentTemplate.Execute(stdin, allData)
		})
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Done!")
}
