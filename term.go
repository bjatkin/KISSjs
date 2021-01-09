package main

import (
	"errors"
)

type kissArgs struct {
	output       string
	entry        string
	globals      string
	viewLocation string
}

func validArgs(args []string) bool {
	if len(args) < 2 {
		// we mucst have at least an entry file
		return false
	}
	if len(args)%2 == 1 {
		// we should have matching argument and values
		// e.g. -g globals
		return false
	}

	for i, arg := range args {
		if i == 0 {
			// this is the dir of the program
			continue
		}

		if (i+1)%2 == 0 {
			// These are the argument values
			if arg[0] == '-' {
				return false
			}
			continue
		}

		if len(arg) != 2 {
			// args should be in the form -O
			return false
		}
		if arg[0] != '-' {
			// args should be in the form -O
			return false
		}
		if arg[1] != 'o' && arg[1] != 'g' && arg[1] != 'v' {
			// only -o, -g, and -v allowed
			return false
		}
	}
	return true
}

func parseArgs(args []string) (kissArgs, error) {
	if !validArgs(args) {
		return kissArgs{}, errors.New("invalid arguments")
	}

	ret := kissArgs{
		entry:  args[1],
		output: getPath(args[1]) + "dist",
	}
	for i, arg := range args {
		if arg == "-o" {
			ret.output = args[i+1]
		}
		if arg == "-g" {
			ret.globals = args[i+1]
		}
		if arg == "-v" {
			ret.viewLocation = args[i+1]
		}
	}
	return ret, nil
}
