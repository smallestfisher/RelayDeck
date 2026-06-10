import { Row, Col, Card, Statistic, Progress, Table, Tag } from 'antd'
import { ArrowUpOutlined, CheckCircleOutlined, WarningOutlined } from '@ant-design/icons'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

const trendData = [
  { time: '00:00', value: 120 },
  { time: '04:00', value: 180 },
  { time: '08:00', value: 250 },
  { time: '12:00', value: 320 },
  { time: '16:00', value: 280 },
  { time: '20:00', value: 350 },
]

const siteData = [
  { key: '1', name: 'GPT站点-01', status: 'online', requests: 12453, avgTime: '120ms', success: 99.2 },
  { key: '2', name: 'Claude站点-01', status: 'online', requests: 8234, avgTime: '95ms', success: 99.8 },
  { key: '3', name: 'Gemini站点-01', status: 'warning', requests: 3421, avgTime: '180ms', success: 95.3 },
  { key: '4', name: 'OpenAI中转-02', status: 'online', requests: 15234, avgTime: '110ms', success: 98.9 },
]

export default function Dashboard() {
  const columns = [
    { 
      title: '站点名称', 
      dataIndex: 'name', 
      key: 'name',
      render: (text: string) => <span style={{ fontWeight: 500 }}>{text}</span>
    },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'online' ? 'success' : 'warning'} icon={status === 'online' ? <CheckCircleOutlined /> : <WarningOutlined />}>
          {status === 'online' ? '在线' : '警告'}
        </Tag>
      )
    },
    { title: '请求数', dataIndex: 'requests', key: 'requests' },
    { title: '平均响应', dataIndex: 'avgTime', key: 'avgTime' },
    { 
      title: '成功率', 
      dataIndex: 'success', 
      key: 'success',
      render: (val: number) => <span style={{ color: val > 98 ? '#52c41a' : '#faad14' }}>{val}%</span>
    },
  ]

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总请求数"
              value={126743}
              prefix="🔥"
              valueStyle={{ color: '#3f8600' }}
              suffix={<ArrowUpOutlined />}
            />
            <div style={{ marginTop: 16, fontSize: 12, color: '#8b92b0' }}>
              较昨日 +12.5%
            </div>
          </Card>
        </Col>
        
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃站点"
              value={12}
              prefix="🌐"
              suffix="/ 15"
            />
            <Progress percent={80} showInfo={false} strokeColor="#1890ff" style={{ marginTop: 16 }} />
          </Card>
        </Col>
        
        <Col span={6}>
          <Card>
            <Statistic
              title="平均响应时间"
              value={124}
              suffix="ms"
              prefix="⚡"
              valueStyle={{ color: '#1890ff' }}
            />
            <div style={{ marginTop: 16, fontSize: 12, color: '#8b92b0' }}>
              较昨日 -8.3%
            </div>
          </Card>
        </Col>
        
        <Col span={6}>
          <Card>
            <Statistic
              title="今日收益"
              value={2845.31}
              prefix="💰"
              precision={2}
              valueStyle={{ color: '#faad14' }}
            />
            <div style={{ marginTop: 16, fontSize: 12, color: '#8b92b0' }}>
              预估月收益 ¥85,359
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={16}>
          <Card title="请求趋势" bordered={false}>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={trendData}>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                <XAxis dataKey="time" stroke="#8b92b0" />
                <YAxis stroke="#8b92b0" />
                <Tooltip 
                  contentStyle={{ background: '#0f1429', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 8 }}
                />
                <Line type="monotone" dataKey="value" stroke="#1890ff" strokeWidth={3} dot={{ fill: '#1890ff', r: 5 }} />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        
        <Col span={8}>
          <Card title="站点健康度" bordered={false}>
            <div style={{ padding: '20px 0' }}>
              <div style={{ marginBottom: 24 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>正常运行</span>
                  <span style={{ color: '#52c41a' }}>92.4%</span>
                </div>
                <Progress percent={92.4} showInfo={false} strokeColor="#52c41a" />
              </div>
              
              <div style={{ marginBottom: 24 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>响应延迟</span>
                  <span style={{ color: '#1890ff' }}>5.8%</span>
                </div>
                <Progress percent={5.8} showInfo={false} strokeColor="#1890ff" />
              </div>
              
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>故障异常</span>
                  <span style={{ color: '#ff4d4f' }}>1.8%</span>
                </div>
                <Progress percent={1.8} showInfo={false} strokeColor="#ff4d4f" />
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Row style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="活跃站点" bordered={false}>
            <Table 
              columns={columns} 
              dataSource={siteData} 
              pagination={false}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
