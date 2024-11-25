package internal

import (
	"bufio"
	"bytes"
	"errors"
	"freyja/internal"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

const FreyjaUnitTestUtilsDir = FreyjaUnitTestDir + "/utils"

func TestGenerateUUID(t *testing.T) {
	uuid := internal.GenerateUUID()
	uuidRegex := "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	match, err := regexp.MatchString(uuidRegex, uuid)
	if err != nil {
		t.Fatalf("Cannot verify regex match between '%s' and '%s'", uuid, uuidRegex)
	}
	if !match {
		t.Logf("UUID string '%s' does not match regex '%s'", uuid, uuidRegex)
		t.Fail()
	}
}

func TestKibToGiB(t *testing.T) {
	kib1 := uint64(33554432)
	expected1 := float64(32)

	res1 := internal.KibToGiB(kib1)
	if res1 != expected1 {
		t.Logf("Expected '%f' but got '%f'", expected1, res1)
		t.Fail()
	}
}

func TestMiBtoKiB(t *testing.T) {
	mib1 := uint64(4096)
	expected1 := uint64(4194304)

	res1 := internal.MiBToKiB(mib1)
	if uint64(res1) != expected1 {
		t.Logf("Expected '%d' but got '%d'", expected1, uint64(res1))
		t.Fail()
	}
}

func TestBytesToGiB(t *testing.T) {
	bytes1 := uint64(2627760128)
	bytes2 := uint64(34359738368)
	expected1 := 2.4472923278808594
	expected2 := 32.0

	res1 := internal.BytesToGiB(bytes1)
	if res1 != expected1 {
		t.Logf("Expected '%f' but got '%f'", expected1, res1)
	}

	res2 := internal.BytesToGiB(bytes2)
	if res2 != expected2 {
		t.Logf("Expected '%f' but got '%f'", expected2, res2)
	}
}

func writeBigDumbFile(t *testing.T, name string, parentDirName string) int64 {
	// init dir
	dir := filepath.Join(FreyjaUnitTestDir, parentDirName)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Panicf("Could not create unit test dir '%s' : %v", dir, err)
	}
	path := filepath.Join(dir, name)
	// random 1024 bytes (1Ko)

	// create file
	f, _ := os.Create(path)
	defer f.Close()
	w := bufio.NewWriter(f)
	// write the random 1Ko bytes content 1024 * 1024 times so the files can be 1Go big
	//for i := 0; i <= 1<<10; i++ {
	for i := 0; i <= 1000; i++ {
		_, err = w.WriteString("hello world !")
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

func TestCopyFile(t *testing.T) {
	// small string file
	name := "test.txt"
	textSourceFile := filepath.Join(FreyjaUnitTestUtilsDir + "/" + name)
	textDestinationFile := filepath.Join(FreyjaUnitTestUtilsDir + "/" + "copy-" + name)
	WriteTempTestFile(name, "utils", []byte("hello world !"))
	copyFileTest(t, textSourceFile, textDestinationFile)
	// large random bytes file
	name = "test.bin"
	binSourceFile := filepath.Join(FreyjaUnitTestUtilsDir + "/" + name)
	binDestinationFile := filepath.Join(FreyjaUnitTestUtilsDir + "/" + "copy-" + name)
	WriteRandomBytesTempFile(t, name, "utils")
	copyFileTest(t, binSourceFile, binDestinationFile)
}

func copyFileTest(t *testing.T, sourceFile string, destinationFile string) {
	// init

	// exec copy
	if err := internal.CopyFile(sourceFile, destinationFile, 0700); err != nil {
		t.Logf("copy failed, reason: %v", err)
		t.FailNow()
	}
	// check if file exists
	if _, err := os.Stat(destinationFile); errors.Is(err, os.ErrNotExist) {
		t.Logf("copy file does not exist, reason: %v", err)
		t.FailNow()
	}
	// compare sizes
	fs, err := os.Stat(sourceFile)
	if err != nil {
		t.Logf("cannot get source file stats: %v", err)
		t.FailNow()
	}
	fd, err := os.Stat(destinationFile)
	if err != nil {
		t.Logf("cannot get copy file stats: %v", err)
		t.FailNow()
	}
	if fs.Size() != fd.Size() {
		t.Logf("copy file size of '%d' do not match source file size of '%d'", fd.Size(), fs.Size())
		t.FailNow()
	}
	// compare content
	sourceBytes, err := os.ReadFile(destinationFile)
	if err != nil {
		t.Logf("cannot read source file, reason: %v", err)
		t.FailNow()
	}
	destinationBytes, err := os.ReadFile(destinationFile)
	if err != nil {
		t.Logf("cannot read copy file, reason: %v", err)
		t.FailNow()
	}
	if bytes.Compare(sourceBytes, destinationBytes) != 0 {
		t.Logf("copy file content and source file content do not match")
		t.FailNow()
	}
}
