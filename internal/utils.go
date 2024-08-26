package internal

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math"
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

func KibToGB(kib uint64) float64 {
	return float64(kib) * 1.024 * math.Pow10(-6)
}

func BytesToGB(by uint64) float64 {
	return float64(by) / math.Pow(2.0, 30)
}

// Truncate truncates a float into a certain decimals after the first one. Ex :
// Truncate(1.123456789, 2) = 1.12
// Truncate(1.123456789, 4) = 1.1234
func Truncate(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	round := int(num + math.Copysign(0.5, num))
	return float64(round) / output
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
