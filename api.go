package main

import (
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
)

// regexes
var loginSuccessfulRegex = regexp.MustCompile("Angemeldet als:")
var pageCountRegex = regexp.MustCompile("([0-9]+) Seiten")
var hiddenFieldRegex = regexp.MustCompile("name=\"\\$Hidden\" value=\"([a-zA-Z0-9]+)\"")
var transferSuccessfulRegex = regexp.MustCompile("tragung war erfolgreich.")

// PapercutAPI object
type PapercutAPI struct {
	client *WebClient
}

func CreatePapercutAPI(user string, pass string, webClient *WebClient) *PapercutAPI {
	pc := new(PapercutAPI)
	pc.client = webClient
	pc.getSession()
	if !pc.loginUser(user, pass) {
		panic("Could not login into PaperCut")
	}
	return pc
}

func (pc *PapercutAPI) getSession() {
	resp := pc.client.Get("https://print.informatik.tu-darmstadt.de/")
	defer resp.Body.Close()
}

func (pc *PapercutAPI) loginUser(user string, pass string) bool {
	resp := pc.client.PostForm("https://print.informatik.tu-darmstadt.de/app", url.Values{
		"service":              {"direct/1/Home/$Form$0"},
		"sp":                   {"S0"},
		"Form0":                {"$Hidden$0,$Hidden$1,inputUsername,inputPassword,$PropertySelection$0,$Submit$0"},
		"$Hidden$0":            {"true"},
		"$Hidden$1":            {"X"},
		"inputUsername":        {user},
		"inputPassword":        {pass},
		"$PropertySelection$0": {"de"},
		"$Submit$0":            {"Anmelden"},
	})
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	ok := loginSuccessfulRegex.Match(body)

	return ok
}

func (pc *PapercutAPI) GetPagesLeft() int {
	resp := pc.client.Get("https://print.informatik.tu-darmstadt.de/app?service=page/UserSummary")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	matches := pageCountRegex.FindSubmatch(body)
	countStr := string(matches[1])

	count64, _ := strconv.ParseInt(countStr, 10, 32)
	count := int(count64)

	return count
}

func (pc *PapercutAPI) TranferPages(receiver string, amount int, comment string) bool {
	// get CSRF token
	resp1 := pc.client.Get("https://print.informatik.tu-darmstadt.de/app?service=page/UserTransfer")
	defer resp1.Body.Close()
	body, _ := ioutil.ReadAll(resp1.Body)
	matches := hiddenFieldRegex.FindSubmatch(body)
	csrfToken := string(matches[1])

	amountStr := strconv.Itoa(amount) + " Seiten"

	// post transfer request
	resp2 := pc.client.PostForm("https://print.informatik.tu-darmstadt.de/app", url.Values{
		"service":         {"direct/1/UserTransfer/transferForm"},
		"sp":              {"S0"},
		"Form0":           {"$Hidden,inputAmount,inputToUsername,inputComment,$Submit"},
		"$Hidden":         {csrfToken},
		"inputAmount":     {amountStr},
		"inputToUsername": {receiver},
		"inputComment":    {comment},
		"$Submit":         {"Übertragung"},
	})
	defer resp2.Body.Close()

	body2, _ := ioutil.ReadAll(resp2.Body)
	ok := transferSuccessfulRegex.Match(body2)

	return ok
}
