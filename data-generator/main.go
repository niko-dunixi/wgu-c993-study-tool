package main

import (
	"data-generator/myTemplates"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

//go:generate go run main-generate.go
func main() {
	initialConnectionString := fmt.Sprintf("sys/%s@//localhost:1521/%s as sysdba",
		mustEnv("ORACLE_PWD"),
		mustEnv("ORACLE_PDB"),
	)
	log.Printf("Using connection string: %s", initialConnectionString)
	allData := map[string]interface{}{
		"Admin": map[string]string{
			"AdminPassword": mustEnv("ADMIN_PASSWORD"),
			"UserPassword":  mustEnv("USER_PASSWORD"),
		},
	}
	// Create initial administrative data
	err := sqlPlus(initialConnectionString, func(stdin io.WriteCloser) {
		defer stdin.Close()
		if err := myTemplates.ExecuteIndividualTemplate(stdin, "initial.sql.tmpl", allData); err != nil {
			log.Fatalln(err)
		}
	})
	if err != nil {
		panic(err)
	}
	log.Println("Done!")
}

func sqlPlus(connection string, stream func(stdin io.WriteCloser)) error {
	command := exec.Command("sqlplus", connection)
	reader, writer := io.Pipe()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = reader
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
