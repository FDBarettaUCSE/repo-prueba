import type {
  RegistrarDTO,
  CambiarContraseniaDTO,
  CambiarRolDTO,
  IngresarDTO,
  UsuarioDTO,
  UsuariosPaginadosDTO,
  RespuestaIngresar,
} from "./types";

import { headersAutenticados } from "./headers";
import { manejarRespuesta } from "./respuestas";
import { API_URL } from "./config";


// Registrar un usuario
export async function registrar(
  dto: RegistrarDTO,
  token: string,
): Promise<UsuarioDTO> {
  const respuesta = await fetch(`${API_URL}/usuarios/registrar`, {
    method: "POST",
    headers: headersAutenticados(token),
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<UsuarioDTO>(respuesta);
}

// Iniciar sesión
export async function ingresar(
  dto: IngresarDTO,
): Promise<RespuestaIngresar> {
  const respuesta = await fetch(`${API_URL}/usuarios/ingresar`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(dto),
  });

  return manejarRespuesta<RespuestaIngresar>(respuesta);
}

// Obtener el perfil del usuario autenticado
export async function obtenerPerfil(
  token: string,
): Promise<UsuarioDTO> {
  const respuesta = await fetch(`${API_URL}/usuarios/perfil`, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<UsuarioDTO>(respuesta);
}

// Listar usuarios de forma paginada
export async function listarUsuariosPaginados(
  token: string,
  pagina = 1,
  tamanioPagina = 100,
): Promise<UsuariosPaginadosDTO> {
  const params = new URLSearchParams({
    pagina: String(pagina),
    tamanioPagina: String(tamanioPagina),
  });
  const respuesta = await fetch(`${API_URL}/usuarios?${params.toString()}`, {
    headers: headersAutenticados(token),
  });

  return manejarRespuesta<UsuariosPaginadosDTO>(respuesta);
}

// Listar todos los usuarios
export async function listarUsuarios(
  token: string,
): Promise<UsuarioDTO[]> {
  const paginado = await listarUsuariosPaginados(token, 1, 100);
  return paginado.items;
}

// Cambiar la contraseña de un usuario
export async function cambiarContrasenia(
  id: string,
  dto: CambiarContraseniaDTO,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/usuarios/${id}/cambiarcontrasenia`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Cambiar el rol de un usuario
export async function cambiarRolUsuario(
  id: string,
  dto: CambiarRolDTO,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/usuarios/${id}/cambiarrol`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
      body: JSON.stringify(dto),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Desactivar un usuario
export async function desactivarUsuario(
  id: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/usuarios/${id}/desactivar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}

// Activar un usuario
export async function activarUsuario(
  id: string,
  token: string,
): Promise<void> {
  const respuesta = await fetch(
    `${API_URL}/usuarios/${id}/activar`,
    {
      method: "PUT",
      headers: headersAutenticados(token),
    },
  );

  await manejarRespuesta<void>(respuesta);
}
