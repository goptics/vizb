const STATIC_ASSET =
  /\.(?:html?|css|m?js|cjs|map|json|wasm|xml|txt|ico|png|jpe?g|gif|svg|webp|avif|woff2?|ttf|otf)$/i

export const extractPathDatasetId = (pathname: string): string | null => {
  const segment = pathname.split('/').filter(Boolean).at(-1)
  if (!segment) return null

  try {
    const decoded = decodeURIComponent(segment)
    if (!decoded || decoded.toLowerCase() === 'index' || STATIC_ASSET.test(decoded)) return null
    return decoded
  } catch {
    return null
  }
}
