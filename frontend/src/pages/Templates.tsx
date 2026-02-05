import React, { useEffect, useState } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, Space, message, Select, Popconfirm } from 'antd'
import { PlusOutlined, DeleteOutlined, PlayCircleOutlined, BuildOutlined, CloudUploadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { templateService } from '../services/api'

interface Template {
  id: string
  name: string
  description: string
  dockerImage: string
  version: number
  status: string
  buildStatus: string
  usageCount: number
  isPublic: boolean
  createdAt: string
}

const Templates: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [templates, setTemplates] = useState<Template[]>([])
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [buildModalOpen, setBuildModalOpen] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null)
  const [form] = Form.useForm()
  const [buildForm] = Form.useForm()
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
      setTemplates([])
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

  const handleBuild = async (values: any) => {
    if (!selectedTemplate) return
    try {
      await templateService.build(selectedTemplate.id, {
        baseImage: values.baseImage,
        pipPackages: values.pipPackages?.split(',').map((s: string) => s.trim()),
        envVars: {},
      })
      message.success('构建任务已提交')
      setBuildModalOpen(false)
      buildForm.resetFields()
    } catch (error) {
      message.error('构建提交失败')
    }
  }

  const handlePublish = async (id: string) => {
    try {
      await templateService.publish(id)
      message.success('模板已发布')
      loadTemplates()
    } catch (error) {
      message.error('发布失败')
    }
  }

  const showBuildModal = (template: Template) => {
    setSelectedTemplate(template)
    setBuildModalOpen(true)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'published':
        return 'success'
      case 'building':
        return 'processing'
      case 'draft':
        return 'default'
      case 'failed':
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
    },
    {
      title: '镜像',
      dataIndex: 'dockerImage',
      key: 'dockerImage',
      render: (v: string) => v || '-',
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      render: (v: number) => `v${v}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (v: string) => (
        <Tag color={getStatusColor(v)}>
          {v === 'published' ? '已发布' : v === 'building' ? '构建中' : v === 'draft' ? '草稿' : v}
        </Tag>
      ),
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
          {record.status === 'draft' && (
            <Button
              type="text"
              icon={<BuildOutlined />}
              onClick={() => showBuildModal(record)}
            >
              构建
            </Button>
          )}
          {record.status === 'draft' && (
            <Popconfirm
              title="确定发布此模板？"
              onConfirm={() => handlePublish(record.id)}
            >
              <Button type="text" icon={<CloudUploadOutlined />}>
                发布
              </Button>
            </Popconfirm>
          )}
          <Popconfirm
            title="确定删除此模板？"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button type="text" danger icon={<DeleteOutlined />} />
          </Popconfirm>
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
        width={600}
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

      <Modal
        title={`构建镜像 - ${selectedTemplate?.name}`}
        open={buildModalOpen}
        onCancel={() => setBuildModalOpen(false)}
        footer={null}
        width={500}
      >
        <Form form={buildForm} layout="vertical" onFinish={handleBuild}>
          <Form.Item name="baseImage" label="基础镜像" rules={[{ required: true }]}>
            <Select
              placeholder="选择 CUDA 版本"
              options={[
                { value: 'nvidia/cuda:12.1-runtime-ubuntu22.04', label: 'CUDA 12.1 Ubuntu 22.04' },
                { value: 'nvidia/cuda:11.8-runtime-ubuntu22.04', label: 'CUDA 11.8 Ubuntu 22.04' },
                { value: 'nvidia/cuda:12.0-runtime-ubuntu22.04', label: 'CUDA 12.0 Ubuntu 22.04' },
                { value: 'nvidia/cuda:11.8-runtime-ubuntu20.04', label: 'CUDA 11.8 Ubuntu 20.04' },
              ]}
            />
          </Form.Item>
          <Form.Item name="pipPackages" label="Python 包 (逗号分隔)">
            <Input.TextArea placeholder="torch, torchvision, tensorboard, opencv-python" rows={3} />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setBuildModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit">开始构建</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Templates
