/**
 * SEO: update document title, meta description, canonical, OG/Twitter, and robots per route.
 * Public pages: index, follow. Dashboard/admin: noindex, nofollow (not for search engines).
 */

const SITE_TITLE = 'UniRoute'
const DEFAULT_DESCRIPTION = 'One unified gateway for every AI model. Route, secure, and manage traffic to any LLM—cloud or local. Plus secure tunneling for any service.'

function getCanonicalUrl(path?: string): string {
  if (typeof window === 'undefined') return 'https://uniroute.co' + (path || '')
  const base = window.location.origin
  const basePath = import.meta.env.BASE_URL?.replace(/\/$/, '') || ''
  const pathPart = path ?? window.location.pathname
  return `${base}${basePath}${pathPart === '' || pathPart === '/' ? '' : pathPart}`
}

export function updateDocumentHead(
  title?: string,
  description?: string,
  path?: string,
  noIndex?: boolean
) {
  const fullTitle = title ? `${title} | ${SITE_TITLE}` : `${SITE_TITLE} - AI Gateway & Secure Tunneling`
  document.title = fullTitle

  const desc = description || DEFAULT_DESCRIPTION
  let metaDesc = document.querySelector('meta[name="description"]') as HTMLMetaElement | null
  if (!metaDesc) {
    metaDesc = document.createElement('meta')
    metaDesc.name = 'description'
    document.head.appendChild(metaDesc)
  }
  metaDesc.content = desc

  let robots = document.querySelector('meta[name="robots"]') as HTMLMetaElement | null
  if (!robots) {
    robots = document.createElement('meta')
    robots.name = 'robots'
    document.head.appendChild(robots)
  }
  robots.content = noIndex ? 'noindex, nofollow' : 'index, follow'

  const canonicalUrl = getCanonicalUrl(path)
  let canonical = document.querySelector('link[rel="canonical"]') as HTMLLinkElement | null
  if (!canonical) {
    canonical = document.createElement('link')
    canonical.rel = 'canonical'
    document.head.appendChild(canonical)
  }
  canonical.href = canonicalUrl

  const ogTitle = document.querySelector('meta[property="og:title"]') as HTMLMetaElement | null
  if (ogTitle) ogTitle.content = fullTitle
  const ogDesc = document.querySelector('meta[property="og:description"]') as HTMLMetaElement | null
  if (ogDesc) ogDesc.content = desc
  const ogUrl = document.querySelector('meta[property="og:url"]') as HTMLMetaElement | null
  if (ogUrl) ogUrl.content = canonicalUrl

  const twTitle = document.querySelector('meta[name="twitter:title"]') as HTMLMetaElement | null
  if (twTitle) twTitle.content = fullTitle
  const twDesc = document.querySelector('meta[name="twitter:description"]') as HTMLMetaElement | null
  if (twDesc) twDesc.content = desc
  const twUrl = document.querySelector('meta[name="twitter:url"]') as HTMLMetaElement | null
  if (twUrl) twUrl.content = canonicalUrl
}
