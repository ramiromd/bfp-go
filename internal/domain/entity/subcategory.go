package entity

// Subcategory es una división de una categoría, por ejemplo "Electricidad"
// dentro de "Servicios básicos". Es el nivel con el que se clasifican los
// movimientos.
type Subcategory struct {
	// Id es el identificador único de la subcategoría, por ejemplo "light".
	Id string
	// CategoryId es el identificador de la categoría a la que pertenece.
	CategoryId string
	// Name es el nombre de la subcategoría tal como se muestra al usuario.
	Name string
	// Description explica qué tipo de movimientos agrupa la subcategoría.
	Description string
}

// NewSubcategory crea una subcategoría asociada a la categoría indicada.
func NewSubcategory(id string, categoryId string, name string, description string) *Subcategory {
	return &Subcategory{
		Id:          id,
		CategoryId:  categoryId,
		Name:        name,
		Description: description,
	}
}