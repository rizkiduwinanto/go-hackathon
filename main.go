package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

func main() {
	GmailIMAP()
}

func GmailIMAP() {
	log.Println("Connecting..")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	defer c.Logout()

	if err := c.Login(os.Getenv("GMAIL_EMAIL"), os.Getenv("GMAIL_PASSWORD")); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Recent
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchBodyStructure}, messages)
	}()

	for msg := range messages {
		if strings.Contains(msg.Envelope.Subject, "AUDIT") {
			fmt.Println(msg.BodyStructure.Filename())
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
