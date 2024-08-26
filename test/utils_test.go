package internal

import (
	"freyja/internal"
	"regexp"
	"testing"
)

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

func TestCheckIfFileExists(t *testing.T) {
	exists := "/dev/null"
	DoesntExist := "dumb"
	if err := internal.CheckIfFileExists(exists); err != nil {
		t.Logf("File '%s' detected as absent but exists : %v", exists, err)
		t.Fail()
	}
	if err := internal.CheckIfFileExists(DoesntExist); err == nil {
		t.Logf("File '%s' detected as present but does not exist", exists)
		t.Fail()
	}
}

func TestKibToGB(t *testing.T) {
	kib1 := uint64(1)
	expected1 := 0.000001024
	kib2 := uint64(4194304)
	expected2 := 4.294967296

	res1 := internal.KibToGB(kib1)
	if res1 != expected1 {
		t.Logf("Expected '%f' but got '%f'", expected1, res1)
	}

	res2 := internal.KibToGB(kib2)
	if res2 != expected2 {
		t.Logf("Expected '%f' but got '%f'", expected2, res2)
	}
}

func TestBytesToGB(t *testing.T) {
	bytes1 := uint64(2627760128)
	bytes2 := uint64(34359738368)
	expected1 := 2.4472923278808594
	expected2 := 32.0

	res1 := internal.BytesToGB(bytes1)
	if res1 != expected1 {
		t.Logf("Expected '%f' but got '%f'", expected1, res1)
	}

	res2 := internal.BytesToGB(bytes2)
	if res2 != expected2 {
		t.Logf("Expected '%f' but got '%f'", expected2, res2)
	}
}
