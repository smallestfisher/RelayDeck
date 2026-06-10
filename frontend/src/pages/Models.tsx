import { useState } from 'react'
import { Table, Button, Space, Tag, Card, Input, Modal, Form, Select, InputNumber, message, Progress } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined, CheckCircleOutlined } from '@ant-design/icons'

interface Model {
  key: string
  name: string
  provider: string
  sites: string[]
  status: 'active' | 'inactive'
  requests: number
  quota: number
  cost: number
}

const initialData: Model[] = [
  { key: '1', name: 'gpt-4', provider: 'OpenAI', sites: ['GPT站点-01', 'OpenAI中转-02'], status: 'active', requests: 8234, quota: 10000, cost: 245.67 },
  { key: '2', name: 'gpt-3.5-turbo', provider: 'OpenAI', sites: ['GPT站点-01'], status: 'active', requests: 15234, quota: 20000, cost: 89.34 },
  { key: '3', name: 'claude-3-opus', provider: 'Anthropic', sites: ['Claude站点-01'], status: 'active', requests: 5421, quota: 8000, cost: 178.23 },
  { key: '4', name: 'gemini-pro', provider: 'Google', sites: ['Gemini站点-01'], status: 'inactive', requests: 2134, quota: 5000, cost: 45.12 },
]

export default function Models() {
  const [data, setData] = useState<Model[]>(initialData)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [form] = Form.useForm()

  const handleAdd = () => {
    form.resetFields()
    setEditingKey(null)
    setIsModalOpen(true)
  }

  const handleEdit = (record: Model) => {
    form.setFieldsValue(record)
    setEditingKey(record.key)
    setIsModalOpen(true)
  }

  const handleDelete = (key: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除该模型配置吗？',
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
        const newModel = { ...values, key: Date.now().toString(), requests: 0, cost: 0 }
        setData([...data, newModel])
        message.success('添加成功')
      }
      setIsModalOpen(false)
    })
  }

  const columns = [
    { 
      title: '模型名称', 
      dataIndex: 'name', 
      key: 'name',
      render: (text: string) => <span style={{ fontWeight: 500, fontFamily: 'monospace' }}>{text}</span>
    },
    { 
      title: '提供商', 
      dataIndex: 'provider', 
      key: 'provider',
      render: (provider: string) => <Tag color="purple">{provider}</Tag>
    },
    { 
      title: '绑定站点', 
      dataIndex: 'sites', 
      key: 'sites',
      render: (sites: string[]) => sites.map(site => <Tag key={site} color="blue">{site}</Tag>)
    },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'} icon={status === 'active' ? <CheckCircleOutlined /> : null}>
          {status === 'active' ? '活跃' : '停用'}
        </Tag>
      )
    },
    { 
      title: '使用量', 
      key: 'usage',
      render: (_: any, record: Model) => {
        const percent = (record.requests / record.quota) * 100
        return (
          <div style={{ width: 150 }}>
            <div style={{ fontSize: 12, marginBottom: 4 }}>{record.requests} / {record.quota}</div>
            <Progress percent={percent} showInfo={false} strokeColor={percent > 80 ? '#ff4d4f' : '#1890ff'} />
          </div>
        )
      }
    },
    { 
      title: '今日成本', 
      dataIndex: 'cost', 
      key: 'cost',
      render: (cost: number) => `¥${cost.toFixed(2)}`
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Model) => (
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
            placeholder="搜索模型名称" 
            prefix={<SearchOutlined />}
            style={{ width: 300 }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加模型
          </Button>
        </div>
        
        <Table columns={columns} dataSource={data} />
      </Card>

      <Modal
        title={editingKey ? '编辑模型' : '添加模型'}
        open={isModalOpen}
        onOk={handleSave}
        onCancel={() => setIsModalOpen(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="模型名称" rules={[{ required: true }]}>
            <Input placeholder="gpt-4, claude-3-opus..." />
          </Form.Item>
          <Form.Item name="provider" label="提供商" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="OpenAI">OpenAI</Select.Option>
              <Select.Option value="Anthropic">Anthropic</Select.Option>
              <Select.Option value="Google">Google</Select.Option>
              <Select.Option value="Other">其他</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="sites" label="绑定站点" rules={[{ required: true }]}>
            <Select mode="multiple" placeholder="选择站点">
              <Select.Option value="GPT站点-01">GPT站点-01</Select.Option>
              <Select.Option value="Claude站点-01">Claude站点-01</Select.Option>
              <Select.Option value="Gemini站点-01">Gemini站点-01</Select.Option>
              <Select.Option value="OpenAI中转-02">OpenAI中转-02</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="quota" label="请求配额" rules={[{ required: true }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="active">活跃</Select.Option>
              <Select.Option value="inactive">停用</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
