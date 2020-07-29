package main

import (
	"data-generator/myTemplates"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

//go:generate go run main-generate.go
func main() {
	allData := map[string]interface{}{
		"Admin": map[string]string{
			"AdminPassword": mustEnv("ADMIN_PASSWORD"),
			"UserPassword":  mustEnv("USER_PASSWORD"),
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
		mustEnv("ORACLE_PWD"),
		mustEnv("ORACLE_PDB"),
	)
	err := sqlPlus(initialConnectionString, func(stdin io.WriteCloser) {
		defer stdin.Close()
		if err := myTemplates.ExecuteIndividualTemplate(stdin, "initial.sql.tmpl", allData); err != nil {
			log.Fatalln(err)
		}
	})
	if err != nil {
		panic(err)
	}
	myAdminConnectionString := fmt.Sprintf("studentadmin/%s@//localhost:1521/%s",
		mustEnv("ADMIN_PASSWORD"),
		mustEnv("ORACLE_PDB"),
	)
	err = myTemplates.ExecuteIterableTemplates(func(currentTemplate *template.Template) error {
		log.Printf("Executing: %s", currentTemplate.ParseName)
		return sqlPlus(myAdminConnectionString, func(stdin io.WriteCloser) {
			defer stdin.Close()
			currentTemplate.Execute(stdin, allData)
		})
	})
	if err != nil {
		panic(err)
	}
	log.Println("Done!")
}

func sqlPlus(connection string, stream func(stdin io.WriteCloser)) error {
	log.Printf("Using connection string: %s", connection)
	command := exec.Command("sqlplus", connection)
	reader, writer := io.Pipe()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = io.TeeReader(reader, os.Stdout)
	if err := command.Start(); err != nil {
		return err
	}
	go stream(writer)
	return command.Wait()
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
