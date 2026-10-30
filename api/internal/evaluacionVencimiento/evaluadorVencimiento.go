package evaluacionvencimiento

import (
	"context"
	"errors"
	"fmt"
	"time"

	alertastock "api/internal/alertaStock"
	"api/internal/auditoria"
	depositoproducto "api/internal/depositoProducto"
	"api/internal/producto"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EvaluadorVencimiento struct {
	productoRepo         producto.ProductoRepositorio
	depositoProductoRepo depositoproducto.DepositoProductoRepositorio
	alertaRepo           alertastock.AlertaStockRepositorio
}

func NuevoEvaluadorVencimiento(
	productoRepo producto.ProductoRepositorio,
	depositoProductoRepo depositoproducto.DepositoProductoRepositorio,
	alertaRepo alertastock.AlertaStockRepositorio,
) *EvaluadorVencimiento {
	return &EvaluadorVencimiento{
		productoRepo:         productoRepo,
		depositoProductoRepo: depositoProductoRepo,
		alertaRepo:           alertaRepo,
	}
}

func (e *EvaluadorVencimiento) EvaluarVencimientos(
	ctx context.Context,
	diasAnticipacion int,
) error {
	if diasAnticipacion < 0 {
		return errors.New("los días de anticipación no pueden ser negativos")
	}

	productos, err := e.productoRepo.EncontrarTodos(ctx)
	if err != nil {
		return fmt.Errorf("error al obtener productos: %w", err)
	}

	stocks, err := e.depositoProductoRepo.EncontrarTodos(ctx)
	if err != nil {
		return fmt.Errorf("error al obtener stocks: %w", err)
	}

	productosPorID := make(map[bson.ObjectID]producto.Producto, len(productos))
	for _, p := range productos {
		productosPorID[p.ID] = p
	}

	ahora := time.Now()
	limite := ahora.AddDate(0, 0, diasAnticipacion)
	usuarioSistema := bson.NilObjectID

	for _, stock := range stocks {
		p, existeProducto := productosPorID[stock.ProductoID]

		tieneVencimiento := existeProducto &&
			p.FechaVencimiento != nil &&
			!p.FechaVencimiento.After(limite)

		correspondeAlerta := tieneVencimiento && stock.Stock > 0

		alerta, err := e.alertaRepo.BuscarActiva(
			ctx,
			stock.DepositoID,
			stock.ProductoID,
			alertastock.AlertaVencimiento,
		)

		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("error al buscar alerta de vencimiento: %w", err)
		}

		if correspondeAlerta {
			if errors.Is(err, mongo.ErrNoDocuments) {
				_, err := e.alertaRepo.Crear(ctx, alertastock.AlertaStock{
					Tipo:             alertastock.AlertaVencimiento,
					ProductoID:       stock.ProductoID,
					DepositoID:       stock.DepositoID,
					StockMinimo:      0,
					StockActual:      stock.Stock,
					FechaVencimiento: p.FechaVencimiento,
					Resuelta:         false,
					Auditoria: auditoria.Auditoria{
						CreadoPor: usuarioSistema,
						CreadoEn:  ahora,
					},
				})
				if err != nil {
					return fmt.Errorf("error al crear alerta de vencimiento: %w", err)
				}
			} else {
				err := e.alertaRepo.ActualizarVencimiento(
					ctx,
					alerta.ID.Hex(),
					*p.FechaVencimiento,
					stock.Stock,
					usuarioSistema,
				)
				if err != nil {
					return fmt.Errorf("error al actualizar alerta de vencimiento: %w", err)
				}
			}
		} else if err == nil {
			_, err := e.alertaRepo.MarcarResuelta(
				ctx,
				alerta.ID.Hex(),
				usuarioSistema,
			)
			if err != nil {
				return fmt.Errorf("error al resolver alerta de vencimiento: %w", err)
			}
		}
	}

	return nil
}
