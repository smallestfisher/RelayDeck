import { Card, Form, Input, Switch, Button, Space, Select, message } from 'antd'

export default function Settings() {
  return (
    <div>
      <Card title="基础设置" style={{ marginBottom: 16 }}>
        <Form layout="vertical">
          <Form.Item label="平台名称">
            <Input defaultValue="RelayDeck" />
          </Form.Item>
          <Form.Item label="默认超时时间（秒）">
            <Input type="number" defaultValue={30} />
          </Form.Item>
          <Form.Item label="启用请求日志">
            <Switch defaultChecked />
          </Form.Item>
        </Form>
      </Card>

      <Card title="路由策略">
        <Form layout="vertical">
          <Form.Item label="负载均衡策略">
            <Select defaultValue="round-robin">
              <Select.Option value="round-robin">轮询</Select.Option>
              <Select.Option value="random">随机</Select.Option>
              <Select.Option value="least-load">最小负载</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="失败重试次数">
            <Input type="number" defaultValue={3} />
          </Form.Item>
          <Form.Item label="自动故障切换">
            <Switch defaultChecked />
          </Form.Item>
        </Form>
      </Card>

      <div style={{ marginTop: 16 }}>
        <Space>
          <Button type="primary" onClick={() => message.success('保存成功')}>保存设置</Button>
          <Button>重置</Button>
        </Space>
      </div>
    </div>
  )
}
