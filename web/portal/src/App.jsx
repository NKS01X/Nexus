import React, { useState, useRef, createContext, useContext, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Navbar from './components/Navbar'
import HeroSection from './components/HeroSection'
import ArchitectureDiagram from './components/ArchitectureDiagram'
import FeatureCards from './components/FeatureCards'
import OnboardingForm from './components/OnboardingForm'
import MerchantsPage from './components/MerchantsPage'
import ApprovalsPage from './components/ApprovalsPage'
import RedTeamPage from './components/RedTeamPage'
import Footer from './components/Footer'

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
  const [token, setToken] = useState(() => sessionStorage.getItem('nexus_admin_token'))

  const login = async (adminKey) => {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ admin_key: adminKey }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || 'Authentication failed')
    sessionStorage.setItem('nexus_admin_token', data.token)
    setToken(data.token)
    return data.token
  }

  const logout = () => {
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

// Landing Page
function LandingPage() {
  const scrollToForm = () => {
    const el = document.getElementById('onboard')
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }

  return (
    <div>
      <Navbar />
      <HeroSection onScrollToForm={scrollToForm} />
      <ArchitectureDiagram />
      <FeatureCards />
      <OnboardingForm />
      <Footer />
    </div>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/merchants" element={<MerchantsPage />} />
            <Route path="/approvals" element={<ApprovalsPage />} />
            <Route path="/redteam" element={<RedTeamPage />} />
          </Routes>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App