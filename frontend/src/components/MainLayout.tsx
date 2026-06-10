import { Layout, Menu, Avatar, Badge, Space, Switch } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  DashboardOutlined,
  CloudServerOutlined,
  ApiOutlined,
  BellOutlined,
  UserOutlined,
  BarChartOutlined,
  FileTextOutlined,
  SettingOutlined,
  KeyOutlined,
  TeamOutlined,
  BulbOutlined,
  BulbFilled
} from '@ant-design/icons'
import { useTheme } from '../contexts/ThemeContext'

const { Header, Sider, Content } = Layout

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { theme, toggleTheme } = useTheme()

  const menuItems = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/sites', icon: <CloudServerOutlined />, label: '站点管理' },
    { key: '/models', icon: <ApiOutlined />, label: '模型管理' },
    { key: '/channels', icon: <ApiOutlined />, label: '渠道管理' },
    { key: '/tokens', icon: <KeyOutlined />, label: '令牌管理' },
    { key: '/statistics', icon: <BarChartOutlined />, label: '统计分析' },
    { key: '/logs', icon: <FileTextOutlined />, label: '日志' },
    { key: '/users', icon: <TeamOutlined />, label: '用户管理' },
    { key: '/settings', icon: <SettingOutlined />, label: '系统设置' },
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={200} theme="dark">
        <div style={{ 
          height: 64, 
          display: 'flex', 
          alignItems: 'center', 
          padding: '0 24px',
          gap: 12,
          fontSize: 18,
          fontWeight: 'bold'
        }}>
          <div style={{
            width: 32,
            height: 32,
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            borderRadius: 8,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>⚡</div>
          RelayDeck
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      
      <Layout>
        <Header style={{
          background: theme === 'dark' ? '#0f1429' : '#fff',
          padding: '0 24px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: theme === 'dark' ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid #f0f0f0'
        }}>
          <div style={{ fontSize: 18, fontWeight: 500 }}>
            统一管理 API 中转
          </div>
          <Space size={24}>
            <Switch
              checked={theme === 'dark'}
              onChange={toggleTheme}
              checkedChildren={<BulbFilled />}
              unCheckedChildren={<BulbOutlined />}
            />
            <Badge count={5} onClick={() => navigate('/notifications')}>
              <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
            </Badge>
            <Avatar icon={<UserOutlined />} style={{ cursor: 'pointer' }} />
          </Space>
        </Header>
        
        <Content style={{
          margin: 24,
          background: theme === 'dark' ? 'rgba(15, 20, 41, 0.5)' : '#f5f5f5',
          borderRadius: 12,
          padding: 24
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
