export const enum ClientOp {
    Ready = 1,
    Tap = 2,
    Swipe = 3,
    DragDrop = 4,
    AckFrame = 5,
}

export const enum ServerOp {
    Init = 101,
    Patch = 102,
    Prompt = 103,
    Progress = 104,
    Result = 105,
}

export interface ClientFrame {
    opcode: ClientOp
    seq: number
    phase: number
    subject?: number
    target?: number
    value?: number
    extra?: number
}

export interface OptionView {
    id: number
    label: string
    icon?: string
    accent?: string
    variant?: string
    muted?: boolean
}

export interface CardView {
    id: number
    title: string
    body: string
    icon?: string
    accent?: string
}

export interface BucketView {
    id: number
    label: string
    accent?: string
    hint?: string
}

export interface SlotView {
    id: number
    label: string
    token?: string
    accent?: string
    active?: boolean
}

export interface SwapView {
    from: number
    to: number
    delayMs?: number
}

export interface LayoutModel {
    type: string
    columns?: number
    options?: OptionView[]
    card?: CardView
    baskets?: BucketView[]
    slots?: SlotView[]
    sequence?: SwapView[]
    allowAnswer?: boolean
    hintLeft?: string
    hintRight?: string
    target?: string
    targetAccent?: string
}

export interface ViewModel {
    mode: string
    theme: string
    title: string
    instruction: string
    progressText: string
    status?: string
    badges?: string[]
    layout: LayoutModel
}

export interface ServerPayload {
    message?: string
    view: ViewModel
}

export interface ServerFrame {
    opcode: number
    seq: number
    phase: number
    entityId: number
    progress: number
    flags: number
    payload: ServerPayload
}

export function encodeClientFrame(frame: ClientFrame): Uint8Array {
    const buf = new Uint8Array(12)
    const view = new DataView(buf.buffer)
    buf[0] = frame.opcode
    view.setUint16(1, frame.seq, true)
    buf[3] = frame.phase & 0xff
    view.setUint16(4, frame.subject ?? 0, true)
    view.setUint16(6, frame.target ?? 0, true)
    view.setInt16(8, frame.value ?? 0, true)
    view.setInt16(10, frame.extra ?? 0, true)
    return buf
}

export function decodeServerFrame(data: Uint8Array): ServerFrame {
    const payload = stripPrefix(data)
    const view = new DataView(payload.buffer, payload.byteOffset, payload.byteLength)
    const size = view.getUint16(8, true)
    const raw = payload.slice(10, 10 + size)

    return {
        opcode: payload[0],
        seq: view.getUint16(1, true),
        phase: payload[3],
        entityId: view.getUint16(4, true),
        progress: payload[6],
        flags: payload[7],
        payload: JSON.parse(new TextDecoder().decode(raw)) as ServerPayload,
    }
}

export function addTransportPrefix(data: Uint8Array): Uint8Array {
    const out = new Uint8Array(data.length + 1)
    out[0] = 0x80
    out.set(data, 1)
    return out
}

export function toUint8Array(input: unknown): Uint8Array | null {
    if (input instanceof Uint8Array) {
        return input
    }
    if (input instanceof ArrayBuffer) {
        return new Uint8Array(input)
    }
    if (ArrayBuffer.isView(input)) {
        return new Uint8Array(input.buffer, input.byteOffset, input.byteLength)
    }
    return null
}

function stripPrefix(data: Uint8Array): Uint8Array {
    if (data.length > 1 && (data[0] & 0x80) !== 0) {
        return data.slice(1)
    }
    return data
}
