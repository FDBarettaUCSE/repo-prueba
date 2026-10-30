
import type {
  SolicitarTransferenciaDTO,
  ResolverTransferenciaDTO,
  CompletarTransferenciaDTO,
  TransferenciaDTO,
  FiltroTransferenciasDTO,
  TransferenciasPaginadasDTO,
} from "./types";

import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import { API_URL } from "./config";

// Listar transferencias con filtros y paginación opcionales
export async function listarTransferencias(
  token: string,
  filtros: FiltroTransferenciasDTO = {},
): Promise<TransferenciasPaginadasDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.estado) {
    parametros.set("estado", filtros.estado);
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  if (filtros.tamanioPagina !== undefined) {
    parametros.set("tamanioPagina", String(filtros.tamanioPagina));
  }

  const query = parametros.toString();
  const url = query
    ? `${API_URL}/transferencias?${query}`
    : `${API_URL}/transferencias`;

  const respuesta = await fetch(url, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<TransferenciasPaginadasDTO>(respuesta);
}

// Buscar una transferencia por ID
export async function buscarTransferenciaPorID(
  id: string,
  token: string,
): Promise<TransferenciaDTO> {
  const respuesta = await fetch(
    `${API_URL}/transferencias/${id}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<TransferenciaDTO>(respuesta);
}

// Solicitar una transferencia entre depósitos
export async function solicitarTransferencia(
  dto: SolicitarTransferenciaDTO,
  token: string,
): Promise<TransferenciaDTO> {
  const respuesta = await fetch(`${API_URL}/transferencias`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<TransferenciaDTO>(respuesta);
}

// Aprobar una transferencia
export async function aprobarTransferencia(
  id: string,
  token: string,
): Promise<{ mensaje: string }> {
  const respuesta = await fetch(
    `${API_URL}/transferencias/${id}/aprobar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<{ mensaje: string }>(respuesta);
}

// Rechazar una transferencia
export async function rechazarTransferencia(
  id: string,
  dto: ResolverTransferenciaDTO,
  token: string,
): Promise<{ mensaje: string }> {
  const respuesta = await fetch(
    `${API_URL}/transferencias/${id}/rechazar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<{ mensaje: string }>(respuesta);
}

// Completar una transferencia
export async function completarTransferencia(
  id: string,
  dto: CompletarTransferenciaDTO,
  token: string,
): Promise<{ mensaje: string }> {
  const respuesta = await fetch(
    `${API_URL}/transferencias/${id}/completar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<{ mensaje: string }>(respuesta);
}
