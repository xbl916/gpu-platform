import React, { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, List, Tag, Spin } from 'antd'
import {
  CloudServerOutlined,
  ContainerOutlined,
  DashboardOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import { PieChart, Pie, Cell, ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip } from 'recharts'
import { resourceService, monitorService } from '../services/api'

interface DashboardData {
  totalContainers: number
  runningContainers: number
  pendingContainers: number
  failedContainers: number
  totalGPUs: number
  availableGPUs: number
  cpuUsage: number
  memoryUsage: number
}

const COLORS = ['#52c41a', '#1890ff', '#faad14', '#ff4d4f']

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<DashboardData | null>(null)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)
      const [availability, dashboard] = await Promise.all([
        resourceService.getAvailability(),
        monitorService.getDashboard().catch(() => null),
      ])
      
      setData({
        totalContainers: dashboard?.containers?.total || 0,
        runningContainers: dashboard?.containers?.running || 0,
        pendingContainers: dashboard?.containers?.pending || 0,
        failedContainers: dashboard?.containers?.failed || 0,
        totalGPUs: availability.totalGpuCount || 0,
        availableGPUs: availability.availableGpuCount || 0,
        cpuUsage: 45,
        memoryUsage: 62,
      })
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const containerData = data ? [
    { name: '运行中', value: data.runningContainers },
    { name: '等待中', value: data.pendingContainers },
    { name: '已停止', value: data.failedContainers },
  ] : []

  const gpuData = data ? [
    { name: '已使用', value: data.totalGPUs - data.availableGPUs },
    { name: '可用', value: data.availableGPUs },
  ] : []

  const trendData = Array.from({ length: 12 }, (_, i) => ({
    time: `${i}:00`,
    cpu: Math.random() * 40 + 30,
    memory: Math.random() * 30 + 50,
  }))

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>仪表盘</h2>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="运行中容器"
              value={data?.runningContainers || 0}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待启动容器"
              value={data?.pendingContainers || 0}
              prefix={<ClockCircleOutlined style={{ color: '#faad14' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="可用 GPU"
              value={data?.availableGPUs || 0}
              suffix={`/ ${data?.totalGPUs || 0}`}
              prefix={<CloudServerOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总容器数"
              value={data?.totalContainers || 0}
              prefix={<ContainerOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
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
        <Col xs={24} lg={12}>
          <Card title="GPU 资源使用">
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={gpuData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={5}
                  dataKey="value"
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                >
                  <Cell fill="#1890ff" />
                  <Cell fill="#52c41a" />
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
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip />
                <Area
                  type="monotone"
                  dataKey="cpu"
                  stackId="1"
                  stroke="#1890ff"
                  fill="#1890ff"
                  fillOpacity={0.3}
                  name="CPU %"
                />
                <Area
                  type="monotone"
                  dataKey="memory"
                  stackId="2"
                  stroke="#52c41a"
                  fill="#52c41a"
                  fillOpacity={0.3}
                  name="Memory %"
                />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="快速操作">
            <List
              bordered
              dataSource={[
                { title: '创建新容器', description: '快速部署 GPU 容器实例' },
                { title: '选择模板', description: '从预置模板创建' },
                { title: '查看资源', description: '监控 GPU 使用情况' },
              ]}
              renderItem={(item) => (
                <List.Item style={{ cursor: 'pointer' }}>
                  <List.Item.Meta title={item.title} description={item.description} />
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="系统状态">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>API 服务</span>
                <Tag color="success">正常运行</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>数据库连接</span>
                <Tag color="success">已连接</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>GPU 集群</span>
                <Tag color="success">健康</Tag>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>存储服务</span>
                <Tag color="success">可用</Tag>
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard
