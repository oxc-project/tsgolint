package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateIsDeterministic(t *testing.T) {
	t.Parallel()

	inputDirectory := t.TempDir()
	files := map[string]string{
		"lib.d.ts":        "/// <reference no-default-lib=\"true\"/>\n",
		"lib.es2025.d.ts": "interface Example {}\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(inputDirectory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(inputDirectory, "ignored.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}

	firstOutput := filepath.Join(t.TempDir(), "first.zip")
	secondOutput := filepath.Join(t.TempDir(), "second.zip")
	if err := generate(inputDirectory, firstOutput); err != nil {
		t.Fatal(err)
	}
	if err := generate(inputDirectory, secondOutput); err != nil {
		t.Fatal(err)
	}

	first, err := os.ReadFile(firstOutput)
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(secondOutput)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("generated archives differ")
	}

	reader, err := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if err != nil {
		t.Fatal(err)
	}
	if len(reader.File) != len(files) {
		t.Fatalf("archive contains %d entries, want %d", len(reader.File), len(files))
	}
	for _, file := range reader.File {
		want, ok := files[filepath.Base(file.Name)]
		if !ok {
			t.Fatalf("unexpected archive entry %q", file.Name)
		}
		entry, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, err := io.ReadAll(entry)
		if err != nil {
			t.Fatal(err)
		}
		if err := entry.Close(); err != nil {
			t.Fatal(err)
		}
		if string(contents) != want {
			t.Fatalf("contents of %q differ", file.Name)
		}
	}
}

func TestGenerateRejectsEmptyInput(t *testing.T) {
	t.Parallel()

	err := generate(t.TempDir(), filepath.Join(t.TempDir(), "libs.zip"))
	if err == nil {
		t.Fatal("generate succeeded without any .d.ts files")
	}
}
