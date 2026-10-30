
import { API_URL } from "./config";
import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";

import type {
  FiltroInventario,
  ReporteInventarioDTO,
  FiltroReporteMovimientos,
  ReporteMovimientosDTO,
  FiltroAlertas,
  ReporteAlertasDTO,
  FiltroPendientes,
  ReportePendientesDTO,
} from "./types";

// ==================== INVENTARIO ====================

export async function obtenerReporteInventario(
  token: string,
  filtros: FiltroInventario = {},
): Promise<ReporteInventarioDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.categoria_id) {
    parametros.set("categoria_id", filtros.categoria_id);
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  const respuesta = await fetch(
    `${API_URL}/reportes/inventario?${parametros}`,
    {
      method: "GET",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ReporteInventarioDTO>(respuesta);
}

// ==================== MOVIMIENTOS ====================

export async function obtenerReporteMovimientos(
  token: string,
  filtros: FiltroReporteMovimientos = {},
): Promise<ReporteMovimientosDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
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

  const respuesta = await fetch(
    `${API_URL}/reportes/movimientos?${parametros}`,
    {
      method: "GET",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ReporteMovimientosDTO>(respuesta);
}

// ==================== ALERTAS ====================

export async function obtenerReporteAlertas(
  token: string,
  filtros: FiltroAlertas = {},
): Promise<ReporteAlertasDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.dias_vencimiento !== undefined) {
    parametros.set(
      "dias_vencimiento",
      String(filtros.dias_vencimiento),
    );
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  const respuesta = await fetch(
    `${API_URL}/reportes/alertas?${parametros}`,
    {
      method: "GET",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ReporteAlertasDTO>(respuesta);
}

// ==================== PENDIENTES ====================

export async function obtenerReportePendientes(
  token: string,
  filtros: FiltroPendientes = {},
): Promise<ReportePendientesDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  const respuesta = await fetch(
    `${API_URL}/reportes/pendientes?${parametros}`,
    {
      method: "GET",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ReportePendientesDTO>(respuesta);
}
