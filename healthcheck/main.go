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
	"context"
	"github.com/docker/docker/api/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)
import "github.com/docker/docker/client"

func main() {
	// Note: overriding previous HEALTHCHECK: [CMD-SHELL "$ORACLE_BASE/$CHECK_DB_FILE" >/dev/null || exit 1]
	// Preserving the original health-check. We want to augment, not overrule

	// chown the unix socket because we can't configure the
	// volume to mount with a specific user.
	// - https://github.com/moby/moby/issues/2259
	chownCommand := exec.Command("sudo", "chown", "oracle", "/var/run/docker.sock")
	chownCommand.Stdin = os.Stdin
	chownCommand.Stdout = os.Stdout
	chownCommand.Stderr = os.Stderr
	err := chownCommand.Run()
	if err != nil {
		log.Fatalln(err)
	}
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
	logBytes, err := ioutil.ReadAll(logStream)
	if err != nil {
		log.Fatalln(err)
	}
	logString := string(logBytes)
	if !strings.Contains(logString, "DATABASE IS READY TO USE!") {
		log.Fatalln("oracle is not yet ready to use")
	}
	// I _could_ create a true database connection, but the driver isn't
	// trivial to implement. I'm going to hedge around it for the time
	// being by opening a connection via sqlplus instead
	log.Println("Oracle is ready")
}
