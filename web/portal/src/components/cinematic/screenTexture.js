import * as THREE from 'three'

export function createScreenTexture() {
  const canvas = document.createElement('canvas')
  canvas.width = 1024
  canvas.height = 640
  const texture = new THREE.CanvasTexture(canvas)
  texture.colorSpace = THREE.SRGBColorSpace
  texture.minFilter = THREE.LinearFilter
  texture.magFilter = THREE.LinearFilter
  return { canvas, texture }
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.arcTo(x + w, y, x + w, y + h, r)
  ctx.arcTo(x + w, y + h, x, y + h, r)
  ctx.arcTo(x, y + h, x, y, r)
  ctx.arcTo(x, y, x + w, y, r)
  ctx.closePath()
}

export function drawLaptopScreen(canvas, { mode, typed, time }) {
  const ctx = canvas.getContext('2d')
  const w = canvas.width
  const h = canvas.height

  ctx.fillStyle = '#0b0d12'
  ctx.fillRect(0, 0, w, h)

  const g = ctx.createLinearGradient(0, 0, w, h)
  g.addColorStop(0, '#12151f')
  g.addColorStop(1, '#0a0c14')
  ctx.fillStyle = g
  ctx.fillRect(0, 0, w, h)

  ctx.fillStyle = 'rgba(255,255,255,0.04)'
  roundRect(ctx, 0, 0, w, 52, 0)
  ctx.fill()

  ctx.fillStyle = '#e8eaef'
  ctx.font = '700 22px Inter, system-ui, sans-serif'
  ctx.fillText('Aegis', 28, 34)
  ctx.font = '500 13px Inter, system-ui, sans-serif'
  ctx.fillStyle = 'rgba(255,255,255,0.35)'
  ctx.fillText('Merchant AI Commerce', 108, 33)

  ctx.fillStyle = '#34d399'
  ctx.beginPath()
  ctx.arc(w - 36, 26, 5, 0, Math.PI * 2)
  ctx.fill()

  if (mode === 'intent') {
    ctx.fillStyle = 'rgba(255,255,255,0.05)'
    roundRect(ctx, 48, 92, w - 96, 420, 18)
    ctx.fill()
    ctx.strokeStyle = 'rgba(255,255,255,0.08)'
    ctx.lineWidth = 1.5
    roundRect(ctx, 48, 92, w - 96, 420, 18)
    ctx.stroke()

    ctx.fillStyle = 'rgba(11,102,239,0.15)'
    roundRect(ctx, 72, 120, 168, 28, 14)
    ctx.fill()
    ctx.fillStyle = '#93c5fd'
    ctx.font = '600 12px Inter, system-ui, sans-serif'
    ctx.fillText('AI BUYER SESSION', 88, 139)

    ctx.fillStyle = 'rgba(255,255,255,0.45)'
    ctx.font = '500 14px Inter, system-ui, sans-serif'
    ctx.fillText('Natural-language purchase', 72, 188)

    const shown = typed || ''
    ctx.fillStyle = '#f8fafc'
    ctx.font = '600 36px Inter, system-ui, sans-serif'
    ctx.fillText(`“${shown}`, 72, 250)
    if (shown.length < 22 && Math.floor(time * 2) % 2 === 0) {
      const metrics = ctx.measureText(`“${shown}`)
      ctx.fillStyle = '#60a5fa'
      ctx.fillRect(72 + metrics.width + 6, 220, 3, 36)
    }
    ctx.fillStyle = '#f8fafc'
    ctx.font = '600 36px Inter, system-ui, sans-serif'
    if (shown.length >= 22) ctx.fillText('”', 72 + ctx.measureText(`“${shown}`).width, 250)

    ctx.fillStyle = 'rgba(255,255,255,0.28)'
    ctx.font = '500 13px Inter, system-ui, sans-serif'
    ctx.fillText('SKU  SHOE-RUN-001-RED  ·  Qty  1  ·  Cap  ₹3,000', 72, 320)

    ctx.fillStyle = 'rgba(96,165,250,0.12)'
    roundRect(ctx, 72, 360, 220, 44, 10)
    ctx.fill()
    ctx.fillStyle = '#93c5fd'
    ctx.font = '600 13px Inter, system-ui, sans-serif'
    ctx.fillText('Routing through Merchant MCP', 88, 387)
  } else {
    ctx.fillStyle = 'rgba(239,68,68,0.12)'
    roundRect(ctx, 48, 88, w - 96, 128, 16)
    ctx.fill()
    ctx.strokeStyle = 'rgba(239,68,68,0.45)'
    ctx.lineWidth = 1.5
    roundRect(ctx, 48, 88, w - 96, 128, 16)
    ctx.stroke()

    ctx.fillStyle = '#fecaca'
    ctx.font = '700 13px Inter, system-ui, sans-serif'
    ctx.fillText('ADMIN  ·  POLICY GATEWAY', 72, 122)

    ctx.fillStyle = '#fff'
    ctx.font = '700 28px Inter, system-ui, sans-serif'
    ctx.fillText('Malicious AI Purchase Blocked', 72, 164)
    ctx.fillStyle = '#fca5a5'
    ctx.font = '600 16px Inter, system-ui, sans-serif'
    ctx.fillText('(10,000× Quantity)', 72, 196)

    const cards = [
      ['Requested', '10,000 shoes'],
      ['Policy cap', '2 per SKU'],
      ['Decision', 'BLOCKED'],
      ['Next', 'Approval queue'],
    ]
    cards.forEach((card, i) => {
      const x = 48 + (i % 2) * 460
      const y = 244 + Math.floor(i / 2) * 150
      ctx.fillStyle = 'rgba(255,255,255,0.04)'
      roundRect(ctx, x, y, 440, 128, 14)
      ctx.fill()
      ctx.strokeStyle = 'rgba(255,255,255,0.07)'
      ctx.stroke()
      ctx.fillStyle = 'rgba(255,255,255,0.38)'
      ctx.font = '600 12px Inter, system-ui, sans-serif'
      ctx.fillText(card[0].toUpperCase(), x + 24, y + 38)
      ctx.fillStyle = i === 2 ? '#f87171' : '#f1f5f9'
      ctx.font = '700 26px Inter, system-ui, sans-serif'
      ctx.fillText(card[1], x + 24, y + 80)
    })
  }
}
