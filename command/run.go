package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sprisa/localhost/util"
	"github.com/urfave/cli/v3"
)

var RunCommand = &cli.Command{
	Name:      "run",
	Usage:     "Run npm script and expose port",
	UsageText: "localhost run [start]",
	Action: func(ctx context.Context, c *cli.Command) error {
		scriptName := c.Args().First()
		if scriptName == "" {
			return cli.ShowSubcommandHelp(c)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return util.WrapError(err, "unable to find current working directory")
		}
		pkgJsonPath := filepath.Join(cwd, "package.json")
		jsonBytes, err := os.ReadFile(pkgJsonPath)
		if err != nil {
			return util.WrapError(err, "error reading package.json")
		}
		pkg := new(PkgJson)
		err = json.Unmarshal(jsonBytes, pkg)
		if err != nil {
			return util.WrapError(err, "error parsing package.json (%s)", pkgJsonPath)
		}

		// util.Log.Info().Msgf("%+v", pkg)

		scriptCommand, hasScriptCommand := pkg.Scripts[scriptName]
		if !hasScriptCommand {
			return fmt.Errorf("unable to find npm script: %s", scriptName)
		}
		scriptCommand = strings.TrimSpace(scriptCommand)

		// shell := os.Getenv("SHELL")
		cmd := exec.Command("sh", "-c", scriptCommand)
		// Pass along existing environment
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()

		return err
	},
}

type PkgJson struct {
	Scripts map[string]string `json:"scripts"`
}
