import * as THREE from 'three'

export const DURATION = 16.8

export const GATE = {
  z: -19.4,
  halfW: 1.9,
  halfH: 2.25,
  depth: 0.55,
}

const v = (x, y, z) => new THREE.Vector3(x, y, z)

export const CAMERA_KEYS = [
  { t: 0.0, pos: v(0.68, 1.22, 1.62), look: v(0.08, 0.86, 0.04), fov: 36, dof: 0.016 },
  { t: 1.6, pos: v(0.42, 1.08, 1.18), look: v(0.1, 0.9, -0.02), fov: 32, dof: 0.012 },
  { t: 3.05, pos: v(0.12, 0.96, 0.48), look: v(0.06, 0.92, -0.38), fov: 26, dof: 0.004 },
  { t: 4.15, pos: v(0.02, 0.55, -1.85), look: v(0, 0.15, -8), fov: 48, dof: 0.07 },
  { t: 5.4, pos: v(0.18, 0.18, -6.4), look: v(0, 0.05, -13), fov: 46, dof: 0.05 },
  { t: 7.0, pos: v(0.55, 0.42, -11.4), look: v(0, 0.25, -18.2), fov: 40, dof: 0.035 },
  { t: 8.35, pos: v(2.15, 0.72, -15.2), look: v(0, 0.45, -19.5), fov: 38, dof: 0.028 },
  { t: 10.4, pos: v(0.15, 0.28, -16.05), look: v(0, 0.55, -19.55), fov: 34, dof: 0.022 },
  { t: 12.15, pos: v(0.0, 1.55, -9.2), look: v(0, 0.2, -17), fov: 52, dof: 0.08 },
  { t: 14.2, pos: v(0.38, 1.1, 1.38), look: v(0.08, 0.88, 0.0), fov: 33, dof: 0.014 },
  { t: 16.8, pos: v(0.55, 1.16, 1.5), look: v(0.08, 0.86, 0.02), fov: 35, dof: 0.016 },
]

export function easeInOutCubic(t) {
  return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2
}

export function sampleCamera(t, outPos, outLook) {
  const keys = CAMERA_KEYS
  if (t <= keys[0].t) {
    outPos.copy(keys[0].pos)
    outLook.copy(keys[0].look)
    return { fov: keys[0].fov, dof: keys[0].dof }
  }
  const last = keys[keys.length - 1]
  if (t >= last.t) {
    outPos.copy(last.pos)
    outLook.copy(last.look)
    return { fov: last.fov, dof: last.dof }
  }
  for (let i = 0; i < keys.length - 1; i++) {
    const a = keys[i]
    const b = keys[i + 1]
    if (t >= a.t && t <= b.t) {
      const u = easeInOutCubic((t - a.t) / (b.t - a.t))
      outPos.lerpVectors(a.pos, b.pos, u)
      outLook.lerpVectors(a.look, b.look, u)
      return {
        fov: a.fov + (b.fov - a.fov) * u,
        dof: a.dof + (b.dof - a.dof) * u,
      }
    }
  }
  outPos.copy(last.pos)
  outLook.copy(last.look)
  return { fov: last.fov, dof: last.dof }
}

export function sceneName(t) {
  if (t < 3) return '01  Real-world intent'
  if (t < 5) return '02  Transition into the machine'
  if (t < 8) return '03  The threat'
  if (t < 12) return '04  Aegis enforcement'
  return '05  Resolution'
}

export function screenMode(t) {
  if (t >= 13.4) return 'blocked'
  return 'intent'
}
