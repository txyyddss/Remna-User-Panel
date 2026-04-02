/// <reference types="vite/client" />

interface Window {
    Telegram?: {
        WebApp: {
            initData: string
            initDataUnsafe: {
                user?: {
                    id: number
                    first_name: string
                    last_name?: string
                    username?: string
                }
            }
            ready: () => void
            expand: () => void
            close: () => void
            openLink: (url: string) => void
            isExpanded: boolean
            viewportHeight: number
            viewportStableHeight: number
            themeParams: {
                bg_color?: string
                text_color?: string
                hint_color?: string
                link_color?: string
                button_color?: string
                button_text_color?: string
                secondary_bg_color?: string
            }
            setBackgroundColor: (color: string) => void
            setHeaderColor: (color: string) => void
            BackButton: {
                show: () => void
                hide: () => void
                onClick: (cb: () => void) => void
            }
            MainButton: {
                text: string
                show: () => void
                hide: () => void
                enable: () => void
                disable: () => void
                isActive: boolean
                isVisible: boolean
                onClick: (cb: () => void) => void
                setText: (text: string) => void
                showProgress: (isLeaveActive: boolean) => void
                hideProgress: () => void
            }
            HapticFeedback: {
                impactOccurred: (style: string) => void
                notificationOccurred: (type: string) => void
                selectionChanged: () => void
            }
        }
    }
}
