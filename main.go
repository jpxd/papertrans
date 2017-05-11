package main

import (
	"flag"
	"fmt"
	"time"
)

const sshHost = "clientssh3.rbg.informatik.tu-darmstadt.de:22"
const sshHostKey = "AAAAE2VjZHNhLXNoYTItbmlzdHAzODQAAAAIbmlzdHAzODQAAABhBH94yoY5H61a9V7FiJOgLyljRZlPP5S2yVa+91nBinXUEfk4SOSUz/Xcg4U5vE/DdP/WADgAa0BtM1Yzay6Iaoq2NRrmxp2QLXvHn+HG1vZ3jHFIYkwBjU04JHfxb0No0g=="
const comment = "Have some pages..."

func main() {
	// parse flags
	configureFlag := flag.Bool("configure", false, "create a new configuration file")
	checkFlag := flag.Bool("check", false, "only check page count")
	flag.Parse()

	// create new config if requested
	if *configureFlag {
		fmt.Println("Creating new config")
		createNewConfig()

		// if the user does not want to check his page count as well, we are done now
		if !(*checkFlag) {
			fmt.Println("Done")
			return
		}
	}

	// get credentials
	fmt.Println("Reading credentials from config")
	config := getConfig()

	// check time window
	if !(*checkFlag) && config.TimeSlotMinutes > 0 {
		fmt.Println("Checking time window")

		now := time.Now()
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()

		firstOfThisMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, location)
		firstOfNextMonth := firstOfThisMonth.AddDate(0, 1, 0)

		windowDuration := time.Duration(config.TimeSlotMinutes) * time.Minute
		beginningOfWindow := firstOfNextMonth.Add(-windowDuration)

		if now.Before(beginningOfWindow) {
			fmt.Println("Not in time window")
			fmt.Println("Done")
			return
		}
	}

	// connect webclient to ssh tunnel
	fmt.Println("Connecting tunnel via SSH")
	ssh, err := createSSHClient(sshHost, config.SSHUser, config.SSHKeyFile)
	if err != nil {
		fmt.Println("Failed to connect to SSH server")
		fmt.Println(err)
		return
	}
	defer ssh.Close()

	client := createTunneledWebClient(ssh)

	// create papercut api
	fmt.Println("Logging into PaperCut")
	pc := createPapercutAPI(config.PaperCutUsername, config.PaperCutPassword, client)

	// get page count
	count := pc.getPagesLeft()
	fmt.Println("Pages left:", count)

	// if we just wanted to check the page count, we are done now
	if *checkFlag {
		fmt.Println("Done")
		return
	}

	// make sure we have enough pages
	if count <= config.MinPagesLeft {
		fmt.Println("Not enough pages left, aborting...")
		fmt.Println("Done")
		return
	}

	// lets tranfer some pages
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
