package main

import (
	"fmt"
	"os"
	"strconv"

	"ramiromd/budget/internal/application"
)

// usage imprime el modo de uso del comando.
func usage() {
	fmt.Fprintf(os.Stderr, "uso: %s <archivo> <tipo de archivo: account|card> <banco> <tiene cabeceras: true|false>\n", os.Args[0])
}

func main() {

	args := os.Args[1:]
	if len(args) != 4 {
		usage()
		os.Exit(1)
	}

	hasHeaders, err := strconv.ParseBool(args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: valor inválido para <tiene cabeceras>: %q\n", args[3])
		usage()
		os.Exit(1)
	}

	input := application.ExportMovementsReq{
		InputPath:  args[0],
		FileType:   args[1],
		Bank:       args[2],
		HasHeaders: hasHeaders,
	}

	useCase := application.NewExportMovements()
	movements, err := useCase.Execute(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf("Movimientos procesados: %v\n", movements)
}
