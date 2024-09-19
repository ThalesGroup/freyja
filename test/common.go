package internal

import (
	"log"
	"os"
	"path/filepath"
)

const FreyjaUnitTestDir = "/tmp/freyja-unit-test"

func WriteTempTestFile(name string, parentDirName string, content []byte) string {
	dir := filepath.Join(FreyjaUnitTestDir, parentDirName)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Panicf("Could not create unit test dir '%s' : %v", dir, err)
	}
	path := filepath.Join(dir, name)
	err = os.WriteFile(path, content, 0660)
	if err != nil {
		log.Panicf("Could not write temp unit test file '%s' : %v", path, err)
	}
	return path
}
