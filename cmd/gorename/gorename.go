package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Accept and validate commandline params:
	// -y (optional) - yes to all
	// -p (optional) - posix regex mode
	// -v (optional) - verbose
	// regex (positional, required) - regex to match
	// replacement (positional, required) - replacement string
	// files (positional, 1 or more, required)

	cmdname := os.Args[0]

	// Parse command-line arguments
	args := os.Args[1:]

	// Check for optional -y flag
	yesToAll := false
	posixMode := false
	verbose := false
	for strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "-y":
			yesToAll = true
		case "-p":
			posixMode = true
		case "-v":
			verbose = true
		case "-h":
			usage(cmdname)
		default:
			fmt.Printf("Unknown option: %s\n", args[0])
			usage(cmdname)
		}
		args = args[1:]
	}

	// Validate remaining arguments
	if len(args) < 3 {
		fmt.Println("Not enough arguments provided.")
		usage(cmdname)
	}

	regexPattern := args[0]
	replacement := args[1]
	files := args[2:]

	// Compile the regex
	var regex *regexp.Regexp
	var err error
	if posixMode {
		regex, err = regexp.CompilePOSIX(regexPattern)
	} else {
		regex, err = regexp.Compile(regexPattern)
	}
	if err != nil {
		fmt.Printf("Invalid regex: %v\n", err)
		os.Exit(1)
	}

	// Process each file
	for _, file := range files {
		err := processFile(file, regex, replacement, yesToAll, verbose)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", file, err)
		}
	}
}

func processFile(file string, regex *regexp.Regexp, replacement string, yesToAll bool, verbose bool) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	nameBefore := info.Name()
	nameAfter := regex.ReplaceAllString(nameBefore, replacement)

	if nameBefore == nameAfter {
		if verbose {
			fmt.Printf("No changes for file: %s\n", nameBefore)
		}
		return nil
	}
	// Check if the new name is valid
	if nameAfter == "" {
		return fmt.Errorf("new name is empty after regex replacement")
	}

	// Check if the new name already exists
	newPath := filepath.Join(filepath.Dir(file), nameAfter)
	if _, err := os.Stat(newPath); err == nil {
		if !yesToAll {
			fmt.Printf("File %s already exists. Overwrite? (y/n): ", newPath)
			var response string
			_, err = fmt.Scanln(&response)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}
			if !strings.HasPrefix(strings.ToLower(response), "y") {
				fmt.Println("Skipping file.")
				return nil
			}
		} else {
			fmt.Printf("File %s already exists. Overwriting...\n", newPath)
		}
	}

	if !yesToAll {
		fmt.Printf("Will rename: %s -> %s\n", nameBefore, nameAfter)
	}

	// If -y flag is not set, confirm with the user
	if !yesToAll {
		fmt.Print("Apply changes? (y/n): ")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		if !strings.HasPrefix(strings.ToLower(response), "y") {
			fmt.Println("Skipping file.")
			return nil
		}
	}

	// Rename the file
	err = os.Rename(file, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	if verbose {
		fmt.Printf("Renamed: %s -> %s\n", nameBefore, nameAfter)
	}

	return nil
}

func usage(arg0 string) {
	fmt.Printf(`Usage: %s [options] <regex> <replacement> <files...>

Options:
  -y          Automatically confirm all changes without prompting.
  -p          Use POSIX regex mode for pattern matching.
  -v          Enable verbose output to display detailed information.
  -h          Show this help message and exit.

Arguments:
  <regex>       A regular expression to match file names.
  <replacement> A replacement string to rename matched files.
  <files...>    One or more file paths to process.

Examples:
  1. Rename files matching "test.*" to "example.*":
     gorename "test.*" "example.*" file1.txt file2.txt

  2. Rename files using POSIX regex mode:
     gorename -p "file[0-9]+" "document" file1.txt file2.txt

  3. Rename files without confirmation:
     gorename -y "old" "new" file1.txt file2.txt

  4. Enable verbose output:
     gorename -v "temp" "final" file1.txt file2.txt

  5. Prefix all filenames with "backup_":
	 gorename "^" "backup_" file1.txt file2.txt

  6. Rename files using capture groups:
	 gorename '^(.*)\.(.*)$' '${1}_backup.${2}' file1.txt file2.txt
`, arg0)
	os.Exit(1)
}
