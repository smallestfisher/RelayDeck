import { useState } from 'react'
import { Table, Button, Space, Tag, Card, Input, Modal, Form, Select, Switch, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, CheckCircleOutlined, CloseCircleOutlined, SearchOutlined } from '@ant-design/icons'

interface Site {
  key: string
  name: string
  url: string
  type: string
  status: 'online' | 'offline'
  priority: number
  requests: number
  success: number
  avgTime: string
}

const initialData: Site[] = [
  { key: '1', name: 'GPT站点-01', url: 'https://api.openai-gpt.com', type: 'OpenAI', status: 'online', priority: 1, requests: 12453, success: 99.2, avgTime: '120ms' },
  { key: '2', name: 'Claude站点-01', url: 'https://api.claude-ai.com', type: 'Anthropic', status: 'online', priority: 2, requests: 8234, success: 99.8, avgTime: '95ms' },
  { key: '3', name: 'Gemini站点-01', url: 'https://api.gemini-pro.com', type: 'Google', status: 'offline', priority: 3, requests: 3421, success: 95.3, avgTime: '180ms' },
  { key: '4', name: 'OpenAI中转-02', url: 'https://relay.openai-api.org', type: 'OpenAI', status: 'online', priority: 1, requests: 15234, success: 98.9, avgTime: '110ms' },
]

export default function Sites() {
  const [data, setData] = useState<Site[]>(initialData)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [form] = Form.useForm()

  const handleAdd = () => {
    form.resetFields()
    setEditingKey(null)
    setIsModalOpen(true)
  }

  const handleEdit = (record: Site) => {
    form.setFieldsValue(record)
    setEditingKey(record.key)
    setIsModalOpen(true)
  }

  const handleDelete = (key: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除该站点吗？',
      onOk: () => {
        setData(data.filter(item => item.key !== key))
        message.success('删除成功')
      }
    })
  }

  const handleSave = () => {
    form.validateFields().then(values => {
      if (editingKey) {
        setData(data.map(item => item.key === editingKey ? { ...item, ...values } : item))
        message.success('更新成功')
      } else {
        const newSite = { ...values, key: Date.now().toString(), requests: 0, success: 100, avgTime: '0ms' }
        setData([...data, newSite])
        message.success('添加成功')
      }
      setIsModalOpen(false)
    })
  }

  const columns = [
    { title: '站点名称', dataIndex: 'name', key: 'name', render: (text: string) => <span style={{ fontWeight: 500 }}>{text}</span> },
    { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
    { 
      title: '类型', 
      dataIndex: 'type', 
      key: 'type',
      render: (type: string) => <Tag color="blue">{type}</Tag>
    },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'online' ? 'success' : 'error'} icon={status === 'online' ? <CheckCircleOutlined /> : <CloseCircleOutlined />}>
          {status === 'online' ? '在线' : '离线'}
        </Tag>
      )
    },
    { title: '优先级', dataIndex: 'priority', key: 'priority' },
    { title: '请求数', dataIndex: 'requests', key: 'requests' },
    { title: '成功率', dataIndex: 'success', key: 'success', render: (val: number) => `${val}%` },
    { title: '平均响应', dataIndex: 'avgTime', key: 'avgTime' },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Site) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.key)}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Input 
            placeholder="搜索站点名称或URL" 
            prefix={<SearchOutlined />}
            style={{ width: 300 }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加站点
          </Button>
        </div>
        
        <Table columns={columns} dataSource={data} />
      </Card>

      <Modal
        title={editingKey ? '编辑站点' : '添加站点'}
        open={isModalOpen}
        onOk={handleSave}
        onCancel={() => setIsModalOpen(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="站点名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="url" label="站点URL" rules={[{ required: true, type: 'url' }]}>
            <Input placeholder="https://" />
          </Form.Item>
          <Form.Item name="type" label="站点类型" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="OpenAI">OpenAI</Select.Option>
              <Select.Option value="Anthropic">Anthropic</Select.Option>
              <Select.Option value="Google">Google</Select.Option>
              <Select.Option value="Other">其他</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="priority" label="优先级" rules={[{ required: true }]}>
            <Select>
              <Select.Option value={1}>高</Select.Option>
              <Select.Option value={2}>中</Select.Option>
              <Select.Option value={3}>低</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="status" label="状态" valuePropName="checked" initialValue={true}>
            <Switch checkedChildren="在线" unCheckedChildren="离线" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
