const TOKEN = import.meta.env.VITE_API_TOKEN as string | undefined

export function apiFetch(input: string, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers)
  if (TOKEN) headers.set('X-Api-Token', TOKEN)
  return fetch(input, { ...init, headers })
}
