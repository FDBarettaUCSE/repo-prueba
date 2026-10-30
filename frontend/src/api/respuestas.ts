//Se incorpora manejarRespuesta, común a todos los archivos usados para fetchear el backend

export class ErrorAPI extends Error {}

// Manejo común de respuestas HTTP
export async function manejarRespuesta<T>(
  respuesta: Response,
): Promise<T> {
  if (!respuesta.ok) {
    const cuerpo = await respuesta.json().catch(() => null);

    throw new ErrorAPI(
      cuerpo?.error ?? `Error ${respuesta.status}`,
    );
  }

  // Los endpoints que responden 204 no devuelven JSON.
  if (respuesta.status === 204) {
    return undefined as T;
  }

  return respuesta.json() as Promise<T>;
}