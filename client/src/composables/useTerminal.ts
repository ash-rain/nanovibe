import { ref, onUnmounted } from 'vue'
import type { Ref } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'

interface UseTerminalOptions {
  projectId: Ref<string | undefined> | string
}

export function useTerminal(options?: UseTerminalOptions) {
  const terminalRef = ref<HTMLElement | null>(null)
  const connected = ref(false)
  const connecting = ref(false)

  let terminal: Terminal | null = null
  let fitAddon: FitAddon | null = null
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000
  let resizeObserver: ResizeObserver | null = null
  let destroyed = false
  let currentProjectId: string | null = null

  function createTerminal(container: HTMLElement) {
    terminal = new Terminal({
      theme: {
        background: '#0a0a14',
        foreground: '#e2e8f0',
        cursor: '#a78bfa',
        cursorAccent: '#0a0a14',
        selectionBackground: 'rgba(124, 58, 237, 0.3)',
        black: '#0a0a14',
        red: '#f87171',
        green: '#34d399',
        yellow: '#fbbf24',
        blue: '#60a5fa',
        magenta: '#a78bfa',
        cyan: '#22d3ee',
        white: '#e2e8f0',
        brightBlack: '#64748b',
        brightRed: '#fca5a5',
        brightGreen: '#6ee7b7',
        brightYellow: '#fcd34d',
        brightBlue: '#93c5fd',
        brightMagenta: '#c4b5fd',
        brightCyan: '#67e8f9',
        brightWhite: '#f1f5f9',
      },
      fontFamily: '"JetBrains Mono", monospace',
      fontSize: 13,
      lineHeight: 1.5,
      cursorBlink: true,
      cursorStyle: 'bar',
      scrollback: 10000,
      allowProposedApi: true,
    })

    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(new WebLinksAddon())
    terminal.open(container)
    fitAddon.fit()

    terminal.onData((data) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data }))
      }
    })

    resizeObserver = new ResizeObserver(() => {
      if (fitAddon) {
        fitAddon.fit()
        const dims = fitAddon.proposeDimensions()
        if (dims && ws && ws.readyState === WebSocket.OPEN) {
          ws.send(
            JSON.stringify({ type: 'resize', cols: dims.cols, rows: dims.rows })
          )
        }
      }
    })
    resizeObserver.observe(container)
  }

  function connect(projectId: string) {
    if (destroyed) return
    currentProjectId = projectId

    if (terminalRef.value && !terminal) {
      createTerminal(terminalRef.value)
    }

    if (ws) {
      ws.close()
    }

    connecting.value = true
    connected.value = false

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    ws = new WebSocket(`${protocol}//${host}/ws/terminal/${projectId}`)

    ws.onopen = () => {
      connected.value = true
      connecting.value = false
      reconnectDelay = 1000
      if (fitAddon && ws) {
        fitAddon.fit()
        const dims = fitAddon.proposeDimensions()
        if (dims) {
          ws.send(
            JSON.stringify({ type: 'resize', cols: dims.cols, rows: dims.rows })
          )
        }
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'output' && terminal) {
          terminal.write(msg.data)
        } else if (msg.type === 'exit' && terminal) {
          terminal.write(`\r\n[Process exited with code ${msg.code}]\r\n`)
          connected.value = false
        }
      } catch {
        // ignore
      }
    }

    ws.onclose = () => {
      connected.value = false
      connecting.value = false
      if (!destroyed && currentProjectId) {
        scheduleReconnect(currentProjectId)
      }
    }

    ws.onerror = () => {
      connected.value = false
      connecting.value = false
    }
  }

  function scheduleReconnect(projectId: string) {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    reconnectTimer = setTimeout(() => {
      if (!destroyed) {
        reconnectDelay = Math.min(reconnectDelay * 2, 30000)
        connect(projectId)
      }
    }, reconnectDelay)
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
    connected.value = false
    connecting.value = false
  }

  function resize() {
    if (fitAddon) {
      fitAddon.fit()
    }
  }

  function clear() {
    if (terminal) {
      terminal.clear()
    }
  }

  function focus() {
    if (terminal) {
      terminal.focus()
    }
  }

  onUnmounted(() => {
    destroyed = true
    disconnect()
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
    if (terminal) {
      terminal.dispose()
      terminal = null
    }
  })

  return {
    terminalRef,
    connected,
    connecting,
    connect,
    disconnect,
    resize,
    clear,
    focus,
  }
}
