// Package pdf prints self-contained CV HTML to searchable PDF through Chromium.
package pdf

import (
	"context"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Renderer struct{ Executable string }

func (r Renderer) Render(ctx context.Context, source string) ([]byte, error) {
	options := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.ExecPath(r.Executable), chromedp.NoSandbox, chromedp.DisableGPU, chromedp.NoFirstRun, chromedp.NoDefaultBrowserCheck, chromedp.Flag("disable-dev-shm-usage", true))
	allocator, closeAllocator := chromedp.NewExecAllocator(ctx, options...)
	defer closeAllocator()
	browser, closeBrowser := chromedp.NewContext(allocator)
	defer closeBrowser()
	timed, closeTimeout := context.WithTimeout(browser, 60*time.Second)
	defer closeTimeout()
	var output []byte
	err := chromedp.Run(timed, chromedp.Navigate("data:text/html;charset=utf-8,"+url.PathEscape(source)), chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		output, _, err = page.PrintToPDF().WithPrintBackground(true).WithPreferCSSPageSize(true).Do(ctx)
		return err
	}))
	return output, err
}
