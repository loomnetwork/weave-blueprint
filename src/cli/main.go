package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"types"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
	"github.com/loomnetwork/go-loom/cli"
)

type MessageData struct {
	Value int
}

func getKeygenCmd() (*cobra.Command) {
	var privFile string
	keygenCmd := &cobra.Command{
		Use:           "genkey",
		Short:         "generate a public and private key pair",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				return errors.Wrapf(err, "Error generating key pair")
			}
			data := base64.StdEncoding.EncodeToString(priv)
			if err := ioutil.WriteFile(privFile, []byte(data), 0664); err != nil {
				return errors.Wrapf(err, "Unable to write private key")
			}
			fmt.Printf("written private key file '%s'\n", privFile)
			return nil
		},
	}
	keygenCmd.Flags().StringVarP(&privFile, "key", "k", "key", "private key file")
	return keygenCmd
}

func main() {
	var user string
	defaultContract := "BluePrint"

	rootCmd := &cobra.Command{
		Use:   "blueprint",
		Short: "Blueprint",
	}

	keygenCmd := getKeygenCmd()

	callCmd := cli.ContractCallCommand()
	rootCmd.AddCommand(callCmd)



	rootCmd.AddCommand(keygenCmd)

	createAccCmd := &cobra.Command{
		Use:           "create-acct",
		Short:         "create-acct create an account used to store data",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := &types.BluePrintCreateAccountTx{
				Version: 1,
				Owner:   user,
				Data:    []byte(user),
			}

			err := cli.CallContract(defaultContract, "CreateAccount", payload, nil)
			if err != nil {
				return errors.Wrap(err, "contract call error")
			}
			fmt.Printf("user %s created successfully!\n", user)
			return nil
		},
	}
	createAccCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

	var value int
	saveStateCmd := &cobra.Command{
		Use:           "set",
		Short:         "set the state",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			// define data
			msgData := MessageData{Value: value}
			data, err := json.Marshal(msgData)
			if err != nil {
				return errors.Wrap(err, "value contains invalid json format")
			}

			msg := &types.BluePrintStateTx{
				Version: 0,
				Owner:   user,
				Data:    data,
			}

			err = cli.CallContract(defaultContract, "SaveState", msg, nil)
			if err != nil {
				return err
			}

			log.Printf("set value %d for user '%s' successfully!\n", value, user)
			return nil
		},
	}
	saveStateCmd.Flags().IntVarP(&value, "value", "v", 0, "integer value")
	saveStateCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

	getStateCmd := &cobra.Command{
		Use:           "get",
		Short:         "get state",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var result types.StateQueryResult
			params := &types.StateQueryParams{
				Owner: user,
			}

			if err := cli.StaticCallContract(defaultContract, "GetState", params, &result); err != nil {
				return err
			}

			var msgData MessageData
			err := json.Unmarshal(result.State, &msgData)
			if err != nil {
				return err
			}
			log.Printf("get value %d from user '%s'\n", msgData.Value, user)
			return nil
		},
	}
	getStateCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

	callCmd.AddCommand(createAccCmd)
	callCmd.AddCommand(saveStateCmd)
	callCmd.AddCommand(getStateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
