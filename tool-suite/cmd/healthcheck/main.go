// I've tried both TNSPING and just plain old pinging the TCP port.
// The problem is that Oracle will start accepting remote connections
// way before it's actually done setting stuff up. The only way I've
// found to reliably work around this is to literally look at the log
// and see the string "DATABASE IS READY TO USE!" which is not automation
// friendly. I've looked around and I can't find where it stores the
// logs locally, which leads me to believe they aren't. The container
// is just depending on docker for the log handling. This means we have
// to do some unorthodox stuff to get things to a happy place in terms
// of usability
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	toolSuite "tool-suite"
)

func main() {
	// Note: overriding previous HEALTHCHECK: [CMD-SHELL "$ORACLE_BASE/$CHECK_DB_FILE" >/dev/null || exit 1]
	// Preserving the original health-check. We want to augment, not overrule
	toolSuite.NativeHealthCheck()
	toolSuite.AugmentedHealthCheck()
	log.Println("Oracle is ready")
	if performUserCheck := shouldCheckUserReady(); !performUserCheck {
		return
	}

}

func shouldCheckUserReady() bool {
	for _, arg := range os.Args {
		if arg == "--check-user-ready" {
			return true
		}
	}
	return false
}

func checkUserIsReady() {
	myUserConnectionString := fmt.Sprintf("student/%s@//localhost:1521/%s",
		toolSuite.MustEnv("USER_PASSWORD"),
		toolSuite.MustEnv("ORACLE_PDB"),
	)
	completionChannel := make(chan error, 1)
	go func() {
		err := toolSuite.SqlPlus(myUserConnectionString, func(stdin io.WriteCloser) error {
			defer stdin.Close()
			_, err := stdin.Write([]byte("select * from dual;"))
			return err
		})
		completionChannel <- err
	}()
	select {
	case sqlErr := <-completionChannel:
		if sqlErr != nil {
			log.Fatalln(sqlErr)
		}
	case <-time.After(5 * time.Second):
		log.Fatalln("Operation timed out, assuming database is not ready")
	}
	log.Printf("Oracle appears ready to use connection string: %s", myUserConnectionString)
}
