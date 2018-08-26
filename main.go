package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"github.com/mgutz/ansi"
	"github.com/spf13/viper"
)

// Structs

type replacement struct {
	Match   string
	Replace string
	Color   string
	Disable bool
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
		newline := line
		for _, n := range defs.Definition {
			for _, f := range n.Filter {
				r := regexp.MustCompile(f.Match)
				newline = r.ReplaceAllStringFunc(newline, func(match string) string {
					if !f.Disable {
						if f.Color != "" {
							if matched, _ := regexp.MatchString("[^\\\\]\\(", f.Match); matched == true {
								match = colorSubstring(match, f.Match, f.Color)
							} else {
								match = colorString(match, f.Match, f.Color)
							}
						}
						if f.Replace != "" {
							match = r.ReplaceAllString(match, f.Replace)
						}
					}
					return match
				})
			}
		}

		// Clean up nested colors
		colorRegExp := "\\x1b\\[(\\d?;?\\d\\dm)([^\\x1b]+)\\x1b\\[(\\d?;?\\d\\dm)([^\\x1b]+)\\x1b\\[(0m)([^\\x1b].+)\\x1b\\[(0m)"
		colorReplace := "\x1b[$1$2\x1c[$3$4\x1c[$1$6\x1b[$7"
		// fmt.Println(">>>>", colorRegExp, "<<<<\n")
		c := regexp.MustCompile(colorRegExp)
		substrings := c.FindAllStringSubmatch(newline, -1)
		for len(substrings) > 0 {
			// fmt.Printf("\r----   %q\n", newline)
			newline = c.ReplaceAllString(newline, colorReplace)
			// for _, s := range substrings {
			// 	fmt.Printf("\rsubstring: %q\n", s)
			// }
			// fmt.Printf("\r++++   %q\n", newline)
			substrings := c.FindAllStringSubmatch(newline, -1)
			if len(substrings) == 0 {
				break
			}
		}
		// fmt.Printf("****   %q\n", newline)
		decode := regexp.MustCompile("\\x1c")
		newline = decode.ReplaceAllString(newline, "\x1b")
		fmt.Printf("\r%s\n", newline)
	})
}

func colorSubstring(line string, find string, color string) string {
	// fmt.Printf("\rcolorSubstring(line: %s, find: %s, color: %s\n", line, find, color)
	// TODO find multiple paren strings
	r := regexp.MustCompile("(.*?)(" + find + ")(.*)")
	replace := r.FindAllStringSubmatch(line, -1)
	// fmt.Printf("TEST: %q\n", replace)
	// for _, s := range replace[0] {
	// 	fmt.Printf("replace-substring: %q\n", s)
	// }
	line = strings.Replace(replace[0][2], replace[0][3], ansi.Color(replace[0][3], color), -1)
	// fmt.Printf("\rcolorSubstring: return %s\n", line)
	return line
}

func colorString(line string, find string, color string) string {
	r := regexp.MustCompile(find)
	line = r.ReplaceAllStringFunc(line, func(match string) string {
		// fmt.Printf("\rcolorString():  %s  |  %s  |  %s\n", match, find, color)
		match = r.ReplaceAllString(match, ansi.Color(match, color))
		return match
	})
	return line
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
// TEST outer inner outer TEST
// TEST outer in(n)er outer TEST