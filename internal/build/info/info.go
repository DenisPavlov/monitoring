package info

import (
	"fmt"
)

func PrintBuildInfo(buildVersion, buildDate, buildCommit string) {
	fmt.Printf(`Build version: %s
Build date: %s
Build commit: %s
`, buildVersion, buildDate, buildCommit)
}
