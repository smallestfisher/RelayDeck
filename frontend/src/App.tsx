import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import MainLayout from './components/MainLayout'
import Dashboard from './pages/Dashboard'
import Sites from './pages/Sites'
import Models from './pages/Models'
import Statistics from './pages/Statistics'
import Logs from './pages/Logs'
import Keys from './pages/Keys'
import Users from './pages/Users'
import Settings from './pages/Settings'
import Notifications from './pages/Notifications'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="sites" element={<Sites />} />
          <Route path="models" element={<Models />} />
          <Route path="statistics" element={<Statistics />} />
          <Route path="logs" element={<Logs />} />
          <Route path="keys" element={<Keys />} />
          <Route path="users" element={<Users />} />
          <Route path="settings" element={<Settings />} />
          <Route path="notifications" element={<Notifications />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
