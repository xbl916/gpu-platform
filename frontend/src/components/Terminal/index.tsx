import React, { useEffect, useRef, useState } from 'react'
import { Button, Space, message } from 'antd'

interface TerminalProps {
  host: string
  port: number
  username?: string
  onDisconnect?: () => void
}

const Terminal: React.FC<TerminalProps> = ({ host, port, username = 'root', onDisconnect }) => {
  const terminalRef = useRef<HTMLDivElement>(null)
  const [connected, setConnected] = useState(false)
  const [output, setOutput] = useState<string[]>([])
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (host && port) {
      connect()
    }
    return () => {
      disconnect()
    }
  }, [host, port])

  const connect = () => {
    setConnected(true)
    setOutput([
      `Connecting to ${username}@${host}:${port}...`,
      'Connection established',
      '',
      `Last login: ${new Date().toLocaleString()}`,
      '',
      'Welcome to GPU Platform Container Terminal',
      'Type "help" for available commands.',
      '',
    ])
  }

  const disconnect = () => {
    setConnected(false)
    setOutput([])
    onDisconnect?.()
  }

  const handleCommand = (command: string) => {
    const trimmed = command.trim()
    if (!trimmed) return

    const newOutput = [...output, `root@gpu-container:~# ${command}`]

    if (trimmed === 'exit' || trimmed === 'logout') {
      newOutput.push('Disconnecting...')
      setOutput(newOutput)
      setTimeout(disconnect, 500)
      return
    }

    if (trimmed === 'clear') {
      setOutput([])
      return
    }

    if (trimmed === 'help') {
      newOutput.push(
        'Available commands:',
        '  help     - Show this help message',
        '  clear    - Clear terminal',
        '  exit     - Disconnect from terminal',
        '  nvidia-smi - Show GPU information',
        '  nvcc -V  - Show CUDA version',
        '  python --version - Show Python version',
        ''
      )
      setOutput(newOutput)
      return
    }

    if (trimmed === 'nvidia-smi') {
      newOutput.push(
        '+-----------------------------------------------------------------------------+',
        '| NVIDIA-SMI 535.154.05       Driver Version: 535.154.05       CUDA Version: 12.2  |',
        '|-------------------------------+----------------------+----------------------+',
        '| GPU  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC |',
        '| Fan  Temp  Perf  Pwr:Usage/Cap|         Memory-Usage | GPU-Util  Compute M. |',
        '|===============================+======================+======================|',
        '|   0  NVIDIA A100-SXM...  Off  | 00000000:00:04.0  On |                  0  |',
        '|  0%   28C    P0    65W / 250W |   1234MiB /  40960MiB |      0%      Default  |',
        '+-------------------------------+----------------------+----------------------+',
        ''
      )
      setOutput(newOutput)
      return
    }

    if (trimmed === 'nvcc -V') {
      newOutput.push(
        'nvcc: NVIDIA (R) Cuda compiler driver',
        'Copyright (c) 2005-2023 NVIDIA Corporation',
        'Built on Wed_Nov_22_2023_23:48:32_PDT',
        'Cuda compilation tools, release 12.3, V12.3.103',
        'Build cuda_12.3.r12.3/compiler.33419258_0',
        ''
      )
      setOutput(newOutput)
      return
    }

    if (trimmed === 'python --version') {
      newOutput.push('Python 3.10.13', '')
      setOutput(newOutput)
      return
    }

    newOutput.push(`bash: ${trimmed}: command not found`, '')
    setOutput(newOutput)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleCommand(e.currentTarget.value)
      e.currentTarget.value = ''
    }
  }

  return (
    <div style={{ fontFamily: 'monospace', fontSize: '14px' }}>
      <div style={{ marginBottom: 8, padding: '8px 12px', background: '#f5f5f5', borderRadius: '4px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <span style={{ color: connected ? '#52c41a' : '#ff4d4f' }}>
            ● {connected ? 'Connected' : 'Disconnected'}
          </span>
          <span>{username}@{host}:{port}</span>
        </Space>
        <Space>
          <Button size="small" onClick={() => setOutput([])}>Clear</Button>
          <Button size="small" danger onClick={disconnect}>Disconnect</Button>
        </Space>
      </div>

      <div
        ref={terminalRef}
        style={{
          background: '#1e1e1e',
          color: '#fff',
          padding: '16px',
          borderRadius: '4px',
          minHeight: '320px',
          maxHeight: '400px',
          overflow: 'auto',
        }}
      >
        {output.map((line, index) => (
          <div key={index} style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {line}
          </div>
        ))}
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <span style={{ color: '#52c41a', marginRight: 8 }}>root@gpu-container:~#</span>
          <input
            ref={inputRef}
            type="text"
            onKeyDown={handleKeyDown}
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: '#fff',
              fontFamily: 'inherit',
              fontSize: 'inherit',
            }}
            autoFocus
            disabled={!connected}
          />
        </div>
      </div>

      <div style={{ marginTop: 8, fontSize: '12px', color: '#999' }}>
        提示: 输入 <code style={{ background: '#f5f5f5', padding: '2px 4px', borderRadius: '2px' }}>help</code> 查看可用命令
      </div>
    </div>
  )
}

export default Terminal
