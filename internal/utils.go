package internal

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
)

// Prompt

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

// ID

func GenerateUUID() string {
	return uuid.New().String()
}

// Conversion

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

// Encoding

func EncodeB64Bytes(b []byte) string {
	return b64.StdEncoding.EncodeToString(b)
}

// Files

func FileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// RemoveIfExists removes a file if it exists.
// Return true if the file was actually removed.
// Return false if no file existed within the given path/
// Return an error if the deletion failed.
func RemoveIfExists(path string) (deleted bool, err error) {
	if FileExists(path) {
		if err = os.Remove(path); err != nil {
			return false, fmt.Errorf("couldn't remove '%s': %w", path, err)
		}
		return true, nil
	}
	Logger.Debug("not removed : doesn't exist", "path", path)
	return false, nil
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
