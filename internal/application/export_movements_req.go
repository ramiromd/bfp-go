package application

// ExportMovementsReq contiene los datos necesarios para exportar un archivo
// de movimientos bancarios a un formato estandarizado.
type ExportMovementsReq struct {
	// InputPath es la ruta del archivo de movimientos a procesar.
	InputPath string
	// FileType indica si el archivo es un extracto de cuenta ("account") o un
	// resumen de tarjeta ("card").
	FileType string
	// Bank indica la entidad bancaria que emitió el archivo.
	Bank string
	// HasHeaders indica si la primera línea del archivo contiene los nombres de
	// las columnas.
	HasHeaders bool
}
