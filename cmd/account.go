package cmd

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	CmdAccount = cli.Command{
		Name:        "account",
		Description: "Manage account for accessing web interface",
		Subcommands: []cli.Command{
			subcmdAddAccount,
			subcmdPrintAccounts,
			subcmdDeleteAccounts,
		},
	}

	subcmdAddAccount = cli.Command{
		Name:        "add",
		Description: "Create new account",
		Action:      runAddAccount,
	}

	subcmdPrintAccounts = cli.Command{
		Name:        "print",
		Description: "List all accounts",
		Aliases:     []string{"list", "ls"},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "search, s",
				Usage: "Search accounts by username",
			},
		},
		Action: runPrintAccount,
	}

	subcmdDeleteAccounts = cli.Command{
		Name: "delete",
		Description: "Delete accounts. " +
			"Accepts space-separated list of usernames. " +
			"If no arguments, all records will be deleted.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "yes, y",
				Usage: "Skip confirmation prompt and delete ALL accounts",
			},
		},
		Action: runDeleteAccount,
	}
)

func runAddAccount(c *cli.Context) error {
	// TODO: check for duplicate account already
	args := c.Args()
	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	username := args[0]
	if username == "" {
		return errors.New(cErrorSprint("Username must not be empty"))
	}

	fmt.Println("Username: " + username)

	// Read and validate password
	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	fmt.Println()
	strPassword := string(bytePassword)
	if len(strPassword) < 8 {
		return errors.New(cErrorSprint("Password must be at least 8 characters"))
	}

	// Save account to database
	err = db.CreateAccount(username, strPassword)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}
	return nil
}

func runPrintAccount(c *cli.Context) error {
	// Parse flags
	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}
	keyword := c.String("search")

	// Fetch list accounts in database
	accounts, err := db.GetAccounts(keyword)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Show list accounts
	for _, account := range accounts {
		cIndex.Print("- ")
		fmt.Println(account.Username)
	}
	return nil
}

func runDeleteAccount(c *cli.Context) error {
	args := c.Args()
	skipConfirm := c.Bool("yes")
	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// If no arguments (i.e all accounts going to be deleted),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmDelete := ""
		fmt.Print("Remove ALL accounts? (y/n): ")
		fmt.Scanln(&confirmDelete)

		if confirmDelete != "y" {
			fmt.Println("No accounts deleted")
			return nil
		}
	}

	// Delete accounts in database
	err = db.DeleteAccounts(args...)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	fmt.Println("Account(s) have been deleted")
	return nil
}
