// Same-origin relay for Portal API requests.
// The browser calls /api/* on the pages.dev origin; this function forwards the
// request to API_BASE_URL (Portal server behind Cloudflare Tunnel) and streams
// the response back. Server-to-server forwarding means no CORS is required.

export async function onRequest(context) {
  const { request, env } = context;
  const incoming = new URL(request.url);

  if (!env.API_BASE_URL) {
    return Response.json({ error: 'API_BASE_URL is not configured' }, { status: 500 });
  }

  const upstreamURL = env.API_BASE_URL.replace(/\/$/, '') + incoming.pathname + incoming.search;

  const upstreamRequest = new Request(upstreamURL, request);
  upstreamRequest.headers.set('X-Forwarded-Host', incoming.hostname);
  upstreamRequest.headers.set('X-Forwarded-Proto', 'https');

  let response;
  try {
    response = await fetch(upstreamRequest);
  } catch {
    return Response.json(
      { error: 'portal api unavailable: upstream tunnel unreachable' },
      { status: 502 }
    );
  }

  const headers = new Headers(response.headers);
  headers.set('Access-Control-Allow-Origin', incoming.origin);
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}
