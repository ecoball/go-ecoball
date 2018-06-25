package main

import (
	"github.com/urfave/cli"
)

var (
	Name     string
	Password string
)

func NewNameFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "name",
		Usage:       "wallet file name",
		Value:       "",
		Destination: &Name,
	}
}

func NewPasswordFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "password",
		Usage:       "wallet password",
		Value:       "",
		Destination: &Password,
	}
}

func checkPassword() {
	/*	var (
				nameTime     = 0
				passwordTime = 0
			)

			//empty name
			if "" == Name {
				fmt.Printf("please input wallet file name:")
				fmt.Scan(&Name)
				goto name
			}

		name:
			if 0 != nameTime {
				fmt.Printf("please input wallet file name:")
				fmt.Scan(&Name)
			}

			//file does not exist
			if _, err := os.Stat(Name); err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				nameTime++
				if nameTime >= 3 {
					fmt.Fprintln(os.Stderr, "More than three times, maybe you didn't create your wallet, exit...")
					os.Exit(1)
				}
				goto name
			}

			//empty password
			if "" == Password {
				fmt.Printf("please input wallet password:")
				fmt.Scan(&Password)
				goto password
			}

		password:
			if 0 != passwordTime {
				fmt.Printf("please input wallet password:")
				fmt.Scan(&Password)
			}

			//worng password
			wallet := account.Open(Name, []byte(Password))
			if nil == wallet {
				fmt.Fprintln(os.Stderr, "open wallet failed!")
				passwordTime++
				if passwordTime >= 3 {
					fmt.Fprintln(os.Stderr, "More than three times, exit...")
					os.Exit(1)
				}
				goto password
			}

			//get account
			if 0 == len(wallet.KeyData.Accounts) {
				fmt.Fprintln(os.Stderr, "empty account, please create account")
				os.Exit(1)
			}

			for _, v := range wallet.KeyData.Accounts {
				common.Account = v
			}*/
}
