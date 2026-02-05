import React, { useEffect, useState } from 'react'
import { Row, Col, Card, Table, Tag, Progress, Statistic, Spin } from 'antd'
import {
  CloudServerOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DashboardOutlined,
} from '@ant-design/icons'
import { resourceService } from '../services/api'

interface GPUServer {
  id: string
  hostname: string
  ipAddress: string
  status: string
  specs: {
    cpuCores: number
    memoryMb: number
    storageGb: number
    nvidiaDriver: string
  }
  gpuDevices: Array<{
    id: string
    name: string
    model: string
    memoryMb: number
    status: string
    temperature: number
  }>
  lastHeartbeat: string
}

interface GPUPool {
  id: string
  name: string
  description: string
  schedulingPolicy: string
  serversCount: number
  totalGPU: number
  availableGPU: number
}

const Resources: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [servers, setServers] = useState<GPUServer[]>([])
  const [pools, setPools] = useState<GPUPool[]>([])
  const [availability, setAvailability] = useState<any>(null)

  useEffect(() => {
    loadResources()
  }, [])

  const loadResources = async () => {
    try {
      setLoading(true)
      const [serverData, poolData, availData] = await Promise.all([
        resourceService.listGPUServers().catch(() => []),
        resourceService.listGPUPools().catch(() => []),
        resourceService.getAvailability().catch(() => null),
      ])
      setServers((serverData as any)?.data || serverData || [])
      setPools((poolData as any)?.data || poolData || [])
      setAvailability(availData)
    } catch (error) {
      console.error('Failed to load resources:', error)
    } finally {
      setLoading(false)
    }
  }

  const serverColumns = [
    {
      title: '主机名',
      dataIndex: 'hostname',
      key: 'hostname',
    },
    {
      title: 'IP 地址',
      dataIndex: 'ipAddress',
      key: 'ipAddress',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'online' ? 'success' : status === 'offline' ? 'error' : 'warning'}>
          {status === 'online' ? '在线' : status === 'offline' ? '离线' : status}
        </Tag>
      ),
    },
    {
      title: 'CPU',
      dataIndex: ['specs', 'cpuCores'],
      key: 'cpu',
      render: (cores: number) => `${cores} 核`,
    },
    {
      title: '内存',
      dataIndex: ['specs', 'memoryMb'],
      key: 'memory',
      render: (mb: number) => `${Math.round(mb / 1024)} GB`,
    },
    {
      title: 'GPU 数量',
      key: 'gpuCount',
      render: (_: any, record: GPUServer) => record.gpuDevices?.length || 0,
    },
    {
      title: 'GPU 型号',
      key: 'gpuModels',
      render: (_: any, record: GPUServer) => (
        <Space wrap>
          {record.gpuDevices?.map((g) => (
            <Tag key={g.id}>{g.model}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '驱动版本',
      dataIndex: ['specs', 'nvidiaDriver'],
      key: 'driver',
    },
    {
      title: '最后心跳',
      dataIndex: 'lastHeartbeat',
      key: 'heartbeat',
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
    },
  ]

  const poolColumns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '描述', dataIndex: 'description', key: 'description' },
    { title: '调度策略', dataIndex: 'schedulingPolicy', key: 'policy' },
    { title: '服务器数', dataIndex: 'serversCount', key: 'servers' },
    { title: '总 GPU', dataIndex: 'totalGPU', key: 'total' },
    {
      title: '可用 GPU',
      dataIndex: 'availableGPU',
      key: 'available',
      render: (v: number) => <Tag color="green">{v}</Tag>,
    },
  ]

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>资源监控</h2>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="GPU 服务器"
              value={servers.length}
              prefix={<CloudServerOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="在线服务器"
              value={servers.filter(s => s.status === 'online').length}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="可用 GPU"
              value={availability?.availableGpuCount || 0}
              suffix={`/ ${availability?.totalGpuCount || 0}`}
              prefix={<DashboardOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {Object.entries(availability?.gpuByModel || {}).map(([model, count]) => (
          <Col xs={12} sm={6} key={model}>
            <Card size="small">
              <Progress
                type="circle"
                percent={Math.round((count as number) / (availability?.totalGpuCount || 1) * 100)}
                format={() => (count as number)}
              />
              <div style={{ marginTop: 8, textAlign: 'center' }}>{model.toUpperCase()}</div>
            </Card>
          </Col>
        ))}
      </Row>

      <Card title="资源池" style={{ marginBottom: 24 }}>
        <Table
          columns={poolColumns}
          dataSource={pools}
          rowKey="id"
          pagination={false}
        />
      </Card>

      <Card title="GPU 服务器列表">
        <Table
          columns={serverColumns}
          dataSource={servers}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  )
}

export default Resources
