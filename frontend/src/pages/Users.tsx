import { Card, Table, Tag, Button, Space } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'

const userData = [
  { key: '1', username: 'admin', email: 'admin@example.com', role: 'admin', status: 'active', lastLogin: '2024-06-10 18:30' },
  { key: '2', username: 'user1', email: 'user1@example.com', role: 'user', status: 'active', lastLogin: '2024-06-10 15:20' },
  { key: '3', username: 'user2', email: 'user2@example.com', role: 'user', status: 'inactive', lastLogin: '2024-06-08 10:15' },
]

export default function Users() {
  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
    { title: '角色', dataIndex: 'role', key: 'role', render: (v: string) => <Tag color={v === 'admin' ? 'red' : 'blue'}>{v}</Tag> },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: string) => <Tag color={v === 'active' ? 'success' : 'default'}>{v === 'active' ? '活跃' : '停用'}</Tag> },
    { title: '最后登录', dataIndex: 'lastLogin', key: 'lastLogin' },
    {
      title: '操作',
      render: () => (
        <Space>
          <Button type="link" icon={<EditOutlined />}>编辑</Button>
          <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
        </Space>
      ),
    },
  ]

  return (
    <Card>
      <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }}>添加用户</Button>
      <Table columns={columns} dataSource={userData} />
    </Card>
  )
}
