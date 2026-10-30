import { API_URL } from "./config";
import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";

import type {
  AlertasPaginadasDTO,
  FiltrosAlertasStock,
} from "./types";

// Lista las alertas de stock con filtros y paginación.
export async function listarAlertasStock(
  token: string,
  filtros: FiltrosAlertasStock = {},
): Promise<AlertasPaginadasDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.solo_pendientes !== undefined) {
    parametros.set(
      "solo_pendientes",
      String(filtros.solo_pendientes),
    );
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  if (filtros.tamanioPagina !== undefined) {
    parametros.set("tamanioPagina", String(filtros.tamanioPagina));
  }

  const query = parametros.toString();
  const url = `${API_URL}/alertas-stock${query ? `?${query}` : ""}`;

  const respuesta = await fetch(url, {
    method: "GET",
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<AlertasPaginadasDTO>(respuesta);
}

