import type { WSMessage } from '$lib/api/types';

type MessageHandler = (msg: WSMessage) => void;
type StatusHandler = (status: 'connecting' | 'connected' | 'disconnected' | 'error') => void;

export class GameSocket {
    private ws: WebSocket | null = null;
    private handlers = new Map<string, MessageHandler[]>();
    private statusHandler?: StatusHandler;
    private reconnectTimer?: ReturnType<typeof setTimeout>;
    private maxReconnects = 5;
    private reconnects = 0;
    private url: string;

    constructor(url: string) {
        this.url = url;
    }

    connect(): void {
        this.statusHandler?.('connecting');

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
            this.reconnects = 0;
            this.statusHandler?.('connected');
        };

        this.ws.onmessage = (event) => {
            try {
                const msg: WSMessage = JSON.parse(event.data);
                const msgHandlers = this.handlers.get(msg.type) ?? [];
                const wildcardHandlers = this.handlers.get('*') ?? [];
                [...msgHandlers, ...wildcardHandlers].forEach(h => h(msg));
            } catch {
                console.error('Failed to parse WebSocket message:', event.data);
            }
        };

        this.ws.onclose = (event) => {
            this.statusHandler?.('disconnected');
            if (!event.wasClean && this.reconnects < this.maxReconnects) {
                const delay = Math.min(1000 * 2 ** this.reconnects, 10000);
                this.reconnects++;
                this.reconnectTimer = setTimeout(() => this.connect(), delay);
            }
        };

        this.ws.onerror = () => {
            this.statusHandler?.('error');
        };
    }

    disconnect(): void {
        clearTimeout(this.reconnectTimer);
        this.maxReconnects = 0; // Prevent reconnect
        this.ws?.close(1000, 'Client closed');
        this.ws = null;
    }

    send(type: string, payload?: unknown): void {
        if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, payload }));
        }
    }

    on(type: string, handler: MessageHandler): () => void {
        const existing = this.handlers.get(type) ?? [];
        this.handlers.set(type, [...existing, handler]);
        return () => {
            const updated = (this.handlers.get(type) ?? []).filter(h => h !== handler);
            this.handlers.set(type, updated);
        };
    }

    onStatus(handler: StatusHandler): void {
        this.statusHandler = handler;
    }

    get readyState(): number {
        return this.ws?.readyState ?? WebSocket.CLOSED;
    }
}
