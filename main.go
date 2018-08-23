package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/mgutz/ansi"
	"github.com/spf13/viper"
)

// Structs

type replacement struct {
	Match   string
	Replace string
	Color   string
}

type definition struct {
	Name   string
	Filter []replacement
}

type configuration struct {
	Definition []definition
}

// Array Flags
//  - https://stackoverflow.com/questions/28322997/how-to-get-a-list-of-values-into-a-flag-in-golang
type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Len() int {
	var n int
	for range *i {
		n++
	}
	return n
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var confFiles arrayFlags

// Functions

// ReadLine reads a line safely into the buffer
func ReadLine(reader io.Reader, f func(string)) {
	buf := bufio.NewReader(reader)
	line, err := buf.ReadBytes('\n')
	for err == nil {
		line = bytes.TrimRight(line, "\n")
		if len(line) > 0 {
			if line[len(line)-1] == 13 { //'\r'
				line = bytes.TrimRight(line, "\r")
			}
			f(string(line))
		}
		line, err = buf.ReadBytes('\n')
	}

	if len(line) > 0 {
		f(string(line))
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Main
func main() {
	// "This line contains - Read - in the center"

	var config, defs configuration
	err := viper.Unmarshal(&config)
	if err != nil {
		panic("Unable to unmarshal config")
	}

	// Debug Configuration
	// for _, n := range config.Definition {
	//     fmt.Printf(" -- Name: %s \n", n.Name)
	//     for _, f := range n.Filter {
	//         fmt.Printf(" -- -- Match: '%s', Color: '%s'\n", f.Match, f.Color)
	//     }
	// }

	// Populate defs list
	if len(confFiles) > 0 {
		// fmt.Printf("Length of confFiles: %d > 0\n", len(confFiles))
		for _, x := range config.Definition {
			if stringInSlice(x.Name, confFiles) {
				defs.Definition = append(defs.Definition, x)
			}
		}
	} else {
		// fmt.Printf("Length of confFiles: %d = 0\n", len(confFiles))
		for _, x := range config.Definition {
			defs.Definition = append(defs.Definition, x)
		}
	}

	// Debug defs
	// for _, n := range defs.Definition {
	//     fmt.Printf(" -- Name: %s \n", n.Name)
	//     for _, f := range n.Filter {
	//         fmt.Printf(" -- -- Match: '%s', Color: '%s'\n", f.Match, f.Color)
	//     }
	// }

	// iterate through defs
	ReadLine(os.Stdin, func(line string) {
		for _, n := range defs.Definition {
			for _, f := range n.Filter {
				r := regexp.MustCompile(f.Match)
				if f.Color != "" {
					line = r.ReplaceAllStringFunc(line, func(match string) string {
						return ansi.Color(match, f.Color)
					})
				}
				if f.Replace != "" {
					line = r.ReplaceAllString(line, f.Replace)
				}
			}
		}
		fmt.Println(line)
	})
}

// Init
func init() {
	viper.SetConfigType("yaml")       // use yaml
	viper.SetConfigName("config")     // name of config file (without extension)
	viper.AddConfigPath("/etc/grc/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.grc") // call multiple times to add many search paths
	viper.AddConfigPath(".")          // optionally look for config in the working directory
	err := viper.ReadInConfig()       // Find and read the config file
	if err != nil {                   // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	flag.Var(&confFiles, "conf", "Some description for this param.")
	flag.Parse()
}
