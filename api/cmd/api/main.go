package main

import (
	alertastock "api/internal/alertaStock"
	"api/internal/auth"
	"api/internal/categoria"
	"api/internal/db"
	"api/internal/deposito"
	depositoproducto "api/internal/depositoProducto"
	evaluacionstock "api/internal/evaluacionStock"
	evaluacionvencimiento "api/internal/evaluacionVencimiento"
	"api/internal/middleware"
	movimientostock "api/internal/movimientoStock"
	novedaddeposito "api/internal/novedadDeposito"
	ordendecompra "api/internal/ordenDeCompra"
	ordendecompraitem "api/internal/ordenDeCompraItem"
	"api/internal/producto"
	productoproveedor "api/internal/productoProveedor"
	"api/internal/proveedor"
	"api/internal/reporte"
	"api/internal/transferencia"
	"api/internal/usuario"
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	// CONFIGURACIÓN Y CONEXIÓN

	if err := auth.RequerirSecretoConfigurado(); err != nil {
		log.Fatal(err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	cliente, err := db.Conectar(mongoURI)
	if err != nil {
		log.Fatal(err)
	}

	database := cliente.Database("gestock")

	// COLECCIONES

	depositoColeccion := database.Collection("depositos")
	usuarioColeccion := database.Collection("usuarios")
	categoriaColeccion := database.Collection("categorias")
	productoColeccion := database.Collection("productos")
	proveedorColeccion := database.Collection("proveedores")
	productoProveedorColeccion := database.Collection("producto_proveedor")
	depositoProductoColeccion := database.Collection("deposito_producto")
	movimientoStockColeccion := database.Collection("movimientos_stock")
	transferenciaColeccion := database.Collection("transferencias")
	novedadDepositoColeccion := database.Collection("novedades_deposito")
	ordenesCompraColeccion := database.Collection("ordenes_compra")
	ordenesCompraItemColeccion := database.Collection("ordenes_compra_items")
	alertaStockColeccion := database.Collection("alertas_stock")

	// REPOSITORIOS

	depositoRepo := deposito.NuevoRepositorioMongo(depositoColeccion)
	usuarioRepo := usuario.NuevoRepositorioMongo(usuarioColeccion)
	categoriaRepo := categoria.NuevoRepositorioMongo(categoriaColeccion)
	productoRepo := producto.NuevoRepositorioMongo(productoColeccion)
	proveedorRepo := proveedor.NuevoRepositorioMongo(proveedorColeccion)
	productoProveedorRepo := productoproveedor.NuevoRepositorioMongo(productoProveedorColeccion)
	depositoProductoRepo := depositoproducto.NuevoRepositorioMongo(
		depositoProductoColeccion,
		productoColeccion,
		categoriaColeccion,
	)
	movimientoStockRepo := movimientostock.NuevoRepositorioMongo(movimientoStockColeccion)
	transferenciaRepo := transferencia.NuevoRepositorioMongo(transferenciaColeccion)
	novedadDepositoRepo := novedaddeposito.NuevoRepositorioMongo(novedadDepositoColeccion)
	ordenCompraRepo := ordendecompra.NuevoRepositorioMongo(ordenesCompraColeccion)
	ordenCompraItemRepo := ordendecompraitem.NuevoRepositorioMongo(ordenesCompraItemColeccion)
	alertaStockRepo := alertastock.NuevoRepositorioMongo(alertaStockColeccion)

	reporteRepo := reporte.NuevoRepositorioMongo(
		depositoProductoColeccion,
		productoColeccion,
		depositoColeccion,
		categoriaColeccion,
		movimientoStockColeccion,
		ordenesCompraColeccion,
		transferenciaColeccion,
	)

	// Funcion utilizada para que Productos pueda consultar stocks en depositoProducto.
	// No se pueden importar mutuamente dado que da error ciclico
	obtenerStocksPorProducto := func(
		ctx context.Context,
		productoID bson.ObjectID,
	) ([]producto.RegistroStockProducto, error) {
		registros, err := depositoProductoRepo.EncontrarPorProducto(ctx, productoID)
		if err != nil {
			return nil, err
		}

		stocks := make([]producto.RegistroStockProducto, 0, len(registros))

		for _, registro := range registros {
			stocks = append(stocks, producto.RegistroStockProducto{
				DepositoID:  registro.DepositoID,
				ProductoID:  registro.ProductoID,
				Stock:       registro.Stock,
				StockMinimo: registro.StockMinimo,
			})
		}

		return stocks, nil
	}

	// SERVICES
	evaluadorVencimiento := evaluacionvencimiento.NuevoEvaluadorVencimiento(
		productoRepo,
		depositoProductoRepo,
		alertaStockRepo,
	)
	evaluadorStock := *evaluacionstock.NuevoEvaluadorStock(alertaStockRepo)
	depositoService := deposito.NuevoDepositoService(
		depositoRepo,
		usuarioRepo,
	)

	usuarioService := usuario.NuevoUsuarioService(usuarioRepo, depositoRepo)

	if err := usuarioService.InicializarAdmin(context.Background()); err != nil {
		log.Fatalf("no se pudo inicializar el administrador inicial: %v", err)
	}

	categoriaService := categoria.NuevaCategoriaService(categoriaRepo)

	productoService := producto.NuevoProductoService(
		productoRepo,
		categoriaRepo,
		obtenerStocksPorProducto,
		evaluadorStock,
	)

	proveedorService := proveedor.NuevoProveedorService(
		proveedorRepo,
		productoProveedorRepo,
		productoRepo,
	)

	alertaStockService := alertastock.NuevoService(
		alertaStockRepo,
	)

	depositoProductoService := depositoproducto.NuevoService(
		depositoProductoRepo,
		depositoRepo,
		productoRepo,
		categoriaRepo,
		usuarioRepo,
		evaluadorStock,
	)

	ordenCompraService := ordendecompra.NuevoOrdenDeCompraService(
		ordenCompraRepo,
		ordenCompraItemRepo,
		proveedorRepo,
		depositoRepo,
		productoProveedorRepo,
		movimientoStockRepo,
		depositoProductoRepo,
		productoRepo,
		evaluadorStock,
	)

	// Función utilizada automáticamente cuando una operación reduce
	// el stock de un depósito.
	generarSugerenciasStock := func(
		ctx context.Context,
		depositoID string,
		usuarioID string,
	) {
		_, err := ordenCompraService.GenerarSugerenciasPorStockBajoMinimo(
			ctx,
			"administrador",
			"",
			depositoID,
			usuarioID,
		)

		if err != nil {
			log.Printf(
				"Aviso: no se pudieron generar sugerencias de orden de compra para el depósito %s: %v",
				depositoID,
				err,
			)
		}
	}

	movimientoStockService := movimientostock.NuevoService(
		movimientoStockRepo,
		depositoProductoRepo,
		productoRepo,
		depositoRepo,
		alertaStockRepo,
		evaluadorStock,
		generarSugerenciasStock,
	)

	transferenciaService := transferencia.NuevoService(
		transferenciaRepo,
		depositoProductoRepo,
		depositoRepo,
		productoRepo,
		alertaStockRepo,
		evaluadorStock,
		movimientoStockRepo,
		generarSugerenciasStock,
	)

	novedadDepositoService := novedaddeposito.NuevoService(
		novedadDepositoRepo,
		depositoRepo,
	)

	reporteService := reporte.NuevoService(reporteRepo)

	// HANDLERS

	depositoHandler := deposito.NuevoDepositoHandler(depositoService)
	usuarioHandler := usuario.NuevoUsuarioHandler(usuarioService)
	categoriaHandler := categoria.NuevoCategoriaHandler(categoriaService)
	productoHandler := producto.NuevoProductoHandler(productoService)
	proveedorHandler := proveedor.NuevoProveedorHandler(proveedorService)
	depositoProductoHandler := depositoproducto.NuevoHandler(depositoProductoService)
	movimientoStockHandler := movimientostock.NuevoHandler(movimientoStockService)
	transferenciaHandler := transferencia.NuevoHandler(transferenciaService)
	novedadDepositoHandler := novedaddeposito.NuevoHandler(novedadDepositoService)
	alertaStockHandler := alertastock.NuevoHandler(alertaStockService)
	reporteHandler := reporte.NuevoHandler(reporteService)
	ordenCompraHandler := ordendecompra.NuevoHandler(ordenCompraService)

	// MIDDLEWARES Y RUTAS

	authMiddleware := middleware.AuthMiddleware(usuarioService)
	rolMiddleware := middleware.RequiereRolMiddleware

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	usuario.RegistrarRutas(router, usuarioHandler, authMiddleware, rolMiddleware)
	deposito.RegistrarRutas(router, depositoHandler, authMiddleware, rolMiddleware)
	categoria.RegistrarRutas(router, categoriaHandler, authMiddleware, rolMiddleware)
	producto.RegistrarRutas(router, productoHandler, authMiddleware, rolMiddleware)
	proveedor.RegistrarRutas(router, proveedorHandler, authMiddleware, rolMiddleware)
	depositoproducto.RegistrarRutas(router, depositoProductoHandler, authMiddleware, rolMiddleware)
	movimientostock.RegistrarRutas(router, movimientoStockHandler, authMiddleware, rolMiddleware)
	transferencia.RegistrarRutas(router, transferenciaHandler, authMiddleware, rolMiddleware)
	novedaddeposito.RegistrarRutas(router, novedadDepositoHandler, authMiddleware, rolMiddleware)
	alertastock.RegistrarRutas(router, alertaStockHandler, authMiddleware, rolMiddleware)
	reporte.RegistrarRutas(router, reporteHandler, authMiddleware, rolMiddleware)
	ordendecompra.RegistrarRutas(router, ordenCompraHandler, authMiddleware, rolMiddleware)

	// Time ticker: Permite ejecutar instrucciones de forma periódica
	// Lo usamos para evaluar periódicamente las alertas por vencimiento.

	go func() {
		const diasAnticipacion = 30

		// Prueba: que se ejecute cada 5 minutos.
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		// Primera evaluación al iniciar el servidor.
		if err := evaluadorVencimiento.EvaluarVencimientos(
			context.Background(),
			diasAnticipacion,
		); err != nil {
			log.Printf("Error al evaluar vencimientos: %v", err)
		}

		for range ticker.C {
			if err := evaluadorVencimiento.EvaluarVencimientos(
				context.Background(),
				diasAnticipacion,
			); err != nil {
				log.Printf("Error al evaluar vencimientos: %v", err)
			}
		}
	}()

	// INICIO DEL SERVIDOR
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
