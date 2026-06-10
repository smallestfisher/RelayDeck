import { Card, List, Tag, Button, Space, Tabs } from 'antd'
import { CheckCircleOutlined, WarningOutlined, InfoCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'

const systemMessages = [
  { id: 1, type: 'info', title: '系统更新通知', content: 'RelayDeck v1.2.0 已发布，新增模型健康检查功能', time: '2小时前' },
  { id: 2, type: 'success', title: '站点上线成功', content: 'GPT站点-03 已成功添加并开始处理请求', time: '5小时前' },
  { id: 3, type: 'warning', title: '请求配额警告', content: 'gpt-4 模型使用量已达80%，请注意调整配额', time: '1天前' },
]

const alerts = [
  { id: 1, type: 'error', title: '站点离线', content: 'Gemini站点-01 响应超时，已自动切换到备用站点', time: '30分钟前', handled: false },
  { id: 2, type: 'warning', title: '响应延迟', content: 'Claude站点-01 平均响应时间超过200ms', time: '2小时前', handled: true },
  { id: 3, type: 'error', title: '认证失败', content: 'OpenAI中转-02 API密钥验证失败', time: '5小时前', handled: false },
]

const getIcon = (type: string) => {
  switch (type) {
    case 'success': return <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 20 }} />
    case 'warning': return <WarningOutlined style={{ color: '#faad14', fontSize: 20 }} />
    case 'error': return <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 20 }} />
    default: return <InfoCircleOutlined style={{ color: '#1890ff', fontSize: 20 }} />
  }
}

export default function Notifications() {
  return (
    <div>
      <Tabs
        items={[
          {
            key: 'system',
            label: '系统消息',
            children: (
              <Card>
                <List
                  itemLayout="horizontal"
                  dataSource={systemMessages}
                  renderItem={item => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={getIcon(item.type)}
                        title={<span style={{ fontWeight: 500 }}>{item.title}</span>}
                        description={
                          <div>
                            <div>{item.content}</div>
                            <div style={{ marginTop: 8, fontSize: 12, opacity: 0.6 }}>{item.time}</div>
                          </div>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            )
          },
          {
            key: 'alerts',
            label: '告警通知',
            children: (
              <Card>
                <List
                  itemLayout="horizontal"
                  dataSource={alerts}
                  renderItem={item => (
                    <List.Item
                      actions={[
                        item.handled ? (
                          <Tag color="success">已处理</Tag>
                        ) : (
                          <Space>
                            <Button type="link" size="small">处理</Button>
                            <Button type="link" size="small" danger>忽略</Button>
                          </Space>
                        )
                      ]}
                    >
                      <List.Item.Meta
                        avatar={getIcon(item.type)}
                        title={<span style={{ fontWeight: 500 }}>{item.title}</span>}
                        description={
                          <div>
                            <div>{item.content}</div>
                            <div style={{ marginTop: 8, fontSize: 12, opacity: 0.6 }}>{item.time}</div>
                          </div>
                        }
                      />
                    </List.Item>
                  )}
                />
              </Card>
            )
          }
        ]}
      />
    </div>
  )
}
