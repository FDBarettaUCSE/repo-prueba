package reporte

import (
	"context"
	"errors"
	"time"

	"api/internal/categoria"
	"api/internal/deposito"
	depositoproducto "api/internal/depositoProducto"
	movimientostock "api/internal/movimientoStock"
	ordendecompra "api/internal/ordenDeCompra"
	"api/internal/producto"
	"api/internal/transferencia"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ErrAccesoDenegado: mismo criterio que en el resto de los módulos —
// Gerente y Operario solo pueden pedir reportes de su propio depósito.
var ErrAccesoDenegado = errors.New("no tenés permiso para operar sobre este depósito")

// Service no persiste nada propio: agrega datos que YA viven en los
// repositorios de los demás módulos. Por eso depende de sus repositorios (de
// solo lectura, en la práctica) y no tiene un repository.go propio con
// lógica — ver ese archivo.
type Service struct {
	depositoRepositorio         deposito.DepositoRepositorio
	categoriaRepositorio        categoria.CategoriaRepositorio
	productoRepositorio         producto.ProductoRepositorio
	depositoProductoRepositorio depositoproducto.DepositoProductoRepositorio
	movimientoStockRepositorio  movimientostock.MovimientoStockRepositorio
	transferenciaRepositorio    transferencia.Repositorio
	ordenCompraRepositorio      ordendecompra.OrdenCompraRepositorio
}

func NuevoService(
	repoDeposito deposito.DepositoRepositorio,
	repoCategoria categoria.CategoriaRepositorio,
	repoProducto producto.ProductoRepositorio,
	repoDepositoProducto depositoproducto.DepositoProductoRepositorio,
	repoMovimientoStock movimientostock.MovimientoStockRepositorio,
	repoTransferencia transferencia.Repositorio,
	repoOrdenCompra ordendecompra.OrdenCompraRepositorio,
) *Service {
	return &Service{
		depositoRepositorio:         repoDeposito,
		categoriaRepositorio:        repoCategoria,
		productoRepositorio:         repoProducto,
		depositoProductoRepositorio: repoDepositoProducto,
		movimientoStockRepositorio:  repoMovimientoStock,
		transferenciaRepositorio:    repoTransferencia,
		ordenCompraRepositorio:      repoOrdenCompra,
	}
}

// resolverAlcanceDeposito decide el depósito EFECTIVO sobre el que va a
// reportar cada consulta: Administrador y Auditor pueden pedir cualquiera
// (o "" = toda la empresa, sin filtrar); Gerente y Operario quedan siempre
// forzados al suyo — si piden explícitamente otro depósito distinto al
// propio, es un acceso denegado, no un filtro que se ignora en silencio.
func resolverAlcanceDeposito(rol, depositoUsuario, depositoIDPedido string) (string, error) {
	if rol == "administrador" || rol == "auditor" {
		return depositoIDPedido, nil
	}
	if depositoIDPedido != "" && depositoIDPedido != depositoUsuario {
		return "", ErrAccesoDenegado
	}
	return depositoUsuario, nil
}

func nombreDeDeposito(depositos []deposito.Deposito, id string) string {
	for _, d := range depositos {
		if d.ID.Hex() == id {
			return d.Nombre
		}
	}
	return ""
}

// ReporteInventario agrega stock y valorización (costo unitario × stock) por
// depósito y por categoría (ver enunciado). Sin filtro de depósito (solo
// disponible para Administrador/Auditor) agrega TODA la empresa.
func (s *Service) ReporteInventario(ctx context.Context, rol, depositoUsuario string, filtro FiltroInventarioDTO) (ReporteInventarioDTO, error) {
	depositoID, err := resolverAlcanceDeposito(rol, depositoUsuario, filtro.DepositoID)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	var filas []depositoproducto.DepositoProducto
	if depositoID != "" {
		oidDeposito, err := bson.ObjectIDFromHex(depositoID)
		if err != nil {
			return ReporteInventarioDTO{}, err
		}
		filas, err = s.depositoProductoRepositorio.EncontrarPorDeposito(ctx, oidDeposito)
		if err != nil {
			return ReporteInventarioDTO{}, err
		}
	} else {
		filas, err = s.depositoProductoRepositorio.EncontrarTodos(ctx)
		if err != nil {
			return ReporteInventarioDTO{}, err
		}
	}

	productos, err := s.productoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}
	productosPorID := make(map[string]producto.Producto, len(productos))
	for _, p := range productos {
		productosPorID[p.ID.Hex()] = p
	}

	depositos, err := s.depositoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	categorias, err := s.categoriaRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}
	categoriasPorID := make(map[string]categoria.Categoria, len(categorias))
	for _, c := range categorias {
		categoriasPorID[c.ID.Hex()] = c
	}

	porDepositoAcc := map[string]*ValorizacionDepositoDTO{}
	porCategoriaAcc := map[string]*ValorizacionCategoriaDTO{}
	var stockTotal int32
	var valorizadoTotal float64

	for _, fila := range filas {
		p, existe := productosPorID[fila.ProductoID.Hex()]
		if !existe {
			continue
		}
		if filtro.CategoriaID != "" && p.Categoria.Hex() != filtro.CategoriaID {
			continue
		}

		valor := float64(fila.Stock) * p.CostoUnitario

		depID := fila.DepositoID.Hex()
		accDep, existeDep := porDepositoAcc[depID]
		if !existeDep {
			accDep = &ValorizacionDepositoDTO{
				DepositoID:     depID,
				DepositoNombre: nombreDeDeposito(depositos, depID),
			}
			porDepositoAcc[depID] = accDep
		}
		accDep.Stock += fila.Stock
		accDep.Valorizado += valor

		catID := p.Categoria.Hex()
		accCat, existeCat := porCategoriaAcc[catID]
		if !existeCat {
			nombreCat := ""
			if c, ok := categoriasPorID[catID]; ok {
				nombreCat = c.Nombre
			}
			accCat = &ValorizacionCategoriaDTO{CategoriaID: catID, CategoriaNombre: nombreCat}
			porCategoriaAcc[catID] = accCat
		}
		accCat.Stock += fila.Stock
		accCat.Valorizado += valor

		stockTotal += fila.Stock
		valorizadoTotal += valor
	}

	porDeposito := make([]ValorizacionDepositoDTO, 0, len(porDepositoAcc))
	for _, v := range porDepositoAcc {
		porDeposito = append(porDeposito, *v)
	}
	porCategoria := make([]ValorizacionCategoriaDTO, 0, len(porCategoriaAcc))
	for _, v := range porCategoriaAcc {
		porCategoria = append(porCategoria, *v)
	}

	return ReporteInventarioDTO{
		PorDeposito:     porDeposito,
		PorCategoria:    porCategoria,
		StockTotal:      stockTotal,
		ValorizadoTotal: valorizadoTotal,
	}, nil
}

// ReporteMovimientos agrupa el historial de movimientos por fecha (día),
// depósito y tipo (ver enunciado). Reutiliza el filtro/consulta que ya expone
// movimientoStock — acá no vuelve a paginar, agrega.
func (s *Service) ReporteMovimientos(ctx context.Context, rol, depositoUsuario string, filtro FiltroReporteMovimientosDTO) (ReporteMovimientosDTO, error) {
	depositoID, err := resolverAlcanceDeposito(rol, depositoUsuario, filtro.DepositoID)
	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	filtroMov := movimientostock.FiltroMovimientosDTO{
		DepositoID: depositoID,
		Tipo:       filtro.Tipo,
		FechaDesde: filtro.FechaDesde,
		FechaHasta: filtro.FechaHasta,
	}

	// depositoPermitido="" porque el alcance YA quedó resuelto arriba (en
	// filtroMov.DepositoID) — no hace falta que el repo lo fuerce de nuevo.
	movimientos, err := s.movimientoStockRepositorio.Listar(ctx, filtroMov, "")
	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	depositos, err := s.depositoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	type clave struct {
		depositoID string
		fecha      string
		tipo       string
	}
	acc := map[clave]*ResumenMovimientoDTO{}

	for _, m := range movimientos {
		depID := m.DepositoID.Hex()
		fecha := m.FechaHora.Format("2006-01-02")
		tipo := string(m.Tipo)

		k := clave{depID, fecha, tipo}
		r, existe := acc[k]
		if !existe {
			r = &ResumenMovimientoDTO{
				DepositoID:     depID,
				DepositoNombre: nombreDeDeposito(depositos, depID),
				Fecha:          fecha,
				Tipo:           tipo,
			}
			acc[k] = r
		}
		r.CantidadEventos++
		r.UnidadesNetas += m.Cantidad
	}

	resumen := make([]ResumenMovimientoDTO, 0, len(acc))
	for _, r := range acc {
		resumen = append(resumen, *r)
	}

	return ReporteMovimientosDTO{
		Resumen:          resumen,
		TotalMovimientos: int64(len(movimientos)),
	}, nil
}

// ReporteAlertas junta, en una sola consulta, los productos bajo mínimo y
// los próximos a vencer (ver enunciado — son los dos tipos de alerta que
// tienen que verse en el dashboard). DiasVencimiento es la ventana
// configurable: default 30 si no se especifica.
func (s *Service) ReporteAlertas(ctx context.Context, rol, depositoUsuario string, filtro FiltroAlertasDTO) (ReporteAlertasDTO, error) {
	depositoID, err := resolverAlcanceDeposito(rol, depositoUsuario, filtro.DepositoID)
	if err != nil {
		return ReporteAlertasDTO{}, err
	}

	diasVencimiento := filtro.DiasVencimiento
	if diasVencimiento <= 0 {
		diasVencimiento = 30
	}

	var filas []depositoproducto.DepositoProducto
	if depositoID != "" {
		oidDeposito, err := bson.ObjectIDFromHex(depositoID)
		if err != nil {
			return ReporteAlertasDTO{}, err
		}
		filas, err = s.depositoProductoRepositorio.EncontrarPorDeposito(ctx, oidDeposito)
		if err != nil {
			return ReporteAlertasDTO{}, err
		}
	} else {
		filas, err = s.depositoProductoRepositorio.EncontrarTodos(ctx)
		if err != nil {
			return ReporteAlertasDTO{}, err
		}
	}

	productos, err := s.productoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReporteAlertasDTO{}, err
	}
	productosPorID := make(map[string]producto.Producto, len(productos))
	for _, p := range productos {
		productosPorID[p.ID.Hex()] = p
	}

	depositos, err := s.depositoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReporteAlertasDTO{}, err
	}

	bajoMinimo := []ProductoBajoMinimoDTO{}
	porVencer := []ProductoPorVencerDTO{}

	ahora := time.Now()
	limite := ahora.Add(time.Duration(diasVencimiento) * 24 * time.Hour)

	for _, fila := range filas {
		p, existe := productosPorID[fila.ProductoID.Hex()]
		if !existe {
			continue
		}

		depNombre := nombreDeDeposito(depositos, fila.DepositoID.Hex())

		stockMinimoEfectivo := p.StockMinimo
		if fila.StockMinimo != nil {
			stockMinimoEfectivo = *fila.StockMinimo
		}

		if fila.Stock < stockMinimoEfectivo {
			bajoMinimo = append(bajoMinimo, ProductoBajoMinimoDTO{
				DepositoID:     fila.DepositoID.Hex(),
				DepositoNombre: depNombre,
				ProductoID:     p.ID.Hex(),
				ProductoNombre: p.Nombre,
				Stock:          fila.Stock,
				StockMinimo:    stockMinimoEfectivo,
			})
		}

		// Solo alertamos vencimiento de lo que efectivamente hay en stock en
		// ESE depósito (fila.Stock > 0): un producto vencido sin stock no es
		// una alerta accionable. !After(limite) incluye tanto lo que vence
		// dentro de la ventana como lo que YA venció.
		if p.FechaVencimiento != nil && fila.Stock > 0 && !p.FechaVencimiento.After(limite) {
			dias := int(p.FechaVencimiento.Sub(ahora).Hours() / 24)
			porVencer = append(porVencer, ProductoPorVencerDTO{
				DepositoID:       fila.DepositoID.Hex(),
				DepositoNombre:   depNombre,
				ProductoID:       p.ID.Hex(),
				ProductoNombre:   p.Nombre,
				Stock:            fila.Stock,
				FechaVencimiento: p.FechaVencimiento.Format("2006-01-02"),
				DiasParaVencer:   dias,
			})
		}
	}

	return ReporteAlertasDTO{BajoMinimo: bajoMinimo, PorVencer: porVencer}, nil
}

// ReportePendientes agrupa por depósito el estado de las órdenes de compra y
// de las transferencias (ver enunciado). Para transferencias, el conteo no
// es un simple "por estado": mira qué le falta HACER a cada depósito
// (ver EstadoTransferenciasDepositoDTO) — que es lo que un dashboard de
// Gerente necesita para saber qué tiene pendiente de aprobar.
func (s *Service) ReportePendientes(ctx context.Context, rol, depositoUsuario string, filtro FiltroPendientesDTO) (ReportePendientesDTO, error) {
	depositoID, err := resolverAlcanceDeposito(rol, depositoUsuario, filtro.DepositoID)
	if err != nil {
		return ReportePendientesDTO{}, err
	}

	depositos, err := s.depositoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReportePendientesDTO{}, err
	}

	ordenes, err := s.ordenCompraRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return ReportePendientesDTO{}, err
	}

	transferencias, err := s.transferenciaRepositorio.Listar(ctx, transferencia.FiltroTransferenciasDTO{}, "")
	if err != nil {
		return ReportePendientesDTO{}, err
	}

	ordenesAcc := map[string]*EstadoOrdenesDepositoDTO{}
	for _, o := range ordenes {
		depID := o.DepositoID.Hex()
		if depositoID != "" && depID != depositoID {
			continue
		}

		acc, existe := ordenesAcc[depID]
		if !existe {
			acc = &EstadoOrdenesDepositoDTO{DepositoID: depID, DepositoNombre: nombreDeDeposito(depositos, depID)}
			ordenesAcc[depID] = acc
		}

		switch o.Estado {
		case ordendecompra.OrdenBorrador:
			acc.Borrador++
		case ordendecompra.OrdenConfirmada:
			acc.Confirmada++
		case ordendecompra.OrdenParcial:
			acc.Parcial++
		case ordendecompra.OrdenCompletada:
			acc.Completada++
		}
	}

	transferenciasAcc := map[string]*EstadoTransferenciasDepositoDTO{}
	obtenerAccTransferencia := func(depID string) *EstadoTransferenciasDepositoDTO {
		acc, existe := transferenciasAcc[depID]
		if !existe {
			acc = &EstadoTransferenciasDepositoDTO{DepositoID: depID, DepositoNombre: nombreDeDeposito(depositos, depID)}
			transferenciasAcc[depID] = acc
		}
		return acc
	}

	for _, t := range transferencias {
		origenID := t.DepositoOrigenID.Hex()
		destinoID := t.DepositoDestinoID.Hex()

		switch t.Estado {
		case transferencia.TransferenciaSolicitada:
			if depositoID == "" || depositoID == origenID {
				obtenerAccTransferencia(origenID).SolicitadasComoOrigen++
			}
			if depositoID == "" || depositoID == destinoID {
				obtenerAccTransferencia(destinoID).PendientesDeAprobar++
			}
		case transferencia.TransferenciaAprobada:
			if depositoID == "" || depositoID == destinoID {
				obtenerAccTransferencia(destinoID).AprobadasPorRecibir++
			}
		}
	}

	ordenesResultado := make([]EstadoOrdenesDepositoDTO, 0, len(ordenesAcc))
	for _, v := range ordenesAcc {
		ordenesResultado = append(ordenesResultado, *v)
	}

	transferenciasResultado := make([]EstadoTransferenciasDepositoDTO, 0, len(transferenciasAcc))
	for _, v := range transferenciasAcc {
		transferenciasResultado = append(transferenciasResultado, *v)
	}

	return ReportePendientesDTO{Ordenes: ordenesResultado, Transferencias: transferenciasResultado}, nil
}
