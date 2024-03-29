package papercut

import (
	"errors"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// DefaultSSHHost is the ssh client login node
const DefaultSSHHost = "clientssh3.rbg.informatik.tu-darmstadt.de:22"

// DefaultSSHHostKey is clientssh3s host key
const DefaultSSHHostKey = "AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBPkohFCPX0tmUqK+w3GWX3B+oGcVPGabuQ9nOZ0bjD37DyIva5qZ3QSr9HfJrz7D1GPIpVONQ8MWEaROisQaYNU="

const apiBase = "https://print.informatik.tu-darmstadt.de"

// regexes
var loginSuccessfulRegex = regexp.MustCompile("<h4>Kontostand</h4>")
var pageCountRegex = regexp.MustCompile("([0-9]+) Seiten")
var hiddenFieldRegex = regexp.MustCompile("name=\"\\$Hidden\" value=\"([a-zA-Z0-9]+)\"")
var transferSuccessfulRegex = regexp.MustCompile("tragung war erfolgreich.")

// PapercutAPI is a papercut api session
type PapercutAPI struct {
	webClient *WebClient
	sshClient *ssh.Client // Only set if created and owned by this instance.
}

// CreatePapercutAPIWithWebClient creates a new papercut api object with a given webclient and logs in to papercut
func CreatePapercutAPIWithWebClient(user string, pass string, webClient *WebClient) *PapercutAPI {
	pc := new(PapercutAPI)
	pc.webClient = webClient
	pc.getSession()
	if !pc.loginUser(user, pass) {
		panic("Could not login into PaperCut")
	}
	return pc
}

// CreatePapercutAPI creates a new papercut api object and logs in to papercut using a new webclient
func CreatePapercutAPI(config *ConfigContainer) (*PapercutAPI, error) {
	ssh, err := CreateSSHClient(DefaultSSHHost, DefaultSSHHostKey, config.SSHUser, config.SSHKeyFile)
	if err != nil {
		return nil, err
	}

	pc := new(PapercutAPI)
	pc.webClient = CreateTunneledWebClient(ssh)
	pc.sshClient = ssh
	pc.getSession()
	if !pc.loginUser(config.PaperCutUsername, config.PaperCutPassword) {
		pc.Close()
		return nil, errors.New("Could not login into PaperCut")
	}
	return pc, nil
}

func (pc *PapercutAPI) getSession() {
	resp := pc.webClient.Get(apiBase)
	defer resp.Body.Close()
}

func (pc *PapercutAPI) loginUser(user string, pass string) bool {
	resp := pc.webClient.PostForm(apiBase+"/app", url.Values{
		"service":              {"direct/1/Home/$Form"},
		"sp":                   {"S0"},
		"Form0":                {"$Hidden$0,$Hidden$1,inputUsername,inputPassword,$Submit$0,$PropertySelection"},
		"$Hidden$0":            {"true"},
		"$Hidden$1":            {"X"},
		"inputUsername":        {user},
		"inputPassword":        {pass},
		"$PropertySelection":   {"de"},
		"$Submit$0":            {"Anmelden"},
	}, apiBase)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	ok := loginSuccessfulRegex.Match(body)

	return ok
}

// GetPagesLeft returns number of papercut page credits for the current user
func (pc *PapercutAPI) GetPagesLeft() int {
	resp := pc.webClient.Get(apiBase + "/app?service=page/UserSummary")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	matches := pageCountRegex.FindSubmatch(body)
	countStr := string(matches[1])

	count64, _ := strconv.ParseInt(countStr, 10, 32)
	count := int(count64)

	return count
}

// TransferPages transfers a given amount of page credits to another user
func (pc *PapercutAPI) TransferPages(receiver string, amount int, comment string) bool {
	// get CSRF token
	resp1 := pc.webClient.Get(apiBase + "/app?service=page/UserTransfer")
	defer resp1.Body.Close()
	body, _ := ioutil.ReadAll(resp1.Body)
	matches := hiddenFieldRegex.FindSubmatch(body)
	csrfToken := string(matches[1])

	amountStr := strconv.Itoa(amount) + " Seiten"

	// post transfer request
	resp2 := pc.webClient.PostForm(apiBase+"/app", url.Values{
		"service":         {"direct/1/UserTransfer/transferForm"},
		"sp":              {"S0"},
		"Form0":           {"$Hidden,inputAmount,inputToUsername,inputComment,$Submit"},
		"$Hidden":         {csrfToken},
		"inputAmount":     {amountStr},
		"inputToUsername": {receiver},
		"inputComment":    {comment},
		"$Submit":         {"Übertragung"},
	}, apiBase)
	defer resp2.Body.Close()

	body2, _ := ioutil.ReadAll(resp2.Body)
	ok := transferSuccessfulRegex.Match(body2)

	return ok
}

// Close closes the api connections
func (pc *PapercutAPI) Close() error {
	if pc.sshClient == nil {
		return nil
	}
	return pc.sshClient.Close()
}
