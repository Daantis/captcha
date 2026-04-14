import { addTransportPrefix, ClientOp, decodeServerFrame, encodeClientFrame, ServerOp, toUint8Array, type ViewModel } from './protocol'
import { renderApp } from './render'

declare global {
    interface Window {
        __CAPTCHA_BOOTSTRAP__?: {
            challengeId: string
            mode: string
            title: string
        }
    }
}

export function boot(mode: string) {
    const root = document.getElementById('app')
    if (!root) {
        return
    }

    const bootstrap = window.__CAPTCHA_BOOTSTRAP__
    let seq = 1
    let phase = 0
    let currentView: ViewModel | null = null
    let currentMessage = 'Подготавливаем задание...'

    const send = (opcode: ClientOp, payload: { subject?: number; target?: number; value?: number; extra?: number } = {}) => {
        const frame = encodeClientFrame({
            opcode,
            seq: seq++,
            phase,
            ...payload,
        })

        window.top.postMessage({
            type: 'captcha:sendData',
            data: addTransportPrefix(frame),
        })
    }

    const rerender = () => {
        renderApp(root, currentView, currentMessage, {
            tap(subject) {
                send(ClientOp.Tap, { subject })
            },
            swipe(subject, direction) {
                send(ClientOp.Swipe, { subject, value: direction })
            },
            dragDrop(subject, target) {
                send(ClientOp.DragDrop, { subject, target })
            },
        })
    }

    window.addEventListener('message', (event) => {
        if (event.data?.type !== 'captcha:serverData') {
            return
        }

        const bytes = toUint8Array(event.data.data)
        if (!bytes) {
            return
        }

        const frame = decodeServerFrame(bytes)
        phase = frame.phase
        currentView = frame.payload.view
        currentMessage = frame.payload.message || currentView.status || ''
        rerender()

        send(ClientOp.AckFrame, { value: frame.seq })

        if (frame.opcode === ServerOp.Result) {
            currentMessage = currentView?.status || 'Ответ отправлен.'
            rerender()
        }
    })

    currentMessage = 'Запрашиваем задание у сервера...'
    rerender()

    if (bootstrap?.mode === mode) {
        send(ClientOp.Ready)
    } else {
        currentMessage = 'Режим задания не совпал с шаблоном.'
        rerender()
    }
}
