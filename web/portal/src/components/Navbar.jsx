import React, { useEffect, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'

export default function Navbar() {
  const location = useLocation()
  const [scrolled, setScrolled] = useState(false)
  const navRef = useRef()

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <nav ref={navRef} className={`navbar ${scrolled ? 'scrolled' : ''}`}>
      <div className="navbar-inner">
        <Link to="/" className="logo">
          <div className="logo-icon">N</div>
          <span className="logo-text">Nexus</span>
        </Link>
        <div className="nav-links">
          <Link to="/" className={`nav-link ${location.pathname === '/' ? 'active' : ''}`}>
            Home
          </Link>
          <Link to="/merchants" className={`nav-link ${location.pathname === '/merchants' ? 'active' : ''}`}>
            Merchants
          </Link>
          <Link to="/approvals" className={`nav-link ${location.pathname === '/approvals' ? 'active' : ''}`}>
            Approvals
          </Link>
          <Link to="/redteam" className={`nav-link ${location.pathname === '/redteam' ? 'active' : ''}`} style={{ color: location.pathname === '/redteam' ? '#ef4444' : '' }}>
            Red Team
          </Link>
        </div>
      </div>
    </nav>
  )
}
