// I've tried both TNSPING and just plain old pinging the TCP port.
// The problem is that Oracle will start accepting remote connections
// way before it's actually done setting stuff up. The only way I've
// found to reliably work around this is to litterally look at the log
// and see the string "DATABASE IS READY TO USE!" which is not automation
// friendly. I've looked around and I can't find where it stores the
// logs locally, which leads me to believe they aren't. The container
// is just depending on docker for the log handling. This means we have
// to do some unorthodox stuff to get things to a happy place in terms
// of usability
package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)
import "github.com/docker/docker/client"

func main() {
	// Note: overriding previous HEALTHCHECK: [CMD-SHELL "$ORACLE_BASE/$CHECK_DB_FILE" >/dev/null || exit 1]
	// Preserving the original health-check. We want to augment, not overrule
	nativeHealthCheck()
	augmentedHealthCheck()
	log.Println("Oracle is ready")
}

func augmentedHealthCheck() {
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

func nativeHealthCheck() {
	oracleDirectory := mustEnv("ORACLE_BASE")
	oracleScriptName := mustEnv("CHECK_DB_FILE")
	fullFilename := path.Join(oracleDirectory, oracleScriptName)
	originalCommand := exec.Command(fullFilename)
	originalCommand.Stdin = os.Stdin
	originalCommand.Stdout = os.Stdout
	originalCommand.Stderr = os.Stderr
	if err := originalCommand.Run(); err != nil {
		log.Fatalln(err)
	}
}

func mustEnv(key string) (value string) {
	value, isPresent := os.LookupEnv(key)
	if !isPresent || value == "" {
		err := fmt.Errorf("there was no value for environment variable: '%s'", key)
		log.Fatalln(err)
	}
	return value
}
