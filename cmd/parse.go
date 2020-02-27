package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	ansi "github.com/k0kubun/go-ansi"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/config"
)

// Parse command
func Parse(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.Instance
	cln := client.Instance
	source := ""
	ext := ""
	if cfg.GenAfterParse {
		if len(cfg.Template) == 0 {
			return errors.New("You have to add at least one code template by `cf config`")
		}
		path := cfg.Template[cfg.Default].Path
		ext = filepath.Ext(path)
		if source, err = readTemplateSource(path, cln); err != nil {
			return err
		}
	}
	parseContest := func(contestID, rootPath string) error {
		problems, err := cln.ParseContest(contestID, rootPath)
		if err == nil && cfg.GenAfterParse {
			for _, problem := range problems {
				problemID := strings.ToLower(problem.ID)
				path := filepath.Join(rootPath, problemID)
				gen(source, path, ext)
			}
		}
		return err
	}
	work := func() error {
		contestID := ""
		problemID := ""
		path := currentPath
		var ok bool
		if contestID, ok = args["<contest-id>"].(string); ok {
			if problemID, ok = args["<problem-id>"].(string); !ok {
				return parseContest(contestID, filepath.Join(currentPath, contestID))
			}
			problemID = strings.ToLower(problemID)
			path = filepath.Join(currentPath, contestID, problemID)
		} else {
			contestID, err = getContestID(args)
			if err != nil {
				return err
			}
			problemID, err = getProblemID(args)
			if err != nil {
				return err
			}
			if problemID == contestID {
				return parseContest(contestID, currentPath)
			}
		}
		samples, standardIO, err := cln.ParseContestProblem(contestID, problemID, path)
		if err != nil {
			color.Red("Failed %v %v", contestID, problemID)
			return err

		}
		warns := ""
		if !standardIO {
			warns = color.YellowString("Non standard input output format.")
		}
		ansi.Printf("%v %v\n", color.GreenString("Parsed %v %v with %v samples.", contestID, problemID, samples), warns)
		if cfg.GenAfterParse {
			gen(source, path, ext)
		}
		return nil
	}
	if err = work(); err != nil {
		if err = loginAgain(cfg, cln, err); err == nil {
			err = work()
		}
	}
	return err
}
