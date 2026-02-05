import React, { useEffect, useState } from 'react'
import { Table, Button, Card, Space, Tag, Input, Modal, Form, Select, Slider, message, Popconfirm } from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  ContainerOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { containerService, templateService } from '../services/api'

interface Container {
  id: string
  name: string
  status: string
  gpuCount: number
  cpuCores: number
  memoryMb: number
  gpuServer: string
  createdAt: string
}

const Containers: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [containers, setContainers] = useState<Container[]>([])
  const [total, setTotal] = useState(0)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [templates, setTemplates] = useState<any[]>([])
  const [form] = Form.useForm()
  const navigate = useNavigate()

  useEffect(() => {
    loadContainers()
    loadTemplates()
  }, [])

  const loadContainers = async () => {
    try {
      setLoading(true)
      const data = await containerService.list({ limit: 100 })
      setContainers(data || [])
      setTotal(Array.isArray(data) ? data.length : 0)
    } catch (error) {
      console.error('Failed to load containers:', error)
      setContainers([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const loadTemplates = async () => {
    try {
      constService.list({ limit: 100 })
      setTemplates data = await template(data || [])
    } catch (error) {
      console.error('Failed to load templates:', error)
    }
  }

  const handleCreate = async (values: any) => {
    try {
      await containerService.create({
        name: values.name,
        resources: {
          gpuCount: values.gpuCount,
          gpuModel: values.gpuModel,
          cpuCores: values.cpuCores,
          memoryMb: values.memoryMb,
          storageGb: values.storageGb,
          image: values.image,
        },
        templateId: values.templateId,
      })
      message.success('容器创建成功')
      setCreateModalOpen(false)
      form.resetFields()
      loadContainers()
    } catch (error: any) {
      message.error(error.response?.data?.error || '创建失败')
    }
  }

  const handleStart = async (id: string) => {
    try {
      await containerService.start(id)
      message.success('容器已启动')
      loadContainers()
    } catch (error) {
      message.error('启动失败')
    }
  }

  const handleStop = async (id: string) => {
    try {
      await containerService.stop(id)
      message.success('容器已停止')
      loadContainers()
    } catch (error) {
      message.error('停止失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await containerService.delete(id)
      message.success('容器已删除')
      loadContainers()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'success'
      case 'pending':
      case 'creating':
        return 'processing'
      case 'stopped':
        return 'default'
      case 'error':
        return 'error'
      default:
        return 'default'
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Container) => (
        <a onClick={() => navigate(`/containers/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status === 'running' ? '运行中' : status === 'pending' ? '等待中' : status}
        </Tag>
      ),
    },
    {
      title: 'GPU',
      dataIndex: 'gpuCount',
      key: 'gpuCount',
      render: (count: number) => `${count} GPU`,
    },
    {
      title: 'CPU',
      dataIndex: 'cpuCores',
      key: 'cpuCores',
      render: (cores: number) => `${cores} 核`,
    },
    {
      title: '内存',
      dataIndex: 'memoryMb',
      key: 'memoryMb',
      render: (mb: number) => `${Math.round(mb / 1024)} GB`,
    },
    {
      title: '服务器',
      dataIndex: 'gpuServer',
      key: 'gpuServer',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Container) => (
        <Space>
          {record.status !== 'running' && (
            <Button
              type="text"
              icon={<PlayCircleOutlined />}
              onClick={() => handleStart(record.id)}
            />
          )}
          {record.status === 'running' && (
            <Button
              type="text"
              icon={<PauseCircleOutlined />}
              onClick={() => handleStop(record.id)}
              danger
            />
          )}
          <Button
            type="text"
            icon={<ContainerOutlined />}
            onClick={() => navigate(`/containers/${record.id}`)}
          />
          <Popconfirm
            title="确定删除此容器？"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button type="text" icon={<DeleteOutlined />} danger />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="容器管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>
            创建容器
          </Button>
        }
      >
        <Table
          loading={loading}
          columns={columns}
          dataSource={containers}
          rowKey="id"
          pagination={{ total, pageSize: 10, showTotal: (total) => `共 ${total} 个容器` }}
        />
      </Card>

      <Modal
        title="创建容器"
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="容器名称" rules={[{ required: true }]}>
            <Input placeholder="my-gpu-container" />
          </Form.Item>

          <Form.Item name="templateId" label="选择模板 (可选)">
            <Select placeholder="选择预置模板" allowClear>
              {templates.map((t) => (
                <Select.Option key={t.id} value={t.id}>
                  {t.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="gpuCount" label="GPU 数量" rules={[{ required: true }]}>
            <Select placeholder="选择 GPU 数量">
              <Select.Option value={1}>1 GPU</Select.Option>
              <Select.Option value={2}>2 GPU</Select.Option>
              <Select.Option value={4}>4 GPU</Select.Option>
              <Select.Option value={8}>8 GPU</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="gpuModel" label="GPU 型号">
            <Select placeholder="选择 GPU 型号 (可选)">
              <Select.Option value="a100">NVIDIA A100</Select.Option>
              <Select.Option value="v100">NVIDIA V100</Select.Option>
              <Select.Option value="rtx3090">RTX 3090</Select.Option>
              <Select.Option value="rtx4090">RTX 4090</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="cpuCores" label="CPU 核心数" rules={[{ required: true }]}>
            <Select placeholder="选择 CPU 核心数">
              <Select.Option value={4}>4 核</Select.Option>
              <Select.Option value={8}>8 核</Select.Option>
              <Select.Option value={16}>16 核</Select.Option>
              <Select.Option value={32}>32 核</Select.Option>
              <Select.Option value={64}>64 核</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="memoryMb" label="内存大小 (GB)" rules={[{ required: true }]}>
            <Select placeholder="选择内存大小">
              <Select.Option value={16384}>16 GB</Select.Option>
              <Select.Option value={32768}>32 GB</Select.Option>
              <Select.Option value={65536}>64 GB</Select.Option>
              <Select.Option value={131072}>128 GB</Select.Option>
              <Select.Option value={262144}>256 GB</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="storageGb" label="存储大小 (GB)">
            <Select placeholder="选择存储大小" defaultValue={100}>
              <Select.Option value={100}>100 GB</Select.Option>
              <Select.Option value={200}>200 GB</Select.Option>
              <Select.Option value={500}>500 GB</Select.Option>
              <Select.Option value={1000}>1 TB</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="image" label="容器镜像">
            <Input placeholder="nvidia/cuda:12.1-runtime-ubuntu22.04" />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setCreateModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit">
                创建
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Containers
