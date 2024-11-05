package internal

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func CheckIfFileExists(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// KibToGiB
// byte to kb = float(n)/(1<<10)
// byte to mb = float(n)/(1<<20)
// byte to gb = float(n)/(1<<30)
// 1<<10 = 2^10. = 1024
// 1<<20 = 2^20 = 1024*1024
// ...
func KibToGiB(kib uint64) float64 {
	return float64(kib) / (1 << 20)
}

func MiBToKiB(GiB uint64) float64 {
	return float64(GiB) * (1 << 10)
}

func BytesToGiB(by uint64) float64 {
	return float64(by) / (1 << 30)
}

// AskUserYesNoConfirmation returns true if user confirmed 'yes', false otherwise.
func AskUserYesNoConfirmation() (bool, error) {
	//Display what you want from user
	fmt.Print("Are you sure ? [Y/n]: ")
	//Declare variable to store input
	var choice string
	//Take input from user
	itemsSCanned, err := fmt.Scan(&choice)
	if itemsSCanned != 1 {
		return false, errors.New("More than one input from user were submitted")
	}
	if err != nil {
		return false, err
	}
	// analyse choice
	switch choice {
	case "Y":
		return true, nil
	case "n":
		return false, nil
	default:
		return false, &UserInputError{Message: fmt.Sprintf("unexpected user input: %s", choice)}
	}

}

func EncodeB64Bytes(b []byte) string {
	return b64.StdEncoding.EncodeToString(b)
}

func CopyFile(source string, destination string, destinationPermissions os.FileMode) error {
	// read
	sourceContent, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("cannot read source file '%s': %w", source, err)
	}
	// write
	if err = os.WriteFile(destination, sourceContent, destinationPermissions); err != nil {
		return fmt.Errorf("cannot write copy file '%s': %w", destination, err)
	}
	return nil
}
