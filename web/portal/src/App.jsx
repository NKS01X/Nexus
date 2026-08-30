import React, { useState, createContext, useContext, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import LandingPage from './components/LandingPage'
import MerchantsPage from './components/MerchantsPage'
import ApprovalsPage from './components/ApprovalsPage'
import RedTeamPage from './components/RedTeamPage'
import AiPurchasePage from './components/AiPurchasePage'
import AegisDemo from './components/AegisDemo.jsx'
import { AnimatePresence, motion } from 'framer-motion'

// Toast Context
const ToastContext = createContext()

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])

  const showToast = (message, type = 'success') => {
    const id = Date.now()
    setToasts((prev) => [...prev, { id, message, type }])
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, 4000)
  }

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="toast-container">
        {toasts.map((toast) => (
          <div key={toast.id} className={`toast toast-${toast.type}`}>
            <span>{toast.message}</span>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  return useContext(ToastContext)
}

// Auth Context
const AuthContext = createContext()

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() =>
    sessionStorage.getItem('aegis_admin_token') || sessionStorage.getItem('nexus_admin_token')
  )

  const login = async (adminKey) => {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ admin_key: adminKey }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Authentication failed')
    sessionStorage.setItem('aegis_admin_token', data.token)
    setToken(data.token)
    return data.token
  }

  const logout = () => {
    sessionStorage.removeItem('aegis_admin_token')
    sessionStorage.removeItem('nexus_admin_token')
    setToken(null)
  }

  const authFetch = async (url, options = {}) => {
    const headers = { ...options.headers }
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    const res = await fetch(url, { ...options, headers })
    if (res.status === 401) {
      logout()
      throw new Error('Session expired')
    }
    return res
  }

  return (
    <AuthContext.Provider value={{ token, login, logout, authFetch, isAuthenticated: !!token }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}

// Page Transition Wrapper
function PageTransition({ children }) {
  const location = useLocation()

  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: 'auto' })
  }, [location.pathname])

  return (
    <AnimatePresence mode="wait">
      <motion.div key={location.pathname} initial="exit" animate="enter" exit="exit" variants={{
        enter: { opacity: 1, y: 0, transition: { duration: 0.5, ease: [0.16, 1, 0.3, 1] } },
        exit: { opacity: 0, y: 20, transition: { duration: 0.3, ease: [0.16, 1, 0.3, 1] } },
      }}>
        {children}
      </motion.div>
    </AnimatePresence>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/merchants" element={<PageTransition><MerchantsPage /></PageTransition>} />
            <Route path="/approvals" element={<PageTransition><ApprovalsPage /></PageTransition>} />
            <Route path="/redteam" element={<PageTransition><RedTeamPage /></PageTransition>} />
            <Route path="/ai-purchase" element={<PageTransition><AiPurchasePage /></PageTransition>} />
              <Route path="/aegis-demo" element={<PageTransition><AegisDemo /></PageTransition>} />
          </Routes>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App