import { Row, Col, Card, Statistic, Progress } from 'antd'
import { PieChart, Pie, Cell, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'

const healthData = [
  { name: '正常', value: 92.4, color: '#52c41a' },
  { name: '延迟', value: 5.8, color: '#1890ff' },
  { name: '异常', value: 1.8, color: '#ff4d4f' },
]

const costData = [
  { date: '06-01', cost: 450 },
  { date: '06-02', cost: 520 },
  { date: '06-03', cost: 480 },
  { date: '06-04', cost: 620 },
  { date: '06-05', cost: 580 },
  { date: '06-06', cost: 680 },
  { date: '06-07', cost: 750 },
]

const modelUsage = [
  { model: 'gpt-4', usage: 8234, percent: 28 },
  { model: 'gpt-3.5-turbo', usage: 15234, percent: 52 },
  { model: 'claude-3-opus', usage: 5421, percent: 18 },
  { model: 'gemini-pro', usage: 2134, percent: 7 },
]

export default function Statistics() {
  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card title="站点健康度">
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie data={healthData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label>
                  {healthData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>

        <Col span={16}>
          <Card title="费用趋势">
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={costData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Line type="monotone" dataKey="cost" stroke="#1890ff" strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={8}>
          <Card>
            <Statistic title="总费用" value={2845.31} prefix="¥" precision={2} />
            <div style={{ marginTop: 16, fontSize: 12, opacity: 0.7 }}>本月累计</div>
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="预估月费用" value={12450} prefix="¥" />
            <div style={{ marginTop: 16, fontSize: 12, opacity: 0.7 }}>基于当前使用量</div>
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="节省成本" value={3280} prefix="¥" suffix="/月" valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: 16, fontSize: 12, opacity: 0.7 }}>通过智能路由</div>
          </Card>
        </Col>
      </Row>

      <Row style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="模型使用统计">
            {modelUsage.map(item => (
              <div key={item.model} style={{ marginBottom: 20 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontFamily: 'monospace', fontWeight: 500 }}>{item.model}</span>
                  <span>{item.usage} 次</span>
                </div>
                <Progress percent={item.percent} strokeColor="#1890ff" />
              </div>
            ))}
          </Card>
        </Col>
      </Row>
    </div>
  )
}
