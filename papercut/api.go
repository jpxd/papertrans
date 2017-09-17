package papercut

import (
	"errors"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// consts
const apiBase = "https://print.informatik.tu-darmstadt.de"

// regexes
var loginSuccessfulRegex = regexp.MustCompile("Angemeldet als:")
var pageCountRegex = regexp.MustCompile("([0-9]+) Seiten")
var hiddenFieldRegex = regexp.MustCompile("name=\"\\$Hidden\" value=\"([a-zA-Z0-9]+)\"")
var transferSuccessfulRegex = regexp.MustCompile("tragung war erfolgreich.")

type PapercutApi struct {
	webClient *WebClient
	sshClient *ssh.Client // Only set if created and owned by this instance.
}

func CreatePapercutApiWithWebClient(user string, pass string, webClient *WebClient) *PapercutApi {
	pc := new(PapercutApi)
	pc.webClient = webClient
	pc.getSession()
	if !pc.loginUser(user, pass) {
		panic("Could not login into PaperCut")
	}
	return pc
}

func CreateRemotePapercutApi(sshHost string, sshHostKey string, config *ConfigContainer) (*PapercutApi, error) {
	ssh, err := CreateSSHClient(sshHost, sshHostKey, config.SSHUser, config.SSHKeyFile)
	if err != nil {
		return nil, err
	}

	pc := new(PapercutApi)
	pc.webClient = CreateTunneledWebClient(ssh)
	pc.sshClient = ssh
	pc.getSession()
	if !pc.loginUser(config.PaperCutUsername, config.PaperCutPassword) {
		pc.Close()
		return nil, errors.New("Could not login into PaperCut")
	}
	return pc, nil
}

func (pc *PapercutApi) getSession() {
	resp := pc.webClient.Get(apiBase)
	defer resp.Body.Close()
}

func (pc *PapercutApi) loginUser(user string, pass string) bool {
	resp := pc.webClient.PostForm(apiBase+"/app", url.Values{
		"service":              {"direct/1/Home/$Form$0"},
		"sp":                   {"S0"},
		"Form0":                {"$Hidden$0,$Hidden$1,inputUsername,inputPassword,$PropertySelection$0,$Submit$0"},
		"$Hidden$0":            {"true"},
		"$Hidden$1":            {"X"},
		"inputUsername":        {user},
		"inputPassword":        {pass},
		"$PropertySelection$0": {"de"},
		"$Submit$0":            {"Anmelden"},
	}, apiBase)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	ok := loginSuccessfulRegex.Match(body)

	return ok
}

func (pc *PapercutApi) GetPagesLeft() int {
	resp := pc.webClient.Get(apiBase + "/app?service=page/UserSummary")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	matches := pageCountRegex.FindSubmatch(body)
	countStr := string(matches[1])

	count64, _ := strconv.ParseInt(countStr, 10, 32)
	count := int(count64)

	return count
}

func (pc *PapercutApi) TranferPages(receiver string, amount int, comment string) bool {
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

func (pc *PapercutApi) Close() error {
	if pc.sshClient == nil {
		return nil
	}
	return pc.sshClient.Close()
}
