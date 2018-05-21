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
)

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

	// create rpc client
	rpcClient := client.NewDAppChainRPCClient(chainID, writeURI, readURI)

	contractAddr, err := rpcClient.Resolve(contractName)
	if err != nil {
		log.Fatal(err)
	}

	// create contract
	contract := client.NewContract(rpcClient, contractAddr.Local)

	//  create account cmd
	createAccCmd := &cobra.Command{
		Use:   "create-acct",
		Short: "create-acct create an account used to store data",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				return err
			}

			privKey, err = base64.StdEncoding.DecodeString(string(privKey))
			if err != nil {
				log.Fatalf("Cannot read priv file: %s", privFile)
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

			privKey, err = base64.StdEncoding.DecodeString(string(privKey))
			if err != nil {
				log.Fatalf("Cannot read priv file: %s", privFile)
			}

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

	rootCmd.AddCommand(createAccCmd)
	rootCmd.AddCommand(saveStateCmd)
	rootCmd.AddCommand(getStateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
