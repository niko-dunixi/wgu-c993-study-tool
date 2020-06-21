package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	mustAgreeTos()
	oracleUsername := mustEnv("ORACLE_USERNAME")
	oraclePassword := mustEnv("ORACLE_PASSWORD")

	downloadCompleteChannel := make(chan struct{})

	//opts := append(chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.Flag("headless", false),
	//	chromedp.Flag("hide-scrollbars", false),
	//)
	opts := chromedp.DefaultExecAllocatorOptions[:]
	allocator, allocatorCancelFunction := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocatorCancelFunction()
	currentContext, cancelFunction := chromedp.NewContext(
		allocator,
		notifyWhenDownloadComplete(downloadCompleteChannel),
	)
	defer cancelFunction()
	err := chromedp.Run(currentContext,
		page.Enable(),
		page.SetDownloadBehavior(page.SetDownloadBehaviorBehaviorAllow).WithDownloadPath("."),
		// Click through the download page, this sets some specific cookies
		chromedp.Navigate("https://www.oracle.com/database/technologies/oracle12c-linux-12201-downloads.html"),
		chromedp.Click(`[data-file="//download.oracle.com/otn/linux/oracle12c/122010/linuxx64_12201_database.zip"]`, chromedp.ByQuery),
		//chromedp.Click(`input[name="licenseAccept"]:visible`, chromedp.ByQuery),
		chromedp.Click(`#w11 > div > div.w11w2.lbdefault > div > div > div > form > ul > li > label > input[type=checkbox]`, chromedp.ByQuery),
		chromedp.Click(`#w11 > div > div.w11w2.lbdefault > div > div > div > form > div > div.oform-bttns.center-bttns > div > div > a`, chromedp.ByQuery),
		chromedp.Sleep(10*time.Second),
		// Perform the sign-in which sets more cookies
		chromedp.Click(`#sso_username`, chromedp.ByQuery),
		chromedp.SendKeys(`#sso_username`, oracleUsername, chromedp.ByQuery),
		chromedp.Click(`#ssopassword`, chromedp.ByQuery),
		chromedp.SendKeys(`#ssopassword`, oraclePassword, chromedp.ByQuery),
		chromedp.Click(`#signin_button`, chromedp.ByQuery),
	)
	if err != nil {
		panic(err)
	}
	for {
		select {
		case _ = <-downloadCompleteChannel:
			return
		default:
			err := chromedp.Run(currentContext, chromedp.Sleep(time.Second*10))
			if err != nil {
				panic(err)
			}
		}
	}
}

func notifyWhenDownloadComplete(channel chan<- struct{}) chromedp.ContextOption {
	return chromedp.WithDebugf(func(s string, rawMessages ...interface{}) {
		go func() {
			for _, rawMessage := range rawMessages {
				var message cdproto.Message
				err := json.Unmarshal([]byte(fmt.Sprintf("%s", rawMessage)), &message)
				if err != nil {
					continue
				}
				isDownloadProgress := message.Method == cdproto.EventPageDownloadProgress
				if !isDownloadProgress {
					return
				}
				eventDownloadProgress := page.EventDownloadProgress{}
				if err != json.Unmarshal(message.Params, &eventDownloadProgress) {
					return
				}
				completed := eventDownloadProgress.State == page.DownloadProgressStateCompleted
				canceled := eventDownloadProgress.State == page.DownloadProgressStateCanceled
				if completed || canceled {
					channel <- struct{}{}
				}
				percentage := eventDownloadProgress.ReceivedBytes / eventDownloadProgress.TotalBytes
				log.Printf("Oracle percent downloaded: %f", percentage)
			}
		}()
	})
}

func mustEnv(key string) string {
	value, isPresent := os.LookupEnv(key)
	if !isPresent || value == "" {
		err := fmt.Errorf("no environment variable for: %s", key)
		panic(err)
	}
	return value
}

func mustAgreeTos() {
	answer := mustEnv("ORACLE_AGREE_TO_TERMS_OF_SERVICE")
	if strings.ToLower(answer) != "i agree" {
		err := fmt.Errorf("ORACLE_AGREE_TO_TERMS_OF_SERVICE environment variable must be equal to 'I AGREE'")
		panic(err)
	}
}
