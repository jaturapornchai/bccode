export function GET() {
  const icon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
  <rect width="64" height="64" rx="12" fill="#3049a5"/>
  <path d="M12 12h40v40H12z" fill="#00897b" opacity=".28"/>
  <text x="32" y="40" text-anchor="middle" font-family="Arial,sans-serif" font-size="22" font-weight="800" fill="white">AI</text>
</svg>`;

  return new Response(icon, {
    headers: {
      "Content-Type": "image/svg+xml",
      "Cache-Control": "public, max-age=31536000, immutable",
    },
  });
}
