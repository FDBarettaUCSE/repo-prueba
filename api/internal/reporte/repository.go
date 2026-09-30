// Este módulo no tiene repository.go propio: no hay una colección de
// "reportes" para leer o escribir. Service.go arma cada reporte combinando,
// en memoria, lo que devuelven los repositorios de deposito, producto,
// categoria, depositoProducto, movimientoStock, transferencia y
// ordenDeCompra — inyectados como dependencias del Service (ver
// NuevoService).
package reporte
