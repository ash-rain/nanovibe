interface FetchResult<T> {
  data: T | null
  error: string | null
}

async function request<T>(
  url: string,
  options: RequestInit = {}
): Promise<FetchResult<T>> {
  try {
    const res = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    })

    if (!res.ok) {
      let message = `HTTP ${res.status}`
      try {
        const body = await res.json()
        message = body.error || body.message || message
      } catch {
        // ignore parse error
      }
      return { data: null, error: message }
    }

    if (res.status === 204) {
      return { data: null, error: null }
    }

    const contentType = res.headers.get('content-type') || ''
    if (contentType.includes('application/json')) {
      const data = await res.json()
      return { data, error: null }
    }

    return { data: null, error: null }
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Network error'
    return { data: null, error: message }
  }
}

export function useFetch() {
  function get<T>(url: string): Promise<FetchResult<T>> {
    return request<T>(url, { method: 'GET' })
  }

  function post<T>(url: string, body?: unknown): Promise<FetchResult<T>> {
    return request<T>(url, {
      method: 'POST',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  }

  function put<T>(url: string, body?: unknown): Promise<FetchResult<T>> {
    return request<T>(url, {
      method: 'PUT',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  }

  function del<T>(url: string): Promise<FetchResult<T>> {
    return request<T>(url, { method: 'DELETE' })
  }

  return { get, post, put, del }
}
