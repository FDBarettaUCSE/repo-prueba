
import { API_URL } from "./config";
import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import type {
  CategoriaDTO,
  CrearCategoriaDTO,
  ActualizarCategoriaDTO,
} from "./types";

// Listar todas las categorías
export async function listarCategorias(
  token: string,
): Promise<CategoriaDTO[]> {
  const respuesta = await fetch(`${API_URL}/categorias`, {
    method: "GET",
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<CategoriaDTO[]>(respuesta);
}

// Buscar una categoría por su ID
export async function buscarCategoriaPorID(
  id: string,
  token: string,
): Promise<CategoriaDTO> {
  const respuesta = await fetch(
    `${API_URL}/categorias/${encodeURIComponent(id)}`,
    {
      method: "GET",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<CategoriaDTO>(respuesta);
}

// Crear una categoría
export async function crearCategoria(
  dto: CrearCategoriaDTO,
  token: string,
): Promise<CategoriaDTO> {
  const respuesta = await fetch(`${API_URL}/categorias`, {
    method: "POST",
    headers: {
      ...headersAutenticados(token),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<CategoriaDTO>(respuesta);
}

// Actualizar una categoría
export async function actualizarCategoria(
  id: string,
  dto: ActualizarCategoriaDTO,
  token: string,
): Promise<CategoriaDTO> {
  const respuesta = await fetch(
    `${API_URL}/categorias/${encodeURIComponent(id)}`,
    {
      method: "PUT",
      headers: {
        ...headersAutenticados(token),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<CategoriaDTO>(respuesta);
}

// Eliminar una categoría
export async function eliminarCategoria(
  id: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/categorias/${encodeURIComponent(id)}`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}
