type ToastType = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
}

let toasts: ToastItem[] = []
let nextId = 0
const listeners = new Set<() => void>()

function emit() {
  listeners.forEach((l) => l())
}

export function subscribe(cb: () => void) {
  listeners.add(cb)
  return () => listeners.delete(cb)
}

export function getToasts() {
  return toasts
}

export function removeToast(id: number) {
  toasts = toasts.filter((t) => t.id !== id)
  emit()
}

export const toast = {
  success(message: string) {
    toasts = [...toasts, { id: ++nextId, type: 'success', message }]
    emit()
  },
  error(message: string) {
    toasts = [...toasts, { id: ++nextId, type: 'error', message }]
    emit()
  },
  info(message: string) {
    toasts = [...toasts, { id: ++nextId, type: 'info', message }]
    emit()
  },
}
