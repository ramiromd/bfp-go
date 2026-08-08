package entity

// Category es una agrupación de alto nivel de los movimientos, por ejemplo
// "Servicios básicos" o "Alimentos". Cada categoría se divide a su vez en
// subcategorías.
type Category struct {
	// Id es el identificador único de la categoría, por ejemplo "services".
	Id string
	// Name es el nombre de la categoría tal como se muestra al usuario.
	Name string
	// Description explica qué tipo de movimientos agrupa la categoría.
	Description string
}

// NewCategory crea una categoría con los datos indicados.
func NewCategory(id string, name string, description string) *Category {
	return &Category{
		Id:          id,
		Name:        name,
		Description: description,
	}
}