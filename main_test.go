package main

import (
	"fmt"
	"testing"
)

func TestColor(t *testing.T) {
	var filters []Filter
	// Color Reset to Previous Testing
	filters = append(filters, Filter{
		Match: "// TEST.+",
		Color: "black+h",
	})
	filters = append(filters, Filter{
		Match: "outer.*outer",
		Color: "green",
	})
	filters = append(filters, Filter{
		Match: "i[n\\(\\)]+er",
		Color: "yellow",
	})
	filters = append(filters, Filter{
		Match: "i(.{2})er",
		Color: "magenta",
	})
	filters = append(filters, Filter{
		Match: "in(\\(.\\))er",
		Color: "red+h",
	})
	// Overlapping RegExps Testing
	filters = append(filters, Filter{
		Match: "PING (.+?) ",
		Color: "magenta+h",
	})
	filters = append(filters, Filter{
		Match: "\\d+\\.\\d+\\.\\d+\\.\\d+",
		Color: "magenta",
	})

	// Definition Setup
	defs := Configuration{}
	defs.Definition = append(defs.Definition, Definition{
		Name:   "testing",
		Filter: filters,
	})

	// Tests
	processed := processLine("// TEST outer zinnerz outer TEST", defs)
	knowngood := "\x1b[90m// TEST \x1b[32mouter z\x1b[33mi\x1b[35mnn\x1b[33mer\x1b[32mz outer\x1b[90m TEST\x1b[0m"
	if processed != knowngood {
		fmt.Printf("expected  : %q\n", knowngood)
		fmt.Printf("actual    : %q\n", processed)
		fmt.Printf("expected  : %s\n", knowngood)
		fmt.Printf("actual    : %s\n", processed)
		t.Fatal("Reset Back to Previous Color Test FAILED")
	}
	processed = processLine("// TEST outer zin(n)erz outer TEST", defs)
	knowngood = "\x1b[90m// TEST \x1b[32mouter z\x1b[33min\x1b[91m(n)\x1b[33mer\x1b[32mz outer\x1b[90m TEST\x1b[0m"
	if processed != knowngood {
		fmt.Printf("expected  : %q\n", knowngood)
		fmt.Printf("actual    : %q\n", processed)
		t.Fatal("Inline (non-regex) Parenthesis Test FAILED")
	}

	processed = processLine("PING dns.public.google.com (8.8.8.8): 56 data bytes", defs)
	knowngood = "PING \x1b[95mdns.public.google.com\x1b[0m (\x1b[35m8.8.8.8\x1b[0m): 56 data bytes"
	if processed != knowngood {
		fmt.Printf("expected  : %q\n", knowngood)
		fmt.Printf("actual    : %q\n", processed)
		t.Fatal("Overlapping Regexes 1 Test FAILED")
	}

	processed = processLine("PING 8.8.8.8 (8.8.8.8): 56 data bytes", defs)
	knowngood = "PING \x1b[95m\x1b[35m8.8.8.8\x1b[0m\x1b[0m (\x1b[35m8.8.8.8\x1b[0m): 56 data bytes"
	//                ^color1 ^color2 ^text  ^reset1^reset2  // I HATE this. #knownissue
	if processed != knowngood {
		fmt.Printf("expected  : %q\n", knowngood)
		fmt.Printf("actual    : %q\n", processed)
		t.Fatal("Overlapping Regexes 2 Test FAILED")
	}

}
