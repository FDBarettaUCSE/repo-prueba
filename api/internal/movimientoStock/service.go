package movimientostock

import (
	"context"
	"errors"
	"log"
	"time"

	alertastock "api/internal/alertaStock"
	"api/internal/auditoria"
	"api/internal/deposito"
	depositoproducto "api/internal/depositoProducto"
	"api/internal/producto"

	evaluacionstock "api/internal/evaluacionStock"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const MaxTamanioPagina = 100

var (
	ErrStockInsuficiente      = errors.New("el stock disponible es insuficiente para realizar esta operación")
	ErrNoAutorizadoDeposito   = errors.New("no tenés autorización para operar sobre este depósito")
	ErrTipoMovimientoInvalido = errors.New("tipo de movimiento no admitido para registro manual")
	ErrCantidadInvalida       = errors.New("la cantidad debe ser mayor a cero")
	ErrMovimientoInexistente  = errors.New("el movimiento solicitado no existe")
	ErrIDMovimientoInvalido   = errors.New("el id del movimiento no es válido")
	ErrProductoInexistente    = errors.New("el producto indicado no existe")
	ErrDepositoInexistente    = errors.New("el depósito indicado no existe")
)

type Service struct {
	repo              MovimientoStockRepositorio
	depositoStockRepo depositoproducto.DepositoProductoRepositorio
	productoRepo      producto.ProductoRepositorio
	depositoRepo      deposito.DepositoRepositorio
	alertaStockRepo   alertastock.AlertaStockRepositorio

	// Componente para evaluar alertas. Necesario porque caeríamos en dependencias circulares.
	evaluadorStock evaluacionstock.EvaluadorStock

	// Funcion callback. Necesario porque sino tendriamos dependencias circulares.
	generarSugerenciasStock func(context.Context, string, string)
}

func NuevoService(
	repo MovimientoStockRepositorio,
	stockRepo depositoproducto.DepositoProductoRepositorio,
	productoRepo producto.ProductoRepositorio,
	depositoRepo deposito.DepositoRepositorio,
	alertaRepo alertastock.AlertaStockRepositorio,
	evaluadorStock evaluacionstock.EvaluadorStock,
	generarSugerenciasStock func(context.Context, string, string),
) *Service {
	return &Service{
		repo:                    repo,
		depositoStockRepo:       stockRepo,
		productoRepo:            productoRepo,
		depositoRepo:            depositoRepo,
		alertaStockRepo:         alertaRepo,
		evaluadorStock:          evaluadorStock,
		generarSugerenciasStock: generarSugerenciasStock,
	}
}

// Registrar aplica el cambio en el inventario y crea el registro inmutable de auditoría.
func (s *Service) Registrar(
	ctx context.Context,
	usuarioID, rol, depUsuario string,
	dto RegistrarMovimientoDTO,
) (MovimientoStockDTO, error) {
	// Operarios y Gerentes solo pueden operar en su depósito asignado.
	if (rol == "operario" || rol == "gerente") && depUsuario != dto.DepositoID {
		return MovimientoStockDTO{}, ErrNoAutorizadoDeposito
	}

	depoOID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return MovimientoStockDTO{}, err
	}

	prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return MovimientoStockDTO{}, err
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return MovimientoStockDTO{}, err
	}

	var delta int32

	switch dto.Tipo {
	case Ingreso:
		if dto.Cantidad <= 0 {
			return MovimientoStockDTO{}, ErrCantidadInvalida
		}
		delta = dto.Cantidad

	case Egreso:
		if dto.Cantidad <= 0 {
			return MovimientoStockDTO{}, ErrCantidadInvalida
		}
		delta = -dto.Cantidad

	case Ajuste:
		// Permite ajuste positivo (aumento de inventario) o negativo (merma/faltante).
		delta = dto.Cantidad

	default:
		return MovimientoStockDTO{}, ErrTipoMovimientoInvalido
	}

	// 0. Validar que el depósito y el producto existan.
	if _, err := s.depositoRepo.EncontrarPorId(ctx, dto.DepositoID); err != nil {
		if errors.Is(err, deposito.ErrDepositoNoEncontrado) {
			return MovimientoStockDTO{}, ErrDepositoInexistente
		}
		return MovimientoStockDTO{}, err
	}

	prod, err := s.productoRepo.EncontrarPorId(ctx, dto.ProductoID)
	if err != nil {
		return MovimientoStockDTO{}, ErrProductoInexistente
	}

	// 1. Asegurar que exista el registro de stock.
	_, err = s.depositoStockRepo.AsegurarExiste(ctx, depoOID, prodOID, uOID)
	if err != nil {
		return MovimientoStockDTO{}, err
	}

	// 2. Modificar el stock.
	_, err = s.depositoStockRepo.AjustarStock(ctx, depoOID, prodOID, delta, uOID)
	if err != nil {
		if errors.Is(err, depositoproducto.ErrStockInsuficiente) {
			return MovimientoStockDTO{}, ErrStockInsuficiente
		}
		return MovimientoStockDTO{}, err
	}

	// 3. Registrar el movimiento.
	mov := MovimientoStock{
		ProductoID:           prodOID,
		DepositoID:           depoOID,
		Cantidad:             delta,
		Tipo:                 dto.Tipo,
		UsuarioResponsableID: uOID,
		FechaHora:            time.Now(),
		Motivo:               dto.Motivo,
		Auditoria: auditoria.Auditoria{
			CreadoPor: uOID,
			CreadoEn:  time.Now(),
		},
	}

	movimiento, err := s.repo.Crear(ctx, mov)
	if err != nil {
		// Compensación: revertir el ajuste si no se pudo registrar el movimiento.
		if _, errRev := s.depositoStockRepo.AjustarStock(
			ctx,
			depoOID,
			prodOID,
			-delta,
			uOID,
		); errRev != nil {
			log.Printf(
				"CRÍTICO: no se pudo revertir el ajuste de stock (depósito %s, producto %s, delta %d) tras fallar el registro del movimiento: %v",
				dto.DepositoID,
				dto.ProductoID,
				delta,
				errRev,
			)
		}

		return MovimientoStockDTO{}, err
	}

	// 4. Evaluar alertas y determinar si corresponde generar una sugerencia.
	stockBajo, err := s.evaluarAlertas(
		ctx,
		dto.DepositoID,
		dto.ProductoID,
		usuarioID,
	)

	if err != nil {
		log.Printf(
			"ERROR: no se pudieron evaluar las alertas del producto %s en el depósito %s: %v",
			dto.ProductoID,
			dto.DepositoID,
			err,
		)
	} else if stockBajo && s.generarSugerenciasStock != nil {
		s.generarSugerenciasStock(ctx, dto.DepositoID, usuarioID)
	}

	return movimiento.ConvertirADTO(prod.Nombre), nil
}

func (s *Service) Listar(
	ctx context.Context,
	rol, depUsuario string,
	filtro FiltroMovimientosDTO,
) (MovimientosPaginadosDTO, error) {
	depositoPermitido := ""

	if rol == "operario" || rol == "gerente" {
		depositoPermitido = depUsuario
	}

	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 20
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	movimientos, total, err := s.repo.Listar(ctx, filtro, depositoPermitido)
	if err != nil {
		return MovimientosPaginadosDTO{}, err
	}

	pagina := filtro.Pagina
	if pagina < 1 {
		pagina = 1
	}

	// Obtener los ID de productos sin repetir.
	idsProductos := make([]bson.ObjectID, 0, len(movimientos))
	vistos := make(map[bson.ObjectID]struct{}, len(movimientos))

	for _, m := range movimientos {
		if _, existe := vistos[m.ProductoID]; !existe {
			idsProductos = append(idsProductos, m.ProductoID)
			vistos[m.ProductoID] = struct{}{}
		}
	}

	// Buscar todos los productos de los movimientos en una sola consulta.
	productos, err := s.productoRepo.EncontrarPorIds(ctx, idsProductos)
	if err != nil {
		return MovimientosPaginadosDTO{}, err
	}

	// Asociar cada ID con el nombre de su producto.
	nombresProductos := make(map[bson.ObjectID]string, len(productos))

	for _, p := range productos {
		nombresProductos[p.ID] = p.Nombre
	}

	// Construir los DTO enriquecidos.
	items := make([]MovimientoStockDTO, 0, len(movimientos))

	for _, m := range movimientos {
		productoNombre := nombresProductos[m.ProductoID]

		if productoNombre == "" {
			productoNombre = "Producto no disponible"
		}

		items = append(items, m.ConvertirADTO(productoNombre))
	}

	return MovimientosPaginadosDTO{
		Items:         items,
		Total:         total,
		Pagina:        pagina,
		TamanioPagina: filtro.TamanioPagina,
	}, nil
}

func (s *Service) ObtenerPorID(
	ctx context.Context,
	id, rol, depUsuario string,
) (MovimientoStockDTO, error) {
	m, err := s.repo.EncontrarPorID(ctx, id)
	if err != nil {
		return MovimientoStockDTO{}, err
	}

	if (rol == "operario" || rol == "gerente") && depUsuario != m.DepositoID.Hex() {
		return MovimientoStockDTO{}, ErrNoAutorizadoDeposito
	}

	prod, err := s.productoRepo.EncontrarPorId(ctx, m.ProductoID.Hex())
	if err != nil {
		return MovimientoStockDTO{}, err
	}

	return m.ConvertirADTO(prod.Nombre), nil
}

const diasAvisoVencimiento = 30

// evaluarAlertas comprueba el stock mínimo y los vencimientos.
// Crea o resuelve alertas de stock según corresponda.
// Devuelve true cuando el stock está por debajo del mínimo.
func (s *Service) evaluarAlertas(
	ctx context.Context,
	depositoID string,
	productoID string,
	usuarioID string,
) (bool, error) {
	depositoOID, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return false, err
	}

	productoOID, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return false, err
	}

	usuarioOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return false, err
	}

	// Obtener producto.
	prod, err := s.productoRepo.EncontrarPorId(ctx, productoID)
	if err != nil {
		return false, err
	}

	// Obtener stock del producto en el depósito.
	stockFila, err := s.depositoStockRepo.EncontrarPorDepositoYProducto(
		ctx,
		depositoOID,
		productoOID,
	)

	var stockActual int32
	minimo := prod.StockMinimo

	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return false, err
		}
		// Si no existe el registro, consideramos que el stock es 0.
	} else {
		stockActual = stockFila.Stock

		if stockFila.StockMinimo != nil {
			minimo = *stockFila.StockMinimo
		}
	}

	ahora := time.Now()
	stockBajo := stockActual < minimo

	// Evaluar alerta de stock mínimo.
	err = s.evaluadorStock.EvaluarStock(
		ctx,
		depositoOID,
		productoOID,
		stockActual,
		minimo,
		usuarioOID,
	)

	if err != nil {
		return stockBajo, err
	}

	// Evaluar vencimiento independientemente del stock mínimo.
	if prod.FechaVencimiento != nil && stockActual > 0 {
		limite := ahora.AddDate(0, 0, diasAvisoVencimiento)

		if !prod.FechaVencimiento.Before(ahora) &&
			!prod.FechaVencimiento.After(limite) {

			alertaExistente, err := s.alertaStockRepo.BuscarActiva(
				ctx,
				depositoOID,
				productoOID,
				alertastock.AlertaVencimiento,
			)

			if errors.Is(err, mongo.ErrNoDocuments) {
				fechaVencimiento := *prod.FechaVencimiento

				_, err = s.alertaStockRepo.Crear(ctx, alertastock.AlertaStock{
					Tipo:             alertastock.AlertaVencimiento,
					ProductoID:       productoOID,
					DepositoID:       depositoOID,
					StockActual:      stockActual,
					FechaVencimiento: &fechaVencimiento,
					Resuelta:         false,
					Auditoria: auditoria.Auditoria{
						CreadoPor: usuarioOID,
						CreadoEn:  ahora,
					},
				})

				if err != nil {
					return stockBajo, err
				}
			} else if err != nil {
				return stockBajo, err
			} else {
				_ = s.alertaStockRepo.ActualizarVencimiento(
					ctx,
					alertaExistente.ID.Hex(),
					*prod.FechaVencimiento,
					stockActual,
					usuarioOID,
				)
			}
		}
	}

	return stockBajo, nil
}
