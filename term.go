package main

import (
	"errors"
)

type kissArgs struct {
	output  string
	entry   string
	globals string
}

func validArgs(args []string) bool {
	if len(args) < 2 {
		return false
	}
	if len(args) > 6 {
		return false
	}

	for i, arg := range args {
		if i == 2 || i == 4 {
			if len(arg) != 2 {
				return false
			}
			if arg[0] != '-' {
				return false
			}
			if arg[1] != 'o' && arg[1] != 'g' {
				return false
			}
			if len(args) < i+2 {
				return false
			}
		}
		if i == 1 || i == 3 {
			if arg[0] == '-' {
				return false
			}
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
	}
	return ret, nil
}
