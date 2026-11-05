package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var zipEpoch = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

func main() {
	input := flag.String("input", "typescript-go/internal/bundled/libs", "directory containing TypeScript library declarations")
	output := flag.String("output", "internal/bundled/libs.zip", "generated archive path")
	flag.Parse()

	if err := generate(*input, *output); err != nil {
		fmt.Fprintf(os.Stderr, "generate bundled libraries: %v\n", err)
		os.Exit(1)
	}
}

func generate(inputDirectory string, outputPath string) error {
	entries, err := os.ReadDir(inputDirectory)
	if err != nil {
		return fmt.Errorf("read input directory: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".d.ts") {
			continue
		}
		names = append(names, entry.Name())
	}
	if len(names) == 0 {
		return fmt.Errorf("no .d.ts files found in %q", inputDirectory)
	}
	slices.Sort(names)

	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	zw.RegisterCompressor(zip.Deflate, func(writer io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(writer, flate.BestCompression)
	})

	for _, name := range names {
		contents, err := os.ReadFile(filepath.Join(inputDirectory, name))
		if err != nil {
			return fmt.Errorf("read %q: %w", name, err)
		}

		header := &zip.FileHeader{
			Name:     "libs/" + name,
			Method:   zip.Deflate,
			Modified: zipEpoch,
		}
		header.SetMode(0o644)
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create archive entry %q: %w", name, err)
		}
		if _, err := writer.Write(contents); err != nil {
			return fmt.Errorf("write archive entry %q: %w", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("close archive: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, archive.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write archive: %w", err)
	}
	return nil
}
