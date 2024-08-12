package internal

import (
	"os"
	"path/filepath"
)

var FreyjaWorkspaceDir = filepath.Join(os.Getenv("HOME"), "freyja-workspace")
