package toolSuite

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

//go:generate go run main-generate.go

func MustEnv(key string) (value string) {
	value, isPresent := os.LookupEnv(key)
	if !isPresent || value == "" {
		err := fmt.Errorf("there was no value for environment variable: '%s'", key)
		log.Fatalln(err)
	}
	return value
}

func NativeHealthCheck() {
	oracleDirectory := MustEnv("ORACLE_BASE")
	oracleScriptName := MustEnv("CHECK_DB_FILE")
	fullFilename := path.Join(oracleDirectory, oracleScriptName)
	originalCommand := exec.Command(fullFilename)
	originalCommand.Stdin = os.Stdin
	originalCommand.Stdout = os.Stdout
	originalCommand.Stderr = os.Stderr
	if err := originalCommand.Run(); err != nil {
		log.Fatalln(err)
	}
}

func AugmentedHealthCheck() {
	fixUnixSocketPermissions()
	// Create a new docker client and poll the logs. If they
	// have our magic string, we're in good shape.
	ctx := context.Background()
	d, err := client.NewEnvClient()
	if err != nil {
		log.Fatalln(err)
	}
	logStream, err := d.ContainerLogs(ctx, "my-oracle", types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Follow:     false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer logStream.Close()
	bufferedLogStream := bufio.NewReader(logStream)
	for {
		currentLine, err := bufferedLogStream.ReadString('\n')
		if err != nil {
			log.Fatalf("oracle is not yet ready to use: %v", err)
		}
		if strings.Contains(currentLine, "DATABASE IS READY TO USE!") {
			return
		}
	}
}

func fixUnixSocketPermissions() {
	// If we don't own the unix socket as the Oracle user, fix it so we can access it
	// - https://github.com/moby/moby/issues/2259
	unixSocketPath := "/var/run/docker.sock"
	fileInfo, err := os.Stat(unixSocketPath)
	if err != nil {
		log.Fatalln(err)
	}
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); !ok {
		log.Fatalf("Could not cast filesystem info to datastructure: %+v", fileInfo.Sys())
	} else {
		if unixSocketUID := int(stat.Uid); os.Getuid() != unixSocketUID {
			chownCommand := exec.Command("sudo", "chown", "oracle", unixSocketPath)
			chownCommand.Stdin = os.Stdin
			chownCommand.Stdout = os.Stdout
			chownCommand.Stderr = os.Stderr
			err = chownCommand.Run()
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func SqlPlus(connection string, stream func(stdin io.WriteCloser) error) error {
	log.Printf("Using connection string: %s", connection)
	command := exec.Command("sqlplus", connection)
	reader, writer := io.Pipe()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = io.TeeReader(reader, os.Stdout)
	if err := command.Start(); err != nil {
		return err
	}
	errorChannel := make(chan error, 2)
	defer close(errorChannel)
	go func() {
		errorChannel <- stream(writer)
	}()
	go func() {
		errorChannel <- command.Wait()
	}()
	for i := 0; i < 2; i++ {
		if err := <-errorChannel; err != nil {
			return err
		}
	}
	return nil
}
