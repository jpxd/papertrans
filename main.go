package main

import "fmt"

const sshHost = "clientssh3.rbg.informatik.tu-darmstadt.de:22"
const comment = "Have some pages..."

func main() {

	config := getConfig()

	// connect webclient to ssh tunnel
	fmt.Println("Connecting tunnel via SSH")
	ssh := createSSHClient(sshHost, config.SSHUser, config.SSHKeyFile)
	client := createTunneledWebClient(ssh)
	defer ssh.Close()

	// get papercut credentials
	fmt.Println("Getting PaperCut credentials")

	// create papercut api
	fmt.Println("Logging into PaperCut")
	pc := createPapercutAPI(config.PaperCutUsername, config.PaperCutPassword, client)

	// lets tranfer some pages
	count := pc.getPagesLeft()
	fmt.Println("Pages left:", count)

	if count <= config.MinPagesLeft {
		fmt.Println("Not enough pages left, aborting..")
		fmt.Println("Done")
		return
	}

	amountToTransfer := count - config.MinPagesLeft

	fmt.Println("Transferring", amountToTransfer, "pages to", config.Receiver)
	if pc.tranferPages(config.Receiver, amountToTransfer, comment) {
		fmt.Println("Transfer was successful")
	} else {
		fmt.Println("Transfer has failed")
	}

	count = pc.getPagesLeft()
	fmt.Println("Pages left:", count)

	fmt.Println("Done")
}
