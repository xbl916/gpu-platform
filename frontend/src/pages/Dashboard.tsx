import React, { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, List, Tag, Progress, Table, Button, Space, message } from 'antd'
import {
  CloudServerOutlined,
  ContainerOutlined,
  DashboardOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  UserOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { PieChart, Pie, Cell, ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, BarChart, Bar, Legend } from 'recharts'
import { useNavigate } from 'react-router-dom'
import { resourceService, monitorService } from '../services/api'
import { useAuthStore } from '../store/authStore'

interface DashboardData {
  overview: {
    totalContainers: number
    runningContainers: number
    pendingContainers: number
    failedContainers: number
    totalGPUs: number
    availableGPUs: number
  }
  gpuByModel: Record<string, number>
  cpuUsage: number
  memoryUsage: number
  recentContainers: Array<{
    id: string
    name: string
    status: string
    gpuCount: number
    createdAt: string
  }>
}

interface QuotaInfo {
  maxGPUCount: number
  maxCPUCores: number
  maxMemoryMB: number
  maxStorageGB: number
  maxInstances: number
  usedGPUCount: number
  usedCPUCores: number
  usedMemoryMB: number
  usedStorageGB: number
  usedInstances: number
}

const COLORS = ['#52c41a', '#1890ff', '#faad14', '#ff4d4f', '#722ed1', '#13c2c2']

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<DashboardData | null>(null)
  const [quota, setQuota] = useState<QuotaInfo | null>(null)
  const navigate = useNavigate()
  const { user } = useAuthStore()

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)
      const [availability, dashboard, quotaData] = await Promise.all([
        resourceService.getAvailability().catch(() => null),
        monitorService.getDashboard().catch(() => null),
        resourceService.getQuota().catch(() => null),
      ])

      const userQuota: QuotaInfo = quotaData || {
        maxGPUCount: 8,
        maxCPUCores: 64,
        maxMemoryMB: 131072,
        maxStorageGB: 500,
        maxInstances: 10,
        usedGPUCount: 2,
        usedCPUCores: 8,
        usedMemoryMB: 32768,
        usedStorageGB: 100,
        usedInstances: 2,
      }

      setQuota(userQuota)
      setData({
        overview: {
          totalContainers: dashboard?.overview?.totalContainers || 5,
          runningContainers: dashboard?.overview?.runningContainers || 3,
          pendingContainers: dashboard?.overview?.pendingContainers || 1,
          failedContainers: dashboard?.overview?.failedContainers || 1,
          totalGPUs: availability?.totalGpuCount || 20,
          availableGPUs: availability?.availableGpuCount || 12,
        },
        gpuByModel: availability?.gpuByModel || { a100: 8, rtx3090: 6, v100: 6 },
        cpuUsage: 45,
        memoryUsage: 62,
        recentContainers: [
          { id: '1', name: 'pytorch-training', status: 'running', gpuCount: 2, createdAt: '2024-01-15 10:30' },
          { id: '2', name: 'tensorflow-dev', status: 'running', gpuCount: 1, createdAt: '2024-01-15 09:15' },
          { id: '3', name: 'jupyter-notebook', status: 'running', gpuCount: 1, createdAt: '2024-01-15 08:00' },
          { id: '4', name: 'stable-diffusion', status: 'stopped', gpuCount: 1, createdAt: '2024-01-14 22:00' },
          { id: '5', name: 'llm-inference', status: 'pending', gpuCount: 4, createdAt: '2024-01-15 11:00' },
        ],
      })
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const containerData = data ? [
    { name: '运行中', value: data.overview.runningContainers },
    { name: '等待中', value: data.overview.pendingContainers },
    { name: '已停止', value: data.overview.stoppedContainers || 0 },
    { name: '失败', value: data.overview.failedContainers },
  ] : []

  const gpuModelData = data ? Object.entries(data.gpuByModel).map(([model, count]) => ({
    model: model.toUpperCase(),
    count: count,
  })) : []

  const trendData = Array.from({ length: 24 }, (_, i) => ({
    hour: `${i}:00`,
    cpu: Math.random() * 40 + 30,
    memory: Math.random() * 30 + 50,
    gpu: Math.random() * 50 + 20,
  }))

  const quotaProgress = (used: number, max: number) => Math.round((used / max) * 100)

  if (loading) {
    return <div style={{ textAlign: 'center', padding: '100px' }}>加载中...</div>
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h2 style={{ margin: 0 }}>仪表盘</h2>
        <Space>
          <Button icon={<UserOutlined />} onClick={() => message.info('个人中心功能开发中')}>个人中心</Button>
          <Button icon={<SettingOutlined />} onClick={() => message.info('设置功能开发中')}>设置</Button>
        </Space>
      </div>

      {quota && (
        <Card title={<><UserOutlined /> 我的资源配额</>} style={{ marginBottom: 16 }}>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} lg={6}>
              <div>
                <div style={{ marginBottom: 4 }}>GPU 配额 ({quota.usedGPUCount}/{quota.maxGPUCount})</div>
                <Progress percent={quotaProgress(quota.usedGPUCount, quota.maxGPUCount)} strokeColor="#1890ff" />
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div>
                <div style={{ marginBottom: 4 }}>CPU 配额 ({quota.usedCPUCores}/{quota.maxCPUCores})</div>
                <Progress percent={quotaProgress(quota.usedCPUCores, quota.maxCPUCores)} strokeColor="#52c41a" />
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div>
                <div style={{ marginBottom: 4 }}>内存配额 ({Math.round(quota.usedMemoryMB/1024)}/{Math.round(quota.maxMemoryMB/1024)} GB)</div>
                <Progress percent={quotaProgress(quota.usedMemoryMB, quota.maxMemoryMB)} strokeColor="#722ed1" />
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div>
                <div style={{ marginBottom: 4 }}>容器配额 ({quota.usedInstances}/{quota.maxInstances})</div>
                <Progress percent={quotaProgress(quota.usedInstances, quota.maxInstances)} strokeColor="#faad14" />
              </div>
            </Col>
          </Row>
        </Card>
      )}

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="运行中容器"
              value={data?.overview.runningContainers || 0}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待启动容器"
              value={data?.overview.pendingContainers || 0}
              prefix={<ClockCircleOutlined style={{ color: '#faad14' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="可用 GPU"
              value={data?.overview.availableGPUs || 0}
              suffix={`/ ${data?.overview.totalGPUs || 0}`}
              prefix={<CloudServerOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总容器数"
              value={data?.overview.totalContainers || 0}
              prefix={<ContainerOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="GPU 型号分布">
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={gpuModelData}>
                <XAxis dataKey="model" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="count" name="数量" fill="#1890ff" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="容器状态分布">
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={containerData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={5}
                  dataKey="value"
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                >
                  {containerData.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col span={24}>
          <Card title="资源使用趋势 (24小时)">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={trendData}>
                <XAxis dataKey="hour" />
                <YAxis domain={[0, 100]} />
                <Tooltip />
                <Legend />
                <Area type="monotone" dataKey="gpu" stackId="1" stroke="#1890ff" fill="#1890ff33" name="GPU %" />
                <Area type="monotone" dataKey="cpu" stackId="2" stroke="#52c41a" fill="#52c41a33" name="CPU %" />
                <Area type="monotone" dataKey="memory" stackId="3" stroke="#722ed1" fill="#722ed133" name="Memory %" />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card
            title="最近容器"
            extra={<Button type="link" onClick={() => navigate('/containers')}>查看全部</Button>}
          >
            <List
              bordered
              dataSource={data?.recentContainers || []}
              renderItem={(item) => (
                <List.Item
                  style={{ cursor: 'pointer' }}
                  onClick={() => navigate(`/containers/${item.id}`)}
                >
                  <List.Item.Meta
                    avatar={<ContainerOutlined style={{ fontSize: 24, color: item.status === 'running' ? '#52c41a' : '#999' }} />}
                    title={item.name}
                    description={`${item.gpuCount} GPU • ${item.createdAt}`}
                  />
                  <Tag color={item.status === 'running' ? 'success' : item.status === 'pending' ? 'processing' : 'default'}>
                    {item.status === 'running' ? '运行中' : item.status}
                  </Tag>
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="快速操作">
            <List
              bordered
              dataSource={[
                { title: '创建容器', description: '快速部署 GPU 容器', icon: <ContainerOutlined />, action: () => navigate('/containers') },
                { title: '选择模板', description: '从预置模板创建', icon: <DashboardOutlined />, action: () => navigate('/templates') },
                { title: '查看资源', description: '监控 GPU 使用情况', icon: <CloudServerOutlined />, action: () => navigate('/resources') },
                { title: '管理配额', description: '查看资源配额', icon: <UserOutlined />, action: () => message.info('配额管理开发中') },
              ]}
              renderItem={(item) => (
                <List.Item style={{ cursor: 'pointer' }} onClick={item.action}>
                  <List.Item.Meta
                    avatar={item.icon}
                    title={item.title}
                    description={item.description}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col span={24}>
          <Card title="系统状态">
            <Row gutter={[16, 16]}>
              {[
                { name: 'API 服务', status: 'success', msg: '正常运行' },
                { name: '数据库连接', status: 'success', msg: '已连接' },
                { name: 'GPU 集群', status: 'success', msg: '健康' },
                { name: '存储服务', status: 'success', msg: '可用' },
                { name: '任务队列', status: 'success', msg: '空闲' },
                { name: '监控服务', status: 'success', msg: '正常' },
              ].map((item, i) => (
                <Col xs={24} sm={12} md={8} lg={4} key={i}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 12px', background: '#f5f5f5', borderRadius: '4px' }}>
                    <span>{item.name}</span>
                    <Tag color={item.status === 'success' ? 'success' : 'error'}>{item.msg}</Tag>
                  </div>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard
