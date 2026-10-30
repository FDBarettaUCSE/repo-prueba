
import type {
  RegistrarMovimientoDTO,
  MovimientoStockDTO,
  FiltroMovimientosDTO,
  MovimientosPaginadosDTO,
} from "./types";

import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import { API_URL } from "./config";

// Registrar un movimiento de stock
export async function registrarMovimiento(
  dto: RegistrarMovimientoDTO,
  token: string,
): Promise<MovimientoStockDTO> {
  const respuesta = await fetch(`${API_URL}/movimientos`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<MovimientoStockDTO>(respuesta);
}

// Listar movimientos con filtros y paginación opcionales
export async function listarMovimientos(
  token: string,
  filtros: FiltroMovimientosDTO = {},
): Promise<MovimientosPaginadosDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.producto_id) {
    parametros.set("producto_id", filtros.producto_id);
  }

  if (filtros.tipo) {
    parametros.set("tipo", filtros.tipo);
  }

  if (filtros.fecha_desde) {
    parametros.set("fecha_desde", filtros.fecha_desde);
  }

  if (filtros.fecha_hasta) {
    parametros.set("fecha_hasta", filtros.fecha_hasta);
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  if (filtros.tamanioPagina !== undefined) {
    parametros.set("tamanioPagina", String(filtros.tamanioPagina));
  }

  const query = parametros.toString();
  const url = query
    ? `${API_URL}/movimientos?${query}`
    : `${API_URL}/movimientos`;

  const respuesta = await fetch(url, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<MovimientosPaginadosDTO>(respuesta);
}

// Buscar un movimiento por ID
export async function buscarMovimientoPorID(
  id: string,
  token: string,
): Promise<MovimientoStockDTO> {
  const respuesta = await fetch(`${API_URL}/movimientos/${id}`, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<MovimientoStockDTO>(respuesta);
}


