import React, { useEffect, useState, useRef } from 'react'
import { Card, Descriptions, Button, Space, Tag, Tabs, Progress, Row, Col, Statistic } from 'antd'
import { ArrowLeftOutlined, PlayCircleOutlined, PauseCircleOutlined, ReloadOutlined, CloudServerOutlined, TerminalOutlined, FolderOutlined, UploadOutlined, DashboardOutlined } from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { AreaChart, Area, CartesianGrid, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { containerService } from '../../services/api'

interface ContainerMetrics {
  cpuUsage: number
  memoryUsageMb: number
  gpuUtilization: number
  gpuMemoryUsedMb: number
  recordedAt: string
}

interface SSHInfo {
  host: string
  port: number
  user: string
}

interface ContainerDetail {
  id: string
  name: string
  status: string
  resources: {
    gpuCount: number
    cpuCores: number
    memoryMb: number
    storageGb: number
    image: string
  }
  gpuServer: string
  webUrl: string
  sshInfo: SSHInfo
  createdAt: string
  startedAt: string
  metrics: ContainerMetrics[]
}

const ContainerDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ContainerDetail | null>(null)
  const [activeTab, setActiveTab] = useState('overview')

  useEffect(() => {
    if (id) loadContainerDetail()
  }, [id])

  const loadContainerDetail = async () => {
    try {
      setLoading(true)
      const data = await containerService.get(id!)
      setDetail(data)
    } catch (error) {
      console.error('Failed to load container:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleStart = async () => {
    try {
      await containerService.start(id!)
      loadContainerDetail()
    } catch (error) {
      console.error('Start failed:', error)
    }
  }

  const handleStop = async () => {
    try {
      await containerService.stop(id!)
      loadContainerDetail()
    } catch (error) {
      console.error('Stop failed:', error)
    }
  }

  const handleRestart = async () => {
    try {
      await containerService.restart(id!)
      loadContainerDetail()
    } catch (error) {
      console.error('Restart failed:', error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'success'
      case 'pending': case 'creating': return 'processing'
      case 'stopped': return 'default'
      case 'error': return 'error'
      default: return 'default'
    }
  }

  const mockMetrics = Array.from({ length: 20 }, (_, i) => ({
    time: new Date(Date.now() - (20 - i) * 5000).toLocaleTimeString('zh-CN'),
    cpu: 30 + Math.random() * 30,
    memory: 60 + Math.random() * 20,
    gpu: 40 + Math.random() * 40,
  }))

  const tabItems = [
    {
      key: 'overview',
      label: '概览',
      children: (
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="GPU 利用率" value={60} suffix="%" precision={1} prefix={<DashboardOutlined />} />
              <Progress percent={60} showInfo={false} strokeColor="#1890ff" />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="CPU 使用率" value={45} suffix="%" precision={1} prefix={<CloudServerOutlined />} />
              <Progress percent={45} showInfo={false} strokeColor="#52c41a" />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="显存使用" value={8} suffix="GB" precision={2} />
              <Progress percent={50} showInfo={false} strokeColor="#722ed1" />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic title="内存使用" value={12.5} suffix="GB" precision={2} />
              <Progress percent={55} showInfo={false} strokeColor="#fa8c16" />
            </Card>
          </Col>
          <Col span={24}>
            <Card title="实时监控 (近 2 分钟)">
              <Row gutter={16}>
                <Col span={12}>
                  <h4>GPU 利用率</h4>
                  <ResponsiveContainer width="100%" height={150}>
                    <AreaChart data={mockMetrics}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="time" />
                      <YAxis domain={[0, 100]} />
                      <Tooltip />
                      <Area type="monotone" dataKey="gpu" stroke="#1890ff" fill="#1890ff33" />
                    </AreaChart>
                  </ResponsiveContainer>
                </Col>
                <Col span={12}>
                  <h4>CPU 利用率</h4>
                  <ResponsiveContainer width="100%" height={150}>
                    <AreaChart data={mockMetrics}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="time" />
                      <YAxis domain={[0, 100]} />
                      <Tooltip />
                      <Area type="monotone" dataKey="cpu" stroke="#52c41a" fill="#52c41a33" />
                    </AreaChart>
                  </ResponsiveContainer>
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>
      ),
    },
    {
      key: 'terminal',
      label: '终端',
      children: (
        <Card title={<Space><TerminalOutlined /> SSH 终端</Space>}>
          <div style={{ background: '#1e1e1e', padding: '16px', borderRadius: '8px', minHeight: '400px' }}>
            <p style={{ color: '#52c41a' }}>root@gpu-container:~#</p>
            <p style={{ color: '#fff' }}>模拟 SSH 终端 - 请使用实际 SSH 客户端连接</p>
            <p style={{ color: '#888', fontSize: '12px' }}>SSH: root@{detail?.sshInfo.host || 'host'} -p {detail?.sshInfo.port || 22}</p>
          </div>
        </Card>
      ),
    },
    {
      key: 'files',
      label: '文件',
      children: (
        <Card title={<Space><FolderOutlined /> 文件管理器</Space>} extra={<Button icon={<UploadOutlined />}>上传</Button>}>
          <p>Workspace: /workspace (50 GB)</p>
          <p>Data: /data (200 GB)</p>
        </Card>
      ),
    },
    {
      key: 'logs',
      label: '日志',
      children: (
        <Card title="容器日志" extra={<Button onClick={loadContainerDetail}>刷新</Button>}>
          <pre style={{ background: '#1e1e1e', color: '#fff', padding: '16px', borderRadius: '8px', maxHeight: '400px', overflow: 'auto' }}>
{`[2024-01-15 10:30:15] Container started successfully
[2024-01-15 10:30:16] NVIDIA GPU Driver: 535.154.05
[2024-01-15 10:30:17] CUDA Version: 12.1
[2024-01-15 10:30:18] GPU Memory: 16384 MB
[2024-01-15 10:30:19] Container is ready
[2024-01-15 10:30:20] SSH service started on port 22`}
          </pre>
        </Card>
      ),
    },
    {
      key: 'info',
      label: '信息',
      children: (
        <Descriptions bordered column={2}>
          <Descriptions.Item label="容器 ID">{detail?.id}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={getStatusColor(detail?.status || '')}>{detail?.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="GPU">{detail?.resources.gpuCount}</Descriptions.Item>
          <Descriptions.Item label="CPU">{detail?.resources.cpuCores} 核</Descriptions.Item>
          <Descriptions.Item label="内存">{Math.round((detail?.resources.memoryMb || 0) / 1024)} GB</Descriptions.Item>
          <Descriptions.Item label="存储">{detail?.resources.storageGb} GB</Descriptions.Item>
          <Descriptions.Item label="镜像" span={2}>{detail?.resources.image}</Descriptions.Item>
          <Descriptions.Item label="GPU 服务器">{detail?.gpuServer}</Descriptions.Item>
          <Descriptions.Item label="SSH">{detail?.sshInfo.user}@{detail?.sshInfo.host}:{detail?.sshInfo.port}</Descriptions.Item>
          <Descriptions.Item label="创建">{new Date(detail?.createdAt || '').toLocaleString('zh-CN')}</Descriptions.Item>
          <Descriptions.Item label="启动">{detail?.startedAt ? new Date(detail.startedAt).toLocaleString('zh-CN') : '-'}</Descriptions.Item>
        </Descriptions>
      ),
    },
  ]

  if (loading && !detail) return <div style={{ padding: 100, textAlign: 'center' }}>加载中...</div>
  if (!detail) return <div>容器不存在</div>

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/containers')}>返回列表</Button>
      </Space>

      <Card
        title={<Space><CloudServerOutlined />{detail.name}<Tag color={getStatusColor(detail.status)}>{detail.status === 'running' ? '运行中' : detail.status}</Tag></Space>}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadContainerDetail}>刷新</Button>
            {detail.status !== 'running' && <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart}>启动</Button>}
            {detail.status === 'running' && <Button danger icon={<PauseCircleOutlined />} onClick={handleStop}>停止</Button>}
            {detail.status === 'running' && <Button onClick={handleRestart}>重启</Button>}
          </Space>
        }
      >
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
      </Card>
    </div>
  )
}

export default ContainerDetail
