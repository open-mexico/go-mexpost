package domain

// Colonia representa la estructura principal de los datos espaciales y postales.
type Colonia struct {
	Codigo      string  `json:"codigo"`
	Nombre      string  `json:"nombre"`
	Tipo        string  `json:"tipo"`
	Ciudad      string  `json:"ciudad"`
	Zona        string  `json:"zona"`
	EstadoID    string  `json:"estado_id"`
	MunicipioID string  `json:"municipio_id"`
	Geometria   *string `json:"geometria,omitempty"` // Es un puntero porque puede ser nulo (NULL en BD)
}