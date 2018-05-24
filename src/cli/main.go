package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"types"

	"github.com/loomnetwork/go-loom/auth"
	"github.com/loomnetwork/go-loom/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

type MessageData struct {
	Value int
}

func main() {
	var writeURI, readURI, contractName, chainID, privFile, user string
	rootCmd := &cobra.Command{
		Use:   "blueprint",
		Short: "blueprint example",
	}
	rootCmd.PersistentFlags().StringVarP(&writeURI, "write", "w", "http://localhost:46658/rpc", "URI for sending txs")
	rootCmd.PersistentFlags().StringVarP(&readURI, "read", "r", "http://localhost:46658/query", "URI for quering app state")
	rootCmd.PersistentFlags().StringVarP(&contractName, "contract", "", "BluePrint", "contract address")
	rootCmd.PersistentFlags().StringVarP(&chainID, "chain", "", "default", "chain ID")

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

	createAccCmd := &cobra.Command{
		Use:           "create-acct",
		Short:         "create-acct create an account used to store data",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				return err
			}

			privKey, err = base64.StdEncoding.DecodeString(string(privKey))
			if err != nil {
				return errors.Wrapf(err, "Cannot read priv file: %s", privFile)
			}

			signer := auth.NewEd25519Signer(privKey)
			payload := &types.BluePrintCreateAccountTx{
				Version: 1,
				Owner:   user,
				Data:    []byte(user),
			}

			contract, err := newContract(chainID, writeURI, readURI, contractName)
			if err != nil {
				return err
			}

			if _, err := contract.Call("CreateAccount", payload, signer, nil); err != nil {
				return errors.Wrap(err, "contract call error")
			}
			fmt.Printf("user %s created successfully!\n", user)
			return nil
		},
	}
	createAccCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")
	createAccCmd.Flags().StringVarP(&privFile, "key", "k", "key", "private key file")

	var value int
	saveStateCmd := &cobra.Command{
		Use:           "set",
		Short:         "set the state",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				return errors.Wrap(err, "private key file not found")
			}

			privKey, err = base64.StdEncoding.DecodeString(string(privKey))
			if err != nil {
				log.Fatalf("Cannot read priv file: %s", privFile)
			}

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

			signer := auth.NewEd25519Signer(privKey)
			contract, err := newContract(chainID, writeURI, readURI, contractName)
			if err != nil {
				return err
			}

			_, err = contract.Call("SaveState", msg, signer, nil)
			if err != nil {
				return err
			}
			fmt.Printf("set value %d for user '%s' successfully!\n", value, user)
			return nil
		},
	}
	saveStateCmd.Flags().StringVarP(&privFile, "key", "k", "key", "private key file")
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

			contract, err := newContract(chainID, writeURI, readURI, contractName)
			if err != nil {
				return err
			}

			if _, err := contract.StaticCall("GetState", params, &result); err != nil {
				return err
			}
			var msgData MessageData
			err = json.Unmarshal(result.State, &msgData)
			if err != nil {
				return err
			}
			fmt.Printf("get value %d from user '%s'\n", msgData.Value, user)
			return nil
		},
	}

	getStateCmd.Flags().StringVarP(&privFile, "key", "k", "key", "private key file")
	getStateCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(createAccCmd)
	rootCmd.AddCommand(saveStateCmd)
	rootCmd.AddCommand(getStateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func newContract(chainID, writeURI, readURI, contractName string) (*client.Contract, error) {
	// create rpc client
	rpcClient := client.NewDAppChainRPCClient(chainID, writeURI, readURI)
	// resolve address
	contractAddr, err := rpcClient.Resolve(contractName)
	if err != nil {
		return nil, err
	}
	// create contract
	contract := client.NewContract(rpcClient, contractAddr.Local)
	return contract, nil
}
