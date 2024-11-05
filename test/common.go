package internal

import (
	"bufio"
	"crypto/rand"
	"log"
	"os"
	"path/filepath"
	"testing"
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

func WriteRandomBytesTempFile(t *testing.T, name string, parentDirName string) int64 {
	// init dir
	dir := filepath.Join(FreyjaUnitTestDir, parentDirName)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Panicf("Could not create unit test dir '%s' : %v", dir, err)
	}
	path := filepath.Join(dir, name)
	// random 1024 bytes (1Ko)
	buf := make([]byte, 1<<10)
	rand.Read(buf)
	// create file
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	w := bufio.NewWriter(f)
	// write the random 1Ko bytes content 1024 * 1024 times so the files can be 1Go big
	for i := 0; i <= 1<<10; i++ {
		_, err = w.Write(buf)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		w.Flush()
	}
	stats, err := f.Stat()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	return stats.Size()
}
