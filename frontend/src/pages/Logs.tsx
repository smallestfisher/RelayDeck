import { Card, Table, Tag, Input, DatePicker, Space } from 'antd'
import { SearchOutlined } from '@ant-design/icons'

const logData = [
  { key: '1', time: '2024-06-10 18:30:15', method: 'POST', path: '/v1/chat/completions', model: 'gpt-4', status: 200, time_cost: '1.2s', site: 'GPT站点-01' },
  { key: '2', time: '2024-06-10 18:29:45', method: 'POST', path: '/v1/chat/completions', model: 'claude-3-opus', status: 200, time_cost: '0.9s', site: 'Claude站点-01' },
  { key: '3', time: '2024-06-10 18:28:32', method: 'POST', path: '/v1/embeddings', model: 'text-embedding-ada-002', status: 200, time_cost: '0.3s', site: 'OpenAI中转-02' },
  { key: '4', time: '2024-06-10 18:27:11', method: 'POST', path: '/v1/chat/completions', model: 'gpt-3.5-turbo', status: 500, time_cost: '3.5s', site: 'GPT站点-01' },
]

export default function Logs() {
  const columns = [
    { title: '时间', dataIndex: 'time', key: 'time', width: 180 },
    { title: '方法', dataIndex: 'method', key: 'method', width: 80, render: (v: string) => <Tag color="blue">{v}</Tag> },
    { title: '路径', dataIndex: 'path', key: 'path', ellipsis: true },
    { title: '模型', dataIndex: 'model', key: 'model', render: (v: string) => <code>{v}</code> },
    { title: '站点', dataIndex: 'site', key: 'site' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: number) => <Tag color={v === 200 ? 'success' : 'error'}>{v}</Tag> },
    { title: '耗时', dataIndex: 'time_cost', key: 'time_cost' },
  ]

  return (
    <Card>
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder="搜索日志" prefix={<SearchOutlined />} style={{ width: 250 }} />
        <DatePicker.RangePicker />
      </Space>
      <Table columns={columns} dataSource={logData} />
    </Card>
  )
}
