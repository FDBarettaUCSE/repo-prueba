export function formatearFecha(fecha: string): string {
  const fechaConvertida = new Date(fecha);

  if (Number.isNaN(fechaConvertida.getTime())) {
    return fecha;
  }

  return fechaConvertida.toLocaleString("es-AR", {
    timeZone: "America/Argentina/Buenos_Aires",
    hourCycle: "h23",
  });
}