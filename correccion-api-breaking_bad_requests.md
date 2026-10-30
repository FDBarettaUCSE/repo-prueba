# Corrección API — Gestock — BREAKING_BAD_REQUESTS

**Integrantes:** Baretta Franco David, Bonino Agustín José, Krause Guillermo Andrés.
**Fecha de corrección:** 2026-10-02
**Commit corregido:** 7c5af8e (2026-10-01)
**Fase 2 (ejecución):** ejecutada (S1, S3, S4, S5, S6, S7, S8, S9, S10; el resto de los flujos no se probó en ejecución porque la lectura del código ya mostraba el problema: ver notas en cada ítem). Mongo 7 y API en contenedores descartables propios (base i=1, Mongo 27101, API 8101), ya eliminados. El `main.go` fija el puerto 8080 dentro del contenedor, por eso la API se corrió con la imagen del propio Dockerfile mapeada a 8101.
**Interpretaciones del equipo:**
- Completar transferencia: personal (operario o gerente) del depósito destino, o administrador.
- Sobrerrecepción en OC: rechazada con 409.
- Ajuste de stock: un solo tipo `ajuste` con cantidad con signo.
- Baja de depósitos/productos: lógica (`fecha_eliminacion`).
- OC sugerida: se dispara en egreso, ajuste negativo y aprobación de transferencia (callback `generarSugerenciasStock` cableado en `main.go`), agrupando por proveedor de menor costo en un borrador por depósito/proveedor.

## Resultado
- **Nivel orientativo:** Aprobado
- **Porcentaje:** 80 % (CUMPLE 33, PARCIAL 19, NO CUMPLE 1, NO VERIFICABLE 0) → 42,5 / 53 = 80,2 %
- **Bloqueantes en falla:** ninguno en `NO CUMPLE`. C2 quedó en `PARCIAL` con un hallazgo CRÍTICO confirmado en ejecución (ver C2).

## Ítems

### A1. Compila 🔒
- **Estado:** CUMPLE
- **Evidencia:** `go build ./...` y `go vet ./...` en `api/` terminan sin errores ni advertencias.

### A2. Capas
- **Estado:** CUMPLE
- **Evidencia:** los handlers (`*/handler.go`) parsean, llaman al service y mapean errores a HTTP; las reglas están en `*/service.go`; el acceso a Mongo está en `*/repository.go` y ningún repository importa Gin ni usa códigos HTTP.

### A3. Repository como interfaz y DI manual
- **Estado:** CUMPLE
- **Evidencia:** cada service recibe interfaces (`UsuarioRepositorio`, `DepositoProductoRepositorio`, etc., con `var _ X = (*RepositorioMongo)(nil)`); todo el grafo se arma en `api/cmd/api/main.go:66-179`; sin clientes globales.

### A4. Model vs DTO
- **Estado:** CUMPLE
- **Evidencia:** models y DTOs separados en todos los dominios; `usuario/dto.go:31-37` (`UsuarioDTO`) no expone el hash; los models no se serializan directo (los handlers responden con `ConvertirADTO()`).

### A5. Errores y códigos HTTP
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** `GET /movimientos/:id` y `GET /ordenes-compra/:id` con un id inexistente o mal formado responden 500 (`handler.go:100-107`, `responderError` de OC) en vez de 404/400, exponiendo el error interno del driver.
- **Evidencia:** hay errores de negocio como valores y mapeo razonable a 400/403/404/409 en la mayoría de los dominios, pero varios caminos devuelven 500 donde corresponde 400/404.
- **Código:** `api/internal/movimientoStock/handler.go:100-107`
```go
m, err := h.service.ObtenerPorID(c.Request.Context(), id, usr.UsuarioRol, usr.DepositoID)
if err != nil {
    if errors.Is(err, ErrNoAutorizadoDeposito) {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
```
- **Detalle:** `GET /movimientos/<id inexistente o mal formado>` responde 500. En OC, `responderError` (`ordenDeCompra/handler.go:20-45`) no mapea los "no encontrado" ni los ObjectID inválidos.
- **Prueba dinámica:** `GET /ordenes-compra/xyz` → `500 {"error":"the provided hex string is not a valid ObjectID"}`. El enunciado/guía piden 400 para IDs inválidos.

### A6. Configuración por entorno
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** `usuario/service.go:290-298` cae a `admin@gestock.com / Admin1234!` cuando faltan `INITIAL_ADMIN_*`; como el compose no las define, un despliegue queda con una credencial de administrador pública y utilizable.
- **Evidencia:** `JWT_SECRET` es obligatorio y sin valor por defecto (`auth/jwt.go:24-29`, corta el arranque) y `MONGO_URI` viene por entorno. Pero el admin inicial cae a credenciales hardcodeadas si no se setean variables.
- **Código:** `api/internal/usuario/service.go:290-298`
```go
adminEmail := os.Getenv("INITIAL_ADMIN_EMAIL")
if adminEmail == "" {
    adminEmail = "admin@gestock.com"
}
adminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
if adminPassword == "" {
    adminPassword = "Admin1234!"
}
```
- **Detalle:** el `docker-compose.yml` no define `INITIAL_ADMIN_*`, así que un despliegue por compose queda con `admin@gestock.com / Admin1234!`, una credencial pública utilizable. Debería exigirse la variable (como se hizo con `JWT_SECRET`) o generarse al azar. Confirmado: el admin con esas credenciales loguea.

### B1. Login
- **Estado:** CUMPLE
- **Evidencia:** `usuario/service.go:101-129`; devuelve JWT y el mismo error `credenciales inválidas` si el usuario no existe o la contraseña falla.
- **Prueba dinámica (S1):** contraseña errónea y usuario inexistente → ambos `401 {"error":"credenciales inválidas"}`.

### B2. Contraseñas con hash 🔒
- **Estado:** CUMPLE
- **Evidencia:** `auth/hash.go` usa `bcrypt.GenerateFromPassword` / `CompareHashAndPassword`; `Usuario.ContraseniaHasheada` tiene `json:"-"` y no hay logs ni DTO con el hash.

### B3. JWT correcto 🔒
- **Estado:** CUMPLE
- **Evidencia:** `auth/jwt.go:43-76`: HS256, `exp` a 2 h, secreto desde entorno, claims `sub`/`rol`/`deposito_id`; `ValidarToken` verifica que el método sea HMAC; `AuthMiddleware` (`middleware/auth_middleware.go:22-83`) devuelve 401 ante falta, formato inválido, firma inválida o vencido.
- **Prueba dinámica (S1):** sin header, token basura, token expirado firmado con el secreto correcto y token `alg:none` → todos 401.

### B4. ABM de usuarios solo Admin
- **Estado:** CUMPLE
- **Evidencia:** `usuario/handler.go:340-403`: registrar, listar, cambiar contraseña, activar, desactivar y cambiar rol llevan `rolMiddleware("administrador")`. Solo `/ingresar` y `/perfil` quedan abiertos (correcto).
- **Prueba dinámica (S3):** gerente → `POST /usuarios/registrar` → 403.

### B5. Usuario desactivado
- **Estado:** PARCIAL
- **Severidad:** MAYOR
- **Hallazgo:** `AuthMiddleware` (`auth_middleware.go:46-75`) valida solo el JWT y nunca consulta si el usuario sigue activo: un usuario desactivado (o con rol/depósito cambiado) sigue operando con su token vigente hasta 2 h.
- **Evidencia:** el login lo bloquea (`usuario/repository.go:41` busca con `"activado": true`), pero el middleware nunca consulta el estado actual del usuario: un token ya emitido sigue operando hasta que vence (2 h).
- **Código:** `api/internal/middleware/auth_middleware.go:46-75`
```go
claims, err := auth.ValidarToken(partes[1])
if err != nil { /* 401 */ }
usuarioID, ok := claims["sub"].(string)
...
usuarioRol, ok := claims["rol"].(string)
...
depositoID, _ := claims["deposito_id"].(string)
setUsuarioContexto(c, UsuarioContexto{UsuarioID: usuarioID, UsuarioRol: usuarioRol, DepositoID: depositoID})
c.Next()
```
- **Detalle:** el enunciado exige que un usuario desactivado "no pueda operar en el sistema"; la guía pide explícitamente que tampoco opere con un token vigente. Rol y depósito también se toman del token (si cambian el rol o el depósito, el token viejo conserva los anteriores). Corrección: buscar el usuario por id en el middleware (o en cada service) y rechazar si `activado == false`. No se probó en ejecución porque el código ya muestra el problema.

### B6. Rol único y depósito
- **Estado:** CUMPLE
- **Evidencia:** `usuario/service.go:44-57` y `176-220`: rol validado contra el conjunto fijo; Gerente/Operario exigen depósito existente; Admin/Auditor no pueden tenerlo.

### C1. Permisos por rol en cada endpoint 🔒
- **Estado:** CUMPLE
- **Evidencia:** se recorrieron todas las rutas de los 12 `RegistrarRutas`. Toda escritura tiene `rolMiddleware` coherente con la matriz (depósitos/categorías/usuarios: admin; productos/proveedores/OC: admin+gerente; movimientos/transferencias: operario+gerente+admin; novedades: gerente/auditor; stock mínimo general: admin). Los GET abiertos a todos los roles autenticados son datos maestros o están filtrados por alcance en el service.
- **Prueba dinámica (S3):** operario → `POST /productos`, `POST /ordenes-compra`, `POST /novedades/difusion`, `POST /depositos/:id/novedades` → 403.
- **Nota:** el permiso de gerente sobre el override de stock mínimo se evalúa en D5.

### C2. Restricción por depósito 🔒
- **Estado:** PARCIAL
- **Severidad:** CRÍTICO
- **Hallazgo:** `DescartarOrdenCompra` y `RegistrarRecepcion` (`ordenDeCompra/service.go:636-674`, `884-903`) no llaman a `validarAccesoDepositoOrden`: un gerente del depósito B puede descartar borradores y registrar recepciones (que suman stock) sobre OC del depósito A (confirmado en S8: stock de A 16 a 21).
- **Evidencia:** el control por depósito con el valor del token está bien resuelto en movimientos, transferencias, stock, alertas, reportes, novedades, creación/edición/confirmación de OC y manejo de ítems. Pero `DescartarOrdenCompra` y `RegistrarRecepcion` no validan el depósito de la orden: cualquier gerente puede descartar borradores y registrar recepciones (que **suman stock**) sobre OC de otros depósitos.
- **Código:** `api/internal/ordenDeCompra/service.go:636-674` (recibe `rol` y `depUsuario` y no los usa) y `service.go:884-903`
```go
func (s *OrdenDeCompraService) DescartarOrdenCompra(ctx context.Context, idUsuario string, id string,
    rol string, depUsuario string) error {
    ...
    orden, err := s.ordenCompraRepositorio.EncontrarPorId(ctx, id)
    ...
    if orden.Estado != OrdenBorrador { return ErrEstadoOrdenInvalido }
    // <- falta validarAccesoDepositoOrden(orden, rol, depUsuario)
```
```go
func (s *OrdenDeCompraService) RegistrarRecepcion(ctx context.Context, ordenCompraID string,
    dto RegistrarRecepcionDTO, idAdmin string) (OrdenCompraDTO, error) {
    // el handler (handler.go:380-385) ni siquiera pasa rol ni depósito
```
- **Detalle:** el resto de los métodos de OC llaman a `validarAccesoDepositoOrden`; estos dos se olvidaron. Escenario: gerente del depósito B llama `POST /ordenes-compra/<OC de A>/recepciones` y el stock de A sube.
- **Prueba dinámica (S8):** `gerB` → `PUT /ordenes-compra/<OC de A>/confirmar` → 403 y `GET` → 403, pero `POST .../recepciones` → 200 (OC pasa a `completada`, stock de A de 16 a 21) y `PUT .../descartar` sobre un borrador de A → 204.
- **Nota de calificación:** `PARCIAL` porque el resto de la superficie está correcta, pero se trata de un IDOR de escritura con efecto sobre stock en un ítem bloqueante: es el punto más importante a corregir. Ver también G7.

### C3. Auditor solo lectura
- **Estado:** CUMPLE
- **Evidencia:** ninguna ruta de escritura incluye `"auditor"` salvo novedades (`novedadDeposito/handler.go:252-256`, `/novedades/difusion`), permitido por enunciado.
- **Prueba dinámica (S3):** auditor → POST movimientos, depósitos, productos, proveedores, OC, transferencias, usuarios, stock-mínimo, escanear alertas → todos 403.

### D1. Depósitos
- **Estado:** CUMPLE
- **Evidencia:** `deposito/dto.go:14-24` (nombre, localidad, provincia obligatorios, responsable validado como usuario existente), rutas de escritura solo admin, stock por depósito en `deposito_producto` (clave depósito+producto).

### D2. Categorías
- **Estado:** CUMPLE
- **Evidencia:** `categoria/handler.go:134-172`: POST/PUT/DELETE solo admin, `binding:"required"` en nombre.

### D3. Productos
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** `CrearProductoDTO` (`producto/dto.go:33-41`) no tiene `binding` y `CrearProducto` no valida rangos: se puede crear un producto con nombre vacío y costo o stock mínimo negativos.
- **Evidencia:** existen todos los campos, la categoría se valida como existente y el vencimiento es opcional, pero no se valida que el costo ni el stock mínimo sean ≥ 0 ni que el nombre venga informado.
- **Código:** `api/internal/producto/dto.go:33-41`
```go
type CrearProductoDTO struct {
    Nombre           string       `json:"nombre"`
    Descripcion      string       `json:"descripcion"`
    CategoriaID      string       `json:"categoriaId"`
    UnidadMedida     UnidadMedida `json:"unidadMedida"`
    CostoUnitario    float64      `json:"costoUnitario"`
    StockMinimo      int32        `json:"stockMinimo"`
    FechaVencimiento *time.Time   `json:"fechaVencimiento,omitempty"`
}
```
- **Detalle:** sin `binding`, y `producto/service.go:208-246` (`CrearProducto`) tampoco chequea rangos. El enunciado/guía exigen costo ≥ 0 y stock mínimo ≥ 0. Se puede crear un producto con costo `-100` y nombre vacío. (Ver también J3.)

### D4. Unidad de medida cerrada
- **Estado:** CUMPLE
- **Evidencia:** `producto/service.go:esUnidadMedidaValida` (unidad, kilogramo, litro, metro), usada en alta y modificación.
- **Prueba dinámica:** `unidadMedida:"caja"` → `400 {"error":"la unidad de medida ingresada no es válida"}`.

### D5. Stock mínimo general + override
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** `PUT /depositos/.../stock/:productoId/stock-minimo` (`depositoProducto/handler.go:179-191`) usa `rolMiddleware("administrador","gerente")`: el gerente puede configurar el override de stock mínimo, que solo debería poder el administrador.
- **Evidencia:** el mínimo efectivo (override si existe, si no el general) está bien calculado en `depositoProducto/service.go:75-80`, en alertas (`alertaStock/service.go:248-252`), en la OC sugerida (`ordenDeCompra/service.go:386-389`) y en el reporte (`$ifNull`). Pero el override lo puede configurar también el Gerente.
- **Código:** `api/internal/depositoProducto/handler.go:179-191`
```go
stock.PUT(
    "/:productoId/stock-minimo",
    authMiddleware,
    rolMiddleware("administrador", "gerente"),
    handler.ConfigurarStockMinimo,
)
```
- **Detalle:** la guía (y el enunciado: "Administrador… configura el stock mínimo de cada producto") indican que lo configura solo el Admin. El stock mínimo general sí es solo admin (`producto/handler.go:310-314`).
- **Prueba dinámica (S9):** con override=2 en B y mínimo general 10, el stock en B (7) y (3) no genera OC, y al caer a 1 sí; en A (sin override) se genera con 7. El efectivo funciona; falla solo el permiso.

### D6. Proveedores
- **Estado:** CUMPLE
- **Evidencia:** `proveedor/handler.go:310-386`: ABM admin+gerente, GET para cualquier autenticado (el auditor solo lee), campos nombre (razón social), CUIT, email, teléfono, y productos que suministra.

### D7. CUIT con dígito verificador
- **Estado:** CUMPLE
- **Evidencia:** `proveedor/cuit.go:14-46`: módulo 11 con pesos 5,4,3,2,7,6,5,4,3,2, acepta guiones, longitud 11 y solo dígitos, tratamiento de dv 11→0 y 10→inválido; se usa en alta y modificación.
- **Prueba dinámica (S4):** `20-12345678-6` → 201 (con y sin guiones); `20-12345678-7` → `400 {"error":"el CUIT ingresado no es válido"}`.

### D8. Productos por proveedor
- **Estado:** CUMPLE
- **Evidencia:** colección `producto_proveedor` (`productoProveedor/model.go`, `repository.go`) con costo y `tiempo_estimado_dias` por vínculo; endpoints `POST/PUT/DELETE /proveedores/:id/productos`; un producto puede tener varios proveedores.

### E1. Todo cambio genera movimiento
- **Estado:** CUMPLE
- **Evidencia:** ingreso/egreso/ajuste en `movimientoStock/service.go:94-111`; transferencia (aprobar y completar) en `transferencia/service.go:145-162, 285-302`; recepción de OC en `ordenDeCompra/service.go:1001-1021`. Todas las rutas que llaman a `AjustarStock` registran el movimiento con producto, depósito, cantidad, tipo, usuario, fecha y motivo.
- **Prueba dinámica (S6/S7/S8):** ajuste, transferencia (dos movimientos) y recepción de OC dejaron su movimiento.

### E2. Stock nunca negativo 🔒
- **Estado:** CUMPLE
- **Evidencia:** `depositoProducto/repository.go:139-168`: `FindOneAndUpdate` con filtro `stock >= -delta` y `$inc` en la misma operación; lo usan egreso, ajuste negativo y aprobación de transferencia.
- **Prueba dinámica (S5):** egreso 10 sobre stock 5 → 409; ajuste -10 → 409; 20 egresos concurrentes de 1 sobre stock 5 → 5 × 201 y 15 × 409, stock final 0.

### E3. Historial inmutable 🔒
- **Estado:** CUMPLE
- **Evidencia:** `movimientoStock/repository.go` solo expone `Crear`, `Listar` y `EncontrarPorID`; solo hay `POST/GET` en `/movimientos`.
- **Prueba dinámica (S6):** `PUT /movimientos/xxx` y `DELETE /movimientos/xxx` → 404; el ajuste crea un movimiento nuevo.

### E4. Validaciones de movimiento
- **Estado:** PARCIAL
- **Severidad:** MAYOR
- **Hallazgo:** `movimientoStock/service.go:46-82` no verifica que el producto ni el depósito existan y `AsegurarExiste` hace upsert: un ingreso con `producto_id` inexistente responde 201 y crea stock y movimiento fantasma.
- **Evidencia:** cantidad (> 0, o con signo en ajuste), motivo obligatorio, tipo en conjunto cerrado y Operario solo en su depósito están bien (`movimientoStock/dto.go:10-16`, `service.go:42-76`). Pero no se verifica que el producto ni el depósito existan: `AsegurarExiste` hace upsert de una fila de stock para cualquier par de ObjectID válidos.
- **Código:** `api/internal/movimientoStock/service.go:46-82`
```go
depoOID, err := bson.ObjectIDFromHex(dto.DepositoID)
...
prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
...
// 1. Asegurar que exista el registro de stock para este producto en el depósito
_, err = s.depositoStockRepo.AsegurarExiste(ctx, depoOID, prodOID, uOID)
```
- **Detalle:** un `ingreso` con `producto_id` inexistente responde 201 y deja stock y movimiento "fantasma" (el reporte de inventario los descarta recién en el `$lookup`). El service no tiene los repos de producto/depósito. Corrección: validar existencia antes de tocar el stock.

### E5. Consistencia stock ↔ movimiento
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** En `movimientoStock/service.go:85-111` el stock se modifica y después se inserta el movimiento, sin reversión si falla la segunda escritura (igual en `transferencia/service.go:131-162` y `ordenDeCompra/service.go:978-1020`): queda stock cambiado sin movimiento que lo respalde.
- **Evidencia:** el stock se modifica primero y el movimiento se inserta después, sin reversión si falla la segunda escritura. La guía acepta ese orden solo "con reversión manual si falla".
- **Código:** `api/internal/movimientoStock/service.go:85-111`
```go
_, err = s.depositoStockRepo.AjustarStock(ctx, depoOID, prodOID, delta, uOID)
...
movimiento, err := s.repo.Crear(ctx, mov)
if err != nil {
    return MovimientoStock{}, err   // el stock ya se movió
}
```
- **Detalle:** mismo patrón en `transferencia/service.go:131-162` (aprobar: descuenta, y si falla el movimiento o `ActualizarEstado` queda stock descontado y la transferencia en `solicitada`) y `ordenDeCompra/service.go:978-1020` (recepción). No hay transacción ni compensación.

### F1. Solicitud
- **Estado:** PARCIAL
- **Severidad:** MAYOR
- **Hallazgo:** `transferencia/service.go:53-82` no valida que el depósito destino exista: una solicitud con `deposito_destino_id` inexistente responde 201 y al completarse acreditaría stock en un depósito que no existe (S7).
- **Evidencia:** valida origen ≠ destino, pertenencia del solicitante al origen, cantidad > 0 y que no toque stock. No valida que el depósito destino exista (el origen y el producto quedan validados de forma indirecta porque se exige fila de stock en origen).
- **Código:** `api/internal/transferencia/service.go:53-82`
```go
if dto.DepositoOrigenID == dto.DepositoDestinoID { return Transferencia{}, ErrMismoDeposito }
...
dp, err := s.depositoStockRepo.EncontrarPorDepositoYProducto(ctx, origOID, prodOID)
if err != nil || dp.Stock < dto.Cantidad { return Transferencia{}, ErrStockInsuficiente }
```
- **Prueba dinámica (S7):** `deposito_destino_id` = `507f1f77bcf86cd799439011` (inexistente) → `201` con estado `solicitada`. Al completarse, `AsegurarExiste` crearía stock en un depósito que no existe.

### F2. Flujo de estados
- **Estado:** CUMPLE
- **Evidencia:** `transferencia/service.go:116-118, 203-205, 244-246`: solo `solicitada→aprobada/rechazada` y `aprobada→completada`; el resto devuelve `ErrEstadoInvalido` (409).
- **Prueba dinámica (S7):** completar sin aprobar → 409; aprobar dos veces → 409; rechazar una aprobada → 409; completar dos veces → 409; aprobar/completar una rechazada → 409.
- **Nota:** la transición no es atómica; ver F3.

### F3. Descuento y acreditación
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** `Aprobar` (`transferencia/service.go:111-142`) lee, chequea el estado y escribe sin filtro por estado (`repository.go:138`): dos aprobaciones simultáneas descuentan dos veces el stock de origen (lo mismo en `Completar`, doble acreditación).
- **Evidencia:** el origen se descuenta al aprobar con update condicional (`AjustarStock`), el destino se acredita al completar y cada paso deja su movimiento (S7: A 10→6, B 0→4, dos movimientos de tipo `transferencia`). Pero la transición de estado es lectura + chequeo + escritura incondicional, por lo que dos aprobaciones simultáneas pasan el chequeo y descuentan dos veces.
- **Código:** `api/internal/transferencia/service.go:111-142` y `repository.go:138`
```go
t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
if t.Estado != TransferenciaSolicitada { return ErrEstadoInvalido }
...
_, err = s.depositoStockRepo.AjustarStock(ctx, t.DepositoOrigenID, t.ProductoID, -t.Cantidad, uOID)
```
```go
r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": setFields})  // sin filtro por estado
```
- **Detalle:** el guard debería ser un `UpdateOne` con filtro `{_id, estado: "solicitada"}` antes de mover stock. Misma causa en `Completar` (doble acreditación). No se ejecutó prueba concurrente.

### F4. Quién aprueba
- **Estado:** CUMPLE
- **Evidencia:** `transferencia/service.go:121-124` y ruta `rolMiddleware("gerente","administrador")`.
- **Prueba dinámica (S7):** operario destino → 403; gerente del origen → 403; gerente de un tercer depósito → 403; gerente del destino → 200.

### F5. Rechazo
- **Estado:** CUMPLE
- **Evidencia:** `service.go:189-228` no toca stock y la transferencia queda en estado final (aprobar/completar posteriores → 409; stock de A sin cambios, S7).

### G1. Creación
- **Estado:** CUMPLE
- **Evidencia:** `ordenDeCompra/service.go:87-139`, `676-747`: Gerente (solo su depósito) o Admin crean la OC en borrador; los ítems exigen que el proveedor de la orden suministre el producto y toman el costo del vínculo proveedor-producto (`item.CostoUnitario = productoProveedor.CostoUnidad`).
- **Prueba dinámica (S8):** ítem agregado con `costoUnitario:7` (el del vínculo); gerente A creando OC para B → 403.

### G2. Borrador y confirmada
- **Estado:** CUMPLE
- **Evidencia:** edición de OC e ítems y descarte solo en `OrdenBorrador`; confirmar exige al menos un ítem; ninguno toca stock.
- **Prueba dinámica (S8):** confirmada: editar ítem, agregar ítem, editar OC y descartar → 409.

### G3. Recepciones
- **Estado:** CUMPLE
- **Evidencia:** `service.go:884-1062`: solo confirmada/parcial; valida todo antes de modificar; `$inc` de `cantidad_recibida` por ítem; genera movimiento de ingreso; estado `parcial` o `completada` según el acumulado por ítem.
- **Prueba dinámica (S8):** recepción 4 de 10 → `parcial` (stock +4); recepción 6 → `completada` (stock +6). El control de acceso por depósito se evalúa en C2/G7.

### G4. Estado completada
- **Estado:** CUMPLE
- **Evidencia:** `service.go:901-904` y validación de acumulado (`939-943`).
- **Prueba dinámica (S8):** recibir sobre `completada` → 409; recibir 11 sobre pedido 10 → 409 (sobrerrecepción rechazada).

### G5. OC sugerida automática
- **Estado:** CUMPLE
- **Evidencia:** `main.go:120-142` + `ordenDeCompra/service.go:315-465`; se invoca desde egreso/ajuste negativo (`movimientoStock/service.go:115-117`) y aprobación de transferencia (`transferencia/service.go:166-172`); usa el mínimo efectivo, elige el proveedor de menor costo y crea/reutiliza un borrador por proveedor y depósito.
- **Prueba dinámica (S9):** en A (sin override, mínimo 10) un egreso que deja 7 genera borrador con `cantidadSolicitada:3, costoUnitario:9`; en B (override 2) con stock 7 y 3 no se genera, y al caer a 1 se genera con cantidad 1.

### G6. Revisión de la sugerida
- **Estado:** CUMPLE
- **Evidencia:** la sugerida es un borrador común: se modifica (PUT de orden e ítems), se confirma o se descarta con las mismas rutas admin+gerente. El hueco de alcance en `descartar` se registra en C2/G7.

### G7. Permisos y alcance de OC
- **Estado:** PARCIAL
- **Severidad:** CRÍTICO (misma causa que C2)
- **Hallazgo:** Por la misma causa que C2, el gerente solo está acotado a su depósito para leer/crear/editar/confirmar OC, pero puede descartar y registrar recepciones de OC de otros depósitos (`ordenDeCompra/service.go:636-674`, `884-903`).
- **Evidencia:** Operario no crea ni confirma (403, S8), Auditor lee todas (200), Gerente solo lee/crea/edita/confirma las de su depósito, pero puede descartar y recibir las de otros depósitos.
- **Código:** ver C2 (`ordenDeCompra/service.go:636-674` y `884-903`).
- **Detalle:** solo se marca lo afectado por esa causa raíz: el alcance del Gerente en descartar y recibir.

### H1. Alerta de stock bajo mínimo
- **Estado:** CUMPLE
- **Evidencia:** `alertaStock/service.go:171-294` (escaneo con mínimo efectivo y alcance por rol, `resolverAlcanceDeposito`), y `GET /depositos/:id/stock?bajoMinimo=true` calcula el mismo criterio en vivo.
- **Prueba dinámica (S9):** escaneo de gerente A → alerta `stock_bajo` con producto/depósito y mínimo 10; operario A pidiendo alertas de B → 403.
- **Nota:** el reporte `/reportes/alertas` devuelve datos mal mapeados; se evalúa en I3 (ver I2).

### H2. Próximos a vencer
- **Estado:** CUMPLE
- **Evidencia:** `alertaStock/service.go:300-348` y `handler.go:114-126`: ventana `dias_vencimiento` configurable por parámetro (default 30, valida ≥ 0), solo productos con fecha de vencimiento.
- **Prueba dinámica (S9):** producto con vencimiento a 10 días → alerta `vencimiento` con ventana 30; con `dias_vencimiento=5` el reporte no lo incluye.

### H3. Envío de novedades
- **Estado:** CUMPLE
- **Evidencia:** `novedadDeposito/service.go:83-169` y rutas (`handler.go:252-288`): gerente solo a su depósito; auditor a un depósito (`POST /depositos/:id/novedades`), varios o todos (`POST /novedades/difusion`); guarda texto y `fecha_envio`; operario no envía.
- **Prueba dinámica (S10):** auditor a A y B → 201 (un documento por depósito); gerente A a B → 403; gerente A a `/novedades/difusion` → 403; operario → 403.

### H4. Consulta de novedades y no leídas
- **Estado:** PARCIAL
- **Severidad:** MAYOR
- **Hallazgo:** `NovedadDeposito` guarda un único booleano `leido` (`novedadDeposito/model.go:26-34`) y el contador cuenta por depósito (`repository.go:122-132`): el contador de no leídas no es por operario (S10: opA2 ve 1 en lugar de 2 tras leer opA una).
- **Evidencia:** el operario ve solo las de su depósito (S10: operario de B pidiendo A → 403), pero el estado de lectura es un único booleano por novedad, compartido por todos los operarios del depósito. El contador de no leídas es global del depósito, no por operario.
- **Código:** `api/internal/novedadDeposito/model.go:26-34` y `repository.go:122-132`
```go
type NovedadDeposito struct {
    ...
    Leido      bool              `bson:"leido"`
    Respuesta  *RespuestaNovedad `bson:"respuesta,omitempty"`
```
```go
return r.coleccion.CountDocuments(ctx, bson.M{"deposito_id": depositoID, "leido": false})
```
- **Detalle:** la guía pide contador por operario. Se necesita guardar los lectores (p. ej. `leido_por: []ObjectID`) y contar según el usuario del token.
- **Prueba dinámica (S10):** 2 novedades para A; opA marca una como leída; opA2 (mismo depósito, no leyó nada) ve `no_leidas:1` en lugar de 2.

### H5. Marcar como leída y responder
- **Estado:** PARCIAL
- **Severidad:** MAYOR (misma causa que H4)
- **Hallazgo:** Marcar como leída o responder hace `$set leido:true` sobre la novedad completa (`novedadDeposito/repository.go:135-175`): la lectura de un operario la marca para todos los del depósito, y una segunda respuesta pisa la primera.
- **Evidencia:** `Responder` guarda la respuesta con `usuario_id` y fecha (`novedadDeposito/service.go:321-336`, cumple). Pero marcar como leída (y responder, que también marca leída) afecta a todos los operarios del depósito.
- **Código:** `api/internal/novedadDeposito/repository.go:135-175`
```go
bson.M{"$set": bson.M{"leido": true}}
```
- **Detalle:** ver H4. Además, una sola `respuesta` embebida: un segundo operario pisa la respuesta del primero. Está aceptado por el enunciado (sin hilo), pero conviene saber que se sobrescribe.

### I1. Stock y valorización
- **Estado:** PARCIAL
- **Severidad:** MAYOR
- **Hallazgo:** Los structs `ResultadoInventarioDeposito/Categoria` (`reporte/model.go:44-56`) no tienen tags `bson`, por lo que `/reportes/inventario` devuelve `deposito_id`/`categoria_id` en cero y nombres vacíos; además el operario obtiene valorización por producto desde `GET /depositos/:id/stock`.
- **Evidencia:** la agregación está en la base (`reporte/repository.go:105-300`, `$lookup`/`$group`/`$multiply`), los totales son correctos y el Operario recibe 403 en `/reportes/inventario`. Pero las filas agrupadas pierden identificación: `ResultadoInventarioDeposito/Categoria` no tienen tags `bson`, y el `$project` emite `deposito_id`/`deposito_nombre`.
- **Código:** `api/internal/reporte/model.go:44-56`
```go
type ResultadoInventarioDeposito struct {
    DepositoID     bson.ObjectID   // sin `bson:"deposito_id"`
    DepositoNombre string
    Stock          int32
    Valorizado     float64
}
```
- **Detalle:** el driver busca la clave `depositoid`, así que `deposito_id`/`deposito_nombre` y `categoria_id`/`categoria_nombre` quedan en valor cero. Además el Operario puede obtener valorización por producto desde `GET /depositos/:id/stock` (`valorizado` y `costoUnitario` en `StockDepositoDTO`, `depositoProducto/dto.go:7-21`, ruta abierta al operario), contra lo que pide la guía (si la puede obtener, queda `PARCIAL`).
- **Prueba dinámica (S3):** `GET /reportes/inventario` → `{"por_deposito":[{"deposito_id":"000000000000000000000000","deposito_nombre":"","stock":3,"valorizado":30}], ...}`; operario → `GET /reportes/inventario` 403 pero `GET /depositos/<A>/stock/<P1>` → 200 con `"valorizado":0` (campo presente).

### I2. Movimientos agrupados
- **Estado:** NO CUMPLE
- **Severidad:** MAYOR
- **Hallazgo:** `ResultadoMovimiento` (`reporte/model.go:58-65`) no tiene tags `bson`: `GET /reportes/movimientos` devuelve `cantidad_eventos:0`, `unidades_netas:0` y depósito sin id ni nombre aun con 20 egresos registrados, por lo que el reporte es inutilizable.
- **Evidencia:** el reporte agrupa por fecha/depósito/tipo con filtro de rango, pero la respuesta sale con los valores perdidos por la misma falta de tags `bson`: `cantidad_eventos` y `unidades_netas` quedan en 0 aunque hay movimientos, y el depósito sin id ni nombre.
- **Código:** `api/internal/reporte/model.go:58-65`
```go
type ResultadoMovimiento struct {
    DepositoID      bson.ObjectID
    DepositoNombre  string
    Fecha           string
    Tipo            string
    CantidadEventos int64   // el pipeline emite "cantidad_eventos"
    UnidadesNetas   int32   // el pipeline emite "unidades_netas"
}
```
- **Detalle:** causa raíz: structs de `reporte/model.go` sin tags `bson` (afecta también I1, I3). El método existe con el nombre correcto pero devuelve datos incorrectos: no se puede construir el dashboard de movimientos. Corrección: agregar `bson:"..."` a todos los campos de los `Resultado*`.
- **Prueba dinámica (S3/S9):** `GET /reportes/movimientos` → `{"deposito_id":"000000000000000000000000","deposito_nombre":"","fecha":"2026-10-02","tipo":"egreso","cantidad_eventos":0,"unidades_netas":0}` con 20 egresos registrados ese día.

### I3. Alertas y pendientes
- **Estado:** PARCIAL
- **Severidad:** MAYOR (misma causa que I2)
- **Hallazgo:** Por la misma falta de tags `bson` (`reporte/model.go:67-101`), `/reportes/alertas` y `/reportes/pendientes` devuelven ids y nombres vacíos, `stock_minimo:0`, `fecha_vencimiento:"0001-01-01"` y `dias_para_vencer:0`.
- **Evidencia:** el pipeline y los filtros de alcance están bien; el resultado sale con ids y nombres en cero, `stock_minimo:0`, `fecha_vencimiento:"0001-01-01"` y `dias_para_vencer:0` en alertas, e ids/nombres vacíos en pendientes (los contadores de una sola palabra como `borrador`, `confirmada` sí salen bien). Existen alternativas funcionales (`/alertas-stock`, H1/H2), por eso `PARCIAL` y no `NO CUMPLE`.
- **Código:** `api/internal/reporte/model.go:67-101` (mismo problema de tags).
- **Prueba dinámica (S9):** `GET /reportes/alertas?dias_vencimiento=30` → `{"bajo_minimo":[{"deposito_id":"0000…","producto_nombre":"","stock":7,"stock_minimo":0}],"por_vencer":[{"fecha_vencimiento":"0001-01-01","dias_para_vencer":0,...}]}`; `GET /reportes/pendientes` → `{"deposito_id":"0000…","deposito_nombre":"","borrador":5,...}`.

### I4. Filtros y paginación
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** `tamanioPagina` no tiene tope máximo (`ordenDeCompra/handler.go:137-143`, patrón repetido): `tamanioPagina=1000000` trae toda la colección; además `GET /usuarios` no pagina y `ListarStockDeposito` carga todo en memoria antes de paginar.
- **Evidencia:** los listados de movimientos, transferencias, OC, productos, alertas, stock y novedades tienen filtros y `pagina/tamanioPagina` con total y default 20, pero ninguno limita el tamaño máximo, y varios listados no paginan.
- **Código:** `api/internal/ordenDeCompra/handler.go:137-143` (patrón repetido en los demás handlers)
```go
tamanioStr := c.Query("tamanioPagina")
if tamanioStr != "" {
    if tamanio, err := strconv.Atoi(tamanioStr); err == nil && tamanio > 0 {
        filtro.TamanioPagina = tamanio   // sin tope
    }
}
```
- **Detalle:** `tamanioPagina=1000000` trae toda la colección (la guía pide "límite máximo"). `GET /usuarios` (`usuario/service.go:240-258`) devuelve toda la colección, y `ListarStockDeposito` carga todos los productos en memoria antes de paginar (`depositoProducto/service.go:170`). El reporte usa tamaño fijo 10 (aceptable).

### I5. Alcance por rol en reportes
- **Estado:** CUMPLE
- **Evidencia:** `reporte/service.go:resolverAlcanceDeposito` y `movimientoStock/service.go:122-130`: Gerente/Operario quedan forzados a su depósito (403 si piden otro por query), Admin/Auditor ven todo.
- **Prueba dinámica (S3):** operario A con `?deposito_id=<B>` en `/reportes/movimientos`, `/reportes/alertas`, `/alertas-stock` → 403; `/movimientos?deposito_id=<B>` devuelve solo datos de A; gerente A en `/reportes/inventario?deposito_id=<B>` → 403.

### J1. Auditoría de documentos
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** Marcar una novedad como leída o responderla (`novedadDeposito/repository.go:149-165, 205-211`) no actualiza `auditoria.modificado_por` ni `auditoria.actualizado_en`, por lo que esos cambios quedan sin rastro de quién y cuándo.
- **Evidencia:** todos los models embeben `auditoria.Auditoria` (creador, fecha de creación, último modificador, última actualización), se completan en el service con el usuario del token y se actualizan en usuarios, productos, proveedores, transferencias, OC, stock. Pero las novedades no actualizan la auditoría al marcarse como leídas o respondidas.
- **Código:** `api/internal/novedadDeposito/repository.go:149-165, 205-211`
```go
bson.M{"$set": bson.M{"leido": true}}
...
update := bson.M{"$set": bson.M{"leido": true, "respuesta": respuesta}}
```
- **Detalle:** falta `auditoria.modificado_por` y `auditoria.actualizado_en` en esas dos modificaciones. Los movimientos y alertas solo guardan creación (son inmutables/automáticos; sin penalización por eso).

### J2. Trazabilidad de operaciones sensibles
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** Los cambios de rol, activación/desactivación y contraseña de usuarios (`usuario/repository.go:124-130, 177-186`) solo sobrescriben el último modificador; no hay log ni historial, así que no queda registro auditable de qué se cambió ni de los cambios anteriores.
- **Evidencia:** los movimientos quedan registrados (cuentan según la guía), las transferencias guardan `usuario_resolucion_id` y `fecha_resolucion`, y los usuarios guardan `modificado_por`/`actualizado_en`. No hay un log o colección de auditoría propia: los cambios de rol, activación/desactivación y contraseña de usuarios sobrescriben el último modificador y no queda historial de qué se cambió.
- **Código:** `api/internal/usuario/repository.go:124-130, 177-186`
```go
setFields := bson.M{"rol": rol, "auditoria.modificado_por": oidAdmin, "auditoria.actualizado_en": time.Now()}
...
"$set": bson.M{"activado": false, "auditoria.modificado_por": oidAdmin, "auditoria.actualizado_en": ahora}
```
- **Detalle:** el enunciado pide altas/bajas/cambios de usuarios "registradas de forma auditable"; con el esquema actual solo se conoce la última modificación y no su naturaleza.

### J3. Validación de entrada
- **Estado:** PARCIAL
- **Severidad:** MENOR
- **Hallazgo:** Productos, proveedores, OC e ítems no tienen `binding` ni validación de rangos (`ordenDeCompra/dto.go:9-12`): `POST /ordenes-compra` con `{}` falla con un 500 por `ObjectIDFromHex`, y se aceptan proveedores sin nombre/email y costos sin rango.
- **Evidencia:** usuarios, depósitos, categorías, movimientos, transferencias y novedades validan con `binding`. Productos, proveedores, OC y ítems no tienen `binding` ni validaciones de rango/obligatoriedad; los ObjectID inválidos terminan en 500.
- **Código:** `api/internal/ordenDeCompra/dto.go:9-12`
```go
type CrearOrdenCompraDTO struct {
    ProveedorID string `json:"proveedorId"`
    DepositoID  string `json:"depositoId"`
}
```
- **Detalle:** `POST /ordenes-compra` con `{}` devuelve 4xx/5xx según el caso (el error de `ObjectIDFromHex` llega como 500). Proveedor sin nombre/email se acepta, costos de vínculo y producto sin rango. No se vio inyección NoSQL: los filtros usan ObjectID parseados o strings tipados.

## Mejoras sugeridas (no puntúan ni bajan la nota)
- Guardar el estado de la transición con `UpdateOne` condicional por estado en transferencias (y evitar doble aprobación/doble completar).
- Generar la OC sugerida de forma atómica: 5 egresos concurrentes crearon 5 borradores distintos para el mismo depósito/proveedor (la guía no exige evitar duplicados).
- La sobrerrecepción se valida en memoria antes del `$inc`; dos recepciones simultáneas podrían pasarse del pedido.
- `GET /movimientos/:id` y similares: 404 para inexistente y 400 para id inválido.
- `Registrar` usuario chequea email duplicado con el filtro `activado:true` (`EncontarPorCorreo`), por lo que se puede volver a registrar el correo de un usuario desactivado y quedar dos usuarios con el mismo email.
- Leer rol/depósito del usuario en cada request (o invalidar tokens al cambiar rol/depósito).
- `EncontrarPorIds` en difusión de novedades no verifica que se hayan encontrado todos los depósitos pedidos; un id inexistente genera una novedad huérfana.

## Observaciones (no puntúan)
- `go.mod` está en `api/` (estructura esperada). Dockerfile/compose/volumen y README no se evaluaron: el compose define un volumen nombrado para Mongo y no define `INITIAL_ADMIN_*`.
- `router.Run(":8080")` está fijo en `main.go`; el `.env.example` incluye un valor de ejemplo de `JWT_SECRET` y la password del admin inicial.
- Código sin uso: `MovimientoStock.ConvertirAModelo` (`movimientoStock/dto.go:64`), `EliminarOrdenCompra`, `OrdenDeCompraService.ActualizarEstadoOrdenCompra`, `OrdenDeCompraService.EncontrarTodos` (listado sin paginar).
- Comentarios que no coinciden con el código: `depositoProducto/repository.go:14-16` dice "responder 400" y el handler responde 409; `ordenDeCompra/service.go:883` ("Revisar si hacer eliminación fisica o logica") está sobre `RegistrarRecepcion`, que no elimina nada.

## Preguntas para el coloquio
1. En `reporte/model.go`, ¿cómo mapea el driver de Mongo un campo `DepositoID` a la clave `deposito_id` que emite el `$project`? ¿Qué devuelve hoy `GET /reportes/movimientos` en `cantidad_eventos` y por qué?
2. En `RegistrarRecepcion` y `DescartarOrdenCompra`, ¿dónde se verifica que el gerente pertenece al depósito de la OC? ¿Qué pasa si el gerente del depósito B llama a `/ordenes-compra/<OC de A>/recepciones`?
3. El middleware JWT toma rol y depósito del token. Si un admin desactiva a un usuario o le cambia el rol, ¿qué ve ese usuario con su token vigente?
4. Si `AjustarStock` funciona y `movimientoRepo.Crear` falla en `Registrar`, ¿en qué estado queda el stock? ¿Qué harían para evitarlo (transacción, compensación)?
5. Dos gerentes aprueban a la vez la misma transferencia: ¿cuántas veces se descuenta el stock de origen y por qué?
6. El campo `leido` de una novedad es un booleano: ¿cómo se obtiene el contador de no leídas "por operario" que pide el enunciado?
7. ¿Por qué el comentario sobre eliminación física/lógica aparece sobre `RegistrarRecepcion`, y por qué el endpoint de OC `/sugerir` y el escaneo de alertas conviven con la generación automática por callback?
