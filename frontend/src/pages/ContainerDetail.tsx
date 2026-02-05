import React, { useEffect, useState } from 'react'
import { Card, Descriptions, Button, Space, Tag, Timeline, Spin, Tabs, Terminal } from 'antd'
import {
  ArrowLeftOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  CloudServerOutlined,
} from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { containerService } from '../services/api'

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
  sshInfo: {
    host: string
    port: number
    user: string
  }
  createdAt: string
  startedAt: string
}

const ContainerDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ContainerDetail | null>(null)

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
      message.success('容器已启动')
      loadContainerDetail()
    } catch (error) {
      message.error('启动失败')
    }
  }

  const handleStop = async () => {
    try {
      await containerService.stop(id!)
      message.success('容器已停止')
      loadContainerDetail()
    } catch (error) {
      message.error('停止失败')
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'success'
      case 'pending': return 'processing'
      case 'stopped': return 'default'
      case 'error': return 'error'
      default: return 'default'
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!detail) {
    return <div>容器不存在</div>
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/containers')}>
          返回列表
        </Button>
      </Space>

      <Card
        title={
          <Space>
            <CloudServerOutlined />
            {detail.name}
            <Tag color={getStatusColor(detail.status)}>
              {detail.status === 'running' ? '运行中' : detail.status}
            </Tag>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadContainerDetail}>刷新</Button>
            {detail.status !== 'running' && (
              <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart}>
                启动
              </Button>
            )}
            {detail.status === 'running' && (
              <Button danger icon={<PauseCircleOutlined />} onClick={handleStop}>
                停止
              </Button>
            )}
          </Space>
        }
      >
        <Tabs
          items={[
            {
              key: 'info',
              label: '基本信息',
              children: (
                <Descriptions bordered column={2}>
                  <Descriptions.Item label="容器 ID">{detail.id}</Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag color={getStatusColor(detail.status)}>{detail.status}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="GPU 数量">{detail.resources.gpuCount}</Descriptions.Item>
                  <Descriptions.Item label="CPU 核心数">{detail.resources.cpuCores}</Descriptions.Item>
                  <Descriptions.Item label="内存">{Math.round(detail.resources.memoryMb / 1024)} GB</Descriptions.Item>
                  <Descriptions.Item label="存储">{detail.resources.storageGb} GB</Descriptions.Item>
                  <Descriptions.Item label="容器镜像" span={2}>{detail.resources.image}</Descriptions.Item>
                  <Descriptions.Item label="GPU 服务器">{detail.gpuServer}</Descriptions.Item>
                  <Descriptions.Item label="创建时间">{new Date(detail.createdAt).toLocaleString('zh-CN')}</Descriptions.Item>
                  <Descriptions.Item label="Web 访问">{detail.webUrl}</Descriptions.Item>
                  <Descriptions.Item label="SSH">
                    {detail.sshInfo.user}@{detail.sshInfo.host}:{detail.sshInfo.port}
                  </Descriptions.Item>
                </Descriptions>
              ),
            },
            {
              key: 'logs',
              label: '日志',
              children: (
                <Card size="small" style={{ background: '#1e1e1e', color: '#fff', minHeight: 300 }}>
                  <pre style={{ margin: 0, color: '#fff' }}>
Container logs will be displayed here...
                  </pre>
                </Card>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}

export default ContainerDetail
