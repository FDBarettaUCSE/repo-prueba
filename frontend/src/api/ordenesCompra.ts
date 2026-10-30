import { API_URL } from "./config";
import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";

import type {
  ActualizarOrdenCompraDTO,
  ActualizarOrdenCompraItemDTO,
  CrearOrdenCompraDTO,
  CrearOrdenCompraItemDTO,
  FiltroOrdenesCompraDTO,
  OrdenCompraDTO,
  OrdenCompraItemDTO,
  OrdenesCompraPaginadasDTO,
  RegistrarRecepcionDTO,
  SugerenciasOrdenCompraDTO,
} from "./types";

// Listar órdenes de compra con filtros y paginación
export async function listarOrdenesCompra(
  token: string,
  filtros: FiltroOrdenesCompraDTO = {},
): Promise<OrdenesCompraPaginadasDTO> {
  const parametros = new URLSearchParams();

  if (filtros.deposito_id) {
    parametros.set("deposito_id", filtros.deposito_id);
  }

  if (filtros.estado) {
    parametros.set("estado", filtros.estado);
  }

  if (filtros.proveedor_id) {
    parametros.set("proveedor_id", filtros.proveedor_id);
  }

  if (filtros.pagina !== undefined) {
    parametros.set("pagina", String(filtros.pagina));
  }

  if (filtros.tamanioPagina !== undefined) {
    parametros.set("tamanioPagina", String(filtros.tamanioPagina));
  }

  const query = parametros.toString();
  const url = `${API_URL}/ordenes-compra${query ? `?${query}` : ""}`;

  const respuesta = await fetch(url, {
    method: "GET",
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<OrdenesCompraPaginadasDTO>(respuesta);
}

// Obtener una orden de compra por ID
export async function obtenerOrdenCompra(
  token: string,
  id: string,
): Promise<OrdenCompraDTO> {
  const respuesta = await fetch(`${API_URL}/ordenes-compra/${id}`, {
    method: "GET",
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<OrdenCompraDTO>(respuesta);
}

// Crear una orden de compra
export async function crearOrdenCompra(
  token: string,
  dto: CrearOrdenCompraDTO,
): Promise<OrdenCompraDTO> {
  const respuesta = await fetch(`${API_URL}/ordenes-compra`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<OrdenCompraDTO>(respuesta);
}

// Actualizar una orden de compra
export async function actualizarOrdenCompra(
  token: string,
  id: string,
  dto: ActualizarOrdenCompraDTO,
): Promise<OrdenCompraDTO> {
  const respuesta = await fetch(`${API_URL}/ordenes-compra/${id}`, {
    method: "PUT",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<OrdenCompraDTO>(respuesta);
}

// Generar sugerencias de órdenes a partir del stock bajo
export async function sugerirOrdenesCompra(
  token: string,
  depositoId?: string,
): Promise<SugerenciasOrdenCompraDTO> {
  const parametros = new URLSearchParams();

  if (depositoId) {
    parametros.set("deposito_id", depositoId);
  }

  const query = parametros.toString();
  const url = `${API_URL}/ordenes-compra/sugerir${query ? `?${query}` : ""}`;

  const respuesta = await fetch(url, {
    method: "POST",
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<SugerenciasOrdenCompraDTO>(respuesta);
}

// Agregar un producto a una orden de compra
export async function agregarItemOrdenCompra(
  token: string,
  ordenCompraId: string,
  dto: CrearOrdenCompraItemDTO,
): Promise<OrdenCompraItemDTO> {
  const respuesta = await fetch(
    `${API_URL}/ordenes-compra/${ordenCompraId}/items`,
    {
      method: "POST",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<OrdenCompraItemDTO>(respuesta);
}

// Actualizar un producto de una orden de compra
export async function actualizarItemOrdenCompra(
  token: string,
  ordenCompraId: string,
  itemId: string,
  dto: ActualizarOrdenCompraItemDTO,
): Promise<OrdenCompraItemDTO> {
  const respuesta = await fetch(
    `${API_URL}/ordenes-compra/${ordenCompraId}/items/${itemId}`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<OrdenCompraItemDTO>(respuesta);
}

// Eliminar un producto de una orden de compra
export async function eliminarItemOrdenCompra(
  token: string,
  ordenCompraId: string,
  itemId: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/ordenes-compra/${ordenCompraId}/items/${itemId}`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<void>(respuesta);
}

// Confirmar una orden de compra
export async function confirmarOrdenCompra(
  token: string,
  id: string,
): Promise<OrdenCompraDTO> {
  const respuesta = await fetch(
    `${API_URL}/ordenes-compra/${id}/confirmar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<OrdenCompraDTO>(respuesta);
}

// Descartar una orden de compra
export async function descartarOrdenCompra(
  token: string,
  id: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/ordenes-compra/${id}/descartar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<void>(respuesta);
}

// Registrar la recepción de productos de una orden de compra
export async function registrarRecepcionOrdenCompra(
  token: string,
  id: string,
  dto: RegistrarRecepcionDTO,
): Promise<OrdenCompraDTO> {
  const respuesta = await fetch(
    `${API_URL}/ordenes-compra/${id}/recepciones`,
    {
      method: "POST",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<OrdenCompraDTO>(respuesta);
}