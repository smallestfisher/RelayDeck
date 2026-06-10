import { Form, Input, Button, Card } from 'antd'
import { UserOutlined, LockOutlined, GoogleOutlined, GithubOutlined, WechatOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import './Login.css'

export default function Login() {
  const navigate = useNavigate()

  const onFinish = (values: any) => {
    console.log('Login:', values)
    navigate('/dashboard')
  }

  return (
    <div className="login-container">
      <div className="login-left">
        <div className="logo">
          <div className="logo-icon">⚡</div>
          <span>RelayDeck</span>
        </div>
        <h1>统一管理，智能调度，<br />让每一次调用更高效</h1>
        <p>集中管理多个大模型站点，统一路由到对应的上游站点，<br />智能选择最优线路，降低成本和管理难度。</p>
        <div className="feature-3d">
          <div className="glow-circle"></div>
          <div className="feature-icon">⚡</div>
        </div>
      </div>
      
      <div className="login-right">
        <Card className="login-card">
          <div className="login-tabs">
            <div className="tab active">登录</div>
            <div className="tab">注册</div>
          </div>
          
          <Form onFinish={onFinish} className="login-form">
            <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input prefix={<UserOutlined />} placeholder="用户名" size="large" />
            </Form.Item>
            
            <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="密码" size="large" />
            </Form.Item>
            
            <div className="login-options">
              <a href="#">忘记密码？</a>
            </div>
            
            <Form.Item>
              <Button type="primary" htmlType="submit" size="large" block>
                登录
              </Button>
            </Form.Item>
          </Form>
          
          <div className="divider">
            <span>或使用第三方登录</span>
          </div>
          
          <div className="social-login">
            <Button icon={<GoogleOutlined />} className="social-btn">Google</Button>
            <Button icon={<GithubOutlined />} className="social-btn">Github</Button>
            <Button icon={<WechatOutlined />} className="social-btn">微信登录</Button>
          </div>
          
          <div className="login-footer">
            登录即表示你同意《服务协议》和《隐私政策》
          </div>
        </Card>
      </div>
    </div>
  )
}
