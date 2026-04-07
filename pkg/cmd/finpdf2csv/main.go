package main

import (
	"os"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdf2csvcli"
)

func main() {
	os.Exit(pdf2csvcli.Run(os.Args[1:], os.Stderr))
}
