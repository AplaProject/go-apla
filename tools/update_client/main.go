package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"github.com/GenesisKernel/go-genesis/tools/update_client/client"
	"github.com/GenesisKernel/go-genesis/tools/update_client/params"
	"github.com/GenesisKernel/go-genesis/tools/update_server/model"
)

type ServerOpt struct {
	Server string `long:"server" description:"updater server address" required:"true"`
}

type ServerCredentialsOpt struct {
	Login    string `long:"login" description:"login for updater server auth" required:"true"`
	Password string `long:"password" description:"password for updater server auth" required:"true"`
}

type VersionOpt struct {
	Version string `long:"version" description:"binary version" required:"true"`
}

type PublicKeyOpt struct {
	PublicKeyPath string `long:"publ-key-path" description:"path to public key" default:"./resources/key.pub" required:"true"`
}

type PrivateKeyOpt struct {
	PrivateKeyPath string `long:"key-path" description:"path to private key for binary signing" default:"./resources/key" required:"true"`
}

var opts struct {
	AddCommand struct {
		ServerOpt
		ServerCredentialsOpt

		Path       string `long:"binary-path" description:"path to binary that will added" required:"true"`
		StartBlock int64  `long:"start-block" description:"block updating from" required:"true"`

		VersionOpt
		PrivateKeyOpt
	} `command:"add-binary"`

	GetCommand struct {
		ServerOpt
		VersionOpt

		Path string `long:"binary-path" description:"path binary will saved to" required:"true"`

		PublicKeyOpt
	} `command:"get-binary"`

	RemoveCommand struct {
		ServerOpt
		ServerCredentialsOpt
		VersionOpt
	} `command:"remove-binary"`

	VersionsCommand struct {
		ServerOpt
		VersionOpt
	} `command:"versions"`

	GenerateKeysCommand struct {
		PublicKeyOpt
		PrivateKeyOpt
	} `command:"generate-keys"`
}

func main() {
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		os.Exit(1)
	}

	c := &client.UpdateClient{}
	var err error

	switch p.Active.Name {
	case "add-binary":
		err = c.AddBinary(
			params.KeyParams{PrivateKeyPath: opts.AddCommand.PrivateKeyPath},
			params.BinaryParams{
				Path:       opts.AddCommand.Path,
				StartBlock: opts.AddCommand.StartBlock,
				Version:    opts.AddCommand.Version,
			},
			params.ServerParams{
				Server:   opts.AddCommand.Server,
				Login:    opts.AddCommand.Login,
				Password: opts.AddCommand.Password,
			})
	case "get-binary":
		var b model.Build
		b, err = c.GetBinary(
			params.ServerParams{
				Server: opts.GetCommand.Server,
			},
			params.KeyParams{
				PublicKeyPath: opts.GetCommand.PublicKeyPath,
			},
			params.BinaryParams{
				Version: opts.GetCommand.Version,
			})
		if err != nil {
			err = errors.Wrapf(err, "getting binary from server")
			break
		}

		p := filepath.Join(opts.GetCommand.Path, b.Name+"_"+b.Version.String())
		err = ioutil.WriteFile(p, b.Body, 0600)
		if err != nil {
			err = errors.Wrapf(err, "writing binary to file")
			break
		}
	case "remove-binary":
		err = c.RemoveBinary(
			params.ServerParams{
				Server:   opts.RemoveCommand.Server,
				Login:    opts.RemoveCommand.Login,
				Password: opts.RemoveCommand.Password,
			},
			params.BinaryParams{
				Version: opts.RemoveCommand.Version,
			})
	case "versions":
		vrs, verr := c.GetVersionList(
			params.ServerParams{
				Server: opts.VersionsCommand.Server,
			},
			params.BinaryParams{
				Version: opts.VersionsCommand.Version,
			})
		if verr != nil {
			err = verr
			break
		}
		for _, v := range vrs {
			fmt.Println(v.Version.String())
		}
	case "generate-keys":
		err = c.GenerateKeys(
			params.KeyParams{
				PrivateKeyPath: opts.GenerateKeysCommand.PrivateKeyPath,
				PublicKeyPath:  opts.GenerateKeysCommand.PublicKeyPath,
			})
	}

	if err != nil {
		fmt.Printf("Error while %s: %s\n", p.Active.Name, err.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Command \"%s\" successfully done\n", p.Active.Name)
	}
}
