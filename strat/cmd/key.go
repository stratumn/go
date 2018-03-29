// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-indigocore/generator"
)

const (
	keyExt       = ".pem"
	pubKeySuffix = ".pub"
)

var (
	keyFilename string
)

// keyCmd represents the info command
var keyCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate keys",
	Long: `Generate key files.

It currently supports 3 key algorithms : ED25519, RSA and ECDSA`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		in := generator.StringSelect{
			InputShared: generator.InputShared{
				Prompt: "What kind of key would you like to generate ?",
			},
			Options: generator.StringSelectOptions{},
		}
		keyOptions := map[x509.PublicKeyAlgorithm]string{
			keys.ED25519: "ED25519",
			x509.RSA:     "RSA",
			x509.ECDSA:   "ECDSA",
		}
		for id, algoName := range keyOptions {
			in.Options[strconv.Itoa(int(id))] = algoName
		}
		ret, err := in.Run()
		if err != nil {
			return err
		}

		algo, err := strconv.Atoi(ret.(string))
		if err != nil {
			return err
		}

		pub, priv, err := keys.GenerateKey(x509.PublicKeyAlgorithm(algo))
		if err != nil {
			return err
		}

		privKeyFilename := keyFilename + keyExt
		err = ioutil.WriteFile(privKeyFilename, priv, 0666)
		if err != nil {
			return err
		}

		pubKeyFilename := keyFilename + pubKeySuffix + keyExt
		err = ioutil.WriteFile(pubKeyFilename, pub, 0666)
		if err != nil {
			return err
		}

		fmt.Printf("Your secret key has been saved in %s.\nYour public key has been saved in %s\n", privKeyFilename, pubKeyFilename)
		return nil
	},
}

func init() {

	RootCmd.AddCommand(keyCmd)
	keyCmd.PersistentFlags().StringVarP(
		&keyFilename,
		"out",
		"o",
		"stratumn_key",
		"Key filename",
	)
}
