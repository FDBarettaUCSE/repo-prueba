import { API_URL } from "./config";
import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";

import type {
  EnviarNovedadDTO,
  EnviarNovedadDifusionDTO,
  FiltrosNovedades,
  NovedadDTO,
  NovedadesPaginadoDTO,
  ResponderNovedadDTO,
} from "./types";

// Envía una novedad a un depósito puntual.
export async function enviarNovedadADeposito(
  token: string,
  depositoId: string,
  dto: EnviarNovedadDTO,
): Promise<NovedadDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/novedades`,
    {
      method: "POST",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<NovedadDTO>(respuesta);
}

// Envía una novedad a varios depósitos o a todos.
export async function enviarNovedadDifusion(
  token: string,
  dto: EnviarNovedadDifusionDTO,
): Promise<NovedadDTO[]> {
  const respuesta = await fetch(`${API_URL}/novedades/difusion`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<NovedadDTO[]>(respuesta);
}

// Lista las novedades de un depósito con filtros y paginación.
export async function listarNovedadesDeposito(
  token: string,
  depositoId: string,
  filtros: FiltrosNovedades = {},
): Promise<NovedadesPaginadoDTO> {
  const parametros = new URLSearchParams();

  if (filtros.solo_no_leidas !== undefined) {
    parametros.set(
      "solo_no_leidas",
      String(filtros.solo_no_leidas),
    );
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  if (filtros.tamanioPagina !== undefined) {
    parametros.set("tamanioPagina", String(filtros.tamanioPagina));
  }

  const query = parametros.toString();
  const url =
    `${API_URL}/depositos/${depositoId}/novedades` +
    (query ? `?${query}` : "");

  const respuesta = await fetch(url, {
    method: "GET",
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<NovedadesPaginadoDTO>(respuesta);
}

// Obtiene una novedad por su ID.
export async function obtenerNovedad(
  token: string,
  depositoId: string,
  id: string,
): Promise<NovedadDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/novedades/${id}`,
    {
      method: "GET",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<NovedadDTO>(respuesta);
}

// Marca una novedad como leída.
export async function marcarNovedadLeida(
  token: string,
  depositoId: string,
  id: string,
): Promise<NovedadDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/novedades/${id}/leida`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<NovedadDTO>(respuesta);
}

// Envía una respuesta a una novedad.
export async function responderNovedad(
  token: string,
  depositoId: string,
  id: string,
  dto: ResponderNovedadDTO,
): Promise<NovedadDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/novedades/${id}/responder`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<NovedadDTO>(respuesta);
}