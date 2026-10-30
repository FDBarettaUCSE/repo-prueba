package reporte

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	movimientostock "api/internal/movimientoStock"
)

// TamanioPagina es fijo según el criterio utilizado en la guía
// de paginación de la cátedra.
const TamanioPagina = 10

// =========================
// Filtros internos
// =========================

type FiltroInventario struct {
	DepositoID  *bson.ObjectID
	CategoriaID *bson.ObjectID
}

type FiltroReporteMovimientos struct {
	DepositoID *bson.ObjectID
	Tipo       *movimientostock.TipoMovimiento
	FechaDesde *time.Time
	FechaHasta *time.Time
}

type FiltroAlertas struct {
	DepositoID      *bson.ObjectID
	DiasVencimiento int
}

type FiltroPendientes struct {
	DepositoID *bson.ObjectID
}

// =========================
// Resultados internos
// =========================

type ResultadoInventarioDeposito struct {
	DepositoID     bson.ObjectID `bson:"deposito_id"`
	DepositoNombre string        `bson:"deposito_nombre"`
	Stock          int32         `bson:"stock"`
	Valorizado     float64       `bson:"valorizado"`
}

type ResultadoInventarioCategoria struct {
	CategoriaID     bson.ObjectID `bson:"categoria_id"`
	CategoriaNombre string        `bson:"categoria_nombre"`
	Stock           int32         `bson:"stock"`
	Valorizado      float64       `bson:"valorizado"`
}

type ResultadoMovimiento struct {
	DepositoID      bson.ObjectID `bson:"deposito_id"`
	DepositoNombre  string        `bson:"deposito_nombre"`
	Fecha           string        `bson:"fecha"`
	Tipo            string        `bson:"tipo"`
	CantidadEventos int64         `bson:"cantidad_eventos"`
	UnidadesNetas   int32         `bson:"unidades_netas"`
}

type ResultadoAlertaMinimo struct {
	DepositoID     bson.ObjectID `bson:"deposito_id"`
	DepositoNombre string        `bson:"deposito_nombre"`
	ProductoID     bson.ObjectID `bson:"producto_id"`
	ProductoNombre string        `bson:"producto_nombre"`
	Stock          int32         `bson:"stock"`
	StockMinimo    int32         `bson:"stock_minimo"`
}

type ResultadoAlertaVencimiento struct {
	DepositoID       bson.ObjectID `bson:"deposito_id"`
	DepositoNombre   string        `bson:"deposito_nombre"`
	ProductoID       bson.ObjectID `bson:"producto_id"`
	ProductoNombre   string        `bson:"producto_nombre"`
	Stock            int32         `bson:"stock"`
	FechaVencimiento time.Time     `bson:"fecha_vencimiento"`
	DiasParaVencer   int           `bson:"dias_para_vencer"`
}

type ResultadoOrdenesDeposito struct {
	DepositoID     bson.ObjectID `bson:"deposito_id"`
	DepositoNombre string        `bson:"deposito_nombre"`
	Borrador       int64         `bson:"borrador"`
	Confirmada     int64         `bson:"confirmada"`
	Parcial        int64         `bson:"parcial"`
	Completada     int64         `bson:"completada"`
}

type ResultadoTransferenciasDeposito struct {
	DepositoID            bson.ObjectID `bson:"deposito_id"`
	DepositoNombre        string        `bson:"deposito_nombre"`
	SolicitadasComoOrigen int64         `bson:"solicitadas_como_origen"`
	PendientesDeAprobar   int64         `bson:"pendientes_de_aprobar"`
	AprobadasPorRecibir   int64         `bson:"aprobadas_por_recibir"`
}

func PaginacionAPaginaDTO(
	pagina int,
	total int64,
) PaginacionDTO {
	return PaginacionDTO{
		Pagina:        pagina,
		TamanioPagina: TamanioPagina,
		Total:         total,
	}
}
