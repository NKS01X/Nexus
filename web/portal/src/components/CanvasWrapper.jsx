import React, { useRef, useState, useEffect } from 'react'
import { Canvas } from '@react-three/fiber'

// Wrapper that only mounts the 3D scene when in viewport
export function LazyCanvas({
  children,
  threshold = 0.1,
  rootMargin = '100px',
  fallback = null,
  canvasProps = {}
}) {
  const ref = useRef(null)
  const [isVisible, setIsVisible] = useState(false)
  const [hasMounted, setHasMounted] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true)
          setHasMounted(true)
          observer.unobserve(ref.current)
        }
      },
      { threshold, rootMargin }
    )

    if (ref.current) {
      observer.observe(ref.current)
    }

    return () => observer.disconnect()
  }, [threshold, rootMargin])

  // Also mount if prefers-reduced-motion is set (for accessibility)
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (mediaQuery.matches && !hasMounted) {
      setIsVisible(true)
      setHasMounted(true)
    }
  }, [hasMounted])

  return (
    <div ref={ref} style={{ width: '100%', height: '100%' }}>
      {isVisible ? (
        <Canvas
          dpr={[0.75, Math.min(window.devicePixelRatio, 1.25)]}
          gl={{
            antialias: false,
            alpha: true,
            powerPreference: 'high-performance',
            preserveDrawingBuffer: false,
          }}
          {...canvasProps}
        >
          {children}
        </Canvas>
      ) : (
        fallback || (
          <div style={{
            width: '100%',
            height: '100%',
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: '20px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--text-tertiary)'
          }}>
            <div style={{ textAlign: 'center', padding: '40px' }}>
              <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ margin: '0 auto 16px', opacity: 0.3 }}>
                <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2"></polygon>
                <line x1="12" y1="22" x2="12" y2="15.5"></line>
                <polyline points="22 8.5 12 15.5 2 8.5"></polyline>
                <polyline points="2 15.5 12 8.5 22 15.5"></polyline>
                <line x1="12" y1="2" x2="12" y2="8.5"></line>
              </svg>
              <p style={{ fontSize: '14px' }}>3D visualization loads when visible</p>
            </div>
          </div>
        )
      )}
    </div>
  )
}

// Hook for checking if we should use 3D (performance/accessibility)
export function useShouldRender3D() {
  const [shouldRender, setShouldRender] = useState(true)

  useEffect(() => {
    // Check for reduced motion preference
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (mediaQuery.matches) {
      setShouldRender(false)
    }

    // Check for low-end devices
    const isLowEnd = navigator.hardwareConcurrency <= 4 ||
                     navigator.deviceMemory <= 4
    if (isLowEnd) {
      setShouldRender(false)
    }

    // Listen for changes
    const handler = (e) => setShouldRender(!e.matches)
    mediaQuery.addEventListener('change', handler)
    return () => mediaQuery.removeEventListener('change', handler)
  }, [])

  return shouldRender
}