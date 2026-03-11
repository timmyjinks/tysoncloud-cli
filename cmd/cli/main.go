package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	uppercase bool
	lowercase bool
	titlecase bool
	wordcount bool
	charcount bool
)

var rootCmd = &cobra.Command{
	Use:   "textfmt [text]",
	Short: "A simple text formatter",
	Long: `textfmt is a CLI tool for formatting and analyzing text.

You can provide text as arguments or pipe it in via stdin.
Multiple formatting options can be applied simultaneously.`,
	Args: cobra.ArbitraryArgs,
	Run:  runTextFormatter,
}

func init() {
	rootCmd.Flags().BoolVarP(&uppercase, "upper", "u", false, "Convert text to uppercase")
	rootCmd.Flags().BoolVarP(&lowercase, "lower", "l", false, "Convert text to lowercase")
	rootCmd.Flags().BoolVarP(&titlecase, "title", "t", false, "Convert text to title case")
	rootCmd.Flags().BoolVar(&wordcount, "words", false, "Count words in text")
	rootCmd.Flags().BoolVar(&charcount, "chars", false, "Count characters in text")
}

func runTextFormatter(cmd *cobra.Command, args []string) {
	var text string

	// Get input text from args or stdin
	if len(args) > 0 {
		text = strings.Join(args, " ")
	} else {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		text = strings.Join(lines, "\n")
	}

	if text == "" {
		fmt.Println("No input text provided")
		return
	}

	result := text

	// Apply formatting transformations
	if uppercase {
		result = strings.ToUpper(result)
	}
	if lowercase {
		result = strings.ToLower(result)
	}
	if titlecase {
		result = strings.Title(result)
	}

	// Output the formatted text
	fmt.Println(result)

	// Show analysis if requested
	if wordcount {
		words := len(strings.Fields(text))
		fmt.Printf("Words: %d\n", words)
	}
	if charcount {
		chars := len(text)
		fmt.Printf("Characters: %d\n", chars)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
