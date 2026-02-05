import React, { useEffect, useState } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, Space, message } from 'antd'
import { PlusOutlined, DeleteOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { templateService } from '../services/api'

interface Template {
  id: string
  name: string
  description: string
  config: {
    baseImage: string
    cudaVersion: string
    pythonVersion: string
    packages: string[]
    pipPackages: string[]
  }
  usageCount: number
  isPublic: boolean
  createdAt: string
}

const Templates: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [templates, setTemplates] = useState<Template[]>([])
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [form] = Form.useForm()
  const navigate = useNavigate()

  useEffect(() => {
    loadTemplates()
  }, [])

  const loadTemplates = async () => {
    try {
      setLoading(true)
      const data = await templateService.list({ limit: 100 })
      setTemplates((data as any)?.data || data || [])
    } catch (error) {
      console.error('Failed to load templates:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (values: any) => {
    try {
      await templateService.create({
        name: values.name,
        description: values.description,
        config: {
          baseImage: values.baseImage || 'nvidia/cuda:12.1-runtime-ubuntu22.04',
          cudaVersion: values.cudaVersion || '12.1',
          pythonVersion: values.pythonVersion || '3.10',
          packages: values.packages?.split(',').map((s: string) => s.trim()) || [],
          pipPackages: values.pipPackages?.split(',').map((s: string) => s.trim()) || [],
        },
      })
      message.success('模板创建成功')
      setCreateModalOpen(false)
      form.resetFields()
      loadTemplates()
    } catch (error) {
      message.error('创建失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await templateService.delete(id)
      message.success('模板已删除')
      loadTemplates()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'CUDA',
      dataIndex: ['config', 'cudaVersion'],
      key: 'cuda',
      render: (v: string) => v || '-',
    },
    {
      title: 'Python',
      dataIndex: ['config', 'pythonVersion'],
      key: 'python',
      render: (v: string) => v || '-',
    },
    {
      title: '使用次数',
      dataIndex: 'usageCount',
      key: 'usageCount',
    },
    {
      title: '公开',
      dataIndex: 'isPublic',
      key: 'isPublic',
      render: (v: boolean) => <Tag color={v ? 'green' : 'blue'}>{v ? '是' : '否'}</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Template) => (
        <Space>
          <Button
            type="text"
            icon={<PlayCircleOutlined />}
            onClick={() => navigate(`/containers?template=${record.id}`)}
          >
            使用
          </Button>
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record.id)}
          />
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="容器模板"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>
            创建模板
          </Button>
        }
      >
        <Table
          loading={loading}
          columns={columns}
          dataSource={templates}
          rowKey="id"
          pagination={{ pageSize: 10, showTotal: (total) => `共 ${total} 个模板` }}
        />
      </Card>

      <Modal
        title="创建模板"
        open={createModalOpen}
        onCancel={() => setCreateModalOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="模板名称" rules={[{ required: true }]}>
            <Input placeholder="pytorch-training" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="PyTorch 训练环境" rows={3} />
          </Form.Item>
          <Form.Item name="baseImage" label="基础镜像">
            <Input placeholder="nvidia/cuda:12.1-runtime-ubuntu22.04" />
          </Form.Item>
          <Form.Item name="cudaVersion" label="CUDA 版本">
            <Input placeholder="12.1" />
          </Form.Item>
          <Form.Item name="pythonVersion" label="Python 版本">
            <Input placeholder="3.10" />
          </Form.Item>
          <Form.Item name="packages" label="系统包 (逗号分隔)">
            <Input placeholder="python3-dev, gcc, cmake" />
          </Form.Item>
          <Form.Item name="pipPackages" label="Python 包 (逗号分隔)">
            <Input placeholder="torch, torchvision, tensorboard" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setCreateModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit">创建</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Templates
