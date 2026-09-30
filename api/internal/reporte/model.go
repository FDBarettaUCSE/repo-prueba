// Package reporte no tiene una colección ni una entidad propia: agrega y
// expone información que ya vive en las colecciones de los demás módulos
// (depósitos, productos, stock, movimientos, transferencias, órdenes de
// compra). Por eso este archivo no define ningún modelo — los DTO de salida
// de cada reporte están en dto.go, y la agregación en sí en service.go.
package reporte
