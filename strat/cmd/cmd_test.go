package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectFile(t *testing.T) {
	dir1, err := ioutil.TempDir("", "cmd")
	defer os.RemoveAll(dir1)
	if err != nil {
		t.Fatal(err)
	}

	// Find in current directory
	want := filepath.Join(dir1, ProjectFile)
	if err := ioutil.WriteFile(want, []byte{}, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	got, err := findProjectFile(dir1)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("findProjectFile(): Found project file is incorrect, got %v, expected %v", got, want)
	}

	// Find in parent directory
	dir2, err := ioutil.TempDir(dir1, "cmd")
	if err != nil {
		t.Fatal(err)
	}
	got, err = findProjectFile(dir2)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("findProjectFile(): Found project file is incorrect, got %v, expected %v", got, want)
	}

	// Not found
	dir3, err := ioutil.TempDir("", "cmd")
	defer os.RemoveAll(dir3)
	if err != nil {
		t.Fatal(err)
	}
	_, err = findProjectFile(dir3)
	if err == nil {
		t.Errorf("findProjectFile(): Expected error not to be nil")
	}
	if got, want = err.Error(), fmt.Sprintf("Could not find %s in %s and its parent directories", ProjectFile, dir3); got != want {
		t.Errorf("findProjectFile(): Wrong error message, got %s expected %s", got, want)
	}
}
