import type { BucketView, CardView, LayoutModel, OptionView, SlotView, ViewModel } from './protocol'

export interface RenderHandlers {
    tap: (subject: number) => void
    swipe: (subject: number, direction: number) => void
    dragDrop: (subject: number, target: number) => void
}

export function renderApp(root: HTMLElement, view: ViewModel | null, message: string, handlers: RenderHandlers) {
    if (!view) {
        root.innerHTML = `
            <main class="shell loading">
                <section class="panel">
                    <div class="eyebrow">Подключение</div>
                    <h1>Загружаем задание</h1>
                    <p>Сервер подготавливает CAPTCHA.</p>
                </section>
            </main>
        `
        return
    }

    document.body.dataset.mode = view.mode
    const note = message || view.status || ''
    root.innerHTML = `
        <main class="shell">
            <section class="panel">
                <header class="hero">
                    <div class="eyebrow">${escapeHtml(view.title)}</div>
                    <div class="progress-row">
                        <h1>${escapeHtml(view.instruction)}</h1>
                        <div class="progress-pill">${escapeHtml(view.progressText)}</div>
                    </div>
                    ${renderBadges(view.badges ?? [])}
                    ${note ? `<p class="status">${escapeHtml(note)}</p>` : ''}
                </header>
                ${renderLayout(view.layout)}
            </section>
        </main>
    `

    bindTap(root, handlers)
    bindSwipe(root, handlers)
    bindDrag(root, handlers)
}

function renderBadges(badges: string[]): string {
    if (!badges.length) {
        return ''
    }
    return `<div class="badges">${badges.map((badge) => `<span>${escapeHtml(badge)}</span>`).join('')}</div>`
}

function renderLayout(layout: LayoutModel): string {
    switch (layout.type) {
        case 'grid':
        case 'letters':
            return renderGrid(layout.options ?? [], layout.columns ?? 2, layout.type === 'letters')
        case 'reality':
            return renderReality(layout.card, layout.hintLeft ?? 'Невозможно', layout.hintRight ?? 'Возможно')
        case 'baskets':
            return renderBaskets(layout.options?.[0], layout.baskets ?? [])
        case 'track':
            return renderTrack(layout)
        case 'complete':
            return renderComplete(layout.card)
        default:
            return '<section class="board"><p>Неизвестный тип задания.</p></section>'
    }
}

function renderGrid(options: OptionView[], columns: number, glyphMode: boolean): string {
    return `
        <section class="board grid-board" style="--columns:${Math.max(2, columns)}">
            ${options.map((option) => `
                <button
                    class="option-card${option.muted ? ' muted' : ''}${glyphMode ? ' glyph' : ''}"
                    data-tap="${option.id}"
                    ${option.muted ? 'disabled' : ''}
                >
                    ${option.icon ? `<span class="option-icon">${escapeHtml(option.icon)}</span>` : ''}
                    <span class="option-label">${escapeHtml(option.label)}</span>
                </button>
            `).join('')}
        </section>
    `
}

function renderReality(card: CardView | undefined, leftLabel: string, rightLabel: string): string {
    if (!card) {
        return ''
    }

    return `
        <section class="board reality-board">
            <article class="swipe-card" data-swipe-card="${card.id}">
                <div class="swipe-icon">${escapeHtml(card.icon ?? '◉')}</div>
                <h2>${escapeHtml(card.title)}</h2>
                <p>${formatMultiline(card.body)}</p>
            </article>
            <div class="swipe-actions">
                <button class="action ghost" data-swipe="${card.id}" data-direction="-1">${escapeHtml(leftLabel)}</button>
                <button class="action" data-swipe="${card.id}" data-direction="1">${escapeHtml(rightLabel)}</button>
            </div>
        </section>
    `
}

function renderBaskets(card: OptionView | undefined, baskets: BucketView[]): string {
    if (!card) {
        return ''
    }

    return `
        <section class="board baskets-board">
            <div
                class="drag-card"
                draggable="true"
                data-draggable="${card.id}"
            >
                ${card.icon ? `<span class="option-icon">${escapeHtml(card.icon)}</span>` : ''}
                <span class="option-label">${escapeHtml(card.label)}</span>
            </div>
            <div class="basket-grid">
                ${baskets.map((basket) => `
                    <div class="basket-target" data-basket="${basket.id}">
                        <strong>${escapeHtml(basket.label)}</strong>
                        <span>${escapeHtml(basket.hint ?? '')}</span>
                        <button class="action slim" data-drop-choice="${basket.id}" data-card="${card.id}">Отправить сюда</button>
                    </div>
                `).join('')}
            </div>
        </section>
    `
}

function renderTrack(layout: LayoutModel): string {
    const slots = layout.slots ?? []
    const sequence = layout.sequence ?? []
    const flash = sequence[0]
    return `
        <section class="board track-board">
            <div class="track-target">
                <span class="eyebrow">Цель</span>
                <strong>${escapeHtml(layout.target ?? 'Следите за фишкой')}</strong>
            </div>
            <div class="track-grid">
                ${slots.map((slot, index) => renderSlot(slot, flash, index)).join('')}
            </div>
        </section>
    `
}

function renderSlot(slot: SlotView, flash: { from: number; to: number } | undefined, index: number): string {
    const isFlash = flash && (flash.from === index + 1 || flash.to === index + 1)
    return `
        <button
            class="track-slot${slot.active ? ' interactive' : ''}${isFlash ? ' flash' : ''}"
            data-slot="${slot.id}"
            ${slot.active ? '' : 'disabled'}
        >
            <span>${escapeHtml(slot.label)}</span>
            <strong>${escapeHtml(slot.token ?? '•')}</strong>
        </button>
    `
}

function renderComplete(card: CardView | undefined): string {
    if (!card) {
        return ''
    }

    return `
        <section class="board complete-board">
            <article class="done-card">
                <h2>${escapeHtml(card.title)}</h2>
                <p>${escapeHtml(card.body)}</p>
            </article>
        </section>
    `
}

function bindTap(root: HTMLElement, handlers: RenderHandlers) {
    root.querySelectorAll<HTMLElement>('[data-tap]').forEach((node) => {
        node.addEventListener('click', () => {
            const subject = Number(node.dataset.tap)
            handlers.tap(subject)
        })
    })

    root.querySelectorAll<HTMLElement>('[data-slot]').forEach((node) => {
        node.addEventListener('click', () => {
            const subject = Number(node.dataset.slot)
            handlers.tap(subject)
        })
    })

    root.querySelectorAll<HTMLElement>('[data-drop-choice]').forEach((node) => {
        node.addEventListener('click', () => {
            handlers.dragDrop(Number(node.dataset.card), Number(node.dataset.dropChoice))
        })
    })
}

function bindSwipe(root: HTMLElement, handlers: RenderHandlers) {
    root.querySelectorAll<HTMLElement>('[data-swipe]').forEach((node) => {
        node.addEventListener('click', () => {
            handlers.swipe(Number(node.dataset.swipe), Number(node.dataset.direction))
        })
    })

    root.querySelectorAll<HTMLElement>('[data-swipe-card]').forEach((node) => {
        let startX = 0
        node.addEventListener('pointerdown', (event) => {
            startX = event.clientX
        })
        node.addEventListener('pointerup', (event) => {
            const delta = event.clientX - startX
            if (Math.abs(delta) < 30) {
                return
            }
            handlers.swipe(Number(node.dataset.swipeCard), delta > 0 ? 1 : -1)
        })
    })
}

function bindDrag(root: HTMLElement, handlers: RenderHandlers) {
    let currentCard = 0

    root.querySelectorAll<HTMLElement>('[data-draggable]').forEach((node) => {
        node.addEventListener('dragstart', (event) => {
            currentCard = Number(node.dataset.draggable)
            event.dataTransfer?.setData('text/plain', String(currentCard))
        })
    })

    root.querySelectorAll<HTMLElement>('[data-basket]').forEach((node) => {
        node.addEventListener('dragover', (event) => {
            event.preventDefault()
        })
        node.addEventListener('drop', (event) => {
            event.preventDefault()
            const data = event.dataTransfer?.getData('text/plain')
            const subject = Number(data || currentCard)
            const target = Number(node.dataset.basket)
            handlers.dragDrop(subject, target)
        })
    })
}

function escapeHtml(value: string): string {
    return value
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
}

function formatMultiline(value: string): string {
    return escapeHtml(value).replaceAll('\n', '<br />')
}
