package racemine

import (
	"context"
	"time"

	"log"

	cdp "github.com/knq/chromedp"
)

const (
	baseURL   = "https://directors.racemine.com"
	loginURL  = baseURL + "/Account/Login"
	exportURL = baseURL + "/events/reports/18391#tab-entrant-export"
)

//Client is the mechanize client for Racemine
type Client struct {
	BaseURL   string
	LoginURL  string
	ExportURL string
	Options   []cdp.Option

	username string
	password string
}

//NewClient returns a new Racemine client
func NewClient(username, password string) *Client {
	c := &Client{
		BaseURL:   baseURL,
		LoginURL:  loginURL,
		ExportURL: exportURL,
		username:  username,
		password:  password,
	}
	c.Options = append(c.Options, cdp.WithErrorf(log.Printf))

	return c
}

//NewExport gets a recent simple export of all signed up runners
func (rc *Client) NewExport() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := cdp.New(ctx, rc.Options...)
	if err != nil {
		return err
	}

	// get latest download
	tasks := rc.login()
	tasks = append(tasks, rc.newExport())
	err = c.Run(ctx, tasks)
	if err != nil {
		return err
	}

	// shutdown chrome
	err = c.Shutdown(ctx)
	if err != nil {
		return err
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		return err
	}

	return nil
}

//GetAllExports returns links for all available exports
func (rc *Client) GetAllExports() ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := cdp.New(ctx, rc.Options...)
	if err != nil {
		return nil, err
	}

	// get latest download
	var attrs []map[string]string
	tasks := rc.login()
	tasks = append(tasks, rc.getAllDownloads(&attrs))
	err = c.Run(ctx, tasks)
	if err != nil {
		return nil, err
	}

	// shutdown chrome
	err = c.Shutdown(ctx)
	if err != nil {
		return nil, err
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		return nil, err
	}

	hrefs := make([]string, len(attrs))
	for i, attr := range attrs {
		hrefs[i] = attr["href"]
	}

	return hrefs, nil
}

func (rc *Client) login() cdp.Tasks {
	return cdp.Tasks{
		cdp.Navigate(rc.LoginURL),
		cdp.WaitVisible(`//input[@name="Email"]`),
		cdp.SendKeys(`//input[@name="Email"]`, rc.username),
		cdp.SendKeys(`//input[@name="Password"]`, rc.password),
		cdp.Submit(`//input[@name="Email"]`),
		cdp.Sleep(10 * time.Second),
	}
}

func (rc *Client) newExport() cdp.Tasks {
	return cdp.Tasks{
		cdp.Navigate(rc.ExportURL),
		cdp.WaitVisible(`#FromDate`),
		cdp.Submit(`//input[@name="FromDate"]`),
	}
}

func (rc *Client) getAllDownloads(attrs *[]map[string]string) cdp.Tasks {
	return cdp.Tasks{
		cdp.Navigate(rc.ExportURL),
		cdp.WaitVisible(`#FromDate`),
		cdp.AttributesAll(`//table[@id='items']//td/a/@href`, attrs, cdp.NodeVisible),
	}
}
