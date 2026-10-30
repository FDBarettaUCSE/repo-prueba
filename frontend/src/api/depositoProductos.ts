
import type {
  StockDepositoDTO,
  ConfigurarStockMinimoDTO,
  ListarStockFiltro,
  StockDepositoPaginadoDTO,
} from "./types";

import { API_URL } from "./config";
import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";

// Listar el stock de un depósito con filtros y paginación
export async function listarStockDeposito(
  depositoId: string,
  token: string,
  filtro: ListarStockFiltro = {},
): Promise<StockDepositoPaginadoDTO> {
  const parametros = new URLSearchParams();

  if (filtro.categoriaId) {
    parametros.set("categoriaId", filtro.categoriaId);
  }

  if (filtro.bajoMinimo !== undefined) {
    parametros.set("bajoMinimo", String(filtro.bajoMinimo));
  }

  if (filtro.pagina !== undefined) {
    parametros.set("pagina", String(filtro.pagina));
  }

  if (filtro.tamanioPagina !== undefined) {
    parametros.set("tamanioPagina", String(filtro.tamanioPagina));
  }

  const consulta = parametros.toString();
  const url =
    `${API_URL}/depositos/${depositoId}/stock` +
    (consulta ? `?${consulta}` : "");

  const respuesta = await fetch(url, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<StockDepositoPaginadoDTO>(respuesta);
}

// Obtener el stock de un producto específico dentro de un depósito
export async function buscarStockProducto(
  depositoId: string,
  productoId: string,
  token: string,
): Promise<StockDepositoDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/stock/${productoId}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<StockDepositoDTO>(respuesta);
}

// Configurar el stock mínimo particular de un producto en un depósito
export async function configurarStockMinimoDeposito(
  depositoId: string,
  productoId: string,
  dto: ConfigurarStockMinimoDTO,
  token: string,
): Promise<StockDepositoDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/stock/${productoId}/stock-minimo`,
    {
      method: "PUT",
      headers: {
        ...headersAutenticados(token),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<StockDepositoDTO>(respuesta);
}

// Quitar el mínimo particular y volver a utilizar el mínimo general del producto
export async function quitarStockMinimoDeposito(
  depositoId: string,
  productoId: string,
  token: string,
): Promise<StockDepositoDTO> {
  const respuesta = await fetch(
    `${API_URL}/depositos/${depositoId}/stock/${productoId}/stock-minimo`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<StockDepositoDTO>(respuesta);
}
