
import type {
  DepositoDTO,
  CrearDepositoDTO,
  ActualizarDepositoDTO,
} from "./types";

import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import { API_URL } from "./config";

// Listar todos los depósitos
export async function listarDepositos(
  token: string,
): Promise<DepositoDTO[]> {
  const respuesta = await fetch(`${API_URL}/depositos`, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<DepositoDTO[]>(respuesta);
}

// Buscar un depósito por ID
export async function buscarDepositoPorID(
  id: string,
  token: string,
): Promise<DepositoDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${id}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<DepositoDTO>(respuesta);
}

// Crear un depósito
export async function crearDeposito(
  dto: CrearDepositoDTO,
  token: string,
): Promise<DepositoDTO> {
  const respuesta = await fetch(`${API_URL}/depositos`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<DepositoDTO>(respuesta);
}

// Actualizar un depósito
export async function actualizarDeposito(
  id: string,
  dto: ActualizarDepositoDTO,
  token: string,
): Promise<DepositoDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${id}`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<DepositoDTO>(respuesta);
}

// Eliminar un depósito
export async function eliminarDeposito(
  id: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${id}`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}
