package application

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"ramiromd/budget/internal/domain/service"
)

// ExportMovements es el caso de uso que lee un archivo de movimientos
// bancarios y lo exporta a un formato estandarizado.
type ExportMovements struct {
	movementConverterFactory *service.MovementConverterFactory
}

// NewExportMovements crea una instancia del caso de uso.
func NewExportMovements() *ExportMovements {
	return &ExportMovements{
		movementConverterFactory: service.NewMovementConverterFactory(),
	}
}

// Execute procesa el archivo indicado en input y devuelve los movimientos
// normalizados.
func (uc *ExportMovements) Execute(input ExportMovementsReq) (ExportMovementsRes, error) {

	movementConverter, err := uc.movementConverterFactory.GetConverter(input.Bank, input.FileType)
	if err != nil {
		return ExportMovementsRes{}, err
	}

	file, err := os.Open(input.InputPath)
	if err != nil {
		return ExportMovementsRes{}, fmt.Errorf("no se pudo abrir el archivo %q: %w", input.InputPath, err)
	}
	defer file.Close()

	result := ExportMovementsRes{}

	// El archivo se lee a nivel de registro y no de línea, porque un mismo
	// registro puede ocupar varias líneas cuando alguno de sus campos está
	// entrecomillado (por ejemplo, la descripción de un movimiento).
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.TrimLeadingSpace = true
	// La cantidad de columnas puede variar entre registros; la validación queda
	// a cargo del conversor de cada banco.
	reader.FieldsPerRecord = -1

	if input.HasHeaders {
		if _, err := reader.Read(); err != nil {
			if errors.Is(err, io.EOF) {
				return result, nil
			}
			return ExportMovementsRes{}, fmt.Errorf("error al leer las cabeceras del archivo %q: %w", input.InputPath, err)
		}
	}

	// Se recorre el archivo registro a registro y se delega la conversión de
	// cada uno en el conversor correspondiente al banco.
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		result.ReadRecords++

		// Un registro mal formado no interrumpe el procesamiento del resto del
		// archivo: se contabiliza como fallido y se continúa.
		if err != nil {
			result.FailedRecords++
			continue
		}

		movement := movementConverter.Convert(record)
		if movement == nil {
			result.FailedRecords++
			continue
		}

		result.ExportedRecords++
		fmt.Printf("Movimiento: %v\n", *movement)
	}

	return result, nil
}
