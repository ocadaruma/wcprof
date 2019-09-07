package wcprof

import (
	"flag"
	"os"
)

func main() {
	backup := flag.Bool("backup", false, "create backup")
	path := flag.String("path", "", "(mandatory) directory to be processed")

	flag.Parse()

	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	config := Config{Backup: *backup}

	InterceptTimer(*path, &config)
}
