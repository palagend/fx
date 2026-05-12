import { ref, onMounted, onUnmounted, provide, inject, type InjectionKey, type Ref } from 'vue'

export interface MobileContext {
  isMobile: Ref<boolean>
  showMobileOverlay: Ref<boolean>
  closeMobileOverlay: () => void
}

export const MobileKey: InjectionKey<MobileContext> = Symbol('mobile')

export function useMobile() {
  const isMobile = ref(false)
  const showMobileOverlay = ref(false)

  const checkMobile = () => {
    const width = window.innerWidth
    const wasMobile = isMobile.value
    isMobile.value = width < 768

    if (isMobile.value && !wasMobile) {
      closeMobileOverlay()
    }
  }

  const closeMobileOverlay = () => {
    showMobileOverlay.value = false
  }

  onMounted(() => {
    checkMobile()
    window.addEventListener('resize', checkMobile)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', checkMobile)
  })

  const provideMobile = () => {
    provide(MobileKey, { isMobile, showMobileOverlay, closeMobileOverlay })
  }

  return {
    isMobile,
    showMobileOverlay,
    checkMobile,
    closeMobileOverlay,
    provideMobile
  }
}

export function injectMobile(): MobileContext {
  const mobile = inject(MobileKey)
  if (!mobile) {
    throw new Error('useMobile must be used within a component that provides mobile context')
  }
  return mobile
}
