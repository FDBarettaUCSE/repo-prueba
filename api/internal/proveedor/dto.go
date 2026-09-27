package proveedor

type ProveedorDTO struct {
	ID       string `json:"id"`
	Nombre   string `json:"nombre"`
	CUIT     string `json:"cuit"`
	Email    string `json:"email"`
	Telefono string `json:"telefono"`
}

type CrearProveedorDTO struct {
	Nombre   string `json:"nombre"`
	CUIT     string `json:"cuit"`
	Email    string `json:"email"`
	Telefono string `json:"telefono"`
}

type ActualizarProveedorDTO struct {
	Nombre   string `json:"nombre"`
	CUIT     string `json:"cuit"`
	Email    string `json:"email"`
	Telefono string `json:"telefono"`
}

func MapeoCrearDTOAProveedor(d CrearProveedorDTO) Proveedor {
	return Proveedor{
		Nombre:   d.Nombre,
		CUIT:     d.CUIT,
		Email:    d.Email,
		Telefono: d.Telefono,
	}
}

func MapeoActualizarDTOAProveedor(d ActualizarProveedorDTO) Proveedor {
	return Proveedor{
		Nombre:   d.Nombre,
		CUIT:     d.CUIT,
		Email:    d.Email,
		Telefono: d.Telefono,
	}
}

// --- Productos que suministra este proveedor ---
// Cada proveedor puede suministrar el mismo producto que otro, pero con su
// propio costo y tiempo de entrega (ver api/internal/productoProveedor).

type AgregarProductoSuministradoDTO struct {
	ProductoID        string  `json:"productoId"`
	Costo             float64 `json:"costo"`
	TiempoEntregaDias int     `json:"tiempoEntregaDias"`
}

type ActualizarProductoSuministradoDTO struct {
	Costo             float64 `json:"costo"`
	TiempoEntregaDias int     `json:"tiempoEntregaDias"`
}

// ProductoSuministradoDTO es la vista enriquecida: además del costo/tiempo
// propios del vínculo, incluye el nombre del producto (para no obligar al
// frontend a pedirlo aparte a /productos por cada ítem de la lista).
type ProductoSuministradoDTO struct {
	ProductoID        string  `json:"productoId"`
	ProductoNombre    string  `json:"productoNombre"`
	Costo             float64 `json:"costo"`
	TiempoEntregaDias int     `json:"tiempoEntregaDias"`
}

// ProveedorDeProductoDTO es la vista inversa: para un producto puntual, qué
// proveedores lo suministran y en qué condiciones — es lo que va a necesitar
// el módulo de Órdenes de Compra para elegir a quién pedirle (por ejemplo, al
// de menor costo).
type ProveedorDeProductoDTO struct {
	ProveedorID       string  `json:"proveedorId"`
	ProveedorNombre   string  `json:"proveedorNombre"`
	Costo             float64 `json:"costo"`
	TiempoEntregaDias int     `json:"tiempoEntregaDias"`
}
