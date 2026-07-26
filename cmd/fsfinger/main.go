package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ccxn/filesystemfinger/fingerprint"
)

type patternsFlag []string

func (p *patternsFlag) String() string { return fmt.Sprint([]string(*p)) }
func (p *patternsFlag) Set(value string) error {
	*p = append(*p, value)
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stderr)
		return errors.New("a command is required")
	}
	switch args[0] {
	case "scan":
		return scan(args[1:], stdout, stderr)
	case "verify":
		return verify(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		usage(stdout)
		return nil
	default:
		usage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func scan(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var patterns patternsFlag
	ignoreFile := flags.String("ignore-file", "", "read ignore patterns from this file")
	noDefaultIgnore := flags.Bool("no-default-ignore", false, "disable built-in temporary-file ignores")
	hashOnly := flags.Bool("hash-only", false, "print only the root hash")
	flags.Var(&patterns, "ignore", "ignore pattern (repeatable)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 1 {
		return errors.New("scan accepts at most one path")
	}
	path := "."
	if flags.NArg() == 1 {
		path = flags.Arg(0)
	}
	manifest, err := fingerprint.Scan(path, fingerprint.Options{
		IgnorePatterns: patterns,
		IgnoreFile:     *ignoreFile,
		UseDefaults:    !*noDefaultIgnore,
		IncludeEntries: !*hashOnly,
	})
	if err != nil {
		return err
	}
	if *hashOnly {
		fmt.Fprintln(stdout, manifest.RootHash)
		return nil
	}
	return writeJSON(stdout, manifest)
}

func verify(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "", "manifest JSON created by scan (required)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" {
		return errors.New("--manifest is required")
	}
	if flags.NArg() > 1 {
		return errors.New("verify accepts at most one path")
	}
	path := "."
	if flags.NArg() == 1 {
		path = flags.Arg(0)
	}
	file, err := os.Open(*manifestPath)
	if err != nil {
		return fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()
	var manifest fingerprint.Manifest
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	result, err := fingerprint.Verify(manifest, path, fingerprint.Options{})
	if err != nil {
		return err
	}
	if err := writeJSON(stdout, result); err != nil {
		return err
	}
	if !result.Match {
		return errors.New("fingerprint mismatch")
	}
	return nil
}

func writeJSON(output io.Writer, value any) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func usage(output io.Writer) {
	fmt.Fprintln(output, `fsfinger - deterministic SHA-256 filesystem fingerprints

Usage:
  fsfinger scan [options] [path]
  fsfinger verify --manifest manifest.json [path]

Run "fsfinger scan -h" for scan options.`)
}
