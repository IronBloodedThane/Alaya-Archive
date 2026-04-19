import { useState, useEffect } from 'react'

const sizeMap = {
  xs: 'w-6 h-6 text-xs',
  sm: 'w-8 h-8 text-sm',
  md: 'w-10 h-10 text-base',
  lg: 'w-16 h-16 text-xl',
  xl: 'w-24 h-24 text-3xl',
}

function initials(displayName, username) {
  const source = (displayName || username || '?').trim()
  if (!source) return '?'
  const parts = source.split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase()
  return source.slice(0, 2).toUpperCase()
}

export default function Avatar({ username, displayName, size = 'md', version, className = '' }) {
  const [broken, setBroken] = useState(false)
  const sizeClass = sizeMap[size] || sizeMap.md

  useEffect(() => {
    setBroken(false)
  }, [username, version])

  const src = username
    ? `/api/v1/users/${encodeURIComponent(username)}/avatar${version ? `?v=${encodeURIComponent(version)}` : ''}`
    : null

  if (!src || broken) {
    return (
      <div
        className={`${sizeClass} rounded-full bg-indigo-600 text-white font-semibold flex items-center justify-center shrink-0 ${className}`}
        aria-label={displayName || username || 'user'}
      >
        {initials(displayName, username)}
      </div>
    )
  }

  return (
    <img
      src={src}
      alt={displayName || username || 'avatar'}
      onError={() => setBroken(true)}
      className={`${sizeClass} rounded-full object-cover shrink-0 ${className}`}
    />
  )
}
