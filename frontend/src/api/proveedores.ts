
import type {
  ProveedorDTO,
  CrearProveedorDTO,
  ActualizarProveedorDTO,
  AgregarProductoSuministradoDTO,
  ActualizarProductoSuministradoDTO,
  ProductoSuministradoDTO,
  ProveedorDeProductoDTO,
} from "./types";

import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import { API_URL } from "./config";

// Listar todos los proveedores
export async function listarProveedores(
  token: string,
): Promise<ProveedorDTO[]> {
  const respuesta = await fetch(`${API_URL}/proveedores`, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<ProveedorDTO[]>(respuesta);
}

// Buscar un proveedor por ID
export async function buscarProveedorPorID(
  id: string,
  token: string,
): Promise<ProveedorDTO> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${id}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ProveedorDTO>(respuesta);
}

// Crear un proveedor
export async function crearProveedor(
  dto: CrearProveedorDTO,
  token: string,
): Promise<ProveedorDTO> {
  const respuesta = await fetch(`${API_URL}/proveedores`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<ProveedorDTO>(respuesta);
}

// Actualizar un proveedor
export async function actualizarProveedor(
  id: string,
  dto: ActualizarProveedorDTO,
  token: string,
): Promise<ProveedorDTO> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${id}`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  return manejarRespuesta<ProveedorDTO>(respuesta);
}

// Eliminar un proveedor
export async function eliminarProveedor(
  id: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${id}`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Listar los productos que suministra un proveedor
export async function listarProductosSuministrados(
  proveedorId: string,
  token: string,
): Promise<ProductoSuministradoDTO[]> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${proveedorId}/productos`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ProductoSuministradoDTO[]>(respuesta);
}

// Listar los proveedores que suministran un producto
export async function listarProveedoresDeProducto(
  productoId: string,
  token: string,
): Promise<ProveedorDeProductoDTO[]> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/producto/${productoId}`,
    {
      headers: headersAutenticados(token),
    },
  );

  return manejarRespuesta<ProveedorDeProductoDTO[]>(respuesta);
}

// Agregar un producto al catálogo de un proveedor
export async function agregarProductoSuministrado(
  proveedorId: string,
  dto: AgregarProductoSuministradoDTO,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${proveedorId}/productos`,
    {
      method: "POST",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Actualizar las condiciones de suministro de un producto
export async function actualizarProductoSuministrado(
  proveedorId: string,
  productoId: string,
  dto: ActualizarProductoSuministradoDTO,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${proveedorId}/productos/${productoId}`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Eliminar un producto del catálogo de un proveedor
export async function eliminarProductoSuministrado(
  proveedorId: string,
  productoId: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/proveedores/${proveedorId}/productos/${productoId}`,
    {
      method: "DELETE",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}
