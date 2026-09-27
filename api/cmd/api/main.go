package main

import (
	"api/internal/categoria"
	"api/internal/db"
	"api/internal/deposito"
	depositoproducto "api/internal/depositoProducto"
	"api/internal/middleware"
	movimientostock "api/internal/movimientoStock"
	"api/internal/producto"
	"api/internal/productoProveedor"
	"api/internal/proveedor"
	"api/internal/transferencia"
	"api/internal/usuario"

	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	cliente, err := db.Conectar(mongoURI)
	if err != nil {
		log.Fatal(err)
	}

	// Definimos el nombre de la base de datos
	database := cliente.Database("gestock")

	// Generamos las colecciones de la DB
	depositoColeccion := database.Collection("depositos")
	usuarioColeccion := database.Collection("usuarios")
	categoriaColeccion := database.Collection("categorias")
	productoColeccion := database.Collection("productos")
	proveedorColeccion := database.Collection("proveedores")
	productoProveedorColeccion := database.Collection("producto_proveedor")
	depositoProductoColeccion := database.Collection("deposito_producto")
	movimientoStockColeccion := database.Collection("movimientos_stock")
	transferenciaColeccion := database.Collection("transferencias")

	// Cargamos los modulos para cada una de las colecciones
	depositoRepo := deposito.NuevoRepositorioMongo(depositoColeccion)
	depositoService := deposito.NuevoDepositoService(depositoRepo)
	depositoHandler := deposito.NuevoDepositoHandler(depositoService)

	usuarioRepo := usuario.NuevoRepositorioMongo(usuarioColeccion)
	usuarioService := usuario.NuevoUsuarioService(usuarioRepo, depositoRepo)
	usuarioHandler := usuario.NuevoUsuarioHandler(usuarioService)

	categoriaRepo := categoria.NuevoRepositorioMongo(categoriaColeccion)
	categoriaService := categoria.NuevaCategoriaService(categoriaRepo)
	categoriaHandler := categoria.NuevoCategoriaHandler(categoriaService)

	productoRepo := producto.NuevoRepositorioMongo(productoColeccion)
	productoService := producto.NuevoProductoService(productoRepo, categoriaRepo)
	productoHandler := producto.NuevoProductoHandler(productoService)

	productoProveedorRepo := productoproveedor.NuevoRepositorioMongo(productoProveedorColeccion)

	proveedorRepo := proveedor.NuevoRepositorioMongo(proveedorColeccion)
	proveedorService := proveedor.NuevoProveedorService(proveedorRepo, productoProveedorRepo, productoRepo)
	proveedorHandler := proveedor.NuevoProveedorHandler(proveedorService)

	depositoProductoRepo := depositoproducto.NuevoRepositorioMongo(depositoProductoColeccion)
	depositoProductoService := depositoproducto.NuevoService(
		depositoProductoRepo, depositoRepo, productoRepo, categoriaRepo, usuarioRepo,
	)
	depositoProductoHandler := depositoproducto.NuevoHandler(depositoProductoService)

	movimientoStockRepo := movimientostock.NuevoRepositorioMongo(movimientoStockColeccion)
	movimientoStockService := movimientostock.NuevoService(movimientoStockRepo, depositoProductoRepo)
	movimientoStockHandler := movimientostock.NuevoHandler(movimientoStockService)

	transferenciaRepo := transferencia.NuevoRepositorioMongo(transferenciaColeccion)
	transferenciaService := transferencia.NuevoService(transferenciaRepo, depositoProductoRepo, movimientoStockRepo)
	transferenciaHandler := transferencia.NuevoHandler(transferenciaService)

	// Definimos los middlewares que van a ser invocados en
	// los registradores de rutas en los handlers
	authMiddleware := middleware.AuthMiddleware()
	rolMiddleware := middleware.RequiereRolMiddleware

	// Definimos el enrutador que se va a encargar de mapear
	router := gin.Default()

	// Definimos el CORS para impedir que se pueda llamar
	// desde cualquier URL de internet
	router.Use(middleware.CORSMiddleware())
	usuario.RegistrarRutas(router, usuarioHandler, authMiddleware, rolMiddleware)
	deposito.RegistrarRutas(router, depositoHandler, authMiddleware, rolMiddleware)
	categoria.RegistrarRutas(router, categoriaHandler, authMiddleware, rolMiddleware)
	producto.RegistrarRutas(router, productoHandler, authMiddleware, rolMiddleware)
	proveedor.RegistrarRutas(router, proveedorHandler, authMiddleware, rolMiddleware)
	depositoproducto.RegistrarRutas(router, depositoProductoHandler, authMiddleware, rolMiddleware)
	movimientostock.RegistrarRutas(router, movimientoStockHandler, authMiddleware, rolMiddleware)
	transferencia.RegistrarRutas(router, transferenciaHandler, authMiddleware, rolMiddleware)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
