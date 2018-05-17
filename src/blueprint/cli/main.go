package main

import (
	"blueprint/types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/auth"
	"github.com/loomnetwork/go-loom/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

func main() {
	var writeURI, readURI, contractHexAddr, chainID, privFile, user string
	rootCmd := &cobra.Command{
		Use:   "blueprint",
		Short: "blueprint example",
	}
	rootCmd.PersistentFlags().StringVarP(&writeURI, "write", "w", "http://localhost:46657", "URI for sending txs")
	rootCmd.PersistentFlags().StringVarP(&readURI, "read", "r", "http://localhost:9999", "URI for quering app state")
	rootCmd.PersistentFlags().StringVarP(&contractHexAddr, "contract", "", "0x005B17864f3adbF53b1384F2E6f2120c6652F779", "contract address")
	rootCmd.PersistentFlags().StringVarP(&chainID, "chain", "", "default", "chain ID")

	contractAddr, err := loom.LocalAddressFromHexString(contractHexAddr)
	if err != nil {
		log.Fatal(err)
	}
	// create rpc client
	rpcClient := client.NewDAppChainRPCClient(chainID, writeURI, readURI)
	// create contract
	contract := client.NewContract(rpcClient, contractAddr, "BluePrint")
	//  create account cmd
	createAccCmd := &cobra.Command{
		Use:   "create-acct",
		Short: "create-acct create an account used to store data",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				return err
			}
			signer := auth.NewEd25519Signer(privKey)
			payload := &types.BluePrintCreateAccountTx{
				Version: 1,
				Owner:   user,
				Data:    []byte("my awesome profile"),
			}
			if _, err := contract.Call("CreateAccount", payload, signer, nil); err != nil {
				return errors.Wrap(err, "contract call error")
			}
			return nil
		},
	}
	createAccCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")
	createAccCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	var value int
	saveStateCmd := &cobra.Command{
		Use:   "set",
		Short: "set the state",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				return errors.Wrap(err, "private key file not found")
			}
			msgData := struct {
				Value int
			}{Value: value}

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
			resp, err := contract.Call("SaveState", msg, signer, nil)
			if err != nil {
				return err
			}
			fmt.Printf("--> resp: %v", resp)

			return nil
		},
	}
	saveStateCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	saveStateCmd.Flags().IntVarP(&value, "value", "v", 0, "integer value")
	saveStateCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

	getStateCmd := &cobra.Command{
		Use:   "get",
		Short: "get state",
		RunE: func(cmd *cobra.Command, args []string) error {
			var result types.StateQueryResult
			params := &types.StateQueryParams{
				Owner: user,
			}
			if _, err := contract.StaticCall("GetState", params, &result); err != nil {
				return err
			}
			fmt.Println(string(result.State))
			return nil
		},
	}

	getStateCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	getStateCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

	keygenCmd := &cobra.Command{
		Use:   "genkey",
		Short: "generate a public and private key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				return errors.Wrap(err, "Error generating key pair")
			}
			if err := ioutil.WriteFile(privFile, priv, 0664); err != nil {
				return errors.Wrap(err, "Unable to write private key file")
			}
			fmt.Printf("generated private key file '%s'\n", privFile)
			return nil
		},
	}
	keygenCmd.Flags().StringVarP(&privFile, "key", "k", "priv_key", "private key file")

	rootCmd.AddCommand(createAccCmd)
	rootCmd.AddCommand(saveStateCmd)
	rootCmd.AddCommand(getStateCmd)
	rootCmd.AddCommand(keygenCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
