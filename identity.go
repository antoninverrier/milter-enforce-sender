package main

import (
	"bufio"
	"fmt"
	"log"
	"mime"
	"os"
	"strings"
)

type User struct {
	Name      string
	Mailbox   string
	Extension string
	Domain    string
}

func (u User) fromHeader() string {
	return fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", u.Name), u.smtpFrom())
}

func (u User) smtpFrom() string {
	if u.Extension != "" {
		return fmt.Sprintf("%s+%s@%s", u.Mailbox, u.Extension, u.Domain)
	} else {
		return fmt.Sprintf("%s@%s", u.Mailbox, u.Domain)

	}
}

type UserDb map[string][]User

func reloadIdentities() error {
	file, err := os.Open(cfg.IdentityDb)

	if err != nil {
		log.Printf("Failed to reload identity database : %s", err)
		return err
	}
	defer file.Close()

	newdb := make(UserDb)

	scan := bufio.NewScanner(file)
	linenum := 0
	login := ""
	user := User{}

	for scan.Scan() {
		line := scan.Text()

		switch linenum % 3 {
		case 0:
			user = User{}
			login = line
		case 1:
			if line == "" {
				log.Printf("Error in database at line <%d> : email address empty", linenum)
				return nil
			}
			user.Mailbox, user.Domain, _ = strings.Cut(line, "@")
		default:
			user.Name = line
			_, user_exists := newdb[login]
			if !user_exists {
				newdb[login] = make([]User, 0)
			}
			newdb[login] = append(newdb[login], user)
		}
		linenum++

	}

	if err = scan.Err(); err != nil {
		log.Printf("Failed to reload identity database : %s", err)
		return err
	}

	users = newdb
	log.Print("Identity database reloaded")

	return nil

}
