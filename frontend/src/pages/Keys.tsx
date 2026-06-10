import { useState } from 'react'
import { Card, Table, Button, Space, Tag, Modal, Form, Input, message } from 'antd'
import { PlusOutlined, CopyOutlined, DeleteOutlined } from '@ant-design/icons'

const initialData = [
  { key: '1', name: '默认密钥', apiKey: 'sk-proj-abc...xyz123', quota: 10000, used: 3245, status: 'active', created: '2024-06-01' },
  { key: '2', name: '测试密钥', apiKey: 'sk-proj-def...uvw456', quota: 5000, used: 1234, status: 'active', created: '2024-06-05' },
]

export default function Keys() {
  const [data] = useState(initialData)
  const [isModalOpen, setIsModalOpen] = useState(false)

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: 'API密钥', dataIndex: 'apiKey', key: 'apiKey', render: (v: string) => <code>{v}</code> },
    { title: '配额', dataIndex: 'quota', key: 'quota' },
    { title: '已使用', dataIndex: 'used', key: 'used' },
    { title: '状态', dataIndex: 'status', key: 'status', render: () => <Tag color="success">活跃</Tag> },
    { title: '创建时间', dataIndex: 'created', key: 'created' },
    {
      title: '操作',
      render: () => (
        <Space>
          <Button type="link" icon={<CopyOutlined />} onClick={() => message.success('已复制')}>复制</Button>
          <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <Card>
      <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalOpen(true)} style={{ marginBottom: 16 }}>
        创建密钥
      </Button>
      <Table columns={columns} dataSource={data} />
      
      <Modal title="创建API密钥" open={isModalOpen} onOk={() => setIsModalOpen(false)} onCancel={() => setIsModalOpen(false)}>
        <Form layout="vertical">
          <Form.Item label="密钥名称" name="name">
            <Input placeholder="为密钥起个名字" />
          </Form.Item>
          <Form.Item label="请求配额" name="quota">
            <Input type="number" placeholder="10000" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
