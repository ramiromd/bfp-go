package application

// ExportMovementsOutput contiene el resultado de exportar un archivo de
// movimientos bancarios a un formato estandarizado.
type ExportMovementsRes struct {
	// ReadRecords es la cantidad de registros leídos del archivo de entrada.
	ReadRecords int
	// ExportedRecords es la cantidad de registros exportados con éxito.
	ExportedRecords int
	// FailedRecords es la cantidad de registros que no se pudieron exportar.
	FailedRecords int
	// OutputPath es la ruta del archivo de salida generado.
	OutputPath string
}
