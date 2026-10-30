
import type {
  ProductoDTO,
  CrearProductoDTO,
  ActualizarProductoDTO,
  StockMinimoDTO,
  FiltroProductosDTO,
  ProductosPaginadosDTO,
} from "./types";

import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import { API_URL } from "./config";

// Listar productos con filtros y paginación
export async function listarProductos(
  token: string,
  filtro: FiltroProductosDTO = {},
): Promise<ProductosPaginadosDTO> {
  const parametros = new URLSearchParams();

  if (filtro.categoria_id) {
    parametros.set("categoria_id", filtro.categoria_id);
  }

  if (filtro.pagina !== undefined) {
    parametros.set("pagina", String(filtro.pagina));
  }

  if (filtro.todos !== undefined) {
    parametros.set("todos", filtro.todos);
  }

  const consulta = parametros.toString();

  const respuesta = await fetch(
    `${API_URL}/productos${consulta ? `?${consulta}` : ""}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ProductosPaginadosDTO>(respuesta);
}

// Buscar un producto por ID
export async function buscarProductoPorID(
  id: string,
  token: string,
): Promise<ProductoDTO> {
  const respuesta = await fetch(
    `${API_URL}/productos/${id}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ProductoDTO>(respuesta);
}

// Crear un producto
export async function crearProducto(
  dto: CrearProductoDTO,
  token: string,
): Promise<ProductoDTO> {
  const respuesta = await fetch(`${API_URL}/productos`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<ProductoDTO>(respuesta);
}

// Actualizar un producto
export async function actualizarProducto(
  id: string,
  dto: ActualizarProductoDTO,
  token: string,
): Promise<ProductoDTO> {
  const respuesta = await fetch(
    `${API_URL}/productos/${id}`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<ProductoDTO>(respuesta);
}

// Configurar el stock mínimo de un producto
export async function configurarStockMinimo(
  id: string,
  dto: StockMinimoDTO,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/productos/${id}/stock-minimo`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Eliminar un producto
export async function eliminarProducto(
  id: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/productos/${id}`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}
