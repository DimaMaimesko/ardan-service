// This program creates a new service project from the template that ships
// with the scaffold-ardan-service skill. It copies the template, replaces the
// placeholder names, formats the Go code, and generates a development signing
// key. It uses only the standard library so it runs with `go run` from any
// directory:
//
//	go run <skill>/scripts/scaffold/main.go -module github.com/acme/orders -name orders -out ../orders
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Placeholders used throughout the template. The module placeholder must be
// replaced before the name placeholder because it contains it.
const (
	placeholderModule = "github.com/tmplorg/tmplsvc"
	placeholderPrefix = "TMPLSVC"
	placeholderName   = "tmplsvc"
	placeholderKID    = "TMPLKID"
	templateSuffix    = ".tmpl"
)

var (
	moduleRegEx = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._~-]*(/[a-zA-Z0-9._~-]+)*$`)
	nameRegEx   = regexp.MustCompile(`^[a-z][a-z0-9-]{0,38}[a-z0-9]$`)
	prefixRegEx = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,39}$`)
)

type config struct {
	module      string
	name        string
	prefix      string
	out         string
	templateDir string
	kid         string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "scaffold:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := parseFlags()
	if err != nil {
		return err
	}

	if err := checkOut(cfg.out); err != nil {
		return err
	}

	n, err := copyTemplate(cfg)
	if err != nil {
		return fmt.Errorf("copy template: %w", err)
	}

	if err := writeDevKey(cfg); err != nil {
		return fmt.Errorf("write dev key: %w", err)
	}

	fmt.Printf("created %s: %d files\n", cfg.out, n)
	fmt.Printf("module:     %s\n", cfg.module)
	fmt.Printf("service:    %s\n", cfg.name)
	fmt.Printf("env prefix: %s_\n", cfg.prefix)
	fmt.Printf("dev key:    zarf/keys/%s.pem\n", cfg.kid)
	fmt.Println()
	fmt.Println("next steps:")
	fmt.Printf("  cd %s\n", cfg.out)
	fmt.Println("  go mod tidy && go mod vendor")
	fmt.Println("  go build ./... && go vet ./...")
	fmt.Println("  make test-only    # needs Docker")

	return nil
}

func parseFlags() (config, error) {
	var cfg config

	flag.StringVar(&cfg.module, "module", "", "Go module path for the new project, e.g. github.com/acme/orders (required)")
	flag.StringVar(&cfg.name, "name", "", "service name: lowercase letters, digits and hyphens, e.g. orders (required)")
	flag.StringVar(&cfg.prefix, "prefix", "", "environment variable prefix (default: the name in upper case, hyphens as underscores)")
	flag.StringVar(&cfg.out, "out", "", "directory to create; it must not exist or must be empty (required)")
	flag.StringVar(&cfg.templateDir, "template", defaultTemplateDir(), "template directory")
	flag.StringVar(&cfg.kid, "kid", "", "key id for the dev signing key (default: a random UUID)")
	flag.Parse()

	var errs []string

	if !moduleRegEx.MatchString(cfg.module) {
		errs = append(errs, fmt.Sprintf("-module %q is not a valid module path", cfg.module))
	}

	if !nameRegEx.MatchString(cfg.name) {
		errs = append(errs, fmt.Sprintf("-name %q must be 2-40 characters of lowercase letters, digits and hyphens, starting with a letter", cfg.name))
	}

	if cfg.prefix == "" {
		cfg.prefix = strings.ToUpper(strings.ReplaceAll(cfg.name, "-", "_"))
	}

	if !prefixRegEx.MatchString(cfg.prefix) {
		errs = append(errs, fmt.Sprintf("-prefix %q must be upper case letters, digits and underscores, starting with a letter", cfg.prefix))
	}

	for _, v := range []string{cfg.module, cfg.name, cfg.prefix} {
		if containsPlaceholder(v) {
			errs = append(errs, fmt.Sprintf("%q must not contain a template placeholder (%s, %s, %s)", v, placeholderName, placeholderPrefix, placeholderKID))
		}
	}

	if cfg.out == "" {
		errs = append(errs, "-out is required")
	}

	if cfg.kid == "" {
		kid, err := newUUID()
		if err != nil {
			return config{}, fmt.Errorf("generate kid: %w", err)
		}
		cfg.kid = kid
	}

	if len(errs) > 0 {
		flag.Usage()
		return config{}, errors.New(strings.Join(errs, "\n  "))
	}

	out, err := filepath.Abs(cfg.out)
	if err != nil {
		return config{}, fmt.Errorf("resolve -out: %w", err)
	}
	cfg.out = out

	return cfg, nil
}

// defaultTemplateDir finds the template folder relative to this source file,
// which works for `go run` from any working directory.
func defaultTemplateDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "template")
}

func containsPlaceholder(v string) bool {
	lower := strings.ToLower(v)
	return strings.Contains(lower, placeholderName) || strings.Contains(lower, strings.ToLower(placeholderKID))
}

// checkOut refuses to write into a directory that already has content, so
// an existing project is never overwritten.
func checkOut(out string) error {
	entries, err := os.ReadDir(out)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("read -out: %w", err)
	case len(entries) > 0:
		return fmt.Errorf("-out %s is not empty", out)
	}

	return nil
}

func copyTemplate(cfg config) (int, error) {
	info, err := os.Stat(cfg.templateDir)
	if err != nil || !info.IsDir() {
		return 0, fmt.Errorf("template directory %q not found; pass -template", cfg.templateDir)
	}

	replacer := strings.NewReplacer(
		placeholderModule, cfg.module,
		placeholderPrefix, cfg.prefix,
		placeholderName, cfg.name,
		placeholderKID, cfg.kid,
	)

	var count int

	fn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(cfg.templateDir, path)
		if err != nil {
			return err
		}

		dest := filepath.Join(cfg.out, strings.ReplaceAll(strings.TrimSuffix(rel, templateSuffix), placeholderName, cfg.name))

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		data = []byte(replacer.Replace(string(data)))

		// Import blocks are sorted by path, so a new module path can change
		// their order. Formatting keeps the output gofmt clean.
		if filepath.Ext(dest) == ".go" {
			data, err = format.Source(data)
			if err != nil {
				return fmt.Errorf("format %s: %w", rel, err)
			}
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(dest, data, fi.Mode().Perm()); err != nil {
			return err
		}

		count++

		return nil
	}

	if err := filepath.WalkDir(cfg.templateDir, fn); err != nil {
		return count, err
	}

	return count, nil
}

// writeDevKey generates the RSA key the service uses to sign tokens in
// development. The file is world readable so the non-root container user can
// read it through the Compose bind mount; it must never be used in production.
func writeDevKey(cfg config) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "PRIVATE KEY", Bytes: der}); err != nil {
		return fmt.Errorf("encode key: %w", err)
	}

	dir := filepath.Join(cfg.out, "zarf", "keys")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, cfg.kid+".pem"), buf.Bytes(), 0o644)
}

// newUUID returns a random version 4 UUID.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
