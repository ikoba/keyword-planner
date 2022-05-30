package keywordplanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

const (
	sleepAfterNavigate = time.Second
	sleepAfterClick    = time.Millisecond * 500
)

var (
	downloadButtonXpath string
	tempDir             string
	downloadDone        = make(chan string)
)

type Request struct {
	Words   []string
	OutDir  string
	Port    int
	Timeout int
	Retry   int
	Debug   bool
}

func (request *Request) Execute() error {
	url := fmt.Sprintf("ws://127.0.0.1:%d/", request.Port)
	allocatorCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), url)
	defer cancel()

	opts := []chromedp.ContextOption{}
	if request.Debug {
		opts = append(opts, chromedp.WithDebugf(log.Printf))
	}
	ctx, cancel := chromedp.NewContext(allocatorCtx, opts...)
	defer cancel()

	var screenWidth int
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.innerWidth;`, &screenWidth),
	)
	if err != nil {
		return fmt.Errorf("chromedp.Run failed: %w", err)
	}
	if screenWidth <= 1560 {
		downloadButtonXpath = `//material-menu[contains(@class, 'download')][2]/material-button`
	} else {
		downloadButtonXpath = `//material-menu[contains(@class, 'download')][1]/material-button`
	}

	chromedp.ListenTarget(ctx, func(v any) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			if ev.State == browser.DownloadProgressStateCompleted {
				downloadDone <- ev.GUID
			}
		}
	})

	tempDir = os.TempDir()
	err = chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(tempDir).
			WithEventsEnabled(true),
	)
	if err != nil {
		return fmt.Errorf("chromedp.Run failed: %w", err)
	}

	for _, word := range request.Words {
		attempts := request.Retry
		for {
			err = request.getKeywords(ctx, word)
			if err == nil {
				break
			}
			if err != context.DeadlineExceeded {
				return err
			}
			attempts--
			if attempts <= 0 {
				return errors.New("maximum number of retry attempts reached")
			}
		}
	}

	return nil
}

func (request *Request) getKeywords(ctx context.Context, word string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(request.Timeout)*time.Second)
	defer cancel()
	done := make(chan error)

	go func() {
		done <- request.getKeywordsSub(ctx, word)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (request *Request) getKeywordsSub(ctx context.Context, word string) error {
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://ads.google.com/aw/keywordplanner/home"),
		chromedp.Sleep(sleepAfterNavigate),
		chromedp.Click(`//div[contains(@class, 'ideas-card')]`, chromedp.NodeVisible),
		chromedp.Sleep(sleepAfterClick),
		chromedp.WaitVisible(`//div[contains(@class, 'input-container')]/input`),
		chromedp.SetValue(`//div[contains(@class, 'input-container')]/input`, word),
		chromedp.Click(`//div[contains(@class, 'submit-button-container')]/material-button/material-ripple`, chromedp.NodeVisible),
		chromedp.Sleep(sleepAfterClick),
	)
	if err != nil {
		return fmt.Errorf("chromedp.Run failed: %w", err)
	}

	var ariaDisabled string
	var ok bool
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(downloadButtonXpath),
		chromedp.AttributeValue(downloadButtonXpath, "aria-disabled", &ariaDisabled, &ok),
	)
	if err != nil {
		return fmt.Errorf("chromedp.Run failed: %w", err)
	}
	if !ok {
		return errors.New("aria-disabled attribute of download button cannot get")
	}
	if ariaDisabled == "true" {
		log.Printf("❎ %s has no keywords\n", word)
		return nil
	}

	err = chromedp.Run(ctx,
		chromedp.Click(downloadButtonXpath, chromedp.NodeVisible),
		chromedp.Sleep(sleepAfterClick),
		chromedp.Click(`//span[text()='.csv']/../..`, chromedp.NodeVisible),
	)
	if err != nil {
		return fmt.Errorf("chromedp.Run failed: %w", err)
	}

	guid, ok := <-downloadDone
	if !ok {
		return errors.New("cannot download csv")
	}
	srcPath := filepath.Join(tempDir, guid)
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", srcPath, err)
	}
	convertedWord := convertInvalidCharacters(word)
	dstPath := filepath.Join(request.OutDir, convertedWord+".csv")
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", dstPath, err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("cannot copy dst: %s, src %s", dstPath, srcPath)
	}

	log.Println(convertedWord + ".csv downloaded")

	return nil
}

var invalidCharacters = []rune{'\\', '/', ':', '*', '?', '"', '>', '<', '|'}

func convertInvalidCharacters(word string) string {
	var converted []rune
	r := []rune(word)
	matched := false
	for _, c := range r {
		matched = false
		for _, i := range invalidCharacters {
			if c == i {
				matched = true
				break
			}
		}
		if matched {
			converted = append(converted, '_')
		} else {
			converted = append(converted, c)
		}
	}
	return string(converted)
}
